package scripts

import "fmt"

// ScriptConnectRuntime triggers connection to the Colab hosted runtime and waits for allocation.
const ScriptConnectRuntime = `(async () => {
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

// ScriptDisconnectRuntime powerwashes and releases the active VM instance.
const ScriptDisconnectRuntime = `(async () => {
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

// ScriptGetAvailableAccelerators reads the hardware accelerator options dynamically from Colab runtime settings.
const ScriptGetAvailableAccelerators = `(async () => {
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

// ScriptSetAccelerator selects a specific accelerator and High-RAM configuration.
func ScriptSetAccelerator(value string, highRAM bool) string {
	return fmt.Sprintf(`(async () => {
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
}

// ScriptInterruptKernel dispatches native SIGINT to the active kernel.
const ScriptInterruptKernel = `(() => {
	const nb = window.colab?.global?.notebook;
	const kernel = nb?.getKernel();
	if (kernel?.interrupt) {
		kernel.interrupt();
		return JSON.stringify({ success: true });
	}
	return JSON.stringify({ error: "kernel.interrupt is not available" });
})()`
