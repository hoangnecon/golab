package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) colabGetKernelStatus(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	rep, err := s.kernel.GetStatusReport(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}

	state := "unknown"
	if rep.Kernel != nil {
		state = rep.Kernel.ExecutionState
	}
	return successResult(rep, latency, state)
}

// KernelTargetInput specifies target kernel (optional).
type KernelTargetInput struct {
	KernelID string `json:"kernel_id,omitempty" jsonschema:"ID of the kernel to target. Leave empty to automatically target the active notebook kernel."`
}

func (s *Server) colabInterruptKernel(ctx context.Context, req *mcp.CallToolRequest, input KernelTargetInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.kernel.InterruptKernel(ctx, input.KernelID)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{
		"interrupted": true,
		"kernel_id":   input.KernelID,
		"message":     "SIGINT interrupt signal dispatched successfully. Kernel state preserved.",
	}, latency, "idle")
}

func (s *Server) colabRestartKernel(ctx context.Context, req *mcp.CallToolRequest, input KernelTargetInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	detail, err := s.kernel.RestartKernel(ctx, input.KernelID)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{
		"restarted": true,
		"kernel":    detail,
		"message":   "Kernel restarted successfully. GPU VRAM cleared.",
	}, latency, detail.ExecutionState)
}

func (s *Server) colabListSessions(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	sessions, err := s.kernel.ListSessions(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(sessions, latency, "")
}

func (s *Server) colabSpawnSecondaryKernel(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	detail, err := s.kernel.SpawnSecondaryKernel(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{
		"spawned": true,
		"kernel":  detail,
		"message": "Secondary Python 3 kernel spawned on Colab VM.",
	}, latency, detail.ExecutionState)
}
