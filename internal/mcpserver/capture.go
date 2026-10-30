package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/hoangnecon/golab/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxPayloadBytes caps each stored request/response body. Larger documents
// (file contents, base64 images, long cell outputs) are kept as a prefix.
const maxPayloadBytes = 64 << 10

const redacted = "[redacted]"

// pendingCall collects what a tool handler reports through logActivity, so the
// capture middleware can persist it together with the request and response.
type pendingCall struct {
	logged    bool
	sessionID string
	latencyMs int64
	err       error
}

type pendingCallKey struct{}

// captureToolCalls is receiving middleware that records the arguments and the
// result of every tools/call whose handler called logActivity. Handlers that
// return before logging (e.g. no session could be resolved) stay unrecorded,
// as before payload capture existed.
func (s *Server) captureToolCalls(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		call, ok := req.(*mcp.CallToolRequest)
		if method != "tools/call" || !ok || call.Params == nil || s.activityStore() == nil {
			return next(ctx, method, req)
		}
		pc := &pendingCall{}
		res, err := next(context.WithValue(ctx, pendingCallKey{}, pc), method, req)
		if pc.logged {
			s.recordCall(ctx, call.Params, pc, res, err)
		}
		return res, err
	}
}

func (s *Server) activityStore() *store.Store {
	if s.sessionMgr == nil {
		return nil
	}
	return s.sessionMgr.Store()
}

func (s *Server) recordCall(ctx context.Context, params *mcp.CallToolParamsRaw, pc *pendingCall, res mcp.Result, callErr error) {
	st := s.activityStore()
	status, errMsg := "success", ""
	if pc.err != nil {
		status, errMsg = "error", pc.err.Error()
	}
	_, _ = st.RecordActivity(ctx, &store.ActivityRecord{
		SessionID:    pc.sessionID,
		ToolName:     params.Name,
		LatencyMs:    pc.latencyMs,
		Status:       status,
		ErrorMessage: errMsg,
		Request:      capturePayload(params.Arguments),
		Response:     capturePayload(resultJSON(res, callErr)),
	})
}

// resultJSON is what the agent received: the structured envelope when present,
// otherwise the text content, otherwise the protocol error.
func resultJSON(res mcp.Result, callErr error) []byte {
	if callErr != nil {
		b, _ := json.Marshal(map[string]string{"error": callErr.Error()})
		return b
	}
	r, ok := res.(*mcp.CallToolResult)
	if !ok || r == nil {
		return nil
	}
	if r.StructuredContent != nil {
		b, _ := json.Marshal(r.StructuredContent)
		return b
	}
	var texts []string
	for _, c := range r.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			texts = append(texts, t.Text)
		}
	}
	if len(texts) == 0 {
		return nil
	}
	joined := strings.Join(texts, "\n")
	if json.Valid([]byte(joined)) {
		return []byte(joined)
	}
	b, _ := json.Marshal(joined)
	return b
}

// capturePayload redacts credential-like fields and caps the size. Input that
// is not JSON is stored as a JSON string.
func capturePayload(raw []byte) *store.Payload {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	body := redactJSON(raw)
	if len(body) <= maxPayloadBytes {
		return &store.Payload{Body: body, Bytes: len(body)}
	}
	cut := maxPayloadBytes
	for cut > 0 && !utf8.RuneStart(body[cut]) {
		cut--
	}
	prefix, _ := marshalNoEscape(string(body[:cut]))
	return &store.Payload{Body: prefix, Bytes: len(body), Truncated: true}
}

func redactJSON(raw []byte) []byte {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil || dec.More() {
		b, _ := marshalNoEscape(string(raw))
		return b
	}
	b, err := marshalNoEscape(redactValue(v))
	if err != nil {
		b, _ = marshalNoEscape(string(raw))
	}
	return b
}

func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if isSensitiveKey(k) {
				t[k] = redacted
			} else {
				t[k] = redactValue(val)
			}
		}
	case []any:
		for i, val := range t {
			t[i] = redactValue(val)
		}
	}
	return v
}

// isSensitiveKey matches field names that conventionally hold credentials.
// Values inside free text (e.g. a secret typed into a code cell) cannot be
// detected this way and are stored as sent.
func isSensitiveKey(key string) bool {
	k := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	switch k {
	case "authorization", "cookie", "set_cookie", "api_key", "apikey", "private_key", "passwd", "pwd":
		return true
	}
	return strings.Contains(k, "password") || strings.Contains(k, "secret") ||
		k == "token" || strings.HasSuffix(k, "_token")
}

// marshalNoEscape keeps <, > and & readable in stored code and HTML.
func marshalNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
