package mq

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"booking_svc/internal/config"
	"booking_svc/internal/events"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer             *kafka.Writer
	topicBookingCreated string
	logger             *slog.Logger
}

func NewProducer(cfg config.Config, logger *slog.Logger) *Producer {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        cfg.TopicBookingCreated,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Producer{
		writer:              w,
		topicBookingCreated: cfg.TopicBookingCreated,
		logger:              logger,
	}
}

func (p *Producer) ProduceBookingCreated(ctx context.Context, evt events.BookingCreated) error {
	value, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg := kafka.Message{
		Key:   []byte(evt.BookingID),
		Value: value,
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}