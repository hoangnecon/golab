package domain

import (
	"context"
	"encoding/json"
)

// FetchOptions configures an internal container tunnel REST request.
type FetchOptions struct {
	Method  string
	Headers map[string]string
	Body    string
}

// FetchResponse represents the parsed response from the internal container tunnel gateway.
type FetchResponse struct {
	Status     int             `json:"status"`
	StatusText string          `json:"statusText"`
	OK         bool            `json:"ok"`
	Body       json.RawMessage `json:"body"`
	RawText    string          `json:"rawText,omitempty"`
}

// AcceleratorOption represents a hardware accelerator (GPU/TPU/CPU).
type AcceleratorOption struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Checked  bool   `json:"checked"`
	Disabled bool   `json:"disabled"`
}

// RuntimeConfigReport details the available and active runtime hardware.
type RuntimeConfigReport struct {
	Accelerators []AcceleratorOption `json:"accelerators"`
	HighRAM      bool                `json:"high_ram"`
}

// RunningCellInfo represents the execution state and live logs of a cell.
type RunningCellInfo struct {
	Index          int     `json:"index"`
	CellID         string  `json:"cell_id"`
	IsRunning      bool    `json:"is_running"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
	ExecutionCount string  `json:"execution_count"`
	Stdout         string  `json:"stdout"`
}

// RunningCellsReport provides an overview of currently executing cells.
type RunningCellsReport struct {
	HasRunningCells bool              `json:"has_running_cells"`
	TotalRunning    int               `json:"total_running"`
	RunningCells    []RunningCellInfo `json:"running_cells"`
}

// GolabDriver abstracts direct CDP and tunnel communication with the notebook page.
type GolabDriver interface {
	// Low-level transport
	FetchKernelXHR(ctx context.Context, path string, opts FetchOptions) (*FetchResponse, error)
	CallMcpTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error)

	// Runtime Lifecycle
	ConnectRuntime(ctx context.Context) error
	DisconnectRuntime(ctx context.Context) error
	GetAvailableAccelerators(ctx context.Context) (*RuntimeConfigReport, error)
	SetAccelerator(ctx context.Context, value string, highRAM bool) error
	InterruptKernel(ctx context.Context) error
	GetKernelStatus(ctx context.Context) (map[string]any, error)

	// Notebook Automation
	GetRunningCells(ctx context.Context) (*RunningCellsReport, error)
	SaveNotebook(ctx context.Context) error
	RunAll(ctx context.Context) error
}
