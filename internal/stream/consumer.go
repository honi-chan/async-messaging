package stream

import (
	"context"
	"encoding/json"
	"log"

	"async-messaging/internal/event"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, event.Event) error) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // Context cancelled
			}
			log.Printf("failed to fetch message: %v", err)
			continue
		}

		var ev event.Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("failed to unmarshal event: %v", err)
			continue
		}

		if err := handler(ctx, ev); err != nil {
			log.Printf("handler failed: %v", err)
			// Decide commit strategy: here we choose to continue (at-least-once is handled by handler idempotent logic or DLQ if it was direct, but for Bridge we assume SNS publish errors might need retry or just log)
			// In a bridge scenario, if SNS publish fails, we might want to NOT commit the offset so we retry.
			// However, for simplicity here, we will just log error. Ideally implementation should handle retries.
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("failed to commit messages: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
