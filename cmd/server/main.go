package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/hoangnecon/golab/internal/api"
	"github.com/hoangnecon/golab/internal/mcpserver"
	"github.com/hoangnecon/golab/internal/process"
	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store"
	"github.com/hoangnecon/golab/ui"
)

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

func main() {
	uiPort := flag.Int("ui-port", 9120, "Port for the Web UI Control Plane dashboard")
	noUI := flag.Bool("no-ui", false, "Disable Web UI server (run MCP only)")
	noMCP := flag.Bool("no-mcp", false, "Disable MCP server (run Web UI only)")
	openFlag := flag.Bool("open", false, "Open Web UI dashboard in default browser on launch")
	dbPath := flag.String("db", "", "Path to SQLite state database (default: ~/.golab/state.db)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Check subcommands or arguments like 'golab ui'
	if flag.NArg() > 0 && flag.Arg(0) == "ui" {
		*openFlag = true
	}

	// 1. Initialize State Store
	st, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer st.Close()

	// Seed default profile if empty
	profiles, _ := st.ListProfiles()
	if len(profiles) == 0 {
		home, _ := os.UserHomeDir()
		defaultProfile := &store.Profile{
			ID:           "default",
			Name:         "Default Google Account",
			Tier:         "free",
			UserDataDir:  filepath.Join(home, ".golab", "profiles", "default"),
			CDPPort:      9222,
			CreatedAt:    time.Now().UTC(),
		}
		_ = st.CreateProfile(defaultProfile)
	}

	// 2. Initialize Process and Session Supervisors
	procMgr := process.NewManager()
	sessionMgr := session.NewManager(st, procMgr)

	// 3. Start Control Plane Web Server (Port 9120)
	if !*noUI {
		uiFS, errFS := ui.FS()
		if errFS != nil {
			log.Printf("[UI] Warning: failed to load embedded UI assets: %v", errFS)
		}

		apiServer := api.NewServer(sessionMgr, uiFS)
		httpServer := &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", *uiPort),
			Handler: apiServer.Handler(),
		}

		go func() {
			log.Printf("[Control Plane] Web UI Dashboard listening on http://127.0.0.1:%d", *uiPort)
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("[Control Plane] HTTP server error: %v", err)
			}
		}()

		if *openFlag {
			go func() {
				time.Sleep(300 * time.Millisecond)
				openBrowser(fmt.Sprintf("http://127.0.0.1:%d", *uiPort))
			}()
		}

		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer shutdownCancel()
			_ = httpServer.Shutdown(shutdownCtx)
		}()
	}

	// 4. Start MCP Server (stdio transport for Agent)
	if !*noMCP {
		mcpSrv := mcpserver.NewServerWithSessionManager(sessionMgr)
		log.Printf("[MCP] GoLab Enterprise Harness MCP Server running on stdio...")
		if err := mcpSrv.Run(ctx); err != nil {
			log.Fatalf("mcp server error: %v", err)
		}
	} else {
		// Keep alive until interrupted if running in UI-only mode
		log.Printf("[Daemon] Running in UI-only mode. Press Ctrl+C to terminate.")
		<-ctx.Done()
	}
}
