# API呼び出しフロー

```
POST /jobs リクエスト
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ cmd/api/main.go                                         │
│   main()                                                │
│     └─ e.POST("/jobs", h.Handle)  ← ルーティング登録    │
│     └─ pubsub.NewPublisher(...)   ← SNS Publisher初期化 │
└─────────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/handler/create_job.go                          │
│   CreateJobHandler.Handle(c echo.Context)               │
│     ├─ c.Bind(&req)           ← リクエストパース        │
│     ├─ uuid.NewString()       ← ジョブID生成            │
│     ├─ json.Marshal(payload)  ← ペイロードJSON化        │
│     ├─ h.publisher.Publish()  ← SNSへ送信 ───────────┐  │
│     └─ return 202 Accepted    ← レスポンス返却       │  │
└──────────────────────────────────────────────────────│──┘
                                                       │
       ┌───────────────────────────────────────────────┘
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/pubsub/publisher.go                            │
│   Publisher.Publish(ctx, ev)                            │
│     └─ sns.Publish()          ← AWS SNS API呼び出し     │
└─────────────────────────────────────────────────────────┘
       │
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
                                                        │
       ┌────────────────────────────────────────────────┘
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/usecase/process_job.go                         │
│   Handler.Handle(ctx, ev)                               │
│     ├─ json.Unmarshal(payload)                          │
│     └─ process()              ← ビジネスロジック        │
│                               (失敗時はerror返却        │
│                                → VisibilityTimeout      │
│                                → SQSが自動再試行)       │
└─────────────────────────────────────────────────────────┘
```

## 呼び出し順序まとめ

| # | ファイル | 関数 | 役割 |
|---|---|---|---|
| 1 | `cmd/api/main.go` | `main()` | サーバー起動、SNS Publisher設定 |
| 2 | `internal/handler/create_job.go` | `Handle()` | イベント作成、Publish呼び出し |
| 3 | `internal/pubsub/publisher.go` | `Publish()` | SNS TopicへPublish |
| 4 | `cmd/worker/main.go` | `processMessages()` | SQSからメッセージ受信 (SNS経由) |
| 5 | `internal/usecase/process_job.go` | `Handle()` | 処理実行 (再試行はSQS任せ) |
