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

	fmt.Printf("Selected Target: %s (URL: %s)\n", selected.Title, selected.URL)
	fmt.Printf("Connecting to WS: %s\n", selected.WebSocketDebuggerURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
					const driveRes = await kernel.getKernelXhr().get('/api/colab/drive');
					let dirExists = false;
					try {
						await kernel.getKernelXhr().get('/api/contents/content/drive');
						dirExists = true;
					} catch (e) {
						dirExists = false;
					}
					return JSON.stringify({
						driveRes: driveRes,
						dirExists: dirExists
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

	fmt.Printf("\nCDP Response:\n%s\n", string(respBytes))
}
