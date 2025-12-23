package queue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Client はSQSキュークライアント
type Client struct {
	sqs      *sqs.Client
	queueURL string
}

// NewClient は新しいSQSクライアントを作成
func NewClient(sqsClient *sqs.Client, queueURL string) *Client {
	return &Client{sqs: sqsClient, queueURL: queueURL}
}

// Send はメッセージをキューに送信
func (c *Client) Send(ctx context.Context, body string) error {
	_, err := c.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &c.queueURL,
		MessageBody: &body,
	})
	return err
}
