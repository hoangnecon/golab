package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SessionOnlyInput struct {
	SessionID string `json:"session_id,omitempty" jsonschema:"Optional session ID to route this command. If omitted, routes to the active notebook session."`
}

func (s *Server) colabGetKernelStatus(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	rep, err := svcs.Kernel.GetStatusReport(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_get_kernel_status", latency, err)
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
	SessionID string `json:"session_id,omitempty" jsonschema:"Optional session ID to route this command. If omitted, routes to the active notebook session."`
	KernelID  string `json:"kernel_id,omitempty" jsonschema:"ID of the kernel to target. Leave empty to automatically target the active notebook kernel."`
}

func (s *Server) colabInterruptKernel(ctx context.Context, req *mcp.CallToolRequest, input KernelTargetInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	err = svcs.Kernel.InterruptKernel(ctx, input.KernelID)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_interrupt_kernel", latency, err)
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
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	detail, err := svcs.Kernel.RestartKernel(ctx, input.KernelID)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_restart_kernel", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{
		"restarted": true,
		"kernel":    detail,
		"message":   "Kernel restarted successfully. GPU VRAM cleared.",
	}, latency, detail.ExecutionState)
}

func (s *Server) colabListSessions(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	sessions, err := svcs.Kernel.ListSessions(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_list_sessions", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(sessions, latency, "")
}

func (s *Server) colabSpawnSecondaryKernel(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	detail, err := svcs.Kernel.SpawnSecondaryKernel(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_spawn_secondary_kernel", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{
		"spawned": true,
		"kernel":  detail,
		"message": "Secondary Python 3 kernel spawned on Colab VM.",
	}, latency, detail.ExecutionState)
}
