package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"async-messaging/internal/event"
	"async-messaging/internal/queue"
)

// CreateJobRequest はジョブ作成リクエスト
type CreateJobRequest struct {
	UserID string `json:"user_id"`
	Input  string `json:"input"`
}

// CreateJobHandler はジョブ作成ハンドラー
type CreateJobHandler struct {
	queue *queue.Client
}

// NewCreateJobHandler は新しいハンドラーを作成
func NewCreateJobHandler(q *queue.Client) *CreateJobHandler {
	return &CreateJobHandler{queue: q}
}

// Handle はジョブ作成リクエストを処理
// 非同期APIの要：処理は行わず、キューへの投入のみ
func (h *CreateJobHandler) Handle(c echo.Context) error {
	var req CreateJobRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	ev := event.JobCreated{
		ID:   uuid.NewString(),
		Type: "job.created",
		Payload: event.Payload{
			UserID: req.UserID,
			Input:  req.Input,
		},
		CreatedAt: time.Now(),
	}

	b, _ := json.Marshal(ev)
	if err := h.queue.Send(c.Request().Context(), string(b)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "queue failed"})
	}

	// 非同期APIの要：200ではなく202 Accepted
	// 成功 = キュー投入成功
	return c.JSON(http.StatusAccepted, map[string]string{
		"job_id": ev.ID,
		"status": "accepted",
	})
}
