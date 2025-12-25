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
	"async-messaging/internal/queue"
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

	// キュークライアントとハンドラー初期化
	queueClient := queue.NewClient(sqsClient, queueURL)
	handler := usecase.NewHandler(queueClient)

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
		var ev event.JobCreated
		if err := json.Unmarshal([]byte(*msg.Body), &ev); err != nil {
			log.Printf("unmarshal error (poison message): %v\n", err)
			continue
		}

		log.Printf("Processing job: %s (type: %s, retry: %d)\n", ev.ID, ev.Type, ev.RetryCount)

		// ハンドラーで処理（失敗時は内部でリトライキューに再投入）
		_ = handler.Handle(ctx, ev)

		// 元メッセージは必ず削除
		_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			log.Printf("delete error: %v\n", err)
		}

		log.Printf("Job processed: %s\n", ev.ID)
	}
}
