package usecase

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"async-messaging/internal/event"
	"async-messaging/internal/queue"
)

// getMaxRetry は環境変数からmaxRetryを取得（デフォルト: 5）
func getMaxRetry() int {
	if v := os.Getenv("MAX_RETRY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 5
}

// Handler はジョブ処理ハンドラー
type Handler struct {
	queue *queue.Client
}

// NewHandler は新しいハンドラーを作成
func NewHandler(q *queue.Client) *Handler {
	return &Handler{queue: q}
}

// Handle はジョブを処理し、失敗時はリトライキューに再投入
func (h *Handler) Handle(ctx context.Context, ev event.JobCreated) error {
	if ev.RetryCount >= getMaxRetry() {
		log.Printf("discard message id=%s after %d retries", ev.ID, ev.RetryCount)
		return nil
	}

	if err := process(ev); err != nil {
		ev.RetryCount++

		body, _ := json.Marshal(ev)
		_ = h.queue.Send(ctx, string(body))

		return nil
	}

	return nil
}

// process は実際のビジネスロジック
func process(ev event.JobCreated) error {
	// TODO: 実際の処理を実装
	// 例: 外部API呼び出し、ファイル生成、DBへの書き込みなど
	return nil
}
