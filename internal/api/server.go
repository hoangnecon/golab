package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hoangnecon/golab/internal/api/controller"
	"github.com/hoangnecon/golab/internal/service"
	"github.com/hoangnecon/golab/internal/session"
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

func writeError(w http.ResponseWriter, status int, message string) {
	controller.WriteError(w, status, message)
}

type ioReadSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Initialize Application Services and Controllers
	st := s.sessionMgr.Store()
	procMgr := s.sessionMgr.ProcessManager()

	projectSvc := service.NewProjectService(st)
	profileSvc := service.NewProfileService(st, procMgr)
	sessionSvc := service.NewSessionService(s.sessionMgr)
	activitySvc := service.NewActivityService(st)

	projectCtrl := controller.NewProjectController(projectSvc)
	projectCtrl.SetSessionService(sessionSvc)
	profileCtrl := controller.NewProfileController(profileSvc)
	sessionCtrl := controller.NewSessionController(sessionSvc)
	activityCtrl := controller.NewActivityController(activitySvc)

	// Profiles
	mux.HandleFunc("GET /api/profiles", profileCtrl.HandleList)
	mux.HandleFunc("POST /api/profiles", profileCtrl.HandleCreate)
	mux.HandleFunc("POST /api/profiles/{id}/launch", profileCtrl.HandleLaunch)
	mux.HandleFunc("POST /api/profiles/{id}/stop", profileCtrl.HandleStop)
	mux.HandleFunc("DELETE /api/profiles/{id}", profileCtrl.HandleDelete)

	// Projects
	mux.HandleFunc("GET /api/projects", projectCtrl.HandleList)
	mux.HandleFunc("POST /api/projects", projectCtrl.HandleCreate)
	mux.HandleFunc("GET /api/projects/{id}", projectCtrl.HandleGet)
	mux.HandleFunc("PUT /api/projects/{id}", projectCtrl.HandleUpdate)
	mux.HandleFunc("DELETE /api/projects/{id}", projectCtrl.HandleDelete)
	mux.HandleFunc("POST /api/projects/{id}/launch", projectCtrl.HandleLaunch)

	// Sessions
	mux.HandleFunc("GET /api/sessions", sessionCtrl.HandleList)
	mux.HandleFunc("POST /api/sessions", sessionCtrl.HandleCreate)
	mux.HandleFunc("POST /api/sessions/{id}/activate", sessionCtrl.HandleActivate)
	mux.HandleFunc("POST /api/sessions/{id}/focus", sessionCtrl.HandleFocus)
	mux.HandleFunc("POST /api/sessions/{id}/interrupt", sessionCtrl.HandleInterrupt)
	mux.HandleFunc("POST /api/sessions/{id}/restart", sessionCtrl.HandleRestart)
	mux.HandleFunc("DELETE /api/sessions/{id}", sessionCtrl.HandleDelete)
	mux.HandleFunc("POST /api/sessions/{id}/lock", sessionCtrl.HandleToggleLock)
	mux.HandleFunc("POST /api/sessions/{id}/connect", sessionCtrl.HandleConnect)
	mux.HandleFunc("POST /api/sessions/{id}/disconnect", sessionCtrl.HandleDisconnect)
	mux.HandleFunc("POST /api/sessions/{id}/save", sessionCtrl.HandleSave)
	mux.HandleFunc("POST /api/sessions/{id}/runall", sessionCtrl.HandleRunAll)
	mux.HandleFunc("GET /api/sessions/{id}/running-cells", sessionCtrl.HandleRunningCells)
	mux.HandleFunc("GET /api/sessions/{id}/accelerators", sessionCtrl.HandleAccelerators)
	mux.HandleFunc("POST /api/sessions/{id}/accelerator", sessionCtrl.HandleSetAccelerator)

	// Activity
	mux.HandleFunc("GET /api/activity", activityCtrl.HandleList)
	mux.HandleFunc("GET /api/activity/{id}", activityCtrl.HandleGet)
	mux.HandleFunc("GET /api/events", s.handleSSE)

	// Static UI assets / SPA handler
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			f, err := s.staticFS.Open(path)
			if err != nil {
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
			controller.WriteJSON(w, http.StatusOK, map[string]any{
				"message": "GoLab Control Plane Daemon running.",
				"version": "v2.0.0-harness",
			})
		})
	}

	return guardMiddleware(mux)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "Streaming unsupported")
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
		recentLogs, _ := s.sessionMgr.Store().ListRecentActivity(ctx, 10)
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
