package mcpserver

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ── Group C — Project Assessment ─────────────────────

type ListDriveInput struct {
	Path     string `json:"path" jsonschema:"Path relative to MyDrive. Empty for root."`
	MaxDepth int    `json:"maxDepth" jsonschema:"Max directory depth to list"`
}

func (s *Server) listDrive(ctx context.Context, req *mcp.CallToolRequest, input ListDriveInput) (*mcp.CallToolResult, Empty, error) {
	path := input.Path
	if path == "" {
		path = "."
	}
	depth := input.MaxDepth
	if depth <= 0 {
		depth = 1
	}
	// Encode path as base64 to prevent injection
	pathB64 := base64.StdEncoding.EncodeToString([]byte(path))
	code := fmt.Sprintf(`import os, json, base64
path = base64.b64decode('%s').decode()
base = '/content/drive/MyDrive/' + path
if not os.path.exists('/content/drive/MyDrive'):
    print(json.dumps({"error": "Drive not mounted"}))
elif not os.path.exists(base):
    print(json.dumps({"error": f"Path not found: {base}"}))
else:
    def scan(p, d):
        items = []
        try:
            for name in sorted(os.listdir(p)):
                full = os.path.join(p, name)
                info = {"name": name, "type": "dir" if os.path.isdir(full) else "file"}
                if info["type"] == "file":
                    info["size"] = os.path.getsize(full)
                elif d > 1:
                    info["children"] = scan(full, d-1)
                items.append(info)
        except PermissionError:
            pass
        return items
    print(json.dumps({"path": base, "items": scan(base, %d)}))
`, pathB64, depth)

	output, err := s.runHiddenCell(ctx, code)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(output)
	return r, Empty{}, nil
}

type ReadFileInput struct {
	Path     string `json:"path" jsonschema:"Absolute path or path relative to /content/"`
	MaxLines int    `json:"maxLines" jsonschema:"Maximum lines to read"`
}

func (s *Server) readFile(ctx context.Context, req *mcp.CallToolRequest, input ReadFileInput) (*mcp.CallToolResult, Empty, error) {
	maxLines := input.MaxLines
	if maxLines <= 0 {
		maxLines = 100
	}
	pathB64 := base64.StdEncoding.EncodeToString([]byte(input.Path))
	code := fmt.Sprintf(`import json, base64
path = base64.b64decode('%s').decode()
if not path.startswith('/'):
    path = '/content/' + path
try:
    with open(path) as f:
        lines = f.readlines()
    numbered = [f"{i+1:3d} | {l.rstrip()}" for i, l in enumerate(lines[:%d])]
    print(json.dumps({"path": path, "totalLines": len(lines), "shown": min(len(lines), %d), "content": "\n".join(numbered)}))
except FileNotFoundError:
    print(json.dumps({"error": f"File not found: {path}"}))
except Exception as e:
    print(json.dumps({"error": str(e)}))
`, pathB64, maxLines, maxLines)

	output, err := s.runHiddenCell(ctx, code)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(output)
	return r, Empty{}, nil
}

type WriteFileInput struct {
	Path    string `json:"path" jsonschema:"Absolute path or path relative to /content/"`
	Content string `json:"content" jsonschema:"File content to write"`
}

func (s *Server) writeFile(ctx context.Context, req *mcp.CallToolRequest, input WriteFileInput) (*mcp.CallToolResult, Empty, error) {
	pathB64 := base64.StdEncoding.EncodeToString([]byte(input.Path))
	contentB64 := base64.StdEncoding.EncodeToString([]byte(input.Content))
	code := fmt.Sprintf(`import os, json, base64
path = base64.b64decode('%s').decode()
if not path.startswith('/'):
    path = '/content/' + path
os.makedirs(os.path.dirname(path) if os.path.dirname(path) else '.', exist_ok=True)
content = base64.b64decode('%s').decode()
with open(path, 'w') as f:
    f.write(content)
print(json.dumps({"written": path, "size": len(content)}))
`, pathB64, contentB64)

	output, err := s.runHiddenCell(ctx, code)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(output)
	return r, Empty{}, nil
}

// ── Group D — Context Management ─────────────────────

func (s *Server) getNotebookOutline(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}

	var outline []map[string]any
	for i, cell := range cells {
		source := strings.Join(cell.Source, "")
		lines := strings.Split(source, "\n")

		summary := ""
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				summary = trimmed
				break
			}
		}
		if len(summary) > 80 {
			summary = summary[:80] + "..."
		}

		var defs []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") {
				name := strings.SplitN(trimmed, "(", 2)[0]
				defs = append(defs, name)
			}
		}

		entry := map[string]any{
			"index": i, "cellId": cell.ID, "type": cell.CellType,
			"lines": len(lines), "summary": summary,
		}
		if len(defs) > 0 {
			entry["definitions"] = defs
		}
		outline = append(outline, entry)
	}
	r, _ := jsonResult(outline)
	return r, Empty{}, nil
}

// ── Group E — Environment ────────────────────────────

func (s *Server) getEnvironment(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	code := `import json, sys, shutil
info = {"python": sys.version.split()[0]}
try:
    import torch
    info["cuda"] = torch.version.cuda or "N/A"
    if torch.cuda.is_available():
        info["gpu"] = torch.cuda.get_device_name(0)
except ImportError:
    info["cuda"] = "N/A"
try:
    import psutil
    info["ram_gb"] = round(psutil.virtual_memory().total / 1e9, 1)
except ImportError:
    pass
disk = shutil.disk_usage('/content')
info["disk_total_gb"] = round(disk.total / 1e9, 1)
info["disk_used_gb"] = round(disk.used / 1e9, 1)
print(json.dumps(info))
`
	output, err := s.runHiddenCell(ctx, code)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(output)
	return r, Empty{}, nil
}

type PackageNameInput struct {
	PackageName string `json:"packageName" jsonschema:"Package name to check"`
}

func (s *Server) checkPackage(ctx context.Context, req *mcp.CallToolRequest, input PackageNameInput) (*mcp.CallToolResult, Empty, error) {
	nameB64 := base64.StdEncoding.EncodeToString([]byte(input.PackageName))
	code := fmt.Sprintf(`import json, importlib, base64
pkg = base64.b64decode('%s').decode()
try:
    mod = importlib.import_module(pkg)
    ver = getattr(mod, '__version__', 'unknown')
    print(json.dumps({"installed": True, "package": pkg, "version": ver}))
except ImportError:
    print(json.dumps({"installed": False, "package": pkg}))
`, nameB64)

	output, err := s.runHiddenCell(ctx, code)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(output)
	return r, Empty{}, nil
}

// ── Group F — Output Tracking ────────────────────────

// GetCellOutputInput extends CellIDInput with an optional flag to include raw image data.
type GetCellOutputInput struct {
	CellID        string `json:"cellId" jsonschema:"Cell ID"`
	IncludeImages bool   `json:"includeImages,omitempty" jsonschema:"If true, include raw base64 image data in response. Default false to save tokens."`
}

func (s *Server) getCellOutput(ctx context.Context, req *mcp.CallToolRequest, input GetCellOutputInput) (*mcp.CallToolResult, Empty, error) {
	cell, err := s.fetchCellWithOutputs(ctx, input.CellID)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}

	stdout, errors, images := s.parseCellOutputs(cell)

	resultMap := map[string]any{
		"cellId":      cell.ID,
		"hasOutput":   len(cell.Outputs) > 0,
		"stdout":      stdout,
		"errors":      errors,
		"hasImages":   len(images) > 0,
		"imageCount":  len(images),
		"outputCount": len(cell.Outputs),
	}

	// Only include heavy base64 data when explicitly requested
	if input.IncludeImages {
		resultMap["images"] = images
	}

	r, _ := jsonResult(resultMap)
	return r, Empty{}, nil
}

// SaveCellImagesInput defines the parameters for the save_cell_images tool.
type SaveCellImagesInput struct {
	CellID     string `json:"cellId" jsonschema:"Cell ID to extract images from"`
	OutputDir  string `json:"outputDir,omitempty" jsonschema:"Local directory path to save images (absolute or relative to CWD). Defaults to './exports'."`
	FilePrefix string `json:"filePrefix,omitempty" jsonschema:"Prefix for saved filenames. Default is the cellId."`
}

func (s *Server) saveCellImages(ctx context.Context, req *mcp.CallToolRequest, input SaveCellImagesInput) (*mcp.CallToolResult, Empty, error) {
	if input.OutputDir == "" {
		input.OutputDir = "exports"
	}

	cell, err := s.fetchCellWithOutputs(ctx, input.CellID)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}

	_, _, images := s.parseCellOutputs(cell)

	if len(images) == 0 {
		r, _ := jsonResult(map[string]any{
			"cellId": input.CellID,
			"saved":  0,
			"files":  []string{},
		})
		return r, Empty{}, nil
	}

	// Create output directory
	if err := os.MkdirAll(input.OutputDir, 0755); err != nil {
		r, _ := errResult("failed to create output directory: " + err.Error())
		return r, Empty{}, nil
	}

	prefix := input.FilePrefix
	if prefix == "" {
		prefix = input.CellID
	}

	var savedFiles []string
	for i, img := range images {
		b64Data, _ := img["data"].(string)
		mimeType, _ := img["mimeType"].(string)
		if b64Data == "" {
			continue
		}

		// Determine file extension from MIME type
		ext := ".png"
		if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
			ext = ".jpg"
		}

		// Decode base64
		imgBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Data))
		if err != nil {
			// Try with padding fix
			imgBytes, err = base64.RawStdEncoding.DecodeString(strings.TrimSpace(b64Data))
			if err != nil {
				continue
			}
		}

		// Generate MD5 hash of image content for de-duplication
		hash := md5.Sum(imgBytes)
		hashStr := fmt.Sprintf("%x", hash)[:8]

		// Build filename
		var filename string
		if len(images) == 1 {
			filename = fmt.Sprintf("%s_%s%s", prefix, hashStr, ext)
		} else {
			filename = fmt.Sprintf("%s_%d_%s%s", prefix, i+1, hashStr, ext)
		}

		fullPath := filepath.Join(input.OutputDir, filename)

		// De-duplication: Skip writing if file with exact same content hash already exists
		if _, err := os.Stat(fullPath); err == nil {
			savedFiles = append(savedFiles, fullPath)
			continue
		}

		// Write file
		if err := os.WriteFile(fullPath, imgBytes, 0644); err != nil {
			r, _ := errResult(fmt.Sprintf("failed to write %s: %s", filename, err.Error()))
			return r, Empty{}, nil
		}

		savedFiles = append(savedFiles, fullPath)
	}

	r, _ := jsonResult(map[string]any{
		"cellId": input.CellID,
		"saved":  len(savedFiles),
		"files":  savedFiles,
	})
	return r, Empty{}, nil
}

// ── Shared Cell Output Helpers ────────────────────────

// cellWithOutputs is the parsed structure of a single cell with its outputs.
type cellWithOutputs struct {
	ID      string
	Outputs []cellOutput
}

type cellOutput struct {
	OutputType string         `json:"output_type"`
	Text       []string       `json:"text"`
	EName      string         `json:"ename"`
	EValue     string         `json:"evalue"`
	Traceback  []string       `json:"traceback"`
	Data       map[string]any `json:"data"`
}

// fetchCellWithOutputs retrieves a single cell by ID with its outputs.
func (s *Server) fetchCellWithOutputs(ctx context.Context, cellID string) (*cellWithOutputs, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		return nil, err
	}

	targetIdx := -1
	for i, c := range cells {
		if c.ID == cellID {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return nil, fmt.Errorf("cell not found: %s", cellID)
	}

	raw, err := s.proxy.CallTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": targetIdx, "cellIndexEnd": targetIdx + 1, "includeOutputs": true,
	})
	if err != nil {
		return nil, err
	}

	result := extractProxyText(raw)
	var resp struct {
		Cells []struct {
			ID      string       `json:"id"`
			Outputs []cellOutput `json:"outputs"`
		} `json:"cells"`
	}
	json.Unmarshal(result, &resp)

	if len(resp.Cells) == 0 {
		return nil, fmt.Errorf("cell not found: %s", cellID)
	}

	return &cellWithOutputs{
		ID:      resp.Cells[0].ID,
		Outputs: resp.Cells[0].Outputs,
	}, nil
}

// parseCellOutputs extracts stdout text, errors, and image data from cell outputs.
func (s *Server) parseCellOutputs(cell *cellWithOutputs) (string, []map[string]any, []map[string]any) {
	var stdout strings.Builder
	var errors []map[string]any
	var images []map[string]any

	for _, o := range cell.Outputs {
		switch o.OutputType {
		case "stream":
			stdout.WriteString(strings.Join(o.Text, ""))
		case "error":
			errInfo := map[string]any{
				"type":    o.EName,
				"message": o.EValue,
			}
			if len(o.Traceback) > 0 {
				errInfo["traceback"] = strings.Join(o.Traceback, "\n")
			}
			errors = append(errors, errInfo)
		case "display_data", "execute_result":
			if o.Data != nil {
				for _, mime := range []string{"image/png", "image/jpeg"} {
					if imgData, ok := o.Data[mime]; ok {
						var b64 string
						switch v := imgData.(type) {
						case string:
							b64 = v
						case []any:
							var sb strings.Builder
							for _, s := range v {
								sb.WriteString(fmt.Sprintf("%v", s))
							}
							b64 = sb.String()
						}
						if b64 != "" {
							images = append(images, map[string]any{
								"mimeType": mime,
								"data":     b64,
							})
						}
					}
				}
				if textData, ok := o.Data["text/plain"]; ok {
					if arr, ok := textData.([]any); ok {
						for _, v := range arr {
							stdout.WriteString(fmt.Sprintf("%v", v))
						}
					} else {
						stdout.WriteString(fmt.Sprintf("%v", textData))
					}
				}
			}
		}
	}

	return stdout.String(), errors, images
}

func (s *Server) getRunningCells(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	var running []map[string]any
	s.runningCells.Range(func(key, value any) bool {
		cellId := key.(string)
		startTime := value.(time.Time)
		running = append(running, map[string]any{
			"cellId":  cellId,
			"elapsed": time.Since(startTime).Round(time.Second).String(),
		})
		return true
	})
	r, _ := jsonResult(map[string]any{"running": running, "count": len(running)})
	return r, Empty{}, nil
}

func (s *Server) getErrorCells(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	raw, err := s.proxy.CallTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": 0, "cellIndexEnd": 500, "includeOutputs": true,
	})
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}

	result := extractProxyText(raw)
	var resp struct {
		Cells []struct {
			ID      string   `json:"id"`
			Source  []string `json:"source"`
			Outputs []struct {
				OutputType string `json:"output_type"`
				EName      string `json:"ename"`
				EValue     string `json:"evalue"`
			} `json:"outputs"`
		} `json:"cells"`
	}
	json.Unmarshal(result, &resp)

	var errors []map[string]any
	for i, cell := range resp.Cells {
		for _, o := range cell.Outputs {
			if o.OutputType == "error" {
				preview := strings.Join(cell.Source, "")
				if len(preview) > 100 {
					preview = preview[:100]
				}
				errors = append(errors, map[string]any{
					"cellIndex": i, "cellId": cell.ID,
					"errorType": o.EName, "message": o.EValue, "preview": preview,
				})
			}
		}
	}
	r, _ := jsonResult(map[string]any{"errorCells": errors, "total": len(errors)})
	return r, Empty{}, nil
}
