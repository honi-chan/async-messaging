package event

import (
	"encoding/json"
	"time"
)

// Event is the common event structure for Pub/Sub
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

// JobPayload is the specific payload for job events
type JobPayload struct {
	UserID string `json:"user_id"`
	Input  string `json:"input"`
}
