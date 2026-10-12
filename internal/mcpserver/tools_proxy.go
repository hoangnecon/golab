package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ProxyForwardInput specifies reverse proxy target inside Colab VM.
type ProxyForwardInput struct {
	Port    int    `json:"port" jsonschema:"Internal port on the Colab VM to route request to (e.g. 7860, 8000, 8501, 3453). Required."`
	Subpath string `json:"subpath,omitempty" jsonschema:"Subpath to request (e.g. 'api/status', 'metrics', or empty for root)."`
	Method  string `json:"method,omitempty" jsonschema:"HTTP method ('GET', 'POST', 'PUT', 'DELETE'). Default 'GET'."`
	Body    string `json:"body,omitempty" jsonschema:"Request body payload (if applicable)."`
}

func (s *Server) colabProxyForward(ctx context.Context, req *mcp.CallToolRequest, input ProxyForwardInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	resp, err := s.proxy.ProxyForward(ctx, input.Port, input.Subpath, input.Method, nil, input.Body)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(resp, latency, "")
}
