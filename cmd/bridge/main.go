package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/joho/godotenv"

	"async-messaging/internal/event"
	"async-messaging/internal/pubsub"
	"async-messaging/internal/stream"
)

func main() {
	// .env読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Kafka設定
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "events"
	}
	kafkaGroupID := "bridge-service"

	// AWS設定
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}
	snsClient := sns.NewFromConfig(cfg)
	topicArn := os.Getenv("SNS_TOPIC_ARN")
	if topicArn == "" {
		topicArn = "arn:aws:sns:ap-northeast-1:000000000000:job-topic"
		log.Println("WARNING: SNS_TOPIC_ARN is not set, using dummy ARN")
	}

	// Initialize Components
	consumer := stream.NewConsumer([]string{kafkaBroker}, kafkaTopic, kafkaGroupID)
	defer consumer.Close()

	publisher := pubsub.NewPublisher(snsClient, topicArn)

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down bridge...")
		cancel()
		consumer.Close()
	}()

	log.Println("Bridge service started. Listening on Kafka topic:", kafkaTopic)

	// Start Consumption
	err = consumer.Consume(ctx, func(ctx context.Context, ev event.Event) error {
		// Filter events if needed (e.g., only "job.created")
		if ev.Type != "job.created" {
			return nil
		}

		log.Printf("Bridging event: %s", ev.ID)
		if err := publisher.Publish(ctx, ev); err != nil {
			log.Printf("Failed to publish to SNS: %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("Consumer stopped with error: %v", err)
	}
}
