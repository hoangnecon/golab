package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReadFileInput specifies arguments for reading a file.
type ReadFileInput struct {
	Path   string `json:"path" jsonschema:"Path to file on Colab (e.g. content/dataset.csv or content/drive/MyDrive/model.py). Required."`
	Format string `json:"format,omitempty" jsonschema:"Data encoding format: 'text' for code/plain text, or 'base64' for images/binary. Default is 'text'."`
}

func (s *Server) colabReadFile(ctx context.Context, req *mcp.CallToolRequest, input ReadFileInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	res, err := s.storage.ReadFile(ctx, input.Path, input.Format)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(res, latency, "")
}

// WriteFileInput specifies arguments for writing/creating a file.
type WriteFileInput struct {
	Path    string `json:"path" jsonschema:"Path to file on Colab (e.g. content/train.py or content/drive/MyDrive/config.yaml). Required."`
	Content string `json:"content" jsonschema:"Content to write into the file. Required."`
	Format  string `json:"format,omitempty" jsonschema:"Content encoding: 'text' or 'base64'. Default is 'text'."`
}

func (s *Server) colabWriteFile(ctx context.Context, req *mcp.CallToolRequest, input WriteFileInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	res, err := s.storage.WriteFile(ctx, input.Path, input.Content, input.Format)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(res, latency, "")
}

// ListDirectoryInput specifies arguments for directory inspection.
type ListDirectoryInput struct {
	Path string `json:"path,omitempty" jsonschema:"Directory path to list (e.g. 'content', 'content/drive/MyDrive', or empty for root)."`
}

func (s *Server) colabListDirectory(ctx context.Context, req *mcp.CallToolRequest, input ListDirectoryInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	res, err := s.storage.ListDirectory(ctx, input.Path)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(res, latency, "")
}

// CreateDirectoryInput specifies directory to create.
type CreateDirectoryInput struct {
	Path string `json:"path" jsonschema:"Directory path to create (e.g. 'content/models' or 'content/logs'). Required."`
}

func (s *Server) colabCreateDirectory(ctx context.Context, req *mcp.CallToolRequest, input CreateDirectoryInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.storage.CreateDirectory(ctx, input.Path)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"created": true, "path": input.Path}, latency, "")
}

// DeleteFileInput specifies file/directory to permanently delete.
type DeleteFileInput struct {
	Path string `json:"path" jsonschema:"Path to file or directory to permanently delete. Required."`
}

func (s *Server) colabDeleteFile(ctx context.Context, req *mcp.CallToolRequest, input DeleteFileInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.storage.DeleteFile(ctx, input.Path)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"deleted": true, "path": input.Path}, latency, "")
}

// MoveFileInput specifies paths for renaming or moving files.
type MoveFileInput struct {
	OldPath string `json:"old_path" jsonschema:"Current path of the file or directory. Required."`
	NewPath string `json:"new_path" jsonschema:"Target destination path. Required."`
}

func (s *Server) colabMoveFile(ctx context.Context, req *mcp.CallToolRequest, input MoveFileInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.storage.MoveFile(ctx, input.OldPath, input.NewPath)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"moved": true, "old_path": input.OldPath, "new_path": input.NewPath}, latency, "")
}
