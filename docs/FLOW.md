# API呼び出しフロー

```
POST /jobs リクエスト
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ cmd/api/main.go                                         │
│   main()                                                │
│     └─ e.POST("/jobs", h.Handle)                        │
│     └─ stream.NewProducer(...)    ← Kafka Producer初期化│
└─────────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/handler/create_job.go                          │
│   CreateJobHandler.Handle(c echo.Context)               │
│     ├─ c.Bind(&req)           ← リクエストパース        │
│     ├─ uuid.NewString()       ← ジョブID生成            │
│     ├─ h.producer.Produce()   ← Kafkaへ送信 ─────────┐  │
│     └─ return 202 Accepted    ← レスポンス返却       │  │
└──────────────────────────────────────────────────────│──┘
                                                       │
       ┌───────────────────────────────────────────────┘
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/stream/producer.go                             │
│   Producer.Produce(ctx, ev)                             │
│     └─ kafka.Writer.WriteMessages() ← Kafkaへ書き込み   │
└─────────────────────────────────────────────────────────┘
       │
       ▼
    [Kafka Topic: events] ──(Consumer Group)──┐
                                              │
       ┌──────────────────────────────────────┘
       ▼
┌─────────────────────────────────────────────────────────┐
│ cmd/bridge/main.go                                      │
│   main()                                                │
│     └─ stream.NewConsumer(...)                          │
│     └─ pubsub.NewPublisher(...)                         │
│     └─ consumer.Consume()     ← 無限ループ              │
│          ├─ publisher.Publish() ──────────────┐         │
│          └─ consumer.CommitMessages()         │         │
└──────────────────────────────────────────────────────│──┘
                                                       │
       ┌───────────────────────────────────────────────┘
       ▼
    [SNS Topic] ──(Subscription)──▶ [SQS Queue]
                                       │
       ┌───────────────────────────────┘
       ▼
┌─────────────────────────────────────────────────────────┐
│ cmd/worker/main.go                                      │
│   main()                                                │
│     └─ processMessages()      ← Long polling            │
│           ├─ client.ReceiveMessage()                    │
│           ├─ json.Unmarshal() ← イベント復元            │
│           ├─ handler.Handle() ← ユースケース呼び出し ─┐ │
│           └─ client.DeleteMessage() ← 削除            │ │
└───────────────────────────────────────────────────────│─┘
```

## 呼び出し順序まとめ

| # | コンポーネント | 技術 | 役割 |
|---|---|---|---|
| 1 | **API** | Kafka Producer | イベントを「事実」としてKafkaに記録 (Source of Truth) |
| 2 | **Bridge** | Kafka Consumer + SNS Publisher | Kafkaからイベントを読み取り、SNSへ即時配信 |
| 3 | **SNS** | AWS SNS | ファンアウト（今回はSQSへ） |
| 4 | **SQS** | AWS SQS | バッファリング、再試行管理 |
| 5 | **Worker** | SQS Consumer | 重い処理を実行 (失敗時はSQSがリトライ) |
