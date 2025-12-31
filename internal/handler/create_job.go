package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"async-messaging/internal/event"
	"async-messaging/internal/stream"
)

// CreateJobRequest はジョブ作成リクエスト
type CreateJobRequest struct {
	UserID string `json:"user_id"`
	Input  string `json:"input"`
}

// CreateJobHandler はジョブ作成ハンドラー
type CreateJobHandler struct {
	producer *stream.Producer
}

// NewCreateJobHandler は新しいハンドラーを作成
func NewCreateJobHandler(p *stream.Producer) *CreateJobHandler {
	return &CreateJobHandler{producer: p}
}

// Handle はジョブ作成リクエストを処理
// 非同期APIの要：処理は行わず、SNSトピックへのPublishのみ
func (h *CreateJobHandler) Handle(c echo.Context) error {
	var req CreateJobRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	payload := event.JobPayload{
		UserID: req.UserID,
		Input:  req.Input,
	}
	payloadBytes, _ := json.Marshal(payload)

	ev := event.Event{
		ID:            uuid.NewString(),
		Type:          "job.created",
		AggregateID:   uuid.NewString(), // New aggregate for each job
		AggregateType: "job",
		SchemaVersion: 1,
		OccurredAt:    time.Now(),
		Payload:       payloadBytes,
	}

	if err := h.producer.Produce(c.Request().Context(), ev); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "produce failed"})
	}

	// 非同期APIの要：200ではなく202 Accepted
	return c.JSON(http.StatusAccepted, map[string]string{
		"job_id": ev.ID,
		"status": "accepted",
	})
}
