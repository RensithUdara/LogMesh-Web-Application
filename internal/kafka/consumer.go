package kafka

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/segmentio/kafka-go"

	"logmesh/internal/model"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokersCSV, topic, groupID string) *Consumer {
	brokers := splitCSV(brokersCSV)
	if len(brokers) == 0 {
		return nil
	}

	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: strings.TrimSpace(groupID),
		}),
	}
}

func (c *Consumer) Fetch(ctx context.Context) (model.LogEvent, error) {
	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return model.LogEvent{}, err
	}

	var event model.LogEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		_ = c.reader.CommitMessages(ctx, message)
		return model.LogEvent{}, err
	}

	if err := c.reader.CommitMessages(ctx, message); err != nil {
		return model.LogEvent{}, err
	}

	return event, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
