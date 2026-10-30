package domain

import (
	"context"
	"time"
)

// PermissionLevel represents user access rights to a notebook.
type PermissionLevel string

const (
	PermissionEdit    PermissionLevel = "edit"
	PermissionView    PermissionLevel = "view"
	PermissionDenied  PermissionLevel = "denied"
	PermissionUnknown PermissionLevel = "unknown"
)

// Session represents an attached notebook worktree running under a specific profile.
type Session struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	ProfileID      string          `json:"profile_id"`
	ProjectID      string          `json:"project_id,omitempty"`
	OwnerProfileID string          `json:"owner_profile_id,omitempty"`
	NotebookURL    string          `json:"notebook_url"`
	NotebookID     string          `json:"notebook_id,omitempty"`
	TabID          string          `json:"tab_id,omitempty"`
	IsActive       bool            `json:"is_active"`
	IsLocked       bool            `json:"is_locked"`
	Permission     PermissionLevel `json:"permission,omitempty"`
	KernelState    string          `json:"kernel_state,omitempty"`
	GPUName        string          `json:"gpu_name,omitempty"`
	GPUMemoryFree  int64           `json:"gpu_memory_free_mb,omitempty"`
	GPUMemoryTotal int64           `json:"gpu_memory_total_mb,omitempty"`
	RAMFreeGB      float64         `json:"ram_free_gb,omitempty"`
	DriveMounted   bool            `json:"drive_mounted"`
	CreatedAt      time.Time       `json:"created_at"`
	LastAttachedAt time.Time       `json:"last_attached_at,omitempty"`
}

// SessionRepository defines persistence operations for Sessions.
type SessionRepository interface {
	SaveSession(ctx context.Context, sess *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	GetActiveSession(ctx context.Context) (*Session, error)
	GetLockedSession(ctx context.Context) (*Session, error)
	ListSessions(ctx context.Context) ([]*Session, error)
	SetActiveSession(ctx context.Context, id string) error
	SetSessionLocked(ctx context.Context, id string, locked bool) error
	SetSessionPermission(ctx context.Context, id string, perm PermissionLevel) error
	DeleteSession(ctx context.Context, id string) error
}
