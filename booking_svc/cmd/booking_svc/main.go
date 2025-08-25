package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"booking_svc/internal/config"
	"booking_svc/internal/db"
	handlerhttp "booking_svc/internal/handler/http"
	"booking_svc/internal/httpserver"
	"booking_svc/internal/logging"
	"booking_svc/internal/mq"
	"booking_svc/internal/repository/postgres"
	"booking_svc/internal/service"
)

var version = "0.1.0"

func main() {
	cfg := config.LoadFromEnv("booking_svc", "8080")
	logger := logging.New(cfg.LogLevel, cfg.ServiceName).With(slog.String("version", version))

	// Signal context for graceful shutdown and consumers
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// DB connect + bootstrap
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

	// Repo + MQ producer + service
	repo := postgres.NewBookingRepo(pool)
	producer := mq.NewProducer(cfg, logger)
	defer func() {
		_ = producer.Close()
	}()
	svc := service.NewBookingService(repo, producer, logger)
	// Consumer: booking.accepted -> mark booking Accepted
	acceptConsumer := mq.NewBookingAcceptedConsumer(cfg, repo, logger)
	defer func() { _ = acceptConsumer.Close() }()
	go func() {
		if err := acceptConsumer.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Error("booking.accepted consumer stopped", slog.String("err", err.Error()))
		}
	}()
	// HTTP server + routes
	srv := httpserver.New(cfg, logger)
	handler := handlerhttp.NewBookingHandler(svc)
	handler.RegisterRoutes(srv.Router())

	// Start and graceful shutdown
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
