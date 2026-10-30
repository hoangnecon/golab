package scripts

import "fmt"

// ScriptFetchKernelXHR routes an authenticated HTTP request through Colab's tunnel gateway to the container VM.
func ScriptFetchKernelXHR(method, headersJSON, bodyJSON, path string) string {
	return fmt.Sprintf(`(async () => {
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
	})()`, method, headersJSON, bodyJSON, path)
}
