package scripts

import "fmt"

// ScriptSaveNotebook commits all notebook modifications and cell outputs to Google Drive (Ctrl+S equivalent).
const ScriptSaveNotebook = `(async () => {
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

// ScriptRunAll executes all cells in the notebook sequentially.
const ScriptRunAll = `(() => {
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

// ScriptGetRunningCells inspects all cells and extracts active execution status and live streaming stdout.
const ScriptGetRunningCells = `(() => {
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

// ScriptCallMcpTool constructs JavaScript to execute one of Colab's native tools.
func ScriptCallMcpTool(field string, argsJSON string) string {
	return fmt.Sprintf(`(async () => {
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
	})()`, field, field, field, argsJSON)
}
