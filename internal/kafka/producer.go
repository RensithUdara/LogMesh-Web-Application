package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"logmesh/internal/model"
)

type Producer interface {
	Publish(ctx context.Context, event model.LogEvent) error
	Close() error
}

type NoopProducer struct{}

func (NoopProducer) Publish(context.Context, model.LogEvent) error { return nil }
func (NoopProducer) Close() error                                  { return nil }

type WriterProducer struct {
	writer *kafka.Writer
}

func NewProducer(brokersCSV, topic string) Producer {
	brokers := splitCSV(brokersCSV)
	if len(brokers) == 0 {
		return NoopProducer{}
	}

	return &WriterProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 50 * time.Millisecond,
		},
	}
}

func (p *WriterProducer) Publish(ctx context.Context, event model.LogEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.Service),
		Value: payload,
		Time:  event.Timestamp,
	})
}

func (p *WriterProducer) Close() error {
	return p.writer.Close()
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
