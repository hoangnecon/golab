package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/hoangnecon/golab/internal/store"
	_ "modernc.org/sqlite"
)

func TestRecordAndGetActivityWithPayloads(t *testing.T) {
	ctx := context.Background()
	st, err := store.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer st.Close()

	id, err := st.RecordActivity(ctx, &store.ActivityRecord{
		SessionID: "sess-1",
		ToolName:  "colab_read_file",
		LatencyMs: 12,
		Status:    "success",
		Request:   &store.Payload{Body: json.RawMessage(`{"path":"content/a.py"}`), Bytes: 23},
		Response:  &store.Payload{Body: json.RawMessage(`"abc"`), Bytes: 9000, Truncated: true},
	})
	if err != nil || id <= 0 {
		t.Fatalf("record: id=%d err=%v", id, err)
	}
	if err := st.LogActivity(ctx, "sess-1", "colab_get_cells", 5, "error", "boom"); err != nil {
		t.Fatalf("log: %v", err)
	}

	list, err := st.ListRecentActivity(ctx, 10)
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %d rows, err=%v", len(list), err)
	}
	byTool := map[string]*store.ActivityLog{}
	for _, l := range list {
		byTool[l.ToolName] = l
	}
	if !byTool["colab_read_file"].HasPayload || byTool["colab_get_cells"].HasPayload {
		t.Fatalf("has_payload wrong: %+v / %+v", byTool["colab_read_file"], byTool["colab_get_cells"])
	}

	d, err := st.GetActivity(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if d.ToolName != "colab_read_file" || d.SessionID != "sess-1" || !d.HasPayload {
		t.Fatalf("unexpected detail: %+v", d)
	}
	if d.Request == nil || string(d.Request.Body) != `{"path":"content/a.py"}` || d.Request.Bytes != 23 || d.Request.Truncated {
		t.Fatalf("unexpected request: %+v", d.Request)
	}
	if d.Response == nil || string(d.Response.Body) != `"abc"` || d.Response.Bytes != 9000 || !d.Response.Truncated {
		t.Fatalf("unexpected response: %+v", d.Response)
	}

	plain, err := st.GetActivity(ctx, byTool["colab_get_cells"].ID)
	if err != nil || plain.Request != nil || plain.Response != nil || plain.ErrorMessage != "boom" {
		t.Fatalf("payload-free row: %+v err=%v", plain, err)
	}

	if _, err := st.GetActivity(ctx, 9999); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing id: want ErrNotFound, got %v", err)
	}
}

// A database created before payload capture must gain the columns on open and
// keep its existing rows readable.
func TestOpenMigratesLegacyActivityTable(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT,
		tool_name TEXT NOT NULL,
		latency_ms INTEGER,
		status TEXT NOT NULL,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	INSERT INTO activity_logs (session_id, tool_name, latency_ms, status) VALUES ('s', 'old_tool', 3, 'success');`)
	db.Close()
	if err != nil {
		t.Fatalf("seed legacy db: %v", err)
	}

	st, err := store.New(path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer st.Close()

	list, err := st.ListRecentActivity(ctx, 10)
	if err != nil || len(list) != 1 || list[0].ToolName != "old_tool" || list[0].HasPayload {
		t.Fatalf("legacy row: %+v err=%v", list, err)
	}
	if _, err := st.RecordActivity(ctx, &store.ActivityRecord{
		ToolName: "new_tool", Status: "success",
		Request: &store.Payload{Body: json.RawMessage(`{}`), Bytes: 2},
	}); err != nil {
		t.Fatalf("record after migration: %v", err)
	}

	// Opening again must be a no-op for the migration.
	st2, err := store.New(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	st2.Close()
}
