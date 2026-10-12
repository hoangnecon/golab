package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
)

func main() {
	port := 9124
	token := "probetoken"
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming WS request from %s, Origin: %s, URL: %s", r.RemoteAddr, r.Header.Get("Origin"), r.URL.String())
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols:   []string{"mcp"},
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Printf("WS accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "done")
		log.Println(">>> COLAB CHROME CONNECTED! <<<")

		ctx := context.Background()

		// Start reader to dump everything Chrome sends
		go func() {
			for {
				_, data, err := conn.Read(ctx)
				if err != nil {
					log.Printf("Read finished/error: %v", err)
					return
				}
				log.Printf("<<< FROM CHROME: %s", string(data))
			}
		}()

		time.Sleep(1 * time.Second)

		// Test 1: tools/list
		log.Println(">>> SENDING: tools/list")
		conn.Write(ctx, websocket.MessageText, []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))

		time.Sleep(2 * time.Second)

		// Test 2: tools/call get_proxy_token
		log.Println(">>> SENDING: tools/call get_runtime_proxy_token")
		conn.Write(ctx, websocket.MessageText, []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_runtime_proxy_token","arguments":{}}}`))

		time.Sleep(2 * time.Second)

		// Test 3: raw custom method getRuntimeProxyToken
		log.Println(">>> SENDING: custom method getRuntimeProxyToken")
		conn.Write(ctx, websocket.MessageText, []byte(`{"jsonrpc":"2.0","id":3,"method":"getRuntimeProxyToken","params":{}}`))

		time.Sleep(2 * time.Second)

		// Test 4: official get_cells
		log.Println(">>> SENDING: official get_cells")
		conn.Write(ctx, websocket.MessageText, []byte(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_cells","arguments":{"cellIndexStart":0,"cellIndexEnd":1,"includeOutputs":false}}}`))

		time.Sleep(5 * time.Second)
		log.Println(">>> PROBE FINISHED <<<")
		os.Exit(0)
	})

	fmt.Printf("\n=======================================================\n")
	fmt.Printf("PROBE SERVER LISTENING ON ws://127.0.0.1:%d\n", port)
	fmt.Printf("Open this URL in Chrome:\n")
	fmt.Printf("https://colab.research.google.com/drive/1cQQ-Ze9PZaWzDYJnl_1DINNsKEKkikkQ#mcpProxyToken=%s&mcpProxyPort=%d\n", token, port)
	fmt.Printf("=======================================================\n\n")

	if err := http.Serve(listener, mux); err != nil {
		log.Fatal(err)
	}
}
