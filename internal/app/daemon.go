package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hoangnecon/golab/internal/api"
	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/mcpserver"
	"github.com/hoangnecon/golab/internal/process"
	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store/sqlite"
)

// Config holds configuration parameters for the Golab Daemon.
type Config struct {
	UIPort   int    `json:"ui_port"`
	NoUI     bool   `json:"no_ui"`
	NoMCP    bool   `json:"no_mcp"`
	OpenFlag bool   `json:"open"`
	DBPath   string `json:"db_path"`
}

// Daemon orchestrates the lifecycle of storage, process supervisor, sessions, Web UI, and MCP server.
type Daemon struct {
	cfg        *Config
	store      *sqlite.Store
	procMgr    *process.Manager
	sessionMgr *session.Manager
	httpServer *http.Server
	mcpServer  *mcpserver.Server
}

// New constructs and wires all dependencies for the daemon.
func New(cfg *Config, uiFS fs.FS) (*Daemon, error) {
	if cfg == nil {
		cfg = &Config{UIPort: 9120}
	}

	st, err := sqlite.New(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("init sqlite store: %w", err)
	}

	// Seed default profile if empty
	ctxInit, cancelInit := context.WithTimeout(context.Background(), 2*time.Second)
	profiles, _ := st.ListProfiles(ctxInit)
	if len(profiles) == 0 {
		home, _ := os.UserHomeDir()
		defaultProfile := &domain.Profile{
			ID:          "default",
			Name:        "Default Google Account",
			Tier:        domain.TierFree,
			UserDataDir: filepath.Join(home, ".golab", "profiles", "default"),
			CDPPort:     9222,
			CreatedAt:   time.Now().UTC(),
		}
		_ = st.CreateProfile(ctxInit, defaultProfile)
	}
	cancelInit()

	procMgr := process.NewManager()
	sessionMgr := session.NewManager(st, procMgr)

	d := &Daemon{
		cfg:        cfg,
		store:      st,
		procMgr:    procMgr,
		sessionMgr: sessionMgr,
	}

	if !cfg.NoUI {
		apiServer := api.NewServer(sessionMgr, uiFS)
		d.httpServer = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", cfg.UIPort),
			Handler: apiServer.Handler(),
		}
	}

	if !cfg.NoMCP {
		d.mcpServer = mcpserver.NewServerWithSessionManager(sessionMgr)
	}

	return d, nil
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", url).Start()
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

// Run starts the daemon components and blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	defer d.store.Close()

	errCh := make(chan error, 2)

	// 1. Start Web UI HTTP Server
	if d.httpServer != nil {
		go func() {
			log.Printf("[Control Plane] Web UI Dashboard listening on http://%s", d.httpServer.Addr)
			if err := d.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("http server: %w", err)
			}
		}()

		if d.cfg.OpenFlag {
			go func() {
				time.Sleep(300 * time.Millisecond)
				openBrowser(fmt.Sprintf("http://127.0.0.1:%d", d.cfg.UIPort))
			}()
		}

		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = d.httpServer.Shutdown(shutdownCtx)
		}()
	}

	// 2. Start MCP Server or wait for shutdown
	if d.mcpServer != nil {
		log.Printf("[MCP] GoLab Enterprise Harness MCP Server running on stdio...")
		if err := d.mcpServer.Run(ctx); err != nil {
			return fmt.Errorf("mcp server error: %w", err)
		}
		return nil
	}

	// If running in UI-only mode, block until context cancelled or HTTP error
	log.Printf("[Daemon] Running in UI-only mode. Press Ctrl+C to terminate.")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
