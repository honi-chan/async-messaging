package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"

	"async-messaging/internal/event"
	"async-messaging/internal/usecase"
)

func main() {
	// .env読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down worker...")
		cancel()
	}()

	// AWS設定の読み込み
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	// 環境変数からキューURLを取得
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		queueURL = "https://sqs.ap-northeast-1.amazonaws.com/466703337425/job-queue"
	}

	// ハンドラー初期化 (Queueへの送信責務は不要になった)
	handler := usecase.NewHandler()

	log.Printf("Worker started. Listening on queue: %s\n", queueURL)

	// メッセージ処理ループ
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
			processMessages(ctx, sqsClient, queueURL, handler)
		}
	}
}

func processMessages(ctx context.Context, client *sqs.Client, queueURL string, handler *usecase.Handler) {
	out, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 10, // バッチ処理
		WaitTimeSeconds:     20, // Long polling
		VisibilityTimeout:   30, // 処理時間の猶予
	})
	if err != nil {
		if ctx.Err() != nil {
			return // Context cancelled
		}
		log.Printf("receive error: %v\n", err)
		return
	}

	for _, msg := range out.Messages {
		if msg.Body == nil {
			continue
		}

		var ev event.Event
		// NOTE: Assuming Raw Message Delivery is ENABLED on SNS subscription.
		// If disabled, we would need to unmarshal the SNS JSON wrapper first -> Message field -> event.Event
		if err := json.Unmarshal([]byte(*msg.Body), &ev); err != nil {
			log.Printf("unmarshal error (poison message?): %v. Body: %s\n", err, *msg.Body)
			// 本来はDLQに送るべきだが、ここではVisibilityTimeout経過後に再試行されるのを防ぐため削除するか、
			// ログに出して無視する。ここでは削除して次へ進む運用とする。
			deleteMessage(ctx, client, queueURL, msg.ReceiptHandle)
			continue
		}

		log.Printf("Processing event: ID=%s Type=%s\n", ev.ID, ev.Type)

		// イベントタイプによる分岐
		switch ev.Type {
		case "job.created":
			if err := handler.Handle(ctx, ev); err != nil {
				log.Printf("Handler failed: %v. Message will be retried (VisibilityTimeout).\n", err)
				// 削除しない = VisibilityTimeout後に再出現
				continue
			}
		default:
			log.Printf("Unknown event type: %s\n", ev.Type)
		}

		// 処理成功または未知のイベントなら削除
		deleteMessage(ctx, client, queueURL, msg.ReceiptHandle)
		log.Printf("Event processed & deleted: %s\n", ev.ID)
	}
}

func deleteMessage(ctx context.Context, client *sqs.Client, queueURL string, receiptHandle *string) {
	_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		log.Printf("delete error: %v\n", err)
	}
}
