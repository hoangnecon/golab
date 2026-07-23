package mcpserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJSONResultSurvivesMCPTransport(t *testing.T) {
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "v1"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "json"}, func(
		context.Context,
		*mcp.CallToolRequest,
		struct{},
	) (*mcp.CallToolResult, any, error) {
		result, err := jsonResult(map[string]any{"ok": true})
		return result, nil, err
	})

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "json",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("StructuredContent type = %T, want map[string]any", result.StructuredContent)
	}
	if structured["ok"] != true {
		t.Fatalf("StructuredContent = %#v, want ok=true", structured)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !json.Valid([]byte(text.Text)) {
		t.Fatalf("text fallback = %#v, want valid JSON", result.Content)
	}
}

func TestJSONResultPopulatesStructuredContentAndJSONFallback(t *testing.T) {
	result, err := jsonResult(map[string]any{
		"ok":    true,
		"count": 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("StructuredContent type = %T, want map[string]any", result.StructuredContent)
	}
	if structured["ok"] != true {
		t.Fatalf("StructuredContent = %#v", structured)
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Content type = %T, want *mcp.TextContent", result.Content[0])
	}
	if !json.Valid([]byte(text.Text)) {
		t.Fatalf("text fallback is not JSON: %q", text.Text)
	}
}

func TestRawJSONResultUnwrapsBrowserMCPEnvelope(t *testing.T) {
	raw := json.RawMessage(`{
		"content": [{"type": "text", "text": "{\"cells\":[{\"id\":\"cell-1\"}]}"}]
	}`)

	result, err := rawJSONResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	structured := result.StructuredContent.(map[string]any)
	if _, ok := structured["cells"]; !ok {
		t.Fatalf("StructuredContent = %#v, want cells", structured)
	}
}

func TestRawJSONResultWrapsPlainTextAsJSON(t *testing.T) {
	result, err := rawJSONResult(json.RawMessage("plain text"))
	if err != nil {
		t.Fatal(err)
	}
	structured := result.StructuredContent.(map[string]any)
	if structured["value"] != "plain text" {
		t.Fatalf("StructuredContent = %#v", structured)
	}
}

func TestJSONResultWrapsArraysInObject(t *testing.T) {
	result, err := jsonResult([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	structured := result.StructuredContent.(map[string]any)
	if _, ok := structured["result"]; !ok {
		t.Fatalf("StructuredContent = %#v, want result wrapper", structured)
	}
}

func TestErrorResultIsStructuredJSON(t *testing.T) {
	result, err := errResult("boom")
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("IsError = false, want true")
	}
	structured := result.StructuredContent.(map[string]any)
	if structured["error"] != "boom" {
		t.Fatalf("StructuredContent = %#v", structured)
	}
}
