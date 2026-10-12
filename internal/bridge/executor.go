package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ColabExecutor executes operations inside the connected Google Colab browser tab.
type ColabExecutor struct {
	cdp *CDPClient
}

// NewColabExecutor creates an executor backed by a CDPClient.
func NewColabExecutor(cdp *CDPClient) *ColabExecutor {
	return &ColabExecutor{cdp: cdp}
}

// FetchOptions configures an internal Colab REST request.
type FetchOptions struct {
	Method  string
	Headers map[string]string
	Body    string
}

// FetchResponse represents the parsed response from Colab's internal gateway.
type FetchResponse struct {
	Status     int             `json:"status"`
	StatusText string          `json:"statusText"`
	OK         bool            `json:"ok"`
	Body       json.RawMessage `json:"body"`
	RawText    string          `json:"rawText,omitempty"`
}

// FetchKernelXHR invokes Colab's native authenticated tunnel fetch method.
func (e *ColabExecutor) FetchKernelXHR(ctx context.Context, path string, opts FetchOptions) (*FetchResponse, error) {
	if opts.Method == "" {
		opts.Method = "GET"
	}

	headersJSON, _ := json.Marshal(opts.Headers)
	bodyJSON, _ := json.Marshal(opts.Body)

	// We execute JavaScript that calls kernel.getKernelXhr().fetch(...)
	// Note: Google responses like /api/colab/resources start with )]}' prefix. We strip it if present.
	jsExpr := fmt.Sprintf(`(async () => {
		const nb = window.colab?.global?.notebook;
		const kernel = nb?.getKernel();
		if (!kernel) {
			return JSON.stringify({ error: "No active Colab kernel instance found. Please connect runtime." });
		}

		const xhr = kernel.getKernelXhr ? kernel.getKernelXhr() : null;
		if (!xhr) {
			return JSON.stringify({ error: "Kernel XHR tunnel interface is unavailable." });
		}

		const method = %q;
		const rawHeaders = %s || {};
		const headers = new Headers();
		for (const [k, v] of Object.entries(rawHeaders)) {
			headers.set(k, v);
		}

		const fetchOpts = { method, headers };
		const rawBody = %s;
		if (rawBody && method !== "GET" && method !== "HEAD") {
			fetchOpts.body = rawBody;
		}

		try {
			const resp = await xhr.fetch(%q, fetchOpts);
			const text = await resp.text();
			
			// Strip Google XSSI prefix )]}' if present
			let cleanText = text;
			if (cleanText.startsWith(")]}'\n")) {
				cleanText = cleanText.substring(5);
			} else if (cleanText.startsWith(")]}'")) {
				cleanText = cleanText.substring(4);
			}
			cleanText = cleanText.trim();

			let parsedBody = null;
			try {
				parsedBody = cleanText ? JSON.parse(cleanText) : null;
			} catch (_) {
				parsedBody = null;
			}

			return JSON.stringify({
				status: resp.status,
				statusText: resp.statusText,
				ok: resp.ok,
				body: parsedBody,
				rawText: parsedBody === null ? cleanText : ""
			});
		} catch (err) {
			return JSON.stringify({ error: err.toString(), stack: err.stack });
		}
	})()`, opts.Method, string(headersJSON), string(bodyJSON), path)

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return nil, fmt.Errorf("eval fetch: %w", err)
	}

	var rawStr string
	if err := json.Unmarshal(val, &rawStr); err == nil {
		val = json.RawMessage(rawStr)
	}

	var resp FetchResponse
	var errCheck struct {
		Error string `json:"error"`
		Stack string `json:"stack,omitempty"`
	}
	if err := json.Unmarshal(val, &errCheck); err == nil && errCheck.Error != "" {
		return nil, errors.New(errCheck.Error)
	}

	if err := json.Unmarshal(val, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal fetch response (%s): %w", string(val), err)
	}

	return &resp, nil
}

// CallColabMcpTool invokes one of Colab's native frontend MCP tools (get_cells, add_code_cell, etc.).
func (e *ColabExecutor) CallColabMcpTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	toolFieldMap := map[string]string{
		"get_cells":      "getCellsTool",
		"add_code_cell":  "addCodeCellTool",
		"add_text_cell":  "addTextCellTool",
		"update_cell":    "updateCellTool",
		"delete_cell":    "deleteCellTool",
		"move_cell":      "moveCellTool",
		"run_code_cell":  "runCodeCellTool",
	}

	field, ok := toolFieldMap[toolName]
	if !ok {
		return nil, fmt.Errorf("unsupported Colab native tool: %s", toolName)
	}

	jsExpr := fmt.Sprintf(`(async () => {
		const nb = window.colab?.global?.notebook;
		const tools = nb?.colabMcpToolsService;
		if (!tools || !tools[%q]) {
			return JSON.stringify({ error: "Tool " + %q + " not found in colabMcpToolsService" });
		}

		try {
			const res = await tools[%q].toolCallback(%s);
			return JSON.stringify({ success: true, result: res });
		} catch (err) {
			return JSON.stringify({ error: err.toString(), stack: err.stack });
		}
	})()`, field, field, field, string(argsJSON))

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return nil, fmt.Errorf("eval tool %s: %w", toolName, err)
	}

	var rawStr string
	if err := json.Unmarshal(val, &rawStr); err == nil {
		val = json.RawMessage(rawStr)
	}

	var resObj struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(val, &resObj); err != nil {
		return nil, fmt.Errorf("unmarshal tool result (%s): %w", string(val), err)
	}
	if resObj.Error != "" {
		return nil, errors.New(resObj.Error)
	}

	return resObj.Result, nil
}

// InterruptKernel triggers the native SIGINT interrupt signal in Colab kernel.
func (e *ColabExecutor) InterruptKernel(ctx context.Context) error {
	jsExpr := `(() => {
		const nb = window.colab?.global?.notebook;
		const kernel = nb?.getKernel();
		if (kernel?.interrupt) {
			kernel.interrupt();
			return JSON.stringify({ success: true });
		}
		return JSON.stringify({ error: "kernel.interrupt is not available" });
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return err
	}

	var rawStr string
	if err := json.Unmarshal(val, &rawStr); err == nil {
		val = json.RawMessage(rawStr)
	}

	var res struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(val, &res); err == nil && res.Error != "" {
		return errors.New(res.Error)
	}
	return nil
}

// GetKernelStatus fetches active session and kernel metadata.
func (e *ColabExecutor) GetKernelStatus(ctx context.Context) (map[string]any, error) {
	resp, err := e.FetchKernelXHR(ctx, "/api/sessions", FetchOptions{Method: "GET"})
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
