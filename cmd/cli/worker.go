package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/hotspot/configs"
	"github.com/mohammadgholizadeh/hotspot/internal/broker"
	"github.com/mohammadgholizadeh/hotspot/internal/cache"
	"github.com/mohammadgholizadeh/hotspot/internal/service"
	"github.com/mohammadgholizadeh/hotspot/internal/storage"
	"github.com/mohammadgholizadeh/hotspot/internal/worker"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the background worker (RabbitMQ consumer)",
	RunE: func(cmd *cobra.Command, args []string) error { return runWorker() },
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

func runWorker() error {
	// Logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Config
	cfg := configs.LoadConfig(CfgPath())

	// DB
	dbPool, err := pgxpool.New(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		return err
	}
	defer dbPool.Close()

	// Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, logger)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	// RabbitMQ
	rabbitClient, err := broker.NewClient(cfg.Broker.URL)
	if err != nil {
		return err
	}
	defer rabbitClient.Close()
	if err := rabbitClient.Setup(); err != nil {
		return err
	}

	store := storage.NewRequestStore(dbPool, logger)
	svc := service.NewRequestService(store, redisClient.Client(), logger)
	cons := worker.NewConsumer(rabbitClient, svc, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cons.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("worker: shutting down...")
	time.Sleep(2 * time.Second)
	return nil
}
