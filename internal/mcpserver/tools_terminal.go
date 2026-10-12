package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) colabListTerminals(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	list, err := svcs.Terminal.ListTerminals(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_list_terminals", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(list, latency, "")
}

func (s *Server) colabCreateTerminal(ctx context.Context, req *mcp.CallToolRequest, input SessionOnlyInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	info, err := svcs.Terminal.CreateTerminal(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_create_terminal", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(info, latency, "")
}

// DeleteTerminalInput specifies terminal name to close.
type DeleteTerminalInput struct {
	SessionID string `json:"session_id,omitempty" jsonschema:"Optional session ID to route this command. If omitted, routes to the active notebook session."`
	Name      string `json:"name" jsonschema:"Identifier name of the terminal session (e.g. '1'). Required."`
}

func (s *Server) colabDeleteTerminal(ctx context.Context, req *mcp.CallToolRequest, input DeleteTerminalInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	err = svcs.Terminal.DeleteTerminal(ctx, input.Name)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_delete_terminal", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"deleted": true, "name": input.Name}, latency, "")
}
