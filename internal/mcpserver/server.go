package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hoangnecon/golab/internal/bridge"
	"github.com/hoangnecon/golab/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const Version = "v1.0.0"

// Server wraps the MCP server with Colab browser proxy.
type Server struct {
	ws    *bridge.WSServer
	proxy *bridge.BrowserProxy
	cfg   *config.Config
	token string
	port  int
	mcp   *mcp.Server
}

func New(ws *bridge.WSServer, cfg *config.Config, token string, port int) *Server {
	s := &Server{
		ws:    ws,
		proxy: bridge.NewBrowserProxy(ws),
		cfg:   cfg,
		token: token,
		port:  port,
	}
	s.mcp = mcp.NewServer(&mcp.Implementation{
		Name:    "golab",
		Version: Version,
	}, nil)
	s.registerTools()
	return s
}

func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	// Meta
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "check_status",
		Description: "Check if a Colab browser is connected. Returns connection status, uptime, and WebSocket port.",
	}, s.checkStatus)

	// Group A — Browser Proxy
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "open_notebook",
		Description: "Opens a Colab notebook. If browser is already connected, reuses existing connection without opening a new tab.",
	}, s.openNotebook)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_cells", Description: "Get cells from the connected notebook. Returns cell IDs, types, source code, and optionally outputs."}, s.getCells)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "add_code_cell", Description: "Add a new code cell at the specified index with the given code."}, s.addCodeCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "add_text_cell", Description: "Add a new markdown/text cell at the specified index."}, s.addTextCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "update_cell", Description: "Replace the entire content of a cell by its ID."}, s.updateCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "delete_cell", Description: "Delete a cell by its ID."}, s.deleteCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "move_cell", Description: "Move a cell to a new index position."}, s.moveCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "run_code_cell", Description: "Execute a code cell and return its outputs. Requires a connected runtime."}, s.runCodeCell)

	// Group B — IDE Cell Editing
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "edit_cell_lines", Description: "Edit specific lines within a cell. Replaces lines from startLine to endLine with newContent. Lines are 1-indexed."}, s.editCellLines)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "find_replace_in_cell", Description: "Find and replace text within a single cell."}, s.findReplaceInCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "find_replace_all", Description: "Find and replace text across all cells in the notebook."}, s.findReplaceAll)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "insert_in_cell", Description: "Insert new code at a specific line number within a cell. Existing code shifts down. Line 0 = top."}, s.insertInCell)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "search_cells", Description: "Search for a pattern across all cells. Returns matches with cell ID, line number, and context."}, s.searchCells)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_cell_with_lines", Description: "Get a cell's content with line numbers for easy reference."}, s.getCellWithLines)

	// Group C — Project Assessment
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "list_drive", Description: "List files and directories in Google Drive. Returns a JSON tree."}, s.listDrive)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "read_file", Description: "Read a text file from the Colab VM or mounted Drive. Returns content with line numbers."}, s.readFile)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "write_file", Description: "Write content to a file on the Colab VM or mounted Drive."}, s.writeFile)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "create_project_structure", Description: "Create a project directory structure on Drive with folders: data/, models/, checkpoints/, logs/, configs/."}, s.createProjectStructure)

	// Group D — Context Management
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_notebook_outline", Description: "Get a compact outline of the notebook: cell index, type, first line summary, function/class definitions, line count."}, s.getNotebookOutline)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_recent_changes", Description: "Detect changes since the last call. Returns added, modified, and deleted cells."}, s.getRecentChanges)

	// Group E — Environment
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_environment", Description: "Get system info: Python version, CUDA version, GPU model, RAM, disk usage."}, s.getEnvironment)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "list_packages", Description: "List installed Python packages with versions."}, s.listPackages)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "check_package", Description: "Check if a specific Python package is installed and its version."}, s.checkPackage)

	// Group F — Output Tracking
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_cell_output", Description: "Get the output of a specific cell by ID. Returns text output, errors, and image data."}, s.getCellOutput)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_running_cells", Description: "Get a list of cells currently being executed."}, s.getRunningCells)
	mcp.AddTool(s.mcp, &mcp.Tool{Name: "get_error_cells", Description: "Get cells that produced errors, with error type and message."}, s.getErrorCells)
}

// ── URL Parsing ──────────────────────────────────────

var (
	colabDriveRe = regexp.MustCompile(`colab\.research\.google\.com/drive/([a-zA-Z0-9_-]+)`)
	driveFileRe  = regexp.MustCompile(`drive\.google\.com/file/d/([a-zA-Z0-9_-]+)`)
	driveOpenRe  = regexp.MustCompile(`drive\.google\.com/open\?id=([a-zA-Z0-9_-]+)`)
	fileIDRe     = regexp.MustCompile(`^[a-zA-Z0-9_-]{20,}$`)
)

func extractNotebookPath(url string) string {
	if url == "" {
		return "/notebooks/empty.ipynb"
	}
	if m := colabDriveRe.FindStringSubmatch(url); len(m) > 1 {
		return "/drive/" + m[1]
	}
	if m := driveFileRe.FindStringSubmatch(url); len(m) > 1 {
		return "/drive/" + m[1]
	}
	if m := driveOpenRe.FindStringSubmatch(url); len(m) > 1 {
		return "/drive/" + m[1]
	}
	if fileIDRe.MatchString(url) {
		return "/drive/" + url
	}
	return "/drive/" + url
}

// ── MCP Response Helpers ─────────────────────────────

// extractProxyText unwraps the MCP content envelope from browser proxy responses.
func extractProxyText(raw json.RawMessage) json.RawMessage {
	var mcpResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &mcpResp); err == nil && len(mcpResp.Content) > 0 {
		return json.RawMessage(mcpResp.Content[0].Text)
	}
	return raw
}

// ── Hidden Cell Execution ────────────────────────────

func (s *Server) runHiddenCell(ctx context.Context, code string) (string, error) {
	cellsRaw, err := s.proxy.CallTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": 0, "cellIndexEnd": 200, "includeOutputs": false,
	})
	if err != nil {
		return "", fmt.Errorf("get_cells: %w", err)
	}

	cellsResult := extractProxyText(cellsRaw)
	var cellsResp struct {
		Cells []json.RawMessage `json:"cells"`
	}
	json.Unmarshal(cellsResult, &cellsResp)
	idx := len(cellsResp.Cells)

	addRaw, err := s.proxy.CallTool(ctx, "add_code_cell", map[string]any{
		"cellIndex": idx, "language": "python", "code": code,
	})
	if err != nil {
		return "", fmt.Errorf("add_code_cell: %w", err)
	}

	addResult := extractProxyText(addRaw)
	var addResp struct {
		NewCellID string `json:"newCellId"`
	}
	json.Unmarshal(addResult, &addResp)
	if addResp.NewCellID == "" {
		return "", fmt.Errorf("failed to create hidden cell: %s", string(addResult))
	}

	runRaw, err := s.proxy.CallTool(ctx, "run_code_cell", map[string]any{"cellId": addResp.NewCellID})
	s.proxy.CallTool(ctx, "delete_cell", map[string]any{"cellId": addResp.NewCellID}) // cleanup

	if err != nil {
		return "", fmt.Errorf("run_code_cell: %w", err)
	}

	runResult := extractProxyText(runRaw)
	var runResp struct {
		Outputs []struct {
			OutputType string   `json:"output_type"`
			Text       []string `json:"text"`
			EName      string   `json:"ename"`
			EValue     string   `json:"evalue"`
		} `json:"outputs"`
	}
	json.Unmarshal(runResult, &runResp)

	var output strings.Builder
	for _, o := range runResp.Outputs {
		switch o.OutputType {
		case "stream":
			output.WriteString(strings.Join(o.Text, ""))
		case "error":
			output.WriteString(fmt.Sprintf("ERROR: %s: %s", o.EName, o.EValue))
		}
	}
	return output.String(), nil
}

// ── Cell Helpers ─────────────────────────────────────

type cellInfo struct {
	ID       string   `json:"id"`
	CellType string   `json:"cell_type"`
	Source   []string `json:"source"`
}

func (s *Server) getAllCells(ctx context.Context, includeOutputs bool) ([]cellInfo, json.RawMessage, error) {
	raw, err := s.proxy.CallTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": 0, "cellIndexEnd": 500, "includeOutputs": includeOutputs,
	})
	if err != nil {
		return nil, nil, err
	}
	result := extractProxyText(raw)
	var resp struct {
		Cells []cellInfo `json:"cells"`
	}
	json.Unmarshal(result, &resp)
	return resp.Cells, result, nil
}

// ── Result Helpers ───────────────────────────────────

func textResult(text string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return textResult(string(data))
}

func errResult(msg string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Error: " + msg}},
		IsError: true,
	}, nil
}

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}
