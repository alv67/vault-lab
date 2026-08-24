package model

import (
	"time"

	"github.com/google/uuid"
)

type HealthEvent struct {
	ID          uuid.UUID `json:"id"`
	AssetID     *uuid.UUID `json:"asset_id,omitempty"`
	EventType   string    `json:"event_type"`
	Status      string    `json:"status"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	DurationMs  int       `json:"duration_ms"`
	ErrorCode   string    `json:"error_code"`
	CreatedAt   time.Time `json:"created_at"`
}

type HealthSummary struct {
	Successes   int     `json:"successes"`
	Failures    int     `json:"failures"`
	SuccessRate float64 `json:"success_rate"`
	RateLimited int     `json:"rate_limited"`
}
