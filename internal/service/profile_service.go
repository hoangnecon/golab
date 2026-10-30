package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/process"
	"github.com/hoangnecon/golab/internal/store"
)

// ProfileService manages Chrome Google account profiles.
type ProfileService struct {
	repo    domain.ProfileRepository
	procMgr *process.Manager
}

func NewProfileService(repo domain.ProfileRepository, procMgr *process.Manager) *ProfileService {
	return &ProfileService{repo: repo, procMgr: procMgr}
}

func (s *ProfileService) CreateProfile(ctx context.Context, name, accountEmail string, tier domain.ProfileTier) (*domain.Profile, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("profile name cannot be empty")
	}
	if tier == "" {
		tier = domain.TierFree
	}

	profID := uuid.New().String()
	home, _ := os.UserHomeDir()
	userDataDir := filepath.Join(home, ".golab", "profiles", profID)

	port, err := s.repo.GetNextAvailableCDPPort(ctx, 9222)
	if err != nil {
		port = 9222
	}

	prof := &domain.Profile{
		ID:           profID,
		Name:         name,
		AccountEmail: strings.TrimSpace(accountEmail),
		Tier:         tier,
		UserDataDir:  userDataDir,
		CDPPort:      port,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateProfile(ctx, prof); err != nil {
		return nil, fmt.Errorf("create profile in store: %w", err)
	}
	return prof, nil
}

func toStoreProfile(p *domain.Profile) *store.Profile {
	if p == nil {
		return nil
	}
	return &store.Profile{
		ID:           p.ID,
		Name:         p.Name,
		AccountEmail: p.AccountEmail,
		Tier:         string(p.Tier),
		UserDataDir:  p.UserDataDir,
		CDPPort:      p.CDPPort,
		IsRunning:    p.IsRunning,
		CreatedAt:    p.CreatedAt,
	}
}

func (s *ProfileService) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	p, err := s.repo.GetProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.procMgr != nil {
		p.IsRunning = s.procMgr.IsProfileRunning(toStoreProfile(p))
	}
	return p, nil
}

func (s *ProfileService) ListProfiles(ctx context.Context) ([]*domain.Profile, error) {
	list, err := s.repo.ListProfiles(ctx)
	if err != nil {
		return nil, err
	}
	if s.procMgr != nil {
		for _, p := range list {
			p.IsRunning = s.procMgr.IsProfileRunning(toStoreProfile(p))
		}
	}
	return list, nil
}

func (s *ProfileService) LaunchProfile(ctx context.Context, id string) error {
	p, err := s.repo.GetProfile(ctx, id)
	if err != nil {
		return err
	}
	if s.procMgr == nil {
		return errors.New("process manager unavailable")
	}
	return s.procMgr.LaunchProfile(ctx, toStoreProfile(p), "")
}

func (s *ProfileService) StopProfile(ctx context.Context, id string) error {
	if s.procMgr == nil {
		return errors.New("process manager unavailable")
	}
	return s.procMgr.StopProfile(id)
}

func (s *ProfileService) DeleteProfile(ctx context.Context, id string) error {
	if s.procMgr != nil {
		_ = s.procMgr.StopProfile(id)
	}
	return s.repo.DeleteProfile(ctx, id)
}
