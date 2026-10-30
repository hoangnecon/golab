package domain_test

import (
	"testing"
	"time"

	"github.com/hoangnecon/golab/internal/domain"
)

func TestProjectEntity(t *testing.T) {
	now := time.Now().UTC()
	p := domain.Project{
		ID:             "p1",
		Name:           "Test Project",
		NotebookURL:    "https://colab.research.google.com/drive/xyz",
		OwnerProfileID: "default",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if p.ID != "p1" || p.OwnerProfileID != "default" {
		t.Fatalf("unexpected project: %+v", p)
	}
}

func TestSessionEntity(t *testing.T) {
	now := time.Now().UTC()
	s := domain.Session{
		ID:             "s1",
		Name:           "Test Session",
		ProfileID:      "default",
		ProjectID:      "p1",
		OwnerProfileID: "default",
		NotebookURL:    "https://colab.research.google.com/drive/xyz",
		IsActive:       true,
		IsLocked:       true,
		Permission:     domain.PermissionEdit,
		CreatedAt:      now,
	}
	if !s.IsLocked || s.Permission != domain.PermissionEdit {
		t.Fatalf("unexpected session: %+v", s)
	}
}

func TestProfileEntity(t *testing.T) {
	now := time.Now().UTC()
	p := domain.Profile{
		ID:          "prof1",
		Name:        "Test Profile",
		Tier:        domain.TierPro,
		UserDataDir: "/tmp/test",
		CDPPort:     9222,
		IsRunning:   true,
		CreatedAt:   now,
	}
	if p.Tier != domain.TierPro || p.CDPPort != 9222 {
		t.Fatalf("unexpected profile: %+v", p)
	}
}
