package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store"
)

func TestGetActivityEndpoint(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	id, err := st.RecordActivity(ctx, &store.ActivityRecord{
		ToolName: "colab_get_cells", Status: "success", LatencyMs: 4,
		Request:  &store.Payload{Body: json.RawMessage(`{"start_index":0}`), Bytes: 17},
		Response: &store.Payload{Body: json.RawMessage(`{"status":"success"}`), Bytes: 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	h := NewServer(session.NewManager(st, nil), nil).Handler()

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "127.0.0.1:9120"
		h.ServeHTTP(rec, req)
		return rec
	}

	rec := get("/api/activity/" + jsonInt(id))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var body struct {
		ID         int64 `json:"id"`
		HasPayload bool  `json:"has_payload"`
		Request    struct {
			Body  map[string]any `json:"body"`
			Bytes int            `json:"bytes"`
		} `json:"request"`
		Response struct {
			Body map[string]any `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (%s)", err, rec.Body)
	}
	if body.ID != id || !body.HasPayload || body.Request.Body["start_index"] != float64(0) || body.Request.Bytes != 17 ||
		body.Response.Body["status"] != "success" {
		t.Fatalf("unexpected body: %s", rec.Body)
	}

	if rec := get("/api/activity/999"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing id: want 404, got %d", rec.Code)
	}
	if rec := get("/api/activity/abc"); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id: want 400, got %d", rec.Code)
	}

	// The list stays payload-free; only the flag is exposed.
	rec = get("/api/activity")
	var list struct {
		Activity []map[string]any `json:"activity"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list.Activity) != 1 {
		t.Fatalf("list: %s", rec.Body)
	}
	if _, has := list.Activity[0]["request"]; has || list.Activity[0]["has_payload"] != true {
		t.Fatalf("list must carry has_payload but no payloads: %v", list.Activity[0])
	}
}

func jsonInt(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}
