package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) colabListTerminals(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	list, err := s.terminal.ListTerminals(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(list, latency, "")
}

func (s *Server) colabCreateTerminal(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	info, err := s.terminal.CreateTerminal(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(info, latency, "")
}

// DeleteTerminalInput specifies terminal name to close.
type DeleteTerminalInput struct {
	Name string `json:"name" jsonschema:"Identifier name of the terminal session (e.g. '1'). Required."`
}

func (s *Server) colabDeleteTerminal(ctx context.Context, req *mcp.CallToolRequest, input DeleteTerminalInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.terminal.DeleteTerminal(ctx, input.Name)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"deleted": true, "name": input.Name}, latency, "")
}
