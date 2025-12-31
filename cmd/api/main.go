package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"async-messaging/internal/handler"
	"async-messaging/internal/stream"
)

func main() {
	// .env読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Kafka設定
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "events"
	}

	// Producer初期化
	producer := stream.NewProducer([]string{kafkaBroker}, kafkaTopic)
	defer producer.Close()

	// Echo初期化(Web サーバーを作っている)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ハンドラー登録
	h := handler.NewCreateJobHandler(producer)
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
