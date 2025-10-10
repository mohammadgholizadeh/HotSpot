package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/hotspot/configs"
	"github.com/mohammadgholizadeh/hotspot/internal/broker"
	"github.com/mohammadgholizadeh/hotspot/internal/cache"
	"github.com/mohammadgholizadeh/hotspot/internal/controller"
	"github.com/mohammadgholizadeh/hotspot/internal/middleware"
	"github.com/mohammadgholizadeh/hotspot/internal/service"
	"github.com/mohammadgholizadeh/hotspot/internal/storage"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	RunE:  func(cmd *cobra.Command, args []string) error { return runHTTPServer() },
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runHTTPServer() error {
	// Logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Config
	cfg := configs.LoadConfig(CfgPath())

	logger.Info("Starting HotSpot HTTP API", zap.String("port", cfg.Server.Port))

	// Database
	dbPool, err := pgxpool.New(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		return err
	}
	defer dbPool.Close()
	if err := dbPool.Ping(context.Background()); err != nil {
		return err
	}

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

	//Layers
	store := storage.NewRequestStore(dbPool, logger)
	svc := service.NewRequestService(store, redisClient.Client(), logger)

	e := echo.New()

	// Middlewares
	lm := middleware.NewLoggerMiddleware(logger)
	bm := middleware.NewRequestBodyMiddleware(BodyLimit())

	v1 := e.Group("/api/v1", lm.Log)

	api := v1.Group("", bm.LimitBodySize)
	controller.RequestRoutes(api, svc, logger, rabbitClient)

	addr := cfg.GetServerAddr()
	if PortOverride() != "" {
		addr = ":" + PortOverride()
	}
	logger.Info("starting http server", zap.String("addr", addr))
	return e.Start(addr)
}
