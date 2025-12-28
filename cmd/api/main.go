package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"async-messaging/internal/handler"
	"async-messaging/internal/pubsub"
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

	// SNSクライアント作成
	snsClient := sns.NewFromConfig(cfg)

	// 環境変数からトピックARNを取得
	topicArn := os.Getenv("SNS_TOPIC_ARN")
	// 開発用ダミーARN (本来は必須エラーにするべきだが、動作確認用に入れている場合もある)
	if topicArn == "" {
		topicArn = "arn:aws:sns:ap-northeast-1:000000000000:job-topic"
		log.Println("WARNING: SNS_TOPIC_ARN is not set, using dummy ARN")
	}

	// Publisher初期化
	publisher := pubsub.NewPublisher(snsClient, topicArn)

	// Echo初期化(Web サーバーを作っている)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ハンドラー登録
	h := handler.NewCreateJobHandler(publisher)
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
