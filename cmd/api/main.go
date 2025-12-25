package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"async-messaging/internal/handler"
	"async-messaging/internal/queue"
)

func main() {
	// .env読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// AWS設定の読み込み
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	// 環境変数からキューURLを取得（本番環境向け）
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		queueURL = "https://sqs.ap-northeast-1.amazonaws.com/466703337425/job-queue"
	}

	queueClient := queue.NewClient(sqsClient, queueURL)

	// Echo初期化
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ハンドラー登録
	h := handler.NewCreateJobHandler(queueClient)
	e.POST("/jobs", h.Handle)

	// ヘルスチェック
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
