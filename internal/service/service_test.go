package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/service"
	"github.com/hoangnecon/golab/internal/store/sqlite"
)

func TestApplicationServices(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "services_test.db")

	st, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("init sqlite store: %v", err)
	}
	defer st.Close()

	// 1. Profile Service
	profileSvc := service.NewProfileService(st, nil)
	prof, err := profileSvc.CreateProfile(ctx, "Test Prof", "test@example.com", domain.TierPro)
	if err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}
	if prof.Name != "Test Prof" || prof.CDPPort != 9222 {
		t.Fatalf("unexpected profile: %+v", prof)
	}

	// 2. Project Service
	projectSvc := service.NewProjectService(st)
	proj, err := projectSvc.CreateProject(ctx, "Test Project", "https://colab.research.google.com/drive/p1", prof.ID)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if proj.OwnerProfileID != prof.ID {
		t.Fatalf("unexpected project owner: %s", proj.OwnerProfileID)
	}

	// Test invalid project url validation
	_, errInvalid := projectSvc.CreateProject(ctx, "Bad", "not-a-url", prof.ID)
	if errInvalid == nil {
		t.Fatalf("expected error on invalid project url, got nil")
	}

	// 3. Activity Service
	activitySvc := service.NewActivityService(st)
	err = activitySvc.LogActivity(ctx, "sess1", "colab_read_file", 15, "success", "")
	if err != nil {
		t.Fatalf("LogActivity failed: %v", err)
	}
	logs, err := activitySvc.ListRecentActivity(ctx, 10)
	if err != nil || len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d, err: %v", len(logs), err)
	}
}
