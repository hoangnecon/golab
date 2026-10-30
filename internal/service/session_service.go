package service

import (
	"context"
	"errors"

	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store"
)

// SessionService coordinates notebook sessions, runtime lifecycle, and guardrails.
type SessionService struct {
	mgr *session.Manager
}

func NewSessionService(mgr *session.Manager) *SessionService {
	return &SessionService{mgr: mgr}
}

func (s *SessionService) ListSessions(ctx context.Context) ([]*store.Session, error) {
	return s.mgr.ListLiveSessions(ctx)
}

func (s *SessionService) OpenSession(ctx context.Context, profileID, notebookURL, name string) (*session.ActiveSession, error) {
	return s.mgr.OpenNotebook(ctx, profileID, notebookURL, name)
}

func (s *SessionService) CloseSession(ctx context.Context, sessionID string) error {
	return s.mgr.CloseSession(ctx, sessionID)
}

func (s *SessionService) SetActiveSession(ctx context.Context, sessionID string) error {
	return s.mgr.Store().SetActiveSession(ctx, sessionID)
}

func (s *SessionService) FocusSession(ctx context.Context, sessionID string) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if act.Session.TabID != "" {
		_ = act.CDP.FocusTab(ctx, act.Session.TabID)
	}
	return nil
}

func (s *SessionService) SetSessionLocked(ctx context.Context, sessionID string, locked bool) error {
	return s.mgr.Store().SetSessionLocked(ctx, sessionID, locked)
}

func (s *SessionService) ConnectRuntime(ctx context.Context, sessionID string) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return act.Services.Kernel.ConnectRuntime(ctx)
}

func (s *SessionService) DisconnectRuntime(ctx context.Context, sessionID string) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return act.Services.Kernel.DisconnectRuntime(ctx)
}

func (s *SessionService) InterruptKernel(ctx context.Context, sessionID string) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return act.Services.Kernel.InterruptKernel(ctx, "")
}

func (s *SessionService) RestartKernel(ctx context.Context, sessionID string) (any, error) {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return act.Services.Kernel.RestartKernel(ctx, "")
}

func (s *SessionService) SaveNotebook(ctx context.Context, sessionID string) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return act.Services.Notebook.SaveNotebook(ctx)
}

func (s *SessionService) RunAll(ctx context.Context, sessionID string) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return act.Services.Notebook.RunAll(ctx)
}

func (s *SessionService) GetRunningCells(ctx context.Context, sessionID string) (any, error) {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return act.Services.Notebook.GetRunningCells(ctx, false)
}

func (s *SessionService) GetAvailableAccelerators(ctx context.Context, sessionID string) (any, error) {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return act.Services.Kernel.GetAvailableAccelerators(ctx)
}

func (s *SessionService) SetAccelerator(ctx context.Context, sessionID, accelerator string, highRAM bool) error {
	act, err := s.mgr.ResolveSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if act.Services == nil || act.Services.Kernel == nil {
		return errors.New("kernel service unavailable")
	}
	return act.Services.Kernel.SetAccelerator(ctx, accelerator, highRAM)
}
