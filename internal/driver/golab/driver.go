package golab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hoangnecon/golab/internal/bridge"
	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/driver/golab/scripts"
)

// Driver implements domain.GolabDriver backed by Chrome DevTools Protocol.
type Driver struct {
	cdp *bridge.CDPClient
}

// NewDriver constructs a GolabDriver wrapping a CDP client.
func NewDriver(cdp *bridge.CDPClient) *Driver {
	return &Driver{cdp: cdp}
}

// CDP returns the underlying CDP client.
func (d *Driver) CDP() *bridge.CDPClient {
	return d.cdp
}

func (d *Driver) evaluateJSON(ctx context.Context, expression string, target any) error {
	val, err := d.cdp.Evaluate(ctx, expression)
	if err != nil {
		return err
	}

	var rawStr string
	if err := json.Unmarshal(val, &rawStr); err == nil {
		val = json.RawMessage(rawStr)
	}

	var errCheck struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(val, &errCheck); err == nil && errCheck.Error != "" {
		return errors.New(errCheck.Error)
	}

	if target != nil {
		if err := json.Unmarshal(val, target); err != nil {
			return fmt.Errorf("unmarshal driver result: %w", err)
		}
	}
	return nil
}

func (d *Driver) FetchKernelXHR(ctx context.Context, path string, opts domain.FetchOptions) (*domain.FetchResponse, error) {
	if opts.Method == "" {
		opts.Method = "GET"
	}
	headersJSON, _ := json.Marshal(opts.Headers)
	bodyJSON, _ := json.Marshal(opts.Body)

	jsExpr := scripts.ScriptFetchKernelXHR(opts.Method, string(headersJSON), string(bodyJSON), path)
	var resp domain.FetchResponse
	if err := d.evaluateJSON(ctx, jsExpr, &resp); err != nil {
		return nil, fmt.Errorf("fetch kernel xhr: %w", err)
	}
	return &resp, nil
}

func (d *Driver) CallMcpTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	toolFieldMap := map[string]string{
		"get_cells":     "getCellsTool",
		"add_code_cell": "addCodeCellTool",
		"add_text_cell": "addTextCellTool",
		"update_cell":   "updateCellTool",
		"delete_cell":   "deleteCellTool",
		"move_cell":     "moveCellTool",
		"run_code_cell": "runCodeCellTool",
	}

	field, ok := toolFieldMap[toolName]
	if !ok {
		return nil, fmt.Errorf("unsupported native tool: %s", toolName)
	}

	jsExpr := scripts.ScriptCallMcpTool(field, string(argsJSON))
	var resObj struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
		Error   string          `json:"error"`
	}
	if err := d.evaluateJSON(ctx, jsExpr, &resObj); err != nil {
		return nil, fmt.Errorf("call mcp tool %s: %w", toolName, err)
	}
	if resObj.Error != "" {
		return nil, errors.New(resObj.Error)
	}
	return resObj.Result, nil
}

func (d *Driver) ConnectRuntime(ctx context.Context) error {
	return d.evaluateJSON(ctx, scripts.ScriptConnectRuntime, nil)
}

func (d *Driver) DisconnectRuntime(ctx context.Context) error {
	return d.evaluateJSON(ctx, scripts.ScriptDisconnectRuntime, nil)
}

func (d *Driver) GetAvailableAccelerators(ctx context.Context) (*domain.RuntimeConfigReport, error) {
	var report domain.RuntimeConfigReport
	if err := d.evaluateJSON(ctx, scripts.ScriptGetAvailableAccelerators, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (d *Driver) SetAccelerator(ctx context.Context, value string, highRAM bool) error {
	return d.evaluateJSON(ctx, scripts.ScriptSetAccelerator(value, highRAM), nil)
}

func (d *Driver) InterruptKernel(ctx context.Context) error {
	return d.evaluateJSON(ctx, scripts.ScriptInterruptKernel, nil)
}

func (d *Driver) GetKernelStatus(ctx context.Context) (map[string]any, error) {
	resp, err := d.FetchKernelXHR(ctx, "/api/sessions", domain.FetchOptions{Method: "GET"})
	if err != nil {
		return nil, err
	}

	var sessions []map[string]any
	if err := json.Unmarshal(resp.Body, &sessions); err != nil {
		return nil, fmt.Errorf("unmarshal sessions: %w", err)
	}

	if len(sessions) == 0 {
		return map[string]any{
			"connected":       false,
			"execution_state": "disconnected",
		}, nil
	}

	first := sessions[0]
	kernelInfo, _ := first["kernel"].(map[string]any)
	notebookInfo, _ := first["notebook"].(map[string]any)

	return map[string]any{
		"connected":       true,
		"session_id":      first["id"],
		"notebook_name":   first["name"],
		"notebook_path":   notebookInfo["path"],
		"kernel_id":       kernelInfo["id"],
		"execution_state": kernelInfo["execution_state"],
		"last_activity":   kernelInfo["last_activity"],
		"connections":     kernelInfo["connections"],
	}, nil
}

func (d *Driver) GetRunningCells(ctx context.Context) (*domain.RunningCellsReport, error) {
	var list []domain.RunningCellInfo
	if err := d.evaluateJSON(ctx, scripts.ScriptGetRunningCells, &list); err != nil {
		return nil, err
	}

	activeCount := 0
	for _, c := range list {
		if c.IsRunning {
			activeCount++
		}
	}
	return &domain.RunningCellsReport{
		HasRunningCells: activeCount > 0,
		TotalRunning:    activeCount,
		RunningCells:    list,
	}, nil
}

func (d *Driver) SaveNotebook(ctx context.Context) error {
	return d.evaluateJSON(ctx, scripts.ScriptSaveNotebook, nil)
}

func (d *Driver) RunAll(ctx context.Context) error {
	return d.evaluateJSON(ctx, scripts.ScriptRunAll, nil)
}
