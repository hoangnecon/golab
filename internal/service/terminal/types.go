package terminal

// TerminalInfo represents an active Jupyter terminal instance.
type TerminalInfo struct {
	Name         string `json:"name"`
	LastActivity string `json:"last_activity"`
}

// CommandResult contains standard output and execution diagnostics.
type CommandResult struct {
	Command   string `json:"command"`
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code"`
	ElapsedMs int64  `json:"elapsed_ms"`
}
