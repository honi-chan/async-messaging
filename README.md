# Go 非同期 API + SQS 実装完了

## 概要

**冪等・再実行前提**の非同期APIテンプレートを実装しました。


## アーキテクチャ

**Event Stream (Kafka) → Pub/Sub (SNS) → Queue (SQS)**

1. **API** -> **Kafka**: イベントを永続化（再計算・監査用）
2. **Bridge** -> **SNS**: リアルタイム通知
3. **SNS** -> **SQS**: ファンアウト・バッファリング
4. **Worker**: 処理（失敗時はSQSがリトライ）

## 環境設定

## 環境設定

1. **Docker (Kafka) 起動**
   ```bash
   docker-compose up -d
   ```
   * ローカルの `localhost:9092` でKafkaが起動します。
   * コンテナ名: `kafka`

2. **Topic 作成 (初回のみ)**
   ```bash
   docker exec -it kafka kafka-topics \
     --create \
     --topic events \
     --bootstrap-server localhost:9092 \
     --partitions 1 \
     --replication-factor 1
   ```

3. **`.env` ファイル設定**
   `.env.example` をコピーして `.env` を作成します。
   ```bash
   cp .env.example .env
   ```
   **推奨設定:**
   ```env
   KAFKA_BROKER=localhost:9092
   KAFKA_TOPIC=events
   ```

### 動作確認（疎通チェック）

**Consumer（受信待機）**
```bash
docker exec -it kafka kafka-console-consumer \
  --topic events \
  --bootstrap-server localhost:9092
```

**Producer（送信テスト）** -- 別ターミナルで実行
```bash
docker exec -it kafka kafka-console-producer \
  --topic events \
  --bootstrap-server localhost:9092
```
入力後にEnterを押すと Consumer 側に表示されます。

## 使い方

1. Kafka起動
   ```bash
   docker-compose up -d
   ```

2. サービス起動（別々のターミナルで）

   ```bash
   # API: ポート8080
   go run ./cmd/api

   # Bridge: Kafka -> SNS
   go run ./cmd/bridge

   # Worker: SQS Consumer
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
