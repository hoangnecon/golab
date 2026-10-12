package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hoangnecon/golab/internal/bridge"
)

// Service manages background Linux PTY terminals.
type Service struct {
	exec *bridge.ColabExecutor
}

// NewService creates a terminal service.
func NewService(exec *bridge.ColabExecutor) *Service {
	return &Service{exec: exec}
}

// ListTerminals returns all active terminal sessions.
func (s *Service) ListTerminals(ctx context.Context) ([]TerminalInfo, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/terminals", bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("list terminals: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("list terminals failed with status %d: %s", resp.Status, resp.RawText)
	}

	var list []TerminalInfo
	if err := json.Unmarshal(resp.Body, &list); err != nil {
		return nil, fmt.Errorf("unmarshal terminals: %w", err)
	}
	return list, nil
}

// CreateTerminal spawns a new bash PTY terminal on the Colab container.
func (s *Service) CreateTerminal(ctx context.Context) (*TerminalInfo, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/terminals", bridge.FetchOptions{Method: "POST"})
	if err != nil {
		return nil, fmt.Errorf("create terminal: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("create terminal failed with status %d: %s", resp.Status, resp.RawText)
	}

	var info TerminalInfo
	if err := json.Unmarshal(resp.Body, &info); err != nil {
		return nil, fmt.Errorf("unmarshal create terminal: %w", err)
	}
	return &info, nil
}

// DeleteTerminal closes a terminal session.
func (s *Service) DeleteTerminal(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("terminal name cannot be empty")
	}
	apiPath := fmt.Sprintf("/api/terminals/%s", name)
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "DELETE"})
	if err != nil {
		return fmt.Errorf("delete terminal %s: %w", name, err)
	}
	if resp.Status != 204 && !resp.OK {
		return fmt.Errorf("delete terminal failed with status %d: %s", resp.Status, resp.RawText)
	}
	return nil
}
