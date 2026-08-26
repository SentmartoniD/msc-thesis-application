package main

import (
	"context"
	"fmt"
	"os"
	"shorten-service/internal/config"
	"shorten-service/internal/handlers"
	"shorten-service/internal/metrics"
	"shorten-service/internal/server"
	"shorten-service/pkg/logger"
	"shorten-service/platform/database"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "shorten-service failed:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("shorten-service")
	if err != nil {
		return err
	}

	if err := logger.Init(logger.Options{
		Level:   cfg.LogLevel,
		Format:  cfg.LogFormat,
		Service: cfg.ServiceName,
		Version: cfg.BuildVersion,
		Commit:  cfg.BuildCommit,
	}); err != nil {
		return err
	}
	defer logger.Sync()

	// Every pod records what it actually loaded, so an anomalous experiment run
	// can be traced back to its exact configuration.
	logger.Log.Info("configuration loaded", zap.Any("config", cfg.GetFields()))

	ctx := context.Background()

	if err := database.Connect(ctx, cfg, metrics.QueryTracer{}); err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer database.Close()

	logger.Log.Info("connected to database")

	metrics.RegisterPoolMetrics()

	health := handlers.NewHealthHandler(database.Ping)

	return server.Run(
		ctx,
		cfg,
		handlers.SetupRouter(cfg),
		handlers.SetupAdminRouter(health),
		health.Drain,
	)
}
