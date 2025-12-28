# AWS インフラ構築手順 (Pub/Sub)

本アプリケーションを正常に動作させるために必要な AWS リソースの設定手順です。

## 1. リソース作成

### A. SNS トピックの作成
1. **サービス**: Amazon SNS
2. **タイプ**: Standard (標準)
3. **名前**: `job-topic` (例)
4. **ARN をメモ**: API 側の `.env` (`SNS_TOPIC_ARN`) に設定します。

### B. SQS キューの作成
1. **サービス**: Amazon SQS
2. **タイプ**: Standard (標準)
3. **名前**: `job-queue` (例)
4. **URL をメモ**: Worker 側の `.env` (`SQS_QUEUE_URL`) に設定します。
5. **ARN をメモ**: 次のサブスクリプション設定で使用します。

---

## 2. サブスクリプション設定 (最重要)

SNS トピックから SQS キューへメッセージを流すための設定です。

1. 作成した SNS トピック (`job-topic`) の画面を開く。
2. 「**サブスクリプションの作成**」をクリック。
3. **プロトコル**: `Amazon SQS` を選択。
4. **エンドポイント**: 作成した **SQS キューの ARN** を入力 (URL ではありません)。
5. **サブスクリプションフィルターポリシー**: (デフォルトのまま)
6. **Redrive policy (dead-letter queue)**: (必要に応じて設定)
7. **Raw message delivery (生のメッセージ配信)**: **有効化 (Enable)** 🟥 **必須**
   - ※ これを有効にしないと、Worker がメッセージを正しくパースできません。

---

## 3. アクセス許可設定

### A. SQS アクセスポリシー (SNS からの受信許可)
SQS キュー (`job-queue`) の「アクセスポリシー」タブで、SNS トピックからの送信を許可します。

```json
{
  "Version": "2012-10-17",
  "Id": "AllowSNS",
  "Statement": [
    {
      "Sid": "Allow-SNS-SendMessage",
      "Effect": "Allow",
      "Principal": {
        "Service": "sns.amazonaws.com"
      },
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:ap-northeast-1:XXXXXXXXXXXX:job-queue",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "arn:aws:sns:ap-northeast-1:XXXXXXXXXXXX:job-topic"
        }
      }
    }
  ]
}
```
※ `Resource` (キューARN) と `aws:SourceArn` (トピックARN) を実際の値に書き換えてください。

### B. API 実行ユーザー (Publisher)
API を実行する IAM ユーザー/ロールに `sns:Publish` 権限が必要です。

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sns:Publish",
            "Resource": "arn:aws:sns:ap-northeast-1:XXXXXXXXXXXX:job-topic"
        }
    ]
}
```

### C. Worker 実行ユーザー (Subscriber)
Worker を実行する IAM ユーザー/ロールに SQS の読み取り/削除権限が必要です。

- `sqs:ReceiveMessage`
- `sqs:DeleteMessage`
- `sqs:GetQueueAttributes`

---

## 4. 動作確認

1. `.env` ファイルの設定を確認。
2. `go run ./cmd/api` と `go run ./cmd/worker` を起動。
3. API にリクエスト送信し、Worker で受信されることを確認。

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test", "input": "hello"}'
```
