package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/store/sqlite"
)

func TestSQLiteRepositories(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test_golab.db")

	st, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize sqlite store: %v", err)
	}
	defer st.Close()

	// 1. Profile Repository
	prof := &domain.Profile{
		ID:           "test_prof",
		Name:         "Test Profile",
		AccountEmail: "test@example.com",
		Tier:         domain.TierPro,
		UserDataDir:  "/tmp/test_dir",
		CDPPort:      9222,
		CreatedAt:    time.Now().UTC(),
	}
	if err := st.CreateProfile(ctx, prof); err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	gotProf, err := st.GetProfile(ctx, "test_prof")
	if err != nil || gotProf.Name != "Test Profile" {
		t.Fatalf("GetProfile mismatch: %+v, err: %v", gotProf, err)
	}

	nextPort, err := st.GetNextAvailableCDPPort(ctx, 9222)
	if err != nil || nextPort != 9223 {
		t.Fatalf("expected nextPort=9223, got %d, err: %v", nextPort, err)
	}

	// 2. Project Repository
	proj := &domain.Project{
		ID:             "proj1",
		Name:           "My Colab Project",
		NotebookURL:    "https://colab.research.google.com/drive/xyz",
		OwnerProfileID: "test_prof",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := st.CreateProject(ctx, proj); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	gotProj, err := st.GetProject(ctx, "proj1")
	if err != nil || gotProj.OwnerProfileID != "test_prof" {
		t.Fatalf("GetProject mismatch: %+v, err: %v", gotProj, err)
	}

	// 3. Session Repository & Lock Guardrail
	sess := &domain.Session{
		ID:             "sess1",
		Name:           "Live Run",
		ProfileID:      "test_prof",
		ProjectID:      "proj1",
		OwnerProfileID: "test_prof",
		NotebookURL:    "https://colab.research.google.com/drive/xyz",
		IsActive:       true,
		IsLocked:       false,
		Permission:     domain.PermissionEdit,
		CreatedAt:      time.Now().UTC(),
	}
	if err := st.SaveSession(ctx, sess); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	if err := st.SetSessionLocked(ctx, "sess1", true); err != nil {
		t.Fatalf("SetSessionLocked failed: %v", err)
	}

	locked, err := st.GetLockedSession(ctx)
	if err != nil || locked.ID != "sess1" {
		t.Fatalf("expected locked session sess1, got %+v, err: %v", locked, err)
	}

	// 4. Activity Repository
	rec := &domain.ActivityRecord{
		SessionID:    "sess1",
		ToolName:     "colab_read_file",
		LatencyMs:    42,
		Status:       "success",
		ErrorMessage: "",
		Request:      &domain.Payload{Body: []byte(`{"path":"test.txt"}`), Bytes: 19},
		Response:     &domain.Payload{Body: []byte(`{"content":"hello"}`), Bytes: 19},
	}
	rowID, err := st.RecordActivity(ctx, rec)
	if err != nil || rowID <= 0 {
		t.Fatalf("RecordActivity failed: %v, rowID: %d", err, rowID)
	}

	detail, err := st.GetActivity(ctx, rowID)
	if err != nil || detail.ToolName != "colab_read_file" || !detail.HasPayload {
		t.Fatalf("GetActivity mismatch: %+v, err: %v", detail, err)
	}
}
