package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type targetInfo struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func main() {
	resp, err := http.Get("http://127.0.0.1:9222/json/list")
	if err != nil {
		log.Fatalf("Failed to fetch targets from CDP: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var targets []targetInfo
	if err := json.Unmarshal(body, &targets); err != nil {
		log.Fatalf("Unmarshal targets error: %v", err)
	}

	var selected *targetInfo
	for i := range targets {
		t := &targets[i]
		if t.Type == "page" && (t.URL != "" && t.URL[:4] == "http") {
			selected = t
			break
		}
	}

	if selected == nil {
		log.Fatalf("No page target found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, selected.WebSocketDebuggerURL, nil)
	if err != nil {
		log.Fatalf("WebSocket Dial error: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	req := map[string]any{
		"id":     1,
		"method": "Runtime.evaluate",
		"params": map[string]any{
			"expression": `(async () => {
				const nb = window.colab?.global?.notebook;
				const kernel = nb?.getKernel();
				if (!kernel) return JSON.stringify({ error: "No kernel" });

				try {
					// 1. Kiểm tra trạng thái kernel trước khi ghi
					const sessionResp = await kernel.getKernelXhr().fetch('/api/sessions');
					const sessionData = await sessionResp.json();
					const stateBefore = sessionData[0]?.kernel?.execution_state;

					// 2. Ghi 5 file song song / liên tiếp
					const files = [
						{ name: "file_1.txt", content: "Nội dung file 1 - Ghi trong lúc Kernel busy!" },
						{ name: "file_2.txt", content: "Nội dung file 2 - Tốc độ mili-giây!" },
						{ name: "file_3.txt", content: "Nội dung file 3 - Không dùng runHiddenCell!" },
						{ name: "file_4.txt", content: "Nội dung file 4 - Non-blocking I/O chuẩn Enterprise!" },
						{ name: "file_5.txt", content: "Nội dung file 5 - GoLab CDP Engine hoàn hảo!" }
					];

					const tStart = Date.now();
					const results = [];

					for (const f of files) {
						const resp = await kernel.getKernelXhr().fetch('/api/contents/content/' + f.name, {
							method: 'PUT',
							headers: new Headers({ 'Content-Type': 'application/json' }),
							body: JSON.stringify({
								content: f.content,
								type: 'file',
								format: 'text'
							})
						});
						const data = await resp.json();
						results.push({ name: f.name, size: data.size, path: data.path });
					}
					const elapsedMs = Date.now() - tStart;

					// 3. Kiểm tra trạng thái kernel sau khi ghi
					const sessionAfter = await (await kernel.getKernelXhr().fetch('/api/sessions')).json();
					const stateAfter = sessionAfter[0]?.kernel?.execution_state;

					return JSON.stringify({
						success: true,
						kernelStateBefore: stateBefore,
						kernelStateAfter: stateAfter,
						totalFilesWritten: results.length,
						totalElapsedMs: elapsedMs,
						files: results
					});
				} catch (e) {
					return JSON.stringify({ error: e.toString(), stack: e.stack });
				}
			})()`,
			"awaitPromise":  true,
			"returnByValue": true,
		},
	}
	reqBytes, _ := json.Marshal(req)

	if err := conn.Write(ctx, websocket.MessageText, reqBytes); err != nil {
		log.Fatalf("Write error: %v", err)
	}

	_, respBytes, err := conn.Read(ctx)
	if err != nil {
		log.Fatalf("Read error: %v", err)
	}

	fmt.Printf("%s\n", string(respBytes))
}
