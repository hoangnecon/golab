package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"nhooyr.io/websocket"
)

// TargetInfo represents a browser tab or devtools target returned by /json/list.
type TargetInfo struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// CDPClient manages connection to Chrome DevTools Protocol.
type CDPClient struct {
	cdpURL    string
	mu        sync.Mutex
	conn      *websocket.Conn
	target    *TargetInfo
	nextID    atomic.Int64
	pending   sync.Map // map[int64]chan *cdpResponse
	closed    chan struct{}
	connected bool
}

type cdpRequest struct {
	ID     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type cdpResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *cdpError       `json:"error,omitempty"`
}

type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// NewCDPClient creates a new CDP client pointing to the given CDP HTTP base URL (e.g. "http://127.0.0.1:9222").
func NewCDPClient(cdpURL string) *CDPClient {
	if cdpURL == "" {
		cdpURL = "http://127.0.0.1:9222"
	}
	return &CDPClient{
		cdpURL: strings.TrimRight(cdpURL, "/"),
		closed: make(chan struct{}),
	}
}

// Connect locates an active Colab tab and establishes a WebSocket connection.
func (c *CDPClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected && c.conn != nil {
		return nil
	}

	target, err := c.findColabTarget(ctx)
	if err != nil {
		return fmt.Errorf("find colab target: %w", err)
	}
	if target.WebSocketDebuggerURL == "" {
		return errors.New("colab target has empty webSocketDebuggerUrl")
	}

	conn, _, err := websocket.Dial(ctx, target.WebSocketDebuggerURL, nil)
	if err != nil {
		return fmt.Errorf("dial cdp websocket: %w", err)
	}
	conn.SetReadLimit(32 * 1024 * 1024) // 32MB max message

	c.conn = conn
	c.target = target
	c.connected = true
	c.closed = make(chan struct{})

	go c.readLoop(conn, c.closed)

	log.Printf("[CDP] Connected to Colab tab: %q (%s)", target.Title, target.URL)
	return nil
}

func (c *CDPClient) readLoop(conn *websocket.Conn, closed chan struct{}) {
	defer func() {
		c.mu.Lock()
		if c.conn == conn {
			c.connected = false
			c.conn = nil
		}
		c.mu.Unlock()
		close(closed)
	}()

	for {
		_, data, err := conn.Read(context.Background())
		if err != nil {
			return
		}

		var resp cdpResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			continue
		}

		if ch, ok := c.pending.LoadAndDelete(resp.ID); ok {
			ch.(chan *cdpResponse) <- &resp
		}
	}
}

func (c *CDPClient) findColabTarget(ctx context.Context) (*TargetInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.cdpURL+"/json/list", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query /json/list: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var targets []TargetInfo
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, err
	}

	// 1. Look for a page target with colab in the URL
	for i := range targets {
		t := &targets[i]
		if t.Type == "page" && strings.Contains(t.URL, "colab.research.google.com") {
			return t, nil
		}
	}

	// 2. Fallback to any http page target if no explicit colab url is found yet
	for i := range targets {
		t := &targets[i]
		if t.Type == "page" && strings.HasPrefix(t.URL, "http") {
			return t, nil
		}
	}

	return nil, errors.New("no active colab page target found in chrome")
}

// Evaluate runs a JavaScript expression in the Colab page and returns the parsed result.
func (c *CDPClient) Evaluate(ctx context.Context, expression string) (json.RawMessage, error) {
	if err := c.ensureConnected(ctx); err != nil {
		return nil, err
	}

	id := c.nextID.Add(1)
	req := cdpRequest{
		ID:     id,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    expression,
			"awaitPromise":  true,
			"returnByValue": true,
		},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *cdpResponse, 1)
	c.pending.Store(id, ch)
	defer c.pending.Delete(id)

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, errors.New("cdp connection closed")
	}

	if err := conn.Write(ctx, websocket.MessageText, reqBytes); err != nil {
		return nil, fmt.Errorf("send cdp request: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("cdp error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		var evalResult struct {
			Result struct {
				Type        string          `json:"type"`
				Value       json.RawMessage `json:"value"`
				Description string          `json:"description,omitempty"`
			} `json:"result"`
			ExceptionDetails *struct {
				Text      string `json:"text"`
				Exception *struct {
					Description string `json:"description"`
				} `json:"exception"`
			} `json:"exceptionDetails,omitempty"`
		}

		if err := json.Unmarshal(resp.Result, &evalResult); err != nil {
			return nil, fmt.Errorf("unmarshal cdp evaluate result: %w", err)
		}

		if evalResult.ExceptionDetails != nil {
			desc := evalResult.ExceptionDetails.Text
			if evalResult.ExceptionDetails.Exception != nil {
				desc = evalResult.ExceptionDetails.Exception.Description
			}
			return nil, fmt.Errorf("js exception: %s", desc)
		}

		return evalResult.Result.Value, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *CDPClient) ensureConnected(ctx context.Context) error {
	c.mu.Lock()
	connected := c.connected && c.conn != nil
	c.mu.Unlock()

	if !connected {
		return c.Connect(ctx)
	}
	return nil
}

// Close terminates the CDP client.
func (c *CDPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "client closing")
		c.conn = nil
	}
	c.connected = false
	return nil
}
