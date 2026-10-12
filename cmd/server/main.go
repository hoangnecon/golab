package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hoangnecon/golab/internal/mcpserver"
)

func main() {
	cdpFlag := flag.String("cdp", "", "Chrome DevTools Protocol URL (default: http://127.0.0.1:9222)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cdpURL := *cdpFlag
	if cdpURL == "" {
		cdpURL = os.Getenv("COLAB_CDP_URL")
	}
	if cdpURL == "" {
		cdpURL = "http://127.0.0.1:9222"
	}

	server := mcpserver.NewServer(cdpURL)
	if err := server.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
