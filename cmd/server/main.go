package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hoangnecon/golab/internal/bridge"
	"github.com/hoangnecon/golab/internal/config"
	"github.com/hoangnecon/golab/internal/mcpserver"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	token := cfg.Token
	if token == "" {
		tokenBytes := make([]byte, 16)
		if _, err := rand.Read(tokenBytes); err != nil {
			log.Fatal(err)
		}
		token = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(tokenBytes)
	}

	ws := bridge.NewWSServer(token)
	port, err := ws.Start(ctx, cfg.WSPort)
	if err != nil {
		log.Fatalf("ws: %v", err)
	}
	defer ws.Stop()
	fmt.Fprintf(os.Stderr, "golab ws://localhost:%d\n", port)

	server := mcpserver.New(ws, cfg, token, port)
	if err := server.Run(ctx); err != nil {
		log.Fatalf("mcp: %v", err)
	}
}
