package mq

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"driver_svc/internal/config"
	"driver_svc/internal/events"
	"driver_svc/internal/repository"

	"github.com/segmentio/kafka-go"
)

type BookingCreatedConsumer struct {
	reader *kafka.Reader
	jobs   repository.JobRepository
	logger *slog.Logger
}

func NewBookingCreatedConsumer(cfg config.Config, jobs repository.JobRepository, logger *slog.Logger) *BookingCreatedConsumer {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        cfg.ConsumerGroupJobs,
		Topic:          cfg.TopicBookingCreated,
		MinBytes:       1,
		MaxBytes:       10e6,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0, // manual commit after DB success
	})
	return &BookingCreatedConsumer{reader: r, jobs: jobs, logger: logger}
}

func (c *BookingCreatedConsumer) Run(ctx context.Context) error {
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

		var evt events.BookingCreated
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			c.logger.Error("invalid booking.created payload", slog.String("err", err.Error()))
			_ = c.reader.CommitMessages(ctx, msg) // skip poison message
			continue
		}

		if err := c.jobs.UpsertOpenJob(ctx, repository.UpsertJobParams{
			BookingID: evt.BookingID,
			PickupLoc: evt.PickupLoc,
			Dropoff:   evt.Dropoff,
			Price:     evt.Price,
		}); err != nil {
			c.logger.Error("upsert job failed", slog.String("booking_id", evt.BookingID), slog.String("err", err.Error()))
			// no commit -> retry later
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit failed", slog.String("err", err.Error()))
		}
	}
}

func (c *BookingCreatedConsumer) Close() error { return c.reader.Close() }
