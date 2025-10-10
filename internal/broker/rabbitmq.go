package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type RabbitMQClient struct {
	Connection *amqp.Connection
}

const (
	ExchangeRideType = "ride_type"
	ExchangeKind     = "direct"
	QueueIntraRide   = "intraRide"
	QueueInterRide   = "interRide"
	KeyIntra         = "intra"
	KeyInter         = "inter"
	LogComponent     = "rabbitmq"
)

func NewClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		zap.L().Error("failed to connect to RabbitMQ",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	client := &RabbitMQClient{
		Connection: conn,
	}

	return client, nil
}

func (c *RabbitMQClient) NewChannel() (*amqp.Channel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		zap.L().Error("failed to open a new channel", zap.Error(err))
		return nil, fmt.Errorf("failed to open a new channel: %w", err)
	}
	return ch, nil
}

func (c *RabbitMQClient) Close() {
	if c.Connection != nil {
		if err := c.Connection.Close(); err != nil {
			zap.L().Error("failed to close amqp connection", zap.Error(err))
		}
	}
	zap.L().Info("amqp connection closed",
		zap.String("entity", LogComponent),
	)
}

func (c *RabbitMQClient) Setup() error {
	ch, err := c.Connection.Channel()
	if err != nil {
		zap.L().Error("failed to open setup channel", zap.Error(err))
		return fmt.Errorf("failed to open setup channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		ExchangeRideType,
		ExchangeKind,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to declare exchange",
			zap.String("exchange", ExchangeRideType),
			zap.Error(err),
		)
		return fmt.Errorf("failed to declare exchange '%s': %w", ExchangeRideType, err)
	}
	zap.L().Debug("declared exchange", zap.String("exchange", ExchangeRideType))

	intraRide, err := ch.QueueDeclare(
		QueueIntraRide,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to declare intraRide queue", zap.Error(err))
		return fmt.Errorf("failed to declare intraRide queue: %w", err)
	}
	zap.L().Debug("declared queue", zap.String("queue", intraRide.Name))

	interRide, err := ch.QueueDeclare(
		QueueInterRide,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to declare interRide queue", zap.Error(err))
		return fmt.Errorf("failed to declare interRide queue: %w", err)
	}
	zap.L().Debug("declared queue", zap.String("queue", interRide.Name))

	err = ch.QueueBind(
		intraRide.Name,
		KeyIntra,
		ExchangeRideType,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to bind intraRide queue",
			zap.String("queue", intraRide.Name),
			zap.String("key", KeyIntra),
			zap.Error(err),
		)
		return fmt.Errorf("failed to bind intraRide queue: %w", err)
	}
	zap.L().Debug("bound queue", zap.String("queue", intraRide.Name), zap.String("key", KeyIntra))

	err = ch.QueueBind(
		interRide.Name,
		KeyInter,
		ExchangeRideType,
		false,
		nil,
	)
	if err != nil {
		zap.L().Error("failed to bind interRide queue",
			zap.String("queue", interRide.Name),
			zap.String("key", KeyInter),
			zap.Error(err),
		)
		return fmt.Errorf("failed to bind interRide queue: %w", err)
	}
	zap.L().Debug("bound queue", zap.String("queue", interRide.Name), zap.String("key", KeyInter))

	return nil
}
