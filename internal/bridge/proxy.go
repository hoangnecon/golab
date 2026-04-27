package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BrowserProxy sends JSON-RPC requests to the Colab browser and waits for responses.
type BrowserProxy struct {
	ws       *WSServer
	nextID   atomic.Int64
	pending  sync.Map // map[int64]chan *JSONRPCResponse
}

func NewBrowserProxy(ws *WSServer) *BrowserProxy {
	p := &BrowserProxy{ws: ws}
	// Start response dispatcher
	go p.dispatchResponses()
	return p
}

func (p *BrowserProxy) dispatchResponses() {
	for msg := range p.ws.FromBrowser {
		var resp JSONRPCResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			continue
		}
		if ch, ok := p.pending.LoadAndDelete(resp.ID); ok {
			ch.(chan *JSONRPCResponse) <- &resp
		}
	}
}

// CallTool sends a tools/call request to the browser and returns the result.
func (p *BrowserProxy) CallTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	if !p.ws.IsConnected() {
		return nil, fmt.Errorf("browser not connected")
	}

	id := p.nextID.Add(1)
	params := map[string]any{
		"name":      toolName,
		"arguments": args,
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params:  paramsJSON,
	}
	reqJSON, _ := json.Marshal(req)

	// Register pending response
	ch := make(chan *JSONRPCResponse, 1)
	p.pending.Store(id, ch)
	defer p.pending.Delete(id)

	// Send to browser
	p.ws.SendToBrowser(reqJSON)

	// Wait for response
	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("browser error: %s", resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(120 * time.Second):
		return nil, fmt.Errorf("timeout waiting for browser response")
	}
}
