package mcpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

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

type ProjectNameInput struct {
	ProjectName string `json:"projectName" jsonschema:"Project name (created under MyDrive)"`
}

func (s *Server) createProjectStructure(ctx context.Context, req *mcp.CallToolRequest, input ProjectNameInput) (*mcp.CallToolResult, Empty, error) {
	nameB64 := base64.StdEncoding.EncodeToString([]byte(input.ProjectName))
	code := fmt.Sprintf(`import os, json, base64
name = base64.b64decode('%s').decode()
base = '/content/drive/MyDrive/' + name
dirs = ['data', 'models', 'checkpoints', 'logs', 'configs']
created = []
for d in dirs:
    p = os.path.join(base, d)
    os.makedirs(p, exist_ok=True)
    created.append(p)
print(json.dumps({"project": base, "created": created}))
`, nameB64)

	output, err := s.runHiddenCell(ctx, code)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(output)
	return r, Empty{}, nil
}

// ── Group D — Context Management ─────────────────────

var (
	lastSnapshotMu sync.Mutex
	lastSnapshot   map[string]string
)

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

func (s *Server) getRecentChanges(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}

	current := make(map[string]string)
	for _, cell := range cells {
		current[cell.ID] = strings.Join(cell.Source, "")
	}

	lastSnapshotMu.Lock()
	prev := lastSnapshot
	lastSnapshot = current
	lastSnapshotMu.Unlock()

	if prev == nil {
		r, _ := jsonResult(map[string]any{"status": "first_snapshot", "cells": len(current)})
		return r, Empty{}, nil
	}

	var added, modified, deleted []string
	for id := range current {
		if _, ok := prev[id]; !ok {
			added = append(added, id)
		} else if current[id] != prev[id] {
			modified = append(modified, id)
		}
	}
	for id := range prev {
		if _, ok := current[id]; !ok {
			deleted = append(deleted, id)
		}
	}
	r, _ := jsonResult(map[string]any{
		"added": added, "modified": modified, "deleted": deleted,
		"unchanged": len(current) - len(added) - len(modified),
	})
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

func (s *Server) listPackages(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	code := `import json, subprocess
r = subprocess.run(['pip', 'list', '--format=json'], capture_output=True, text=True)
pkgs = json.loads(r.stdout)
print(json.dumps({"packages": pkgs, "total": len(pkgs)}))
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

func (s *Server) getCellOutput(ctx context.Context, req *mcp.CallToolRequest, input CellIDInput) (*mcp.CallToolResult, Empty, error) {
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
			ID      string `json:"id"`
			Outputs []struct {
				OutputType string   `json:"output_type"`
				Text       []string `json:"text"`
				EName      string   `json:"ename"`
				EValue     string   `json:"evalue"`
			} `json:"outputs"`
		} `json:"cells"`
	}
	json.Unmarshal(result, &resp)

	for _, cell := range resp.Cells {
		if cell.ID != input.CellID {
			continue
		}
		var sb strings.Builder
		for _, o := range cell.Outputs {
			switch o.OutputType {
			case "stream":
				sb.WriteString(strings.Join(o.Text, ""))
			case "error":
				sb.WriteString(fmt.Sprintf("ERROR: %s: %s\n", o.EName, o.EValue))
			default:
				sb.WriteString(strings.Join(o.Text, ""))
			}
		}
		r, _ := jsonResult(map[string]any{
			"cellId": cell.ID, "hasOutput": len(cell.Outputs) > 0,
			"outputCount": len(cell.Outputs), "content": sb.String(),
		})
		return r, Empty{}, nil
	}
	r, _ := errResult("cell not found: " + input.CellID)
	return r, Empty{}, nil
}

func (s *Server) getRunningCells(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": 0, "cellIndexEnd": 500, "includeOutputs": false,
	})
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
