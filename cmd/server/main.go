package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/hoangnecon/golab/internal/app"
	"github.com/hoangnecon/golab/ui"
)

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

	uiFS, errFS := ui.FS()
	if errFS != nil {
		log.Printf("[UI] Warning: failed to load embedded UI assets: %v", errFS)
	}

	daemon, err := app.New(&app.Config{
		UIPort:   *uiPort,
		NoUI:     *noUI,
		NoMCP:    *noMCP,
		OpenFlag: *openFlag,
		DBPath:   *dbPath,
	}, uiFS)
	if err != nil {
		log.Fatalf("failed to initialize daemon: %v", err)
	}

	if err := daemon.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("daemon error: %v", err)
	}
}
