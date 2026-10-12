package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ProxyForwardInput specifies reverse proxy target inside Colab VM.
type ProxyForwardInput struct {
	SessionID string `json:"session_id,omitempty" jsonschema:"Optional session ID to route this command. If omitted, routes to the active notebook session."`
	Port      int    `json:"port" jsonschema:"Internal port on the Colab VM to route request to (e.g. 7860, 8000, 8501, 3453). Required."`
	Subpath   string `json:"subpath,omitempty" jsonschema:"Subpath to request (e.g. 'api/status', 'metrics', or empty for root)."`
	Method    string `json:"method,omitempty" jsonschema:"HTTP method ('GET', 'POST', 'PUT', 'DELETE'). Default 'GET'."`
	Body      string `json:"body,omitempty" jsonschema:"Request body payload (if applicable)."`
}

func (s *Server) colabProxyForward(ctx context.Context, req *mcp.CallToolRequest, input ProxyForwardInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	svcs, sessID, err := s.getServices(ctx, input.SessionID)
	if err != nil {
		return errorResult(err, 0)
	}
	resp, err := svcs.Proxy.ProxyForward(ctx, input.Port, input.Subpath, input.Method, nil, input.Body)
	latency := time.Since(start).Milliseconds()
	s.logActivity(sessID, "colab_proxy_forward", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(resp, latency, "")
}
