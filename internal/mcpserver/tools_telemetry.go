package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) colabGetHardwareTelemetry(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	telem, err := s.telemetry.GetHardwareTelemetry(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(telem, latency, "")
}

func (s *Server) colabGetBuildInfo(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	info, err := s.telemetry.GetBuildInfo(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(info, latency, "")
}

func (s *Server) colabCheckDriveStatus(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	status, err := s.telemetry.CheckDriveStatus(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(status, latency, "")
}
