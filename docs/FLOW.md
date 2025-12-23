# API呼び出しフロー

```
POST /jobs リクエスト
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ cmd/api/main.go                                         │
│   main()                                                │
│     └─ e.POST("/jobs", h.Handle)  ← ルーティング登録    │
└─────────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/handler/create_job.go                          │
│   CreateJobHandler.Handle(c echo.Context)               │
│     ├─ c.Bind(&req)           ← リクエストパース        │
│     ├─ uuid.NewString()       ← ジョブID生成            │
│     ├─ json.Marshal(ev)       ← イベントJSON化          │
│     ├─ h.queue.Send(...)      ← SQSへ送信 ───────────┐  │
│     └─ return 202 Accepted    ← レスポンス返却       │  │
└──────────────────────────────────────────────────────│──┘
                                                       │
       ┌───────────────────────────────────────────────┘
       ▼
┌─────────────────────────────────────────────────────────┐
│ internal/queue/sqs.go                                   │
│   Client.Send(ctx, body string)                         │
│     └─ sqs.SendMessage()      ← AWS SQS API呼び出し     │
└─────────────────────────────────────────────────────────┘
       │
       ▼
    [SQS Queue]
       │
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
│     ├─ ev.RetryCount >= maxRetry → discard              │
│     ├─ process(ev)            ← ビジネスロジック        │
│     └─ (失敗時) h.queue.Send() ← リトライキュー再投入   │
└─────────────────────────────────────────────────────────┘
```

## 呼び出し順序まとめ

| # | ファイル | 関数 | 役割 |
|---|---|---|---|
| 1 | `cmd/api/main.go` | `main()` | サーバー起動、ルーティング |
| 2 | `internal/handler/create_job.go` | `Handle()` | リクエスト処理、イベント作成 |
| 3 | `internal/queue/sqs.go` | `Send()` | SQSへメッセージ送信 |
| 4 | `cmd/worker/main.go` | `processMessages()` | SQSからメッセージ受信 |
| 5 | `internal/usecase/process_job.go` | `Handle()` | 冪等処理実行 |
