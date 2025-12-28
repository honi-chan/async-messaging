package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"async-messaging/internal/event"
)

// Handler はジョブ処理ハンドラー
type Handler struct{}

// NewHandler は新しいハンドラーを作成
func NewHandler() *Handler {
	return &Handler{}
}

// Handle はジョブを処理
// 失敗時はエラーを返し、SQSの標準リトライ機能を活用する (VisibilityTimeout)
func (h *Handler) Handle(ctx context.Context, ev event.Event) error {
	// ペイロードをパース
	var payload event.JobPayload
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return process(ev.ID, payload)
}

// process は実際のビジネスロジック
func process(jobID string, payload event.JobPayload) error {
	// TODO: 実際の処理を実装
	// 例: 外部API呼び出し、ファイル生成、DBへの書き込みなど
	fmt.Printf("Processing job logic: ID=%s, UserID=%s\n", jobID, payload.UserID)
	return nil
}
