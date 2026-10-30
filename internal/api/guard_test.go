package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hoangnecon/golab/internal/session"
	"github.com/hoangnecon/golab/internal/store"
)

func newGuardedHandler(t *testing.T) (http.Handler, *store.Store) {
	t.Helper()
	st, err := store.New(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return NewServer(session.NewManager(st, nil), nil).Handler(), st
}

func TestGuardAllowsOwnOriginAndNonBrowserClients(t *testing.T) {
	h, _ := newGuardedHandler(t)
	cases := []struct {
		name    string
		host    string
		headers map[string]string
	}{
		{"curl without Origin", "127.0.0.1:9120", nil},
		{"same origin ipv4", "127.0.0.1:9120", map[string]string{"Origin": "http://127.0.0.1:9120", "Sec-Fetch-Site": "same-origin"}},
		{"localhost host and origin", "localhost:9120", map[string]string{"Origin": "http://localhost:9120"}},
		{"ipv6 loopback", "[::1]:9120", map[string]string{"Origin": "http://[::1]:9120"}},
		{"address bar navigation", "127.0.0.1:9120", map[string]string{"Sec-Fetch-Site": "none"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/activity", nil)
			req.Host = c.host
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Fatalf("CORS must not be enabled, got Access-Control-Allow-Origin=%q", got)
			}
		})
	}
}

func TestGuardRejectsForeignRequests(t *testing.T) {
	h, _ := newGuardedHandler(t)
	cases := []struct {
		name    string
		host    string
		headers map[string]string
	}{
		{"DNS rebinding host", "evil.example.com", nil},
		{"DNS rebinding host with port", "evil.example.com:9120", map[string]string{"Origin": "http://evil.example.com:9120"}},
		{"foreign origin", "127.0.0.1:9120", map[string]string{"Origin": "http://evil.example.com"}},
		{"vite dev origin the proxy did not rewrite", "127.0.0.1:9120", map[string]string{"Origin": "http://127.0.0.1:5173"}},
		{"other loopback port", "127.0.0.1:9120", map[string]string{"Origin": "http://127.0.0.1:5199"}},
		{"localhost page calling 127.0.0.1", "127.0.0.1:9120", map[string]string{"Origin": "http://localhost:9120"}},
		{"null origin", "127.0.0.1:9120", map[string]string{"Origin": "null"}},
		{"file origin", "127.0.0.1:9120", map[string]string{"Origin": "file://"}},
		{"cross-site fetch metadata", "127.0.0.1:9120", map[string]string{"Sec-Fetch-Site": "cross-site"}},
		{"same-site fetch metadata", "127.0.0.1:9120", map[string]string{"Sec-Fetch-Site": "same-site"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/activity", nil)
			req.Host = c.host
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("want 403, got %d: %s", rec.Code, rec.Body)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Fatalf("CORS header leaked on rejection: %q", got)
			}
		})
	}
}

// The attack seen in the browser probe: a foreign page sends a "simple" POST
// (text/plain, no preflight) that the handler decodes as JSON.
func TestGuardBlocksForeignWritesBeforeTheHandlerRuns(t *testing.T) {
	h, st := newGuardedHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"planted"}`))
	req.Host = "127.0.0.1:9120"
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Origin", "http://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
	profiles, err := st.ListProfiles(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range profiles {
		if p.Name == "planted" {
			t.Fatal("profile was created by a foreign origin")
		}
	}

	own := httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"legit"}`))
	own.Host = "127.0.0.1:9120"
	own.Header.Set("Content-Type", "application/json")
	own.Header.Set("Origin", "http://127.0.0.1:9120")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, own)
	if rec.Code != http.StatusCreated {
		t.Fatalf("own-origin create: want 201, got %d: %s", rec.Code, rec.Body)
	}
}

func TestGuardPreflightFromForeignOriginIsRefused(t *testing.T) {
	h, _ := newGuardedHandler(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/profiles", nil)
	req.Host = "127.0.0.1:9120"
	req.Header.Set("Origin", "http://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("preflight: want 403 and no CORS headers, got %d %v", rec.Code, rec.Header())
	}
}
