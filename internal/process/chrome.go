package process

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/hoangnecon/golab/internal/store"
)

// Manager oversees external Chrome processes and their CDP ports.
type Manager struct {
	mu        sync.RWMutex
	processes map[string]*exec.Cmd
	ports     map[string]int
}

func NewManager() *Manager {
	return &Manager{
		processes: make(map[string]*exec.Cmd),
		ports:     make(map[string]int),
	}
}

// FindChromePath locates the Google Chrome binary on the system.
func FindChromePath() (string, error) {
	if env := os.Getenv("CHROME_PATH"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env, nil
		}
	}

	switch runtime.GOOS {
	case "darwin":
		candidates := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			filepath.Join(os.Getenv("HOME"), "Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	case "linux":
		candidates := []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"}
		for _, c := range candidates {
			if p, err := exec.LookPath(c); err == nil {
				return p, nil
			}
		}
	case "windows":
		candidates := []string{
			os.Getenv("ProgramFiles") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("ProgramFiles(x86)") + `\Google\Chrome\Application\chrome.exe`,
			os.Getenv("LocalAppData") + `\Google\Chrome\Application\chrome.exe`,
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}

	return "", fmt.Errorf("google chrome binary not found on %s", runtime.GOOS)
}

// IsPortListening checks if a local TCP port is responding.
func IsPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsCDPAlive verifies that Chrome's CDP endpoint is healthy.
func IsCDPAlive(port int) bool {
	client := http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// LaunchProfile starts Google Chrome with an isolated profile and allocated CDP port.
func (m *Manager) LaunchProfile(ctx context.Context, p *store.Profile, targetURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if IsCDPAlive(p.CDPPort) {
		// Already running
		return nil
	}

	chromePath, err := FindChromePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(p.UserDataDir, 0755); err != nil {
		return fmt.Errorf("create profile user data dir: %w", err)
	}

	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", p.CDPPort),
		fmt.Sprintf("--user-data-dir=%s", p.UserDataDir),
		"--no-first-run",
		"--no-default-browser-check",
	}

	if targetURL != "" {
		args = append(args, targetURL)
	} else {
		args = append(args, "https://colab.research.google.com")
	}

	cmd := exec.Command(chromePath, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start chrome: %w", err)
	}

	m.processes[p.ID] = cmd
	m.ports[p.ID] = p.CDPPort

	// Wait up to 5 seconds for CDP to become responsive
	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		delete(m.processes, p.ID)
		delete(m.ports, p.ID)
		m.mu.Unlock()
	}()

	t0 := time.Now()
	for time.Since(t0) < 5*time.Second {
		if IsCDPAlive(p.CDPPort) {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}

	return nil
}

// StopProfile shuts down the managed Chrome process for a profile.
func (m *Manager) StopProfile(profileID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd, exists := m.processes[profileID]
	if !exists || cmd.Process == nil {
		return nil
	}

	_ = cmd.Process.Signal(os.Interrupt)
	time.Sleep(300 * time.Millisecond)
	_ = cmd.Process.Kill()

	delete(m.processes, profileID)
	delete(m.ports, profileID)
	return nil
}

// IsProfileRunning checks if Chrome for a given profile is actively listening.
func (m *Manager) IsProfileRunning(p *store.Profile) bool {
	return IsCDPAlive(p.CDPPort)
}
