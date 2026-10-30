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

// ConnectRuntime automates connection to the Colab hosted runtime.
func (e *ColabExecutor) ConnectRuntime(ctx context.Context) error {
	jsExpr := `(async () => {
		const nb = window.colab?.global?.notebook;
		const btn = document.querySelector("colab-connect-button");
		if (btn?.connectState === "connected") {
			return JSON.stringify({ success: true, message: "already_connected" });
		}

		try {
			nb.commandHandler.execute({id: "connect", isDisabled: () => false});
		} catch (e) {
			const inner = btn?.shadowRoot?.querySelector("#connect");
			if (inner) inner.click();
			else return JSON.stringify({ error: e.toString() });
		}

		const t0 = Date.now();
		while (Date.now() - t0 < 35000) {
			await new Promise(r => setTimeout(r, 1000));
			if (btn?.connectState === "connected") {
				return JSON.stringify({ success: true, connectState: btn.connectState });
			}
		}
		return JSON.stringify({ error: "Connect timed out after 35 seconds", connectState: btn?.connectState });
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return fmt.Errorf("eval connect runtime: %w", err)
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

// DisconnectRuntime terminates and powerwashes the active Colab VM instance.
func (e *ColabExecutor) DisconnectRuntime(ctx context.Context) error {
	jsExpr := `(async () => {
		const nb = window.colab?.global?.notebook;
		const kernel = nb?.getKernel ? nb.getKernel() : null;
		const btn = document.querySelector("colab-connect-button");

		if (btn?.connectState === "disconnected") {
			return JSON.stringify({ success: true, message: "already_disconnected" });
		}

		try {
			nb.commandHandler.execute({id: "powerwash-current-vm", isDisabled: () => false});
		} catch (e) {
			if (kernel?.unassignCurrentVm) {
				await kernel.unassignCurrentVm();
			} else {
				return JSON.stringify({ error: e.toString() });
			}
		}

		const t0 = Date.now();
		while (Date.now() - t0 < 3000) {
			await new Promise(r => setTimeout(r, 200));
			const mwc = document.querySelector("mwc-dialog[open]");
			if (mwc) {
				const yesBtn = mwc.querySelector("[slot='primaryAction']") || Array.from(mwc.querySelectorAll("button, md-text-button")).find(b => b.innerText?.trim() === "Yes");
				if (yesBtn) {
					yesBtn.click();
					break;
				}
			}
		}

		const t1 = Date.now();
		while (Date.now() - t1 < 5000) {
			await new Promise(r => setTimeout(r, 500));
			if (btn?.connectState === "disconnected") {
				return JSON.stringify({ success: true });
			}
		}
		return JSON.stringify({ success: true });
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return fmt.Errorf("eval disconnect runtime: %w", err)
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

// GetAvailableAccelerators reads the hardware accelerator options from Colab runtime settings.
func (e *ColabExecutor) GetAvailableAccelerators(ctx context.Context) (json.RawMessage, error) {
	jsExpr := `(async () => {
		const nb = window.colab?.global?.notebook;
		try {
			nb.commandHandler.execute({id: "change-runtime-type", isDisabled: () => false});
		} catch (e) {
			return JSON.stringify({ error: e.toString() });
		}

		await new Promise(r => setTimeout(r, 600));

		const mwc = document.querySelector("mwc-dialog[open]");
		if (!mwc) return JSON.stringify({ error: "Dialog failed to open" });

		const accelerators = [];
		const walk = (node) => {
			if (!node) return;
			if (node.nodeType === 1) {
				const tag = node.tagName.toLowerCase();
				if (tag === "input" && (node.name === "accelerator" || node.getAttribute("name") === "accelerator")) {
					accelerators.push({
						name: node.getAttribute("aria-label") || node.value || "CPU",
						value: node.value,
						checked: node.checked,
						disabled: node.disabled || node.hasAttribute("disabled")
					});
				}
				if (node.shadowRoot) walk(node.shadowRoot);
			}
			for (const child of node.childNodes || []) {
				walk(child);
			}
		};
		walk(mwc);

		const highRamInput = mwc.querySelector("input#switch") || mwc.querySelector("input[aria-label='High-RAM']");
		const highRamChecked = highRamInput ? highRamInput.checked : false;

		const cancelBtn = mwc.querySelector("[slot='secondaryAction']");
		if (cancelBtn) cancelBtn.click();
		else mwc.open = false;

		return JSON.stringify({
			accelerators: accelerators,
			high_ram: highRamChecked
		});
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return nil, fmt.Errorf("eval get accelerators: %w", err)
	}

	var rawStr string
	if err := json.Unmarshal(val, &rawStr); err == nil {
		val = json.RawMessage(rawStr)
	}

	var errCheck struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(val, &errCheck); err == nil && errCheck.Error != "" {
		return nil, errors.New(errCheck.Error)
	}
	return val, nil
}

// SetAccelerator selects a specific hardware accelerator and high-RAM configuration.
func (e *ColabExecutor) SetAccelerator(ctx context.Context, value string, highRAM bool) error {
	jsExpr := fmt.Sprintf(`(async () => {
		const nb = window.colab?.global?.notebook;
		try {
			nb.commandHandler.execute({id: "change-runtime-type", isDisabled: () => false});
		} catch (e) {
			return JSON.stringify({ error: e.toString() });
		}

		await new Promise(r => setTimeout(r, 600));

		const mwc = document.querySelector("mwc-dialog[open]");
		if (!mwc) return JSON.stringify({ error: "Dialog failed to open" });

		const targetVal = %q;
		let found = false;
		const walk = (node) => {
			if (!node) return;
			if (node.nodeType === 1) {
				const tag = node.tagName.toLowerCase();
				if (tag === "input" && (node.name === "accelerator" || node.getAttribute("name") === "accelerator")) {
					if (node.value === targetVal) {
						if (node.disabled || node.hasAttribute("disabled")) {
							throw new Error("Accelerator " + targetVal + " is disabled for this account plan.");
						}
						node.checked = true;
						node.dispatchEvent(new Event("change", { bubbles: true }));
						node.click();
						found = true;
					}
				}
				if (node.shadowRoot) walk(node.shadowRoot);
			}
			for (const child of node.childNodes || []) {
				walk(child);
			}
		};
		try {
			walk(mwc);
		} catch (err) {
			const cancel = mwc.querySelector("[slot='secondaryAction']");
			if (cancel) cancel.click();
			return JSON.stringify({ error: err.message });
		}

		if (!found) {
			const cancel = mwc.querySelector("[slot='secondaryAction']");
			if (cancel) cancel.click();
			return JSON.stringify({ error: "Accelerator value not found: " + targetVal });
		}

		const highRamInput = mwc.querySelector("input#switch") || mwc.querySelector("input[aria-label='High-RAM']");
		if (highRamInput && %t !== highRamInput.checked) {
			highRamInput.click();
		}

		const saveBtn = mwc.querySelector("[slot='primaryAction']");
		if (saveBtn) {
			saveBtn.click();
		}

		return JSON.stringify({ success: true, selected: targetVal });
	})()`, value, highRAM)

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return fmt.Errorf("eval set accelerator: %w", err)
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

// RunAll triggers execution of all cells in the notebook.
func (e *ColabExecutor) RunAll(ctx context.Context) error {
	jsExpr := `(() => {
		const nb = window.colab?.global?.notebook;
		try {
			nb.commandHandler.execute({id: "runall", isDisabled: () => false});
			return JSON.stringify({ success: true });
		} catch (e) {
			if (nb?.runAll) {
				nb.runAll();
				return JSON.stringify({ success: true });
			}
			return JSON.stringify({ error: e.toString() });
		}
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return fmt.Errorf("eval run all: %w", err)
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

// GetRunningCells inspects all cells and returns live execution status and streamed outputs.
func (e *ColabExecutor) GetRunningCells(ctx context.Context) (json.RawMessage, error) {
	jsExpr := `(() => {
		const nb = window.colab?.global?.notebook;
		const cells = nb ? nb.getCells() : [];

		const running = [];
		for (let idx = 0; idx < cells.length; idx++) {
			const c = cells[idx];
			const isRunning = c.isRunning ? c.isRunning() : !!c.busy;
			const cellId = c.cellId || (c.getCellId ? c.getCellId() : "");

			let stdout = "";
			const model = c.outputArea?.outputs;
			if (model?.outputs) {
				for (const outObj of Object.values(model.outputs)) {
					if (outObj.outputItems) {
						for (const item of Object.values(outObj.outputItems)) {
							if (item.text) {
								if (Array.isArray(item.text)) stdout += item.text.join("");
								else stdout += item.text;
							}
						}
					}
				}
			}

			let elapsedSec = 0;
			if (c.executionStateChangeTime) {
				elapsedSec = Math.round((Date.now() - c.executionStateChangeTime) / 100) / 10;
			}

			running.push({
				index: idx,
				cell_id: cellId,
				is_running: isRunning,
				elapsed_seconds: elapsedSec,
				execution_count: c.executionStatusText || String(c.getExecutionCount ? c.getExecutionCount() : ""),
				stdout: stdout.trim()
			});
		}

		return JSON.stringify(running);
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return nil, fmt.Errorf("eval get running cells: %w", err)
	}

	var rawStr string
	if err := json.Unmarshal(val, &rawStr); err == nil {
		val = json.RawMessage(rawStr)
	}
	return val, nil
}

// SaveNotebook commits all notebook modifications and cell outputs to Google Drive (equivalent to Ctrl+S).
func (e *ColabExecutor) SaveNotebook(ctx context.Context) error {
	jsExpr := `(async () => {
		const nb = window.colab?.global?.notebook;
		try {
			if (nb?.saveNotebook) {
				await nb.saveNotebook();
				return JSON.stringify({ success: true });
			}
			nb.commandHandler.execute({ id: "save", isDisabled: () => false });
			return JSON.stringify({ success: true });
		} catch (e) {
			return JSON.stringify({ error: e.toString() });
		}
	})()`

	val, err := e.cdp.Evaluate(ctx, jsExpr)
	if err != nil {
		return fmt.Errorf("eval save notebook: %w", err)
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
