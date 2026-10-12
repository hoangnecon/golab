package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) colabGetHardwareTelemetry(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	telem, err := svcs.Telemetry.GetHardwareTelemetry(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_get_hardware_telemetry", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(telem, latency, "")
}

func (s *Server) colabGetBuildInfo(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	info, err := svcs.Telemetry.GetBuildInfo(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_get_build_info", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(info, latency, "")
}

func (s *Server) colabCheckDriveStatus(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	status, err := svcs.Telemetry.CheckDriveStatus(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_check_drive_status", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(status, latency, "")
}
