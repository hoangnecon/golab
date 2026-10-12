package mcpserver

import (
	"encoding/json"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResponseMetadata tracks execution metrics for agent context.
type ResponseMetadata struct {
	LatencyMs   int64  `json:"latency_ms"`
	KernelState string `json:"kernel_state,omitempty"`
	Timestamp   string `json:"timestamp"`
}

// ResponseEnvelope is the standardized payload returned by all GoLab tools.
type ResponseEnvelope struct {
	Status   string           `json:"status"` // "success" or "error"
	Data     any              `json:"data,omitempty"`
	Error    string           `json:"error,omitempty"`
	Metadata ResponseMetadata `json:"metadata"`
}

func successResult(data any, latencyMs int64, kernelState string) (*mcp.CallToolResult, any, error) {
	envelope := ResponseEnvelope{
		Status: "success",
		Data:   data,
		Metadata: ResponseMetadata{
			LatencyMs:   latencyMs,
			KernelState: kernelState,
			Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		},
	}

	dataBytes, _ := json.MarshalIndent(envelope, "", "  ")
	var structured map[string]any
	_ = json.Unmarshal(dataBytes, &structured)

	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(dataBytes)}},
		StructuredContent: structured,
	}, nil, nil
}

func errorResult(err error, latencyMs int64) (*mcp.CallToolResult, any, error) {
	errMsg := "unknown error"
	if err != nil {
		errMsg = err.Error()
	}

	envelope := ResponseEnvelope{
		Status: "error",
		Error:  errMsg,
		Metadata: ResponseMetadata{
			LatencyMs: latencyMs,
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		},
	}

	dataBytes, _ := json.MarshalIndent(envelope, "", "  ")
	var structured map[string]any
	_ = json.Unmarshal(dataBytes, &structured)

	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(dataBytes)}},
		StructuredContent: structured,
		IsError:           true,
	}, nil, nil
}
