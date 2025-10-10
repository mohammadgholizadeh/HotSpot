package worker

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/hotspot/internal/broker"
	"github.com/mohammadgholizadeh/hotspot/internal/domain"
)

type Consumer struct {
	rabbit *broker.RabbitMQClient
	svc    domain.RequestService
	logger *zap.Logger
}

func NewConsumer(rabbit *broker.RabbitMQClient, svc domain.RequestService, logger *zap.Logger) *Consumer {
	return &Consumer{rabbit: rabbit, svc: svc, logger: logger}
}

func (c *Consumer) Start(ctx context.Context) {
	go c.consume(ctx, broker.QueueIntraRide, func(ctx context.Context, req domain.Request) error {
		return c.svc.CreateByIntraType(ctx, req)
	})
	go c.consume(ctx, broker.QueueInterRide, func(ctx context.Context, req domain.Request) error {
		return c.svc.CreateByInterType(ctx, req)
	})
}

func (c *Consumer) consume(ctx context.Context, queue string, handler func(context.Context, domain.Request) error) {
	backoff := time.Second
	for {
		ch, err := c.rabbit.NewChannel()
		if err != nil {
			c.logger.Error("worker: new channel failed", zap.String("queue", queue), zap.Error(err))
			time.Sleep(backoff)
			backoff = min(backoff*2, 30*time.Second)
			continue
		}

		msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
		if err != nil {
			c.logger.Error("worker: register consumer failed", zap.String("queue", queue), zap.Error(err))
			_ = ch.Close()
			time.Sleep(backoff)
			backoff = min(backoff*2, 30*time.Second)
			continue
		}
		c.logger.Info("worker: consuming", zap.String("queue", queue))
		backoff = time.Second

		for {
			select {
			case <-ctx.Done():
				_ = ch.Close()
				c.logger.Info("worker: stopped", zap.String("queue", queue))
				return
			case msg, ok := <-msgs:
				if !ok {
					_ = ch.Close()
					c.logger.Warn("worker: delivery channel closed, reconnecting", zap.String("queue", queue))
					time.Sleep(time.Second)
					break
				}
				var req domain.Request
				if err := json.Unmarshal(msg.Body, &req); err != nil {
					c.logger.Error("worker: unmarshal failed", zap.String("queue", queue), zap.Error(err))
					continue
				}
				if err := handler(ctx, req); err != nil {
					c.logger.Error("worker: processing failed", zap.String("queue", queue), zap.Error(err))
				}
			}
		}
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
