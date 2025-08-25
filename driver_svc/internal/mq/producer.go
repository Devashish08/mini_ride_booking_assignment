package mq

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"driver_svc/internal/config"
	"driver_svc/internal/events"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer               *kafka.Writer
	topicBookingAccepted string
	logger               *slog.Logger
}

func NewProducer(cfg config.Config, logger *slog.Logger) *Producer {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        cfg.TopicBookingAccepted,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Producer{
		writer:               w,
		topicBookingAccepted: cfg.TopicBookingAccepted,
		logger:               logger,
	}
}

func (p *Producer) ProduceBookingAccepted(ctx context.Context, evt events.BookingAccepted) error {
	value, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.BookingID), // preserves per-booking ordering
		Value: value,
	})
}

func (p *Producer) Close() error { return p.writer.Close() }
