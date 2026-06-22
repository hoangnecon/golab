package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var allowedOrigins = []string{
	"https://colab.research.google.com",
	"https://colab.google.com",
}

// ConnectionStatus holds metadata about the current browser connection.
type ConnectionStatus struct {
	Connected   bool      `json:"connected"`
	ConnectedAt time.Time `json:"connectedAt,omitempty"`
	RemoteAddr  string    `json:"remoteAddr,omitempty"`
	Uptime      string    `json:"uptime,omitempty"`
	WSPort      int       `json:"wsPort"`
}

// WSServer accepts a single WebSocket connection from a Colab browser tab.
type WSServer struct {
	token      string
	mu         sync.Mutex
	conn       *websocket.Conn
	connCancel context.CancelFunc // cancel function for the active connection

	connectedAt time.Time
	remoteAddr  string
	connected   chan struct{} // closed when a browser connects
	server      *http.Server
	listener    net.Listener
	port        int

	FromBrowser chan json.RawMessage
	ToBrowser   chan json.RawMessage
}

func NewWSServer(token string) *WSServer {
	return &WSServer{
		token:       token,
		connected:   make(chan struct{}),
		FromBrowser: make(chan json.RawMessage, 64),
		ToBrowser:   make(chan json.RawMessage, 64),
	}
}

// Start begins listening on the specified port (0 = random). Returns the actual port number.
func (s *WSServer) Start(ctx context.Context, port int) (int, error) {
	var err error
	// Use explicit 127.0.0.1 — on Windows, "localhost" can resolve to ::1 (IPv6)
	// while browser extensions connect via 127.0.0.1 (IPv4).
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	if port != 0 {
		// Configured port: retry a few times in case a previous GoLab is still
		// shutting down. Do NOT fall back to a random port — that causes the
		// browser to connect to the wrong port.
		for i := 0; i < 5; i++ {
			s.listener, err = net.Listen("tcp4", addr)
			if err == nil {
				break
			}
			log.Printf("Port %d in use, retrying in 500ms (%d/5)...", port, i+1)
			time.Sleep(500 * time.Millisecond)
		}
	} else {
		s.listener, err = net.Listen("tcp4", "127.0.0.1:0")
	}
	if err != nil {
		return 0, fmt.Errorf("port %d is still in use after retries — kill the old GoLab process: %w", port, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWS)
	s.server = &http.Server{Handler: mux}

	go func() {
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			log.Printf("WS server error: %v", err)
		}
	}()

	s.port = s.listener.Addr().(*net.TCPAddr).Port
	return s.port, nil
}

func (s *WSServer) Stop() {
	if s.server != nil {
		s.server.Close()
	}
}

// DisconnectAndRotateToken closes the current connection and updates the auth token.
func (s *WSServer) DisconnectAndRotateToken(newToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if newToken != "" {
		s.token = newToken
	}
	s.closeActiveLocked()
}

// IsConnected returns true if a browser is connected.
func (s *WSServer) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn != nil
}

// Status returns connection metadata.
func (s *WSServer) Status() ConnectionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := ConnectionStatus{
		Connected: s.conn != nil,
		WSPort:    s.port,
	}
	if s.conn != nil {
		st.ConnectedAt = s.connectedAt
		st.RemoteAddr = s.remoteAddr
		st.Uptime = time.Since(s.connectedAt).Round(time.Second).String()
	}
	return st
}

// WaitConnected returns a channel that closes when a browser connects.
func (s *WSServer) WaitConnected() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

// closeActiveLocked force-closes the active connection. Must hold s.mu.
func (s *WSServer) closeActiveLocked() {
	if s.connCancel != nil {
		s.connCancel()
		s.connCancel = nil
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.connectedAt = time.Time{}
	s.remoteAddr = ""
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
	Subprotocols: []string{"mcp"},
}

func (s *WSServer) handleWS(w http.ResponseWriter, r *http.Request) {
	log.Printf("[WS] incoming request from=%s origin=%q proto=%q query=%s",
		r.RemoteAddr, r.Header.Get("Origin"), r.Header.Get("Sec-WebSocket-Protocol"), r.URL.RawQuery)

	if !s.validateAuth(r) {
		log.Printf("[WS] AUTH FAILED from=%s", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP to WebSocket — gorilla/websocket does NOT depend on r.Context()
	// after upgrade, so this works reliably on Windows.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(10 * 1024 * 1024) // 10MB

	// Create a per-connection context for clean shutdown.
	ctx, cancel := context.WithCancel(context.Background())

	// Replace any existing connection — force-close it cleanly.
	s.mu.Lock()
	if s.conn != nil {
		log.Println("Force-closing stale browser connection")
		s.closeActiveLocked()
	}

	// Drain stale messages from previous sessions. Without this, the writer
	// goroutine sends old messages to the new browser, which causes Colab
	// to close the connection immediately (wrong message IDs / unexpected data).
	drainCount := 0
	for {
		select {
		case <-s.ToBrowser:
			drainCount++
		default:
			goto drained
		}
	}
drained:
	for {
		select {
		case <-s.FromBrowser:
			drainCount++
		default:
			goto ready
		}
	}
ready:
	if drainCount > 0 {
		log.Printf("Drained %d stale messages from channels", drainCount)
	}

	s.conn = conn
	s.connCancel = cancel
	s.connectedAt = time.Now()
	s.remoteAddr = r.RemoteAddr

	// Signal WaitConnected() callers.
	ch := s.connected
	s.connected = make(chan struct{})
	s.mu.Unlock()
	close(ch)

	log.Println("Colab browser connected")

	// --- Cleanup on exit ---
	defer func() {
		cancel()
		s.mu.Lock()
		if s.conn == conn { // Only clear if we're still the active connection
			s.conn = nil
			s.connCancel = nil
			s.connectedAt = time.Time{}
			s.remoteAddr = ""
		}
		s.mu.Unlock()
		conn.Close()
		log.Println("Colab browser disconnected")
	}()

	// --- Ping/Pong keepalive ---
	// Detect dead connections within 45s (ping every 30s, pong deadline 45s).
	conn.SetReadDeadline(time.Now().Add(45 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(45 * time.Second))
		return nil
	})

	var wg sync.WaitGroup

	// Writer goroutine: sends messages to browser + ping keepalive.
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case msg := <-s.ToBrowser:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Printf("[WS] write error: %v", err)
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("[WS] ping error: %v", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Reader goroutine: reads messages from browser.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel() // When read fails (browser closed), cancel ctx to stop writer.
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[WS] read error: %v", err)
				return
			}
			log.Printf("[WS] recv type=%d len=%d", msgType, len(data))
			select {
			case s.FromBrowser <- json.RawMessage(data):
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

func (s *WSServer) validateAuth(r *http.Request) bool {
	if strings.Contains(r.URL.RawQuery, "access_token="+s.token) {
		return true
	}
	if strings.Contains(r.URL.Path, "access_token="+s.token) {
		return true
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return false
	}
	return parts[1] == s.token
}

// SendToBrowser sends a JSON-RPC message to the browser.
func (s *WSServer) SendToBrowser(msg json.RawMessage) {
	s.ToBrowser <- msg
}
