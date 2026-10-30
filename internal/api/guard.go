package api

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// guardMiddleware keeps the control plane reachable only from its own pages.
//
// The daemon listens on loopback without credentials, so the browser is the
// attack path: any website the user opens can send requests to 127.0.0.1.
// There is deliberately no CORS: the UI is served by this server (or by the
// Vite dev proxy, which keeps the Host header), so every legitimate browser
// request is same-origin. The guard enforces three things:
//
//   - Host must be a loopback name (defeats DNS rebinding, where a foreign
//     name resolves to 127.0.0.1).
//   - Sec-Fetch-Site, when the browser sends it, must be same-origin or none.
//   - Origin, when present (browsers send it on cross-origin and on
//     non-GET requests), must equal the request's own origin.
//
// Requests without Origin and Sec-Fetch-Site come from non-browser clients
// such as curl, which an attacking web page cannot impersonate.
func guardMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reason := rejectReason(r); reason != "" {
			writeError(w, http.StatusForbidden, reason)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func rejectReason(r *http.Request) string {
	if !isLoopbackHost(r.Host) {
		return "forbidden host"
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
	default:
		return "cross-site request blocked"
	}
	if origin := r.Header.Get("Origin"); origin != "" && !sameOrigin(origin, r.Host) {
		return "cross-origin request blocked"
	}
	return ""
}

func isLoopbackHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// sameOrigin reports whether an Origin header names this server (same host and
// port as the Host header). The literal "null" origin never matches.
func sameOrigin(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	return strings.EqualFold(u.Host, host)
}
