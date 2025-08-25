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
	handlerhttp "driver_svc/internal/handler/http"
	"driver_svc/internal/httpserver"
	"driver_svc/internal/logging"
	"driver_svc/internal/mq"
	"driver_svc/internal/repository/postgres"
	"driver_svc/internal/seed"
	"driver_svc/internal/service"
)

var version = "0.1.0"

func main() {
	cfg := config.LoadFromEnv("driver_svc", "8081")
	logger := logging.New(cfg.LogLevel, cfg.ServiceName).With(slog.String("version", version))

	// Shared shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// DB connect + bootstrap + seed
	startupCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
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

	// Repos
	driverRepo := postgres.NewDriverRepo(pool) // kept for next parts (GET /drivers)
	jobRepo := postgres.NewJobRepo(pool)

	// Producer for booking.accepted
	producer := mq.NewProducer(cfg, logger)
	defer func() { _ = producer.Close() }()

	// Service + HTTP
	jobsSvc := service.NewJobsService(driverRepo, jobRepo, producer, logger)
	srv := httpserver.New(cfg, logger)
	h := handlerhttp.NewJobsHandler(jobsSvc)
	h.RegisterRoutes(srv.Router())

	// Kafka consumer: booking.created -> upsert Open job
	consumer := mq.NewBookingCreatedConsumer(cfg, jobRepo, logger)
	defer func() { _ = consumer.Close() }()
	go func() {
		if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Error("booking.created consumer stopped", slog.String("err", err.Error()))
		}
	}()

	// HTTP server
	errCh := srv.Start()

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
