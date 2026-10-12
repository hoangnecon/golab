package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/hoangnecon/golab/internal/bridge"
)

// Service provides real-time hardware telemetry and environment diagnostics.
type Service struct {
	exec *bridge.ColabExecutor
}

// NewService creates a telemetry service.
func NewService(exec *bridge.ColabExecutor) *Service {
	return &Service{exec: exec}
}

func bytesToGB(b int64) float64 {
	return math.Round((float64(b)/(1024*1024*1024))*100) / 100
}

func bytesToMB(b int64) float64 {
	return math.Round((float64(b)/(1024*1024))*10) / 10
}

// GetHardwareTelemetry fetches instant RAM, Disk and Tesla GPU utilization via /api/colab/resources.
func (s *Service) GetHardwareTelemetry(ctx context.Context) (*HardwareTelemetry, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/colab/resources", bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("fetch colab resources: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("resources request failed with status %d: %s", resp.Status, resp.RawText)
	}

	var raw struct {
		Memory struct {
			TotalBytes int64 `json:"totalBytes"`
			FreeBytes  int64 `json:"freeBytes"`
		} `json:"memory"`
		Disks []struct {
			Filesystem struct {
				Label      string `json:"label"`
				TotalBytes int64  `json:"totalBytes"`
				UsedBytes  int64  `json:"usedBytes"`
			} `json:"filesystem"`
		} `json:"disks"`
		GPUs []struct {
			Name              string  `json:"name"`
			MemoryUsedBytes   int64   `json:"memoryUsedBytes"`
			MemoryTotalBytes  int64   `json:"memoryTotalBytes"`
			GPUUtilization    float64 `json:"gpuUtilization"`
			MemoryUtilization float64 `json:"memoryUtilization"`
			EverUsed          bool    `json:"everUsed"`
		} `json:"gpus"`
	}

	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal resources response: %w", err)
	}

	usedRAM := raw.Memory.TotalBytes - raw.Memory.FreeBytes
	res := &HardwareTelemetry{
		RAM: RAMTelemetry{
			TotalBytes: raw.Memory.TotalBytes,
			FreeBytes:  raw.Memory.FreeBytes,
			UsedBytes:  usedRAM,
			TotalGB:    bytesToGB(raw.Memory.TotalBytes),
			FreeGB:     bytesToGB(raw.Memory.FreeBytes),
			UsedGB:     bytesToGB(usedRAM),
		},
		Disks: make([]DiskTelemetry, 0, len(raw.Disks)),
		GPUs:  make([]GPUTelemetry, 0, len(raw.GPUs)),
	}

	for _, d := range raw.Disks {
		fs := d.Filesystem
		var pct float64
		if fs.TotalBytes > 0 {
			pct = math.Round((float64(fs.UsedBytes)/float64(fs.TotalBytes))*1000) / 10
		}
		res.Disks = append(res.Disks, DiskTelemetry{
			Label:       fs.Label,
			TotalBytes:  fs.TotalBytes,
			UsedBytes:   fs.UsedBytes,
			TotalGB:     bytesToGB(fs.TotalBytes),
			UsedGB:      bytesToGB(fs.UsedBytes),
			UsedPercent: pct,
		})
	}

	for _, g := range raw.GPUs {
		usedMB := bytesToMB(g.MemoryUsedBytes)
		totMB := bytesToMB(g.MemoryTotalBytes)
		res.GPUs = append(res.GPUs, GPUTelemetry{
			Name:              g.Name,
			MemoryUsedBytes:   g.MemoryUsedBytes,
			MemoryTotalBytes:  g.MemoryTotalBytes,
			MemoryUsedMB:      usedMB,
			MemoryTotalMB:     totMB,
			MemoryFreeMB:      totMB - usedMB,
			GPUUtilization:    g.GPUUtilization,
			MemoryUtilization: g.MemoryUtilization,
			EverUsed:          g.EverUsed,
		})
	}

	return res, nil
}

// GetBuildInfo fetches container build and release information.
func (s *Service) GetBuildInfo(ctx context.Context) (*BuildInfo, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/colab/build-info", bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("fetch build info: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("build info failed with status %d: %s", resp.Status, resp.RawText)
	}

	var raw struct {
		ReleaseTag string `json:"release_tag"`
	}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal build info: %w", err)
	}
	return &BuildInfo{ReleaseTag: raw.ReleaseTag}, nil
}

// CheckDriveStatus inspects Google Drive health and FUSE timeouts.
func (s *Service) CheckDriveStatus(ctx context.Context) (*DriveStatus, error) {
	resp, err := s.exec.FetchKernelXHR(ctx, "/api/colab/drive", bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("fetch drive status: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("drive status failed with status %d: %s", resp.Status, resp.RawText)
	}

	var raw struct {
		DFS []any `json:"dfs"`
	}
	_ = json.Unmarshal(resp.Body, &raw)

	mounted := len(raw.DFS) > 0
	if !mounted {
		return &DriveStatus{
			Mounted:        false,
			Status:         "disconnected",
			Message:        "Google Drive is not mounted in this Colab session.",
			ActionRequired: "Run 'from google.colab import drive; drive.mount(\"/content/drive\")' in a code cell to mount Google Drive.",
			DFS:            raw.DFS,
		}, nil
	}

	return &DriveStatus{
		Mounted:   true,
		Status:    "connected",
		MountPath: "/content/drive",
		Message:   "Google Drive is mounted and healthy.",
		DFS:       raw.DFS,
	}, nil
}
