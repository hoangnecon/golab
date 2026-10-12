package mcpserver

import (
	"context"

	"github.com/hoangnecon/golab/internal/bridge"
	"github.com/hoangnecon/golab/internal/service/kernel"
	"github.com/hoangnecon/golab/internal/service/notebook"
	"github.com/hoangnecon/golab/internal/service/proxy"
	"github.com/hoangnecon/golab/internal/service/storage"
	"github.com/hoangnecon/golab/internal/service/telemetry"
	"github.com/hoangnecon/golab/internal/service/terminal"
	"github.com/hoangnecon/golab/internal/session"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const ServerVersion = "v2.0.0-harness"

// Server wraps the enterprise Colab MCP Server.
type Server struct {
	mcp        *mcp.Server
	sessionMgr *session.Manager
	cdp        *bridge.CDPClient
	exec       *bridge.ColabExecutor
	storage    *storage.Service
	kernel     *kernel.Service
	telemetry  *telemetry.Service
	terminal   *terminal.Service
	notebook   *notebook.Service
	proxy      *proxy.Service
}

// NewServer creates and initializes a complete MCP server with all domain services.
func NewServer(cdpURL string) *Server {
	cdp := bridge.NewCDPClient(cdpURL)
	exec := bridge.NewColabExecutor(cdp)

	s := &Server{
		cdp:       cdp,
		exec:      exec,
		storage:   storage.NewService(exec),
		kernel:    kernel.NewService(exec),
		telemetry: telemetry.NewService(exec),
		terminal:  terminal.NewService(exec),
		notebook:  notebook.NewService(exec),
		proxy:     proxy.NewService(exec),
	}

	s.mcp = mcp.NewServer(&mcp.Implementation{
		Name:    "golab-enterprise",
		Version: ServerVersion,
	}, nil)

	s.registerTools()
	return s
}

// NewServerWithSessionManager creates an MCP server wired to a dynamic session manager.
func NewServerWithSessionManager(sm *session.Manager) *Server {
	s := &Server{
		sessionMgr: sm,
	}

	s.mcp = mcp.NewServer(&mcp.Implementation{
		Name:    "golab-enterprise",
		Version: ServerVersion,
	}, nil)

	s.registerTools()
	return s
}

func (s *Server) getServices(ctx context.Context, sessionID string) (*session.DomainServices, string, error) {
	if s.sessionMgr != nil {
		act, err := s.sessionMgr.ResolveSession(ctx, sessionID)
		if err != nil {
			return nil, "", err
		}
		return act.Services, act.Session.ID, nil
	}
	return &session.DomainServices{
		Storage:   s.storage,
		Kernel:    s.kernel,
		Telemetry: s.telemetry,
		Notebook:  s.notebook,
		Terminal:  s.terminal,
		Proxy:     s.proxy,
	}, "default", nil
}

func (s *Server) logActivity(sessionID, toolName string, latencyMs int64, err error) {
	if s.sessionMgr != nil && s.sessionMgr.Store() != nil {
		status := "success"
		errMsg := ""
		if err != nil {
			status = "error"
			errMsg = err.Error()
		}
		_ = s.sessionMgr.Store().LogActivity(sessionID, toolName, latencyMs, status, errMsg)
	}
}

// Run starts the MCP server over standard input/output.
func (s *Server) Run(ctx context.Context) error {
	if s.cdp != nil {
		_ = s.cdp.Connect(ctx)
	}
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	// Group 1 — Storage & File Operations (Non-blocking)
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_read_file",
		Description: "Reads the complete contents of a file on the Colab container or mounted Google Drive in 20-50ms. Completely non-blocking, executes via Colab Public Tunnel REST Gateway without touching the Python kernel.",
	}, s.colabReadFile)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_write_file",
		Description: "Creates or overwrites a file atomically in the Colab container or Google Drive. Non-blocking, executes in ~40ms while the model is training without queue delays.",
	}, s.colabWriteFile)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_list_directory",
		Description: "Lists files and subdirectories with sizes, permissions, and timestamps. Use 'content' for VM root or 'content/drive/MyDrive' for Google Drive.",
	}, s.colabListDirectory)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_create_directory",
		Description: "Creates a new directory in the Colab filesystem or mounted Google Drive.",
	}, s.colabCreateDirectory)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_delete_file",
		Description: "Permanently deletes a file or directory on the Colab container.",
	}, s.colabDeleteFile)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_move_file",
		Description: "Renames or moves a file or directory to a new path in the Colab filesystem.",
	}, s.colabMoveFile)

	// Group 2 — Kernel Execution Lifecycle
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_get_kernel_status",
		Description: "Returns real-time status of all Colab kernels: execution state ('idle', 'busy', 'starting'), active notebook session, uptime, and connections.",
	}, s.colabGetKernelStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_interrupt_kernel",
		Description: "Dispatches an out-of-band SIGINT signal to stop a running or looping cell immediately within 1ms, preserving all variables in RAM.",
	}, s.colabInterruptKernel)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_restart_kernel",
		Description: "Restarts the Python kernel, terminating stalled processes and completely freeing GPU VRAM.",
	}, s.colabRestartKernel)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_list_sessions",
		Description: "Lists all active notebook sessions and their kernel bindings.",
	}, s.colabListSessions)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_spawn_secondary_kernel",
		Description: "Spawns a secondary Python kernel on the Colab VM to run parallel jobs on the Tesla GPU.",
	}, s.colabSpawnSecondaryKernel)

	// Group 3 — Hardware & Environment Telemetry
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_get_hardware_telemetry",
		Description: "Retrieves instant hardware telemetry in 2ms: Nvidia Tesla GPU VRAM (allocated, free, total in MB and bytes), GPU/memory utilization percentages, system host RAM, and disk storage. Non-blocking.",
	}, s.colabGetHardwareTelemetry)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_get_build_info",
		Description: "Retrieves Google Colab container image release tags, build dates, and runtime environment metadata.",
	}, s.colabGetBuildInfo)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_check_drive_status",
		Description: "Checks whether Google Drive is currently mounted in this runtime (/content/drive). Returns mounted: true/false with explicit status and mounting instructions. Always check this before reading or writing to Google Drive to prevent file-not-found hallucinations.",
	}, s.colabCheckDriveStatus)

	// Group 4 — Linux PTY Terminals
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_list_terminals",
		Description: "Lists all active Linux PTY terminal sessions running on the Colab container.",
	}, s.colabListTerminals)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_create_terminal",
		Description: "Creates a new background bash PTY terminal on the Colab container.",
	}, s.colabCreateTerminal)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_delete_terminal",
		Description: "Terminates a background bash terminal session.",
	}, s.colabDeleteTerminal)

	// Group 5 — Notebook DOM Precision Editing
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_get_notebook_outline",
		Description: "Returns a compact skeletal outline of the notebook: cell indices, cell types, line counts, and first-line summaries.",
	}, s.colabGetNotebookOutline)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_get_cells",
		Description: "Fetches code and markdown cells from the connected notebook, with optional outputs.",
	}, s.colabGetCells)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_add_cell",
		Description: "Inserts a new code or markdown cell at a specified index in the notebook.",
	}, s.colabAddCell)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_edit_cell_lines",
		Description: "Performs precision line-level editing within a cell (start_line to end_line, 1-based index) avoiding whole-cell replacement.",
	}, s.colabEditCellLines)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_update_cell",
		Description: "Replaces the entire content of an existing cell by its ID.",
	}, s.colabUpdateCell)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_delete_cell",
		Description: "Deletes a cell from the notebook by its ID.",
	}, s.colabDeleteCell)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_move_cell",
		Description: "Moves a cell to a new index position in the notebook.",
	}, s.colabMoveCell)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_run_code_cell",
		Description: "Executes a code cell by its ID and returns its output trace.",
	}, s.colabRunCodeCell)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_get_cell_with_lines",
		Description: "Returns a cell's source code annotated with line numbers for reference.",
	}, s.colabGetCellWithLines)

	// Group 6 — Reverse Proxy Router
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_proxy_forward",
		Description: "Routes an HTTP request directly to any internal port running on the Colab VM via Google's /_proxy/<port>/ gateway (e.g. Gradio, TensorBoard, FastAPI).",
	}, s.colabProxyForward)

	// Group 7 — Multi-Session Worktree Management
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_list_all_sessions",
		Description: "Lists all registered and active Google Colab notebook sessions across all Chrome profiles, along with current GPU VRAM and kernel states.",
	}, s.colabListAllSessions)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_switch_active_session",
		Description: "Switches the active Google Colab notebook context. Subsequent MCP commands without an explicit session_id will route to this session.",
	}, s.colabSwitchActiveSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "colab_create_session",
		Description: "Opens and attaches a new Colab notebook session under a specified Google account profile.",
	}, s.colabCreateSession)
}
