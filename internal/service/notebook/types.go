package notebook

// Cell represents a single notebook cell.
type Cell struct {
	ID             string           `json:"id"`
	CellType       string           `json:"cell_type"` // "code" or "markdown"
	Source         []string         `json:"source"`
	ExecutionCount *int             `json:"execution_count,omitempty"`
	Outputs        []map[string]any `json:"outputs,omitempty"`
}

// OutlineItem represents a concise overview of a single cell.
type OutlineItem struct {
	Index     int    `json:"index"`
	CellID    string `json:"cell_id"`
	CellType  string `json:"cell_type"`
	Summary   string `json:"summary"`
	LineCount int    `json:"line_count"`
}

// NotebookOutline contains structural metadata of the notebook.
type NotebookOutline struct {
	TotalCells    int           `json:"total_cells"`
	CodeCells     int           `json:"code_cells"`
	MarkdownCells int           `json:"markdown_cells"`
	Outline       []OutlineItem `json:"outline"`
}

// EditLinesResult reports the outcome of a line-level edit.
type EditLinesResult struct {
	CellID       string `json:"cell_id"`
	Modified     bool   `json:"modified"`
	LinesChanged int    `json:"lines_changed"`
	TotalLines   int    `json:"total_lines"`
}
