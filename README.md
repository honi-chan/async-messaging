# Go 非同期 API + SQS 実装完了

## 概要

**冪等・再実行前提**の非同期APIテンプレートを実装しました。

## プロジェクト構成

## プロジェクト構成

```
api/
├── cmd/
│   ├── api/main.go       # HTTP API サーバー
│   └── worker/main.go    # SQS Consumer (SNS Subscriber)
└── internal/
    ├── event/event.go    # 共通イベント定義
    ├── handler/create_job.go  # ジョブ作成ハンドラー
    ├── pubsub/publisher.go    # SNS Publisher
    └── usecase/process_job.go # 処理ロジック
```

## 実装のポイント

| コンポーネント | 責務 |
|---|---|
| **API** | SNS トピックへの Publish のみ（疎結合） |
| **Worker** | SQS から受信。再試行は SQS の標準機能（VisibilityTimeout + DLQ）に任せる |

### 非同期API設計原則 (Pub/Sub)

- ✅ Publisher は Subscriber を知らない
- ✅ イベントは事実として定義

### Worker設計原則

- ✅ SQS 標準のリトライ機構を利用
- ✅ 手動での再エンキューは行わない (Basic No Philosophy)

## 環境変数 (.env)

プロジェクトルートに `.env` ファイルを作成して設定します。

```env
# AWS認証情報
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=ap-northeast-1

# SNS Topic ARN (API用)
SNS_TOPIC_ARN="arn:aws:sns:ap-northeast-1:000000000000:job-topic"

# SQS Queue URL (Worker用)
SQS_QUEUE_URL="https://sqs.ap-northeast-1.amazonaws.com/×/job-queue"

# アプリケーション設定
PORT=8080
```

## 使い方

```bash
# API起動
go run ./cmd/api

# Worker起動（別ターミナル）
go run ./cmd/worker
```
※ `.env` ファイルが自動的に読み込まれます。


## API呼び出し例

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "input": "test data"}'
```

レスポンス:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "accepted"
}
```

## ドキュメント

- [API呼び出しフロー](./docs/FLOW.md)

## ビルド検証

✅ `go build ./...` 成功
