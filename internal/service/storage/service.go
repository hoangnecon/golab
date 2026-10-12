package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hoangnecon/golab/internal/bridge"
)

// Service provides high-speed, non-blocking file and directory operations via colab-fileshim.
type Service struct {
	exec *bridge.ColabExecutor
}

// NewService creates a storage service.
func NewService(exec *bridge.ColabExecutor) *Service {
	return &Service{exec: exec}
}

func cleanColabPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}
	segments := strings.Split(p, "/")
	escaped := make([]string, len(segments))
	for i, s := range segments {
		escaped[i] = url.PathEscape(s)
	}
	return strings.Join(escaped, "/")
}

func isDrivePath(cleaned string) bool {
	return strings.HasPrefix(cleaned, "content/drive") || strings.HasPrefix(cleaned, "drive")
}

func unmountedDriveError(path string) error {
	return fmt.Errorf("google drive is not mounted: path %q is inaccessible. Please mount Google Drive first by executing: from google.colab import drive; drive.mount('/content/drive')", path)
}

// ReadFile reads the full contents of a file on the Colab container. Format can be "text" or "base64".
func (s *Service) ReadFile(ctx context.Context, filePath string, format string) (*FileReadResult, error) {
	cleaned := cleanColabPath(filePath)
	if format != "base64" {
		format = "text"
	}

	apiPath := fmt.Sprintf("/api/contents/%s?content=1&format=%s", cleaned, format)
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", filePath, err)
	}
	if !resp.OK {
		if resp.Status == 404 && isDrivePath(cleaned) {
			return nil, unmountedDriveError(filePath)
		}
		return nil, fmt.Errorf("read file failed with status %d: %s", resp.Status, resp.RawText)
	}

	var rawEntry struct {
		Name         string  `json:"name"`
		Path         string  `json:"path"`
		Type         string  `json:"type"`
		Size         int64   `json:"size"`
		Format       string  `json:"format"`
		Content      string  `json:"content"`
		LastModified *string `json:"last_modified"`
	}
	if err := json.Unmarshal(resp.Body, &rawEntry); err != nil {
		return nil, fmt.Errorf("unmarshal file read response: %w", err)
	}

	if rawEntry.Type == "directory" {
		return nil, fmt.Errorf("path %q is a directory, not a file", filePath)
	}

	return &FileReadResult{
		Path:    rawEntry.Path,
		Name:    rawEntry.Name,
		Type:    rawEntry.Type,
		Size:    rawEntry.Size,
		Format:  rawEntry.Format,
		Content: rawEntry.Content,
	}, nil
}

// WriteFile creates or overwrites a file atomically in the Colab container.
func (s *Service) WriteFile(ctx context.Context, filePath string, content string, format string) (*FileWriteResult, error) {
	cleaned := cleanColabPath(filePath)
	if cleaned == "" {
		return nil, errors.New("file path cannot be empty")
	}
	if format != "base64" {
		format = "text"
	}

	bodyObj := map[string]any{
		"name":    cleaned,
		"path":    cleaned,
		"type":    "file",
		"format":  format,
		"content": content,
	}
	bodyBytes, err := json.Marshal(bodyObj)
	if err != nil {
		return nil, err
	}

	apiPath := "/api/contents/" + cleaned
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{
		Method:  "PUT",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    string(bodyBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("write file %q: %w", filePath, err)
	}
	if !resp.OK {
		if resp.Status == 404 && isDrivePath(cleaned) {
			return nil, unmountedDriveError(filePath)
		}
		return nil, fmt.Errorf("write file failed with status %d: %s", resp.Status, resp.RawText)
	}

	var rawEntry struct {
		Name   string `json:"name"`
		Path   string `json:"path"`
		Type   string `json:"type"`
		Size   int64  `json:"size"`
	}
	if err := json.Unmarshal(resp.Body, &rawEntry); err != nil {
		return nil, fmt.Errorf("unmarshal write response: %w", err)
	}

	return &FileWriteResult{
		Path:    rawEntry.Path,
		Name:    rawEntry.Name,
		Size:    rawEntry.Size,
		Type:    rawEntry.Type,
		Created: true,
	}, nil
}

// ListDirectory lists all files and subdirectories inside a directory.
func (s *Service) ListDirectory(ctx context.Context, dirPath string) (*DirectoryListing, error) {
	cleaned := cleanColabPath(dirPath)
	apiPath := "/api/contents/" + cleaned
	if cleaned != "" {
		apiPath += "?content=1"
	} else {
		apiPath = "/api/contents?content=1"
	}

	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, fmt.Errorf("list directory %q: %w", dirPath, err)
	}
	if !resp.OK {
		if resp.Status == 404 && isDrivePath(cleaned) {
			return nil, unmountedDriveError(dirPath)
		}
		return nil, fmt.Errorf("list directory failed with status %d: %s", resp.Status, resp.RawText)
	}

	var rawDir struct {
		Name     string      `json:"name"`
		Path     string      `json:"path"`
		Type     string      `json:"type"`
		Writable bool        `json:"writable"`
		Content  []FileEntry `json:"content"`
	}
	if err := json.Unmarshal(resp.Body, &rawDir); err != nil {
		return nil, fmt.Errorf("unmarshal list directory response: %w", err)
	}

	return &DirectoryListing{
		Path:         rawDir.Path,
		Name:         rawDir.Name,
		Type:         rawDir.Type,
		Writable:     rawDir.Writable,
		Entries:      rawDir.Content,
		TotalEntries: len(rawDir.Content),
	}, nil
}

// CreateDirectory creates a new directory.
func (s *Service) CreateDirectory(ctx context.Context, dirPath string) error {
	cleaned := cleanColabPath(dirPath)
	if cleaned == "" {
		return errors.New("directory path cannot be empty")
	}

	bodyBytes, _ := json.Marshal(map[string]any{"type": "directory"})
	apiPath := "/api/contents/" + cleaned
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{
		Method:  "PUT",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    string(bodyBytes),
	})
	if err != nil {
		return fmt.Errorf("create directory %q: %w", dirPath, err)
	}
	if !resp.OK {
		return fmt.Errorf("create directory failed with status %d: %s", resp.Status, resp.RawText)
	}
	return nil
}

// DeleteFile permanently removes a file or directory.
func (s *Service) DeleteFile(ctx context.Context, path string) error {
	cleaned := cleanColabPath(path)
	if cleaned == "" {
		return errors.New("path cannot be empty")
	}

	apiPath := "/api/contents/" + cleaned
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{Method: "DELETE"})
	if err != nil {
		return fmt.Errorf("delete %q: %w", path, err)
	}
	if resp.Status != 204 && !resp.OK {
		return fmt.Errorf("delete failed with status %d: %s", resp.Status, resp.RawText)
	}
	return nil
}

// MoveFile renames or moves a file/directory to a new path.
func (s *Service) MoveFile(ctx context.Context, oldPath, newPath string) error {
	cleanedOld := cleanColabPath(oldPath)
	cleanedNew := cleanColabPath(newPath)
	if cleanedOld == "" || cleanedNew == "" {
		return errors.New("oldPath and newPath cannot be empty")
	}

	bodyBytes, _ := json.Marshal(map[string]any{"path": cleanedNew})
	apiPath := "/api/contents/" + cleanedOld
	resp, err := s.exec.FetchKernelXHR(ctx, apiPath, bridge.FetchOptions{
		Method:  "PATCH",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    string(bodyBytes),
	})
	if err != nil {
		return fmt.Errorf("move %q -> %q: %w", oldPath, newPath, err)
	}
	if !resp.OK {
		return fmt.Errorf("move failed with status %d: %s", resp.Status, resp.RawText)
	}
	return nil
}
