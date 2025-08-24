package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"driver_svc/internal/config"
	"driver_svc/internal/db"
	"driver_svc/internal/httpserver"
	"driver_svc/internal/logging"
	"driver_svc/internal/seed"
)

var version = "0.1.0"

func main() {
	cfg := config.LoadFromEnv("driver_svc", "8081")
	logger := logging.New(cfg.LogLevel, cfg.ServiceName).With(slog.String("version", version))

	// DB connect + bootstrap + seed
	startupCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool, err := db.Connect(startupCtx, db.Params{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
	})
	if err != nil {
		logger.Error("db connect failed", slog.String("err", err.Error()))
		return
	}
	defer pool.Close()

	if err := db.Bootstrap(startupCtx, pool); err != nil {
		logger.Error("db bootstrap failed", slog.String("err", err.Error()))
		return
	}
	if err := seed.SeedDrivers(startupCtx, pool); err != nil {
		logger.Error("seed drivers failed", slog.String("err", err.Error()))
		return
	}

	srv := httpserver.New(cfg, logger)
	errCh := srv.Start()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GracefulTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", slog.String("err", err.Error()))
		}
	}
	logger.Info("exit")
}
