package mcpserver

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Empty is a no-op output type for tools that return results via CallToolResult.
type Empty struct{}

// proxyTool forwards a tool call to the browser and returns the raw result.
func (s *Server) proxyTool(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, Empty, error) {
	result, err := s.proxy.CallTool(ctx, name, args)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	r, _ := textResult(string(result))
	return r, Empty{}, nil
}

// openBrowser opens a URL in the default browser, cross-platform.
func openBrowser(ctx context.Context, url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Run()
}

// ── Meta Tools ───────────────────────────────────────

func (s *Server) checkStatus(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Empty, error) {
	st := s.ws.Status()
	r, _ := jsonResult(st)
	return r, Empty{}, nil
}

// ── Group A — Browser Proxy ──────────────────────────

type OpenNotebookInput struct {
	NotebookURL string `json:"notebook_url" jsonschema:"Colab URL or Drive URL or file ID. Empty for scratchpad."`
	ForceNew    bool   `json:"force_new,omitempty" jsonschema:"If true, disconnects any existing notebook to open this one."`
}

func (s *Server) openNotebook(ctx context.Context, req *mcp.CallToolRequest, input OpenNotebookInput) (*mcp.CallToolResult, Empty, error) {
	if s.ws.IsConnected() {
		if !input.ForceNew {
			st := s.ws.Status()
			r, _ := jsonResult(map[string]any{
				"status":  "already_connected",
				"message": "Browser already connected, reusing existing connection.",
				"uptime":  st.Uptime,
				"wsPort":  st.WSPort,
			})
			return r, Empty{}, nil
		}

		// Force new connection: simply disconnect the existing browser
		// so the new tab takes over. We DO NOT rotate the token, otherwise 
		// IDE restarts will lose the ephemeral token and break the connection!
		s.ws.DisconnectAndRotateToken(s.token)
	}

	path := extractNotebookPath(input.NotebookURL)
	url := fmt.Sprintf("https://%s%s#mcpProxyToken=%s&mcpProxyPort=%d", s.cfg.ColabBaseURL, path, s.token, s.port)

	if err := openBrowser(ctx, url); err != nil {
		r, _ := errResult(fmt.Sprintf("failed to open browser: %v", err))
		return r, Empty{}, nil
	}

	select {
	case <-s.ws.WaitConnected():
		r, _ := jsonResult(map[string]any{
			"status":  "connected",
			"message": "Browser connected successfully.",
			"wsPort":  s.port,
		})
		return r, Empty{}, nil
	case <-ctx.Done():
		r, _ := errResult("cancelled")
		return r, Empty{}, nil
	}
}

type GetCellsInput struct {
	CellIndexStart int  `json:"cellIndexStart" jsonschema:"Start index (0-based)"`
	CellIndexEnd   int  `json:"cellIndexEnd" jsonschema:"End index (exclusive)"`
	IncludeOutputs bool `json:"includeOutputs" jsonschema:"Include cell outputs in response"`
}

func (s *Server) getCells(ctx context.Context, req *mcp.CallToolRequest, input GetCellsInput) (*mcp.CallToolResult, Empty, error) {
	end := input.CellIndexEnd
	if end == 0 {
		end = 100
	}
	return s.proxyTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": input.CellIndexStart, "cellIndexEnd": end, "includeOutputs": input.IncludeOutputs,
	})
}

type AddCodeCellInput struct {
	CellIndex int    `json:"cellIndex" jsonschema:"Position to insert cell (0-based)"`
	Code      string `json:"code" jsonschema:"Python code content"`
}

func (s *Server) addCodeCell(ctx context.Context, req *mcp.CallToolRequest, input AddCodeCellInput) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "add_code_cell", map[string]any{
		"cellIndex": input.CellIndex, "language": "python", "code": input.Code,
	})
}

type AddTextCellInput struct {
	CellIndex int    `json:"cellIndex" jsonschema:"Position to insert cell"`
	Text      string `json:"text" jsonschema:"Markdown content"`
}

func (s *Server) addTextCell(ctx context.Context, req *mcp.CallToolRequest, input AddTextCellInput) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "add_text_cell", map[string]any{"cellIndex": input.CellIndex, "content": input.Text})
}

type UpdateCellInput struct {
	CellID  string `json:"cellId" jsonschema:"Cell ID to update"`
	Content string `json:"content" jsonschema:"New cell content (replaces entire cell)"`
}

func (s *Server) updateCell(ctx context.Context, req *mcp.CallToolRequest, input UpdateCellInput) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "update_cell", map[string]any{"cellId": input.CellID, "content": input.Content})
}

type CellIDInput struct {
	CellID string `json:"cellId" jsonschema:"Cell ID"`
}

func (s *Server) deleteCell(ctx context.Context, req *mcp.CallToolRequest, input CellIDInput) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "delete_cell", map[string]any{"cellId": input.CellID})
}

type MoveCellInput struct {
	CellID       string `json:"cellId" jsonschema:"Cell ID to move"`
	NewCellIndex int    `json:"newCellIndex" jsonschema:"New position index"`
}

func (s *Server) moveCell(ctx context.Context, req *mcp.CallToolRequest, input MoveCellInput) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "move_cell", map[string]any{"cellId": input.CellID, "newCellIndex": input.NewCellIndex})
}

func (s *Server) runCodeCell(ctx context.Context, req *mcp.CallToolRequest, input CellIDInput) (*mcp.CallToolResult, Empty, error) {
	return s.proxyTool(ctx, "run_code_cell", map[string]any{"cellId": input.CellID})
}

// ── Group B — IDE Cell Editing ───────────────────────

type EditCellLinesInput struct {
	CellID     string `json:"cellId" jsonschema:"Cell ID to edit"`
	StartLine  int    `json:"startLine" jsonschema:"First line to replace (1-indexed)"`
	EndLine    int    `json:"endLine" jsonschema:"Last line to replace (1-indexed inclusive)"`
	NewContent string `json:"newContent" jsonschema:"Replacement content for the specified lines"`
}

func (s *Server) editCellLines(ctx context.Context, req *mcp.CallToolRequest, input EditCellLinesInput) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	for _, cell := range cells {
		if cell.ID != input.CellID {
			continue
		}
		source := strings.Join(cell.Source, "")
		lines := strings.Split(source, "\n")
		if input.StartLine < 1 || input.EndLine > len(lines) || input.StartLine > input.EndLine {
			r, _ := errResult(fmt.Sprintf("invalid line range %d-%d (cell has %d lines)", input.StartLine, input.EndLine, len(lines)))
			return r, Empty{}, nil
		}
		newLines := make([]string, 0, len(lines))
		newLines = append(newLines, lines[:input.StartLine-1]...)
		newLines = append(newLines, strings.Split(input.NewContent, "\n")...)
		newLines = append(newLines, lines[input.EndLine:]...)
		s.proxy.CallTool(ctx, "update_cell", map[string]any{"cellId": input.CellID, "content": strings.Join(newLines, "\n")})
		r, _ := jsonResult(map[string]any{"modified": true, "linesChanged": input.EndLine - input.StartLine + 1, "totalLines": len(newLines)})
		return r, Empty{}, nil
	}
	r, _ := errResult("cell not found: " + input.CellID)
	return r, Empty{}, nil
}

type FindReplaceInput struct {
	CellID     string `json:"cellId" jsonschema:"Cell ID"`
	Find       string `json:"find" jsonschema:"Text to find"`
	Replace    string `json:"replace" jsonschema:"Replacement text"`
	ReplaceAll bool   `json:"replaceAll" jsonschema:"Replace all occurrences"`
}

func (s *Server) findReplaceInCell(ctx context.Context, req *mcp.CallToolRequest, input FindReplaceInput) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	for _, cell := range cells {
		if cell.ID != input.CellID {
			continue
		}
		source := strings.Join(cell.Source, "")
		count := strings.Count(source, input.Find)
		if count == 0 {
			r, _ := jsonResult(map[string]any{"replacements": 0})
			return r, Empty{}, nil
		}
		var newSource string
		if input.ReplaceAll {
			newSource = strings.ReplaceAll(source, input.Find, input.Replace)
		} else {
			newSource = strings.Replace(source, input.Find, input.Replace, 1)
			count = 1
		}
		s.proxy.CallTool(ctx, "update_cell", map[string]any{"cellId": input.CellID, "content": newSource})
		r, _ := jsonResult(map[string]any{"replacements": count})
		return r, Empty{}, nil
	}
	r, _ := errResult("cell not found: " + input.CellID)
	return r, Empty{}, nil
}

type FindReplaceAllInput struct {
	Find    string `json:"find" jsonschema:"Text to find"`
	Replace string `json:"replace" jsonschema:"Replacement text"`
}

func (s *Server) findReplaceAll(ctx context.Context, req *mcp.CallToolRequest, input FindReplaceAllInput) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	var results []map[string]any
	for _, cell := range cells {
		source := strings.Join(cell.Source, "")
		count := strings.Count(source, input.Find)
		if count == 0 {
			continue
		}
		s.proxy.CallTool(ctx, "update_cell", map[string]any{"cellId": cell.ID, "content": strings.ReplaceAll(source, input.Find, input.Replace)})
		results = append(results, map[string]any{"cellId": cell.ID, "replacements": count})
	}
	r, _ := jsonResult(map[string]any{"affectedCells": results, "totalCells": len(results)})
	return r, Empty{}, nil
}

type InsertInCellInput struct {
	CellID     string `json:"cellId" jsonschema:"Cell ID"`
	LineNumber int    `json:"lineNumber" jsonschema:"Line number to insert at (0=top)"`
	Content    string `json:"content" jsonschema:"Code to insert"`
}

func (s *Server) insertInCell(ctx context.Context, req *mcp.CallToolRequest, input InsertInCellInput) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	for _, cell := range cells {
		if cell.ID != input.CellID {
			continue
		}
		source := strings.Join(cell.Source, "")
		lines := strings.Split(source, "\n")
		insertLines := strings.Split(input.Content, "\n")
		pos := input.LineNumber
		if pos < 0 {
			pos = 0
		}
		if pos > len(lines) {
			pos = len(lines)
		}
		newLines := make([]string, 0, len(lines)+len(insertLines))
		newLines = append(newLines, lines[:pos]...)
		newLines = append(newLines, insertLines...)
		newLines = append(newLines, lines[pos:]...)
		_, err = s.proxy.CallTool(ctx, "update_cell", map[string]any{"cellId": cell.ID, "content": strings.Join(newLines, "\n")})
		if err != nil {
			r, _ := errResult(fmt.Sprintf("update_cell failed: %v", err))
			return r, Empty{}, nil
		}
		r, _ := jsonResult(map[string]any{"inserted": true, "linesInserted": len(insertLines), "totalLines": len(newLines)})
		return r, Empty{}, nil
	}
	r, _ := errResult("cell not found: " + input.CellID)
	return r, Empty{}, nil
}

type SearchCellsInput struct {
	Query string `json:"query" jsonschema:"Text to search for"`
}

func (s *Server) searchCells(ctx context.Context, req *mcp.CallToolRequest, input SearchCellsInput) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	var matches []map[string]any
	for i, cell := range cells {
		source := strings.Join(cell.Source, "")
		for lineIdx, line := range strings.Split(source, "\n") {
			if strings.Contains(line, input.Query) {
				matches = append(matches, map[string]any{
					"cellIndex": i, "cellId": cell.ID, "cellType": cell.CellType,
					"line": lineIdx + 1, "content": strings.TrimSpace(line),
				})
			}
		}
	}
	r, _ := jsonResult(map[string]any{"matches": matches, "total": len(matches)})
	return r, Empty{}, nil
}

func (s *Server) getCellWithLines(ctx context.Context, req *mcp.CallToolRequest, input CellIDInput) (*mcp.CallToolResult, Empty, error) {
	cells, _, err := s.getAllCells(ctx, false)
	if err != nil {
		r, _ := errResult(err.Error())
		return r, Empty{}, nil
	}
	for _, cell := range cells {
		if cell.ID != input.CellID {
			continue
		}
		source := strings.Join(cell.Source, "")
		lines := strings.Split(source, "\n")
		var sb strings.Builder
		for i, line := range lines {
			sb.WriteString(fmt.Sprintf("%3d | %s\n", i+1, line))
		}
		r, _ := jsonResult(map[string]any{"cellId": cell.ID, "cellType": cell.CellType, "lines": len(lines), "content": sb.String()})
		return r, Empty{}, nil
	}
	r, _ := errResult("cell not found: " + input.CellID)
	return r, Empty{}, nil
}
