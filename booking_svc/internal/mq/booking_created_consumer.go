package mq

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"booking_svc/internal/config"
	"booking_svc/internal/events"
	"booking_svc/internal/repository"

	"github.com/segmentio/kafka-go"
)

type BookingAcceptedConsumer struct {
	reader *kafka.Reader
	repo   repository.BookingRepository
	logger *slog.Logger
}

func NewBookingAcceptedConsumer(cfg config.Config, repo repository.BookingRepository, logger *slog.Logger) *BookingAcceptedConsumer {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        cfg.ConsumerGroupAccepts,
		Topic:          cfg.TopicBookingAccepted,
		MinBytes:       1,
		MaxBytes:       10e6,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0, // commit after DB success
	})
	return &BookingAcceptedConsumer{reader: r, repo: repo, logger: logger}
}

func (c *BookingAcceptedConsumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Error("kafka fetch failed", slog.String("err", err.Error()))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var evt events.BookingAccepted
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			c.logger.Error("invalid booking.accepted payload", slog.String("err", err.Error()))
			_ = c.reader.CommitMessages(ctx, msg) // skip poison
			continue
		}

		updated, err := c.repo.MarkAccepted(ctx, evt.BookingID, evt.DriverID)
		if err != nil {
			c.logger.Error("db update failed", slog.String("booking_id", evt.BookingID), slog.String("err", err.Error()))
			// no commit -> retry later
			continue
		}
		if !updated {
			// already accepted or missing — idempotent no-op
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit failed", slog.String("err", err.Error()))
		}
	}
}

func (c *BookingAcceptedConsumer) Close() error { return c.reader.Close() }
