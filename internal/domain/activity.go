package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// ErrNotFound indicates a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ActivityLog records Agent actions for real-time audit streaming.
type ActivityLog struct {
	ID           int64     `json:"id"`
	SessionID    string    `json:"session_id,omitempty"`
	ToolName     string    `json:"tool_name"`
	LatencyMs    int64     `json:"latency_ms"`
	Status       string    `json:"status"` // "success" or "error"
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	HasPayload   bool      `json:"has_payload"`
}

// Payload is a captured tool request or response body.
type Payload struct {
	Body      json.RawMessage `json:"body"`
	Bytes     int             `json:"bytes"`
	Truncated bool            `json:"truncated"`
}

// ActivityRecord is a tool call to persist together with its payloads.
type ActivityRecord struct {
	SessionID    string
	ToolName     string
	LatencyMs    int64
	Status       string
	ErrorMessage string
	Request      *Payload
	Response     *Payload
}

// ActivityDetail is one tool call with its captured payloads.
type ActivityDetail struct {
	ActivityLog
	Request  *Payload `json:"request"`
	Response *Payload `json:"response"`
}

// ActivityRepository defines audit logging persistence.
type ActivityRepository interface {
	LogActivity(ctx context.Context, sessionID, toolName string, latencyMs int64, status, errorMessage string) error
	RecordActivity(ctx context.Context, rec *ActivityRecord) (int64, error)
	ListRecentActivity(ctx context.Context, limit int) ([]*ActivityLog, error)
	GetActivity(ctx context.Context, id int64) (*ActivityDetail, error)
}
