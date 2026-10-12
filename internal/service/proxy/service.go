package proxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hoangnecon/golab/internal/bridge"
)

// Service provides reverse proxy forwarding to arbitrary ports on the Colab VM (/_proxy/<port>/...).
type Service struct {
	exec *bridge.ColabExecutor
}

// NewService creates a reverse proxy service.
func NewService(exec *bridge.ColabExecutor) *Service {
	return &Service{exec: exec}
}

// ProxyForward sends a request through Colab's reverse proxy router to localhost:<port> inside the VM.
func (s *Service) ProxyForward(ctx context.Context, port int, subpath string, method string, headers map[string]string, body string) (*bridge.FetchResponse, error) {
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", port)
	}

	subpath = strings.TrimPrefix(subpath, "/")
	proxyPath := fmt.Sprintf("/_proxy/%d/%s", port, subpath)

	return s.exec.FetchKernelXHR(ctx, proxyPath, bridge.FetchOptions{
		Method:  method,
		Headers: headers,
		Body:    body,
	})
}
