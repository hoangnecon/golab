package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) colabGetNotebookOutline(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	outline, err := s.notebook.GetNotebookOutline(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(outline, latency, "")
}

// GetCellsInput specifies cell range.
type GetCellsInput struct {
	StartIndex     int  `json:"start_index,omitempty" jsonschema:"Start index (0-based inclusive). Default 0."`
	EndIndex       int  `json:"end_index,omitempty" jsonschema:"End index (exclusive). Default 100."`
	IncludeOutputs bool `json:"include_outputs,omitempty" jsonschema:"Include executed cell outputs. Default false."`
}

func (s *Server) colabGetCells(ctx context.Context, req *mcp.CallToolRequest, input GetCellsInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	cells, err := s.notebook.GetCells(ctx, input.StartIndex, input.EndIndex, input.IncludeOutputs)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(cells, latency, "")
}

// AddCellInput specifies new cell content.
type AddCellInput struct {
	Index    int    `json:"index" jsonschema:"Position index where cell is inserted (0-based). Required."`
	CellType string `json:"cell_type" jsonschema:"Cell type: 'code' or 'markdown'. Required."`
	Content  string `json:"content" jsonschema:"Code or markdown content for the new cell. Required."`
}

func (s *Server) colabAddCell(ctx context.Context, req *mcp.CallToolRequest, input AddCellInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	var cellID string
	var err error
	if input.CellType == "markdown" || input.CellType == "text" {
		cellID, err = s.notebook.AddTextCell(ctx, input.Index, input.Content)
	} else {
		cellID, err = s.notebook.AddCodeCell(ctx, input.Index, input.Content)
	}
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"cell_id": cellID, "index": input.Index, "type": input.CellType}, latency, "")
}

// EditCellLinesInput specifies surgical line replacements.
type EditCellLinesInput struct {
	CellID     string `json:"cell_id" jsonschema:"ID of the cell to edit. Required."`
	StartLine  int    `json:"start_line" jsonschema:"First line to replace (1-based index). Required."`
	EndLine    int    `json:"end_line" jsonschema:"Last line to replace (1-based index inclusive). Required."`
	NewContent string `json:"new_content" jsonschema:"New code/text replacing the line range. Required."`
}

func (s *Server) colabEditCellLines(ctx context.Context, req *mcp.CallToolRequest, input EditCellLinesInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	res, err := s.notebook.EditCellLines(ctx, input.CellID, input.StartLine, input.EndLine, input.NewContent)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(res, latency, "")
}

// UpdateCellInput overwrites full cell.
type UpdateCellInput struct {
	CellID  string `json:"cell_id" jsonschema:"ID of the cell. Required."`
	Content string `json:"content" jsonschema:"Complete replacement content. Required."`
}

func (s *Server) colabUpdateCell(ctx context.Context, req *mcp.CallToolRequest, input UpdateCellInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.notebook.UpdateCell(ctx, input.CellID, input.Content)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"updated": true, "cell_id": input.CellID}, latency, "")
}

// CellTargetInput specifies target cell ID.
type CellTargetInput struct {
	CellID string `json:"cell_id" jsonschema:"Cell ID. Required."`
}

func (s *Server) colabDeleteCell(ctx context.Context, req *mcp.CallToolRequest, input CellTargetInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.notebook.DeleteCell(ctx, input.CellID)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"deleted": true, "cell_id": input.CellID}, latency, "")
}

// MoveCellInput specifies move target.
type MoveCellInput struct {
	CellID   string `json:"cell_id" jsonschema:"Cell ID. Required."`
	NewIndex int    `json:"new_index" jsonschema:"New index position (0-based). Required."`
}

func (s *Server) colabMoveCell(ctx context.Context, req *mcp.CallToolRequest, input MoveCellInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	err := s.notebook.MoveCell(ctx, input.CellID, input.NewIndex)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{"moved": true, "cell_id": input.CellID, "new_index": input.NewIndex}, latency, "")
}

func (s *Server) colabRunCodeCell(ctx context.Context, req *mcp.CallToolRequest, input CellTargetInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	rawOutput, err := s.notebook.RunCodeCell(ctx, input.CellID)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(rawOutput, latency, "")
}

func (s *Server) colabGetCellWithLines(ctx context.Context, req *mcp.CallToolRequest, input CellTargetInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	res, err := s.notebook.GetCellWithLines(ctx, input.CellID)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(res, latency, "")
}
