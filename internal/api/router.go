package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store"
)

// Server provides the HTTP Control Plane API and UI static assets.
type Server struct {
	sessionMgr *session.Manager
	staticFS   fs.FS
	mu         sync.RWMutex
	clients    map[chan []byte]bool
}

func NewServer(sm *session.Manager, staticFS fs.FS) *Server {
	s := &Server{
		sessionMgr: sm,
		staticFS:   staticFS,
		clients:    make(map[chan []byte]bool),
	}
	go s.eventBroadcaster()
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /api/profiles", s.handleListProfiles)
	mux.HandleFunc("POST /api/profiles", s.handleCreateProfile)
	mux.HandleFunc("POST /api/profiles/{id}/launch", s.handleLaunchProfile)
	mux.HandleFunc("POST /api/profiles/{id}/stop", s.handleStopProfile)
	mux.HandleFunc("DELETE /api/profiles/{id}", s.handleDeleteProfile)

	mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	mux.HandleFunc("POST /api/sessions", s.handleCreateSession)
	mux.HandleFunc("POST /api/sessions/{id}/activate", s.handleActivateSession)
	mux.HandleFunc("POST /api/sessions/{id}/focus", s.handleFocusSession)
	mux.HandleFunc("POST /api/sessions/{id}/interrupt", s.handleInterruptSession)
	mux.HandleFunc("POST /api/sessions/{id}/restart", s.handleRestartSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", s.handleDeleteSession)

	mux.HandleFunc("GET /api/activity", s.handleListActivity)
	mux.HandleFunc("GET /api/events", s.handleSSE)

	// Static UI assets / SPA handler
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Serve API or static
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			// Check if file exists in FS
			f, err := s.staticFS.Open(path)
			if err != nil {
				// SPA fallback to index.html
				indexFile, errIndex := s.staticFS.Open("index.html")
				if errIndex == nil {
					defer indexFile.Close()
					http.ServeContent(w, r, "index.html", time.Now(), indexFile.(ioReadSeeker))
					return
				}
			} else {
				_ = f.Close()
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "GoLab Control Plane Daemon running.",
				"version": "v2.0.0-harness",
			})
		})
	}

	return s.corsMiddleware(mux)
}

type ioReadSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"status":  "error",
		"message": message,
	})
}

// --- Profile Handlers ---

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := s.sessionMgr.Store().ListProfiles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, p := range profiles {
		p.IsRunning = s.sessionMgr.ProcessManager().IsProfileRunning(p)
	}

	writeJSON(w, http.StatusOK, map[string]any{"profiles": profiles})
}

func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string `json:"name"`
		AccountEmail string `json:"account_email"`
		Tier         string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if input.Tier == "" {
		input.Tier = "free"
	}

	port, err := s.sessionMgr.Store().GetNextAvailableCDPPort(9222)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	profileID := uuid.New().String()
	home, _ := os.UserHomeDir()
	userDataDir := filepath.Join(home, ".golab", "profiles", profileID)

	p := &store.Profile{
		ID:           profileID,
		Name:         input.Name,
		AccountEmail: input.AccountEmail,
		Tier:         input.Tier,
		UserDataDir:  userDataDir,
		CDPPort:      port,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.sessionMgr.Store().CreateProfile(p); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleLaunchProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.sessionMgr.Store().GetProfile(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	if err := s.sessionMgr.ProcessManager().LaunchProfile(r.Context(), p, "https://colab.research.google.com"); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("launch chrome: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"launched": true, "cdp_port": p.CDPPort})
}

func (s *Server) handleStopProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_ = s.sessionMgr.ProcessManager().StopProfile(id)
	writeJSON(w, http.StatusOK, map[string]any{"stopped": true})
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_ = s.sessionMgr.ProcessManager().StopProfile(id)
	if err := s.sessionMgr.Store().DeleteProfile(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// --- Session Handlers ---

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.sessionMgr.ListLiveSessions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProfileID   string `json:"profile_id"`
		NotebookURL string `json:"notebook_url"`
		Name        string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.ProfileID == "" || input.NotebookURL == "" {
		writeError(w, http.StatusBadRequest, "profile_id and notebook_url are required")
		return
	}

	act, err := s.sessionMgr.OpenNotebook(r.Context(), input.ProfileID, input.NotebookURL, input.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("open notebook: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, act.Session)
}

func (s *Server) handleActivateSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.sessionMgr.Store().SetActiveSession(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"activated": true, "session_id": id})
}

func (s *Server) handleFocusSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	act, err := s.sessionMgr.ResolveSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if act.Session.TabID != "" {
		_ = act.CDP.FocusTab(r.Context(), act.Session.TabID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"focused": true})
}

func (s *Server) handleInterruptSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	act, err := s.sessionMgr.ResolveSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := act.Services.Kernel.InterruptKernel(r.Context(), ""); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("interrupt failed: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"interrupted": true, "message": "SIGINT signal dispatched to kernel."})
}

func (s *Server) handleRestartSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	act, err := s.sessionMgr.ResolveSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	detail, err := act.Services.Kernel.RestartKernel(r.Context(), "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("restart failed: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"restarted": true, "kernel": detail})
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.sessionMgr.CloseSession(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"closed": true})
}

// --- Activity Handlers ---

func (s *Server) handleListActivity(w http.ResponseWriter, r *http.Request) {
	logs, err := s.sessionMgr.Store().ListRecentActivity(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"activity": logs})
}

// --- SSE Stream ---

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msgChan := make(chan []byte, 10)
	s.mu.Lock()
	s.clients[msgChan] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, msgChan)
		close(msgChan)
		s.mu.Unlock()
	}()

	// Send initial ping
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgChan:
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) eventBroadcaster() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.RLock()
		hasClients := len(s.clients) > 0
		s.mu.RUnlock()
		if !hasClients {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		sessions, _ := s.sessionMgr.ListLiveSessions(ctx)
		recentLogs, _ := s.sessionMgr.Store().ListRecentActivity(10)
		cancel()

		payload, _ := json.Marshal(map[string]any{
			"type":      "tick",
			"timestamp": time.Now().UTC(),
			"sessions":  sessions,
			"activity":  recentLogs,
		})

		s.mu.RLock()
		for ch := range s.clients {
			select {
			case ch <- payload:
			default:
			}
		}
		s.mu.RUnlock()
	}
}
