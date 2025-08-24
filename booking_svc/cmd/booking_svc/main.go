package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"booking_svc/internal/config"
	"booking_svc/internal/httpserver"
	"booking_svc/internal/logging"
)

var version = "0.1.0"

func main() {
	cfg := config.LoadFromEnv("booking_svc", "8080")
	logger := logging.New(cfg.LogLevel, cfg.ServiceName).With(slog.String("version", version))

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
