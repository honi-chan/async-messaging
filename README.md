# Go 非同期 API + SQS 実装完了

## 概要

**冪等・再実行前提**の非同期APIテンプレートを実装しました。

## プロジェクト構成

```
api/
├── cmd/
│   ├── api/main.go       # HTTP API サーバー
│   └── worker/main.go    # SQS Consumer
└── internal/
    ├── event/message.go  # イベントDTO
    ├── handler/create_job.go  # ジョブ作成ハンドラー
    ├── queue/sqs.go      # SQSクライアント
    └── usecase/process_job.go # 冪等処理ロジック
```

## 実装のポイント

| コンポーネント | 責務 |
|---|---|
| **API** | 受付のみ、キュー投入だけ → `202 Accepted` |
| **Worker** | 実際の処理、冪等性チェック、再実行対応 |

### 非同期API設計原則

- ✅ 処理しない
- ✅ DB書かない（極力）
- ✅ キュー送信だけ

### Worker設計原則

- ✅ 再実行前提
- ✅ 失敗しても壊れない
- ✅ DLQ 前提

## 環境変数 (.env)

プロジェクトルートに `.env` ファイルを作成して設定します。

```env
# AWS認証情報
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=ap-northeast-1

# SQS Queue URL
SQS_QUEUE_URL="https://sqs.ap-northeast-1.amazonaws.com/×/job-queue"

# アプリケーション設定
MAX_RETRY=5
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
