package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCapturePayloadRedactsCredentialFields(t *testing.T) {
	p := capturePayload([]byte(`{"path":"a<b>.py","API-Key":"k","headers":{"Authorization":"Bearer x","accept":"*/*"},
		"items":[{"refresh_token":"r","n":12345678901234567890}],"token_count":3,"db_password":"p"}`))
	if p == nil || p.Truncated {
		t.Fatalf("unexpected payload: %+v", p)
	}
	var got map[string]any
	if err := json.Unmarshal(p.Body, &got); err != nil {
		t.Fatalf("body is not JSON: %v (%s)", err, p.Body)
	}
	if got["API-Key"] != redacted || got["db_password"] != redacted {
		t.Errorf("top-level secrets not redacted: %s", p.Body)
	}
	headers := got["headers"].(map[string]any)
	if headers["Authorization"] != redacted || headers["accept"] != "*/*" {
		t.Errorf("nested headers wrong: %s", p.Body)
	}
	item := got["items"].([]any)[0].(map[string]any)
	if item["refresh_token"] != redacted {
		t.Errorf("secret inside array not redacted: %s", p.Body)
	}
	if got["token_count"] != float64(3) || got["path"] != "a<b>.py" {
		t.Errorf("non-secret fields changed: %s", p.Body)
	}
	if !strings.Contains(string(p.Body), "12345678901234567890") || !strings.Contains(string(p.Body), "a<b>.py") {
		t.Errorf("numbers or HTML characters were rewritten: %s", p.Body)
	}
}

func TestCapturePayloadTruncatesLargeBodies(t *testing.T) {
	big := `{"content":"` + strings.Repeat("é", maxPayloadBytes) + `"}`
	p := capturePayload([]byte(big))
	if p == nil || !p.Truncated || p.Bytes != len(big) {
		t.Fatalf("expected truncation of %d bytes, got %+v", len(big), p)
	}
	var prefix string
	if err := json.Unmarshal(p.Body, &prefix); err != nil {
		t.Fatalf("truncated body must be a JSON string: %v", err)
	}
	if len(prefix) > maxPayloadBytes || !strings.HasPrefix(big, prefix) {
		t.Fatalf("prefix wrong: len=%d", len(prefix))
	}
}

func TestCapturePayloadNonJSONAndEmpty(t *testing.T) {
	if capturePayload(nil) != nil || capturePayload([]byte("  ")) != nil {
		t.Fatal("empty input must not produce a payload")
	}
	p := capturePayload([]byte("not json"))
	if p == nil || string(p.Body) != `"not json"` {
		t.Fatalf("non-JSON input must be stored as a JSON string, got %+v", p)
	}
}

// End to end over the MCP protocol: the middleware stores the request arguments
// and the envelope the agent received, and leaves unlogged calls alone.
func TestToolCallsAreRecordedWithPayloads(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	srv := NewServerWithSessionManager(session.NewManager(st, nil))

	ctx := context.Background()
	serverT, clientT := mcp.NewInMemoryTransports()
	ss, err := srv.mcp.Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "colab_list_all_sessions", Arguments: map[string]any{}}); err != nil {
		t.Fatalf("call list sessions: %v", err)
	}
	// No session exists, so this returns before logActivity and stays unrecorded.
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "colab_get_kernel_status", Arguments: map[string]any{"session_id": "nonexistent"}}); err != nil {
		t.Fatalf("call kernel status: %v", err)
	}

	list, err := st.ListRecentActivity(10)
	if err != nil || len(list) != 1 {
		t.Fatalf("want exactly 1 recorded call, got %d (err %v): %+v", len(list), err, list)
	}
	if list[0].ToolName != "colab_list_all_sessions" || list[0].Status != "success" || !list[0].HasPayload {
		t.Fatalf("unexpected row: %+v", list[0])
	}

	d, err := st.GetActivity(list[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if d.Request == nil || string(d.Request.Body) != `{}` {
		t.Fatalf("request not captured: %+v", d.Request)
	}
	var envelope struct {
		Status string `json:"status"`
		Data   struct {
			TotalSessions int `json:"total_sessions"`
		} `json:"data"`
	}
	if d.Response == nil || json.Unmarshal(d.Response.Body, &envelope) != nil || envelope.Status != "success" {
		t.Fatalf("response not captured: %+v", d.Response)
	}
}

// Outside the middleware (direct handler use) logActivity still writes a row.
func TestLogActivityWithoutMiddlewareWritesDirectly(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	srv := NewServerWithSessionManager(session.NewManager(st, nil))
	srv.logActivity(context.Background(), "s1", "colab_read_file", 7, nil)

	list, err := st.ListRecentActivity(10)
	if err != nil || len(list) != 1 || list[0].HasPayload || list[0].LatencyMs != 7 {
		t.Fatalf("unexpected rows: %+v err=%v", list, err)
	}
}
