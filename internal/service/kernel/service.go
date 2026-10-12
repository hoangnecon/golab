package kernel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hoangnecon/golab/internal/bridge"
)

// Service controls Kernel execution lifecycle via jupyter-server port 9000.
type Service struct {
	exec *bridge.ColabExecutor
}

// NewService creates a kernel service.
func NewService(exec *bridge.ColabExecutor) *Service {
	return &Service{exec: exec}
}

// ListKernels lists all active kernels.
func (s *Service) ListKernels(ctx context.Context) ([]KernelDetail, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/kernels", bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("list kernels: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("list kernels failed with status %d: %s", resp.Status, resp.RawText)
	}

	var kernels []KernelDetail
	if err := json.Unmarshal(resp.Body, &kernels); err != nil {
		return nil, fmt.Errorf("unmarshal kernels: %w", err)
	}
	return kernels, nil
}

// ListSessions returns all active notebook sessions.
func (s *Service) ListSessions(ctx context.Context) ([]SessionDetail, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/sessions", bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("list sessions failed with status %d: %s", resp.Status, resp.RawText)
	}

	var sessions []SessionDetail
	if err := json.Unmarshal(resp.Body, &sessions); err != nil {
		return nil, fmt.Errorf("unmarshal sessions: %w", err)
	}
	return sessions, nil
}

// GetStatusReport aggregates current session and all kernel states.
func (s *Service) GetStatusReport(ctx context.Context) (*KernelStatusReport, error) {
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	kernels, err := s.ListKernels(ctx)
	if err != nil {
		return nil, err
	}

	report := &KernelStatusReport{
		IsConnected:  len(sessions) > 0,
		AllKernels:   kernels,
		TotalKernels: len(kernels),
	}

	if len(sessions) > 0 {
		report.ActiveSession = &sessions[0]
		report.Kernel = sessions[0].Kernel
	} else if len(kernels) > 0 {
		report.Kernel = &kernels[0]
	}

	return report, nil
}

// InterruptKernel sends an immediate SIGINT to the active kernel or the specified kernel ID.
func (s *Service) InterruptKernel(ctx context.Context, kernelID string) error {
	// If kernelID is omitted, first try fast native DOM interrupt
	if kernelID == "" {
		if err := s.exec.InterruptKernel(ctx); err == nil {
			return nil
		}
		// Otherwise resolve active kernel ID
		rep, err := s.GetStatusReport(ctx)
		if err != nil {
			return err
		}
		if rep.Kernel == nil || rep.Kernel.ID == "" {
			return errors.New("no active kernel found to interrupt")
		}
		kernelID = rep.Kernel.ID
	}

	apiPath := fmt.Sprintf("/api/kernels/%s/interrupt", kernelID)
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "POST"})
	if err != nil {
		return fmt.Errorf("interrupt kernel %s: %w", kernelID, err)
	}
	if resp.Status != 204 && !resp.OK {
		return fmt.Errorf("interrupt failed with status %d: %s", resp.Status, resp.RawText)
	}
	return nil
}

// RestartKernel restarts the specified kernel or active kernel.
func (s *Service) RestartKernel(ctx context.Context, kernelID string) (*KernelDetail, error) {
	if kernelID == "" {
		rep, err := s.GetStatusReport(ctx)
		if err != nil {
			return nil, err
		}
		if rep.Kernel == nil || rep.Kernel.ID == "" {
			return nil, errors.New("no active kernel found to restart")
		}
		kernelID = rep.Kernel.ID
	}

	apiPath := fmt.Sprintf("/api/kernels/%s/restart", kernelID)
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "POST"})
	if err != nil {
		return nil, fmt.Errorf("restart kernel %s: %w", kernelID, err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("restart failed with status %d: %s", resp.Status, resp.RawText)
	}

	var detail KernelDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		return nil, fmt.Errorf("unmarshal restart response: %w", err)
	}
	return &detail, nil
}

// SpawnSecondaryKernel creates a new parallel Python 3 kernel.
func (s *Service) SpawnSecondaryKernel(ctx context.Context) (*KernelDetail, error) {
	bodyBytes, _ := json.Marshal(map[string]any{"name": "python3"})
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/kernels", bridge.FetchOptions{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    string(bodyBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("spawn secondary kernel: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("spawn kernel failed with status %d: %s", resp.Status, resp.RawText)
	}

	var detail KernelDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		return nil, fmt.Errorf("unmarshal spawn response: %w", err)
	}
	return &detail, nil
}

// KillKernel deletes and terminates a kernel instance.
func (s *Service) KillKernel(ctx context.Context, kernelID string) error {
	if kernelID == "" {
		return errors.New("kernel ID cannot be empty")
	}
	apiPath := fmt.Sprintf("/api/kernels/%s", kernelID)
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "DELETE"})
	if err != nil {
		return fmt.Errorf("kill kernel %s: %w", kernelID, err)
	}
	if resp.Status != 204 && !resp.OK {
		return fmt.Errorf("kill failed with status %d: %s", resp.Status, resp.RawText)
	}
	return nil
}
