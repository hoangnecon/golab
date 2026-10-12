package store

import "time"

// Profile represents an isolated Google account / Chrome user profile.
type Profile struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	AccountEmail string    `json:"account_email,omitempty"`
	Tier         string    `json:"tier"` // "free", "pro", "pro_plus"
	UserDataDir  string    `json:"user_data_dir"`
	CDPPort      int       `json:"cdp_port"`
	IsRunning    bool      `json:"is_running"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session represents a connected Google Colab notebook worktree.
type Session struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ProfileID      string    `json:"profile_id"`
	NotebookURL    string    `json:"notebook_url"`
	NotebookID     string    `json:"notebook_id,omitempty"`
	TabID          string    `json:"tab_id,omitempty"`
	IsActive       bool      `json:"is_active"`
	KernelState    string    `json:"kernel_state,omitempty"`
	GPUName        string    `json:"gpu_name,omitempty"`
	GPUMemoryFree  int64     `json:"gpu_memory_free_mb,omitempty"`
	GPUMemoryTotal int64     `json:"gpu_memory_total_mb,omitempty"`
	RAMFreeGB      float64   `json:"ram_free_gb,omitempty"`
	DriveMounted   bool      `json:"drive_mounted"`
	CreatedAt      time.Time `json:"created_at"`
	LastAttachedAt time.Time `json:"last_attached_at,omitempty"`
}

// ActivityLog records Agent actions for real-time audit streaming.
type ActivityLog struct {
	ID           int64     `json:"id"`
	SessionID    string    `json:"session_id,omitempty"`
	ToolName     string    `json:"tool_name"`
	LatencyMs    int64     `json:"latency_ms"`
	Status       string    `json:"status"` // "success" or "error"
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
