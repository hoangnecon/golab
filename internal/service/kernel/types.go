package kernel

// KernelDetail represents an active Jupyter kernel in the Colab container.
type KernelDetail struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastActivity   string `json:"last_activity"`
	ExecutionState string `json:"execution_state"` // "idle", "busy", "starting"
	Connections    int    `json:"connections"`
}

// SessionDetail represents an active notebook session mapping.
type SessionDetail struct {
	ID       string        `json:"id"`
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Kernel   *KernelDetail `json:"kernel,omitempty"`
	Notebook map[string]any `json:"notebook,omitempty"`
}

// KernelStatusReport gives a complete real-time view of the kernel state.
type KernelStatusReport struct {
	IsConnected    bool          `json:"is_connected"`
	ActiveSession  *SessionDetail `json:"active_session,omitempty"`
	Kernel         *KernelDetail  `json:"kernel,omitempty"`
	AllKernels     []KernelDetail `json:"all_kernels"`
	TotalKernels   int           `json:"total_kernels"`
}
