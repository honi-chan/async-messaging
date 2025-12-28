package pubsub

import (
	"context"
	"encoding/json"

	"async-messaging/internal/event"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// Publisher handles sending events to SNS
type Publisher struct {
	client   *sns.Client
	topicArn string
}

// NewPublisher creates a new SNS publisher
func NewPublisher(client *sns.Client, topicArn string) *Publisher {
	return &Publisher{
		client:   client,
		topicArn: topicArn,
	}
}

// Publish sends an event to the SNS topic
func (p *Publisher) Publish(ctx context.Context, ev event.Event) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: &p.topicArn,
		Message:  aws.String(string(b)),
	})
	return err
}
