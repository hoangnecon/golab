package telemetry

// RAMTelemetry contains system RAM metrics.
type RAMTelemetry struct {
	TotalBytes int64   `json:"total_bytes"`
	FreeBytes  int64   `json:"free_bytes"`
	UsedBytes  int64   `json:"used_bytes"`
	TotalGB    float64 `json:"total_gb"`
	FreeGB     float64 `json:"free_gb"`
	UsedGB     float64 `json:"used_gb"`
}

// DiskTelemetry contains disk usage metrics.
type DiskTelemetry struct {
	Label       string  `json:"label"`
	TotalBytes  int64   `json:"total_bytes"`
	UsedBytes   int64   `json:"used_bytes"`
	TotalGB     float64 `json:"total_gb"`
	UsedGB      float64 `json:"used_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// GPUTelemetry contains detailed NVIDIA GPU metrics (e.g. Tesla T4).
type GPUTelemetry struct {
	Name              string  `json:"name"`
	MemoryUsedBytes   int64   `json:"memory_used_bytes"`
	MemoryTotalBytes  int64   `json:"memory_total_bytes"`
	MemoryUsedMB      float64 `json:"memory_used_mb"`
	MemoryTotalMB     float64 `json:"memory_total_mb"`
	MemoryFreeMB      float64 `json:"memory_free_mb"`
	GPUUtilization    float64 `json:"gpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
	EverUsed          bool    `json:"ever_used"`
}

// HardwareTelemetry encapsulates full machine telemetry.
type HardwareTelemetry struct {
	RAM   RAMTelemetry    `json:"ram"`
	Disks []DiskTelemetry `json:"disks"`
	GPUs  []GPUTelemetry  `json:"gpus"`
}

// BuildInfo contains Colab container release metadata.
type BuildInfo struct {
	ReleaseTag string `json:"release_tag"`
}

// DriveStatus contains Google Drive health and mount info.
type DriveStatus struct {
	Mounted bool           `json:"mounted"`
	Details map[string]any `json:"details,omitempty"`
}
