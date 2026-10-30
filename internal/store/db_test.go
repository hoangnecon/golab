package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoangnecon/golab/internal/store"
)

func TestStoreOperations(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	// 1. Profile operations
	p := &store.Profile{
		ID:          "prof-1",
		Name:        "Test Profile",
		Tier:        "pro",
		UserDataDir: "/tmp/data",
		CDPPort:     9222,
		CreatedAt:   time.Now().UTC(),
	}
	if err := st.CreateProfile(ctx, p); err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	gotP, err := st.GetProfile(ctx, "prof-1")
	if err != nil || gotP.Name != "Test Profile" {
		t.Fatalf("unexpected profile: %+v, err: %v", gotP, err)
	}

	nextPort, err := st.GetNextAvailableCDPPort(ctx, 9222)
	if err != nil || nextPort != 9223 {
		t.Fatalf("expected next port 9223, got %d (err: %v)", nextPort, err)
	}

	// 2. Session operations
	s := &store.Session{
		ID:          "sess-1",
		Name:        "Notebook A",
		ProfileID:   "prof-1",
		NotebookURL: "https://colab.research.google.com/drive/xyz",
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
	}
	if err := st.SaveSession(ctx, s); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	active, err := st.GetActiveSession(ctx)
	if err != nil || active.ID != "sess-1" {
		t.Fatalf("unexpected active session: %+v, err: %v", active, err)
	}

	// 3. Activity logging
	if err := st.LogActivity(ctx, "sess-1", "colab_read_file", 25, "success", ""); err != nil {
		t.Fatalf("failed to log activity: %v", err)
	}

	logs, err := st.ListRecentActivity(ctx, 10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("expected at least 1 log, got %d (err: %v)", len(logs), err)
	}
	if logs[0].ToolName != "colab_read_file" || logs[0].LatencyMs != 25 {
		t.Fatalf("unexpected log entry: %+v", logs[0])
	}
}
