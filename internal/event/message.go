package event

import "time"

// JobCreated はジョブ作成イベントのDTO
type JobCreated struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"` // ex: "job.created"
	Payload    Payload   `json:"payload"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// Payload はイベントのペイロード
type Payload struct {
	UserID string `json:"user_id"`
	Input  string `json:"input"`
}
