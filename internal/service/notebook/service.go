package notebook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hoangnecon/golab/internal/bridge"
)

// Service provides rich, surgical control over notebook cells and execution.
type Service struct {
	exec *bridge.ColabExecutor
}

// NewService creates a notebook service.
func NewService(exec *bridge.ColabExecutor) *Service {
	return &Service{exec: exec}
}

// GetCells fetches a range of cells with optional execution outputs.
func (s *Service) GetCells(ctx context.Context, start, end int, includeOutputs bool) ([]Cell, error) {
	if end <= 0 {
		end = 500
	}

	raw, err := s.exec.CallColabMcpTool(ctx, "get_cells", map[string]any{
		"cellIndexStart": start,
		"cellIndexEnd":   end,
		"includeOutputs": includeOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("get cells: %w", err)
	}

	var res struct {
		Cells []Cell `json:"cells"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("unmarshal cells: %w", err)
	}
	return res.Cells, nil
}

// AddCodeCell inserts a new Python code cell at cellIndex.
func (s *Service) AddCodeCell(ctx context.Context, index int, code string) (string, error) {
	raw, err := s.exec.CallColabMcpTool(ctx, "add_code_cell", map[string]any{
		"cellIndex": index,
		"language":  "python",
		"code":      code,
	})
	if err != nil {
		return "", fmt.Errorf("add code cell: %w", err)
	}

	var res struct {
		NewCellID string `json:"newCellId"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", fmt.Errorf("unmarshal add cell response: %w", err)
	}
	return res.NewCellID, nil
}

// AddTextCell inserts a new Markdown text cell at cellIndex.
func (s *Service) AddTextCell(ctx context.Context, index int, content string) (string, error) {
	raw, err := s.exec.CallColabMcpTool(ctx, "add_text_cell", map[string]any{
		"cellIndex": index,
		"content":   content,
	})
	if err != nil {
		return "", fmt.Errorf("add text cell: %w", err)
	}

	var res struct {
		NewCellID string `json:"newCellId"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", fmt.Errorf("unmarshal add text cell response: %w", err)
	}
	return res.NewCellID, nil
}

// UpdateCell replaces the complete content of a cell.
func (s *Service) UpdateCell(ctx context.Context, cellID, content string) error {
	if cellID == "" {
		return errors.New("cellId cannot be empty")
	}
	_, err := s.exec.CallColabMcpTool(ctx, "update_cell", map[string]any{
		"cellId":  cellID,
		"content": content,
	})
	if err != nil {
		return fmt.Errorf("update cell %s: %w", cellID, err)
	}
	return nil
}

// DeleteCell removes a cell by its ID.
func (s *Service) DeleteCell(ctx context.Context, cellID string) error {
	if cellID == "" {
		return errors.New("cellId cannot be empty")
	}
	_, err := s.exec.CallColabMcpTool(ctx, "delete_cell", map[string]any{
		"cellId": cellID,
	})
	if err != nil {
		return fmt.Errorf("delete cell %s: %w", cellID, err)
	}
	return nil
}

// MoveCell moves a cell to a new position.
func (s *Service) MoveCell(ctx context.Context, cellID string, newIndex int) error {
	if cellID == "" {
		return errors.New("cellId cannot be empty")
	}
	_, err := s.exec.CallColabMcpTool(ctx, "move_cell", map[string]any{
		"cellId":    cellID,
		"cellIndex": newIndex,
	})
	if err != nil {
		return fmt.Errorf("move cell %s: %w", cellID, err)
	}
	return nil
}

// RunCodeCell triggers execution of a code cell and returns output when finished.
func (s *Service) RunCodeCell(ctx context.Context, cellID string) (json.RawMessage, error) {
	if cellID == "" {
		return nil, errors.New("cellId cannot be empty")
	}
	raw, err := s.exec.CallColabMcpTool(ctx, "run_code_cell", map[string]any{
		"cellId": cellID,
	})
	if err != nil {
		return nil, fmt.Errorf("run cell %s: %w", cellID, err)
	}
	return raw, nil
}

// EditCellLines replaces lines in a specific range [startLine, endLine] (1-based index).
func (s *Service) EditCellLines(ctx context.Context, cellID string, startLine, endLine int, newContent string) (*EditLinesResult, error) {
	cells, err := s.GetCells(ctx, 0, 500, false)
	if err != nil {
		return nil, err
	}

	var target *Cell
	for i := range cells {
		if cells[i].ID == cellID {
			target = &cells[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("cell not found: %s", cellID)
	}

	fullSource := strings.Join(target.Source, "")
	lines := strings.Split(fullSource, "\n")
	totalLines := len(lines)

	if startLine < 1 || endLine > totalLines || startLine > endLine {
		return nil, fmt.Errorf("invalid line range %d-%d (cell has %d lines)", startLine, endLine, totalLines)
	}

	replacementLines := strings.Split(newContent, "\n")
	newLines := make([]string, 0, totalLines-(endLine-startLine+1)+len(replacementLines))
	newLines = append(newLines, lines[:startLine-1]...)
	newLines = append(newLines, replacementLines...)
	newLines = append(newLines, lines[endLine:]...)

	newSource := strings.Join(newLines, "\n")
	if err := s.UpdateCell(ctx, cellID, newSource); err != nil {
		return nil, err
	}

	return &EditLinesResult{
		CellID:       cellID,
		Modified:     true,
		LinesChanged: endLine - startLine + 1,
		TotalLines:   len(newLines),
	}, nil
}

// GetCellWithLines returns cell content with 1-based line numbers.
func (s *Service) GetCellWithLines(ctx context.Context, cellID string) (map[string]any, error) {
	cells, err := s.GetCells(ctx, 0, 500, true)
	if err != nil {
		return nil, err
	}

	for _, c := range cells {
		if c.ID == cellID {
			source := strings.Join(c.Source, "")
			lines := strings.Split(source, "\n")
			var sb strings.Builder
			for i, line := range lines {
				sb.WriteString(fmt.Sprintf("%3d | %s\n", i+1, line))
			}
			return map[string]any{
				"cell_id":         c.ID,
				"cell_type":       c.CellType,
				"total_lines":     len(lines),
				"numbered_source": sb.String(),
				"raw_source":      source,
				"outputs":         c.Outputs,
			}, nil
		}
	}
	return nil, fmt.Errorf("cell not found: %s", cellID)
}

// GetNotebookOutline returns a compact structural overview of the entire notebook.
func (s *Service) GetNotebookOutline(ctx context.Context) (*NotebookOutline, error) {
	cells, err := s.GetCells(ctx, 0, 500, false)
	if err != nil {
		return nil, err
	}

	outline := make([]OutlineItem, 0, len(cells))
	codeCount := 0
	mdCount := 0

	for i, c := range cells {
		source := strings.Join(c.Source, "")
		lines := strings.Split(source, "\n")
		summary := ""
		if len(lines) > 0 {
			summary = strings.TrimSpace(lines[0])
		}

		if c.CellType == "code" {
			codeCount++
		} else {
			mdCount++
		}

		outline = append(outline, OutlineItem{
			Index:     i,
			CellID:    c.ID,
			CellType:  c.CellType,
			Summary:   summary,
			LineCount: len(lines),
		})
	}

	return &NotebookOutline{
		TotalCells:    len(cells),
		CodeCells:     codeCount,
		MarkdownCells: mdCount,
		Outline:       outline,
	}, nil
}
