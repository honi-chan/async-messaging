package stream

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"async-messaging/internal/event"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Produce(ctx context.Context, ev event.Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(ev.AggregateID),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("failed to write messages: %v", err)
		return err
	}

	log.Printf("Event produced to Kafka: ID=%s Type=%s", ev.ID, ev.Type)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
