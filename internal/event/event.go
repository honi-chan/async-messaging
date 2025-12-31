package event

import (
	"encoding/json"
	"time"
)

// Event is the common event structure for Pub/Sub and Event Sourcing
type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	AggregateID   string          `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	SchemaVersion int             `json:"schema_version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

// JobPayload is the specific payload for job events
type JobPayload struct {
	UserID string `json:"user_id"`
	Input  string `json:"input"`
}
