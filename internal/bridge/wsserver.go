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

	"nhooyr.io/websocket"
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
	token       string
	mu          sync.Mutex
	conn        *websocket.Conn
	connectedAt time.Time
	remoteAddr  string
	connected   chan struct{}
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
	addr := fmt.Sprintf("localhost:%d", port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		if port != 0 {
			log.Printf("Port %d in use, falling back to random port", port)
			s.listener, err = net.Listen("tcp", "localhost:0")
		}
		if err != nil {
			return 0, fmt.Errorf("listen: %w", err)
		}
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

	if s.conn != nil {
		s.conn.Close(websocket.StatusPolicyViolation, "Token rotated")
	}
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
	return s.connected
}

func (s *WSServer) handleWS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	originOK := false
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			originOK = true
			break
		}
	}
	if !originOK {
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return
	}

	if !s.validateAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	s.mu.Lock()
	if s.conn != nil {
		s.mu.Unlock()
		http.Error(w, "already connected", http.StatusConflict)
		return
	}
	s.mu.Unlock()

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:   []string{"mcp"},
		OriginPatterns: []string{"colab.research.google.com", "colab.google.com"},
	})
	if err != nil {
		log.Printf("WS accept error: %v", err)
		return
	}
	conn.SetReadLimit(10 * 1024 * 1024)

	s.mu.Lock()
	s.conn = conn
	s.connectedAt = time.Now()
	s.remoteAddr = r.RemoteAddr
	s.mu.Unlock()

	select {
	case <-s.connected:
		s.connected = make(chan struct{})
		close(s.connected)
	default:
		close(s.connected)
	}

	log.Println("Colab browser connected")

	ctx := r.Context()
	defer func() {
		s.mu.Lock()
		s.conn = nil
		s.connectedAt = time.Time{}
		s.remoteAddr = ""
		s.connected = make(chan struct{})
		s.mu.Unlock()
		conn.Close(websocket.StatusNormalClosure, "bye")
		log.Println("Colab browser disconnected")
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			select {
			case s.FromBrowser <- json.RawMessage(data):
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case msg := <-s.ToBrowser:
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
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
