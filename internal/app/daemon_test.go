package app_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoangnecon/golab/internal/app"
)

func TestDaemonLifecycle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "daemon_test.db")

	cfg := &app.Config{
		UIPort: 0, // OS allocates free port
		NoUI:   true,
		NoMCP:  true,
		DBPath: dbPath,
	}

	d, err := app.New(cfg, nil)
	if err != nil {
		t.Fatalf("failed to create daemon: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() {
		runErr <- d.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel() // Trigger graceful shutdown

	select {
	case err := <-runErr:
		if err != nil && err != context.Canceled {
			t.Fatalf("daemon run error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not stop gracefully within 2 seconds")
	}
}
