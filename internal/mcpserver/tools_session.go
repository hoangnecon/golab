package mcpserver

import (
	"context"
	"errors"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListSessionsInput specifies arguments for listing active sessions.
type ListSessionsInput struct{}

func (s *Server) colabListAllSessions(ctx context.Context, req *mcp.CallToolRequest, input ListSessionsInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	if s.sessionMgr == nil {
		return errorResult(errors.New("session manager is not initialized"), 0)
	}

	sessions, err := s.sessionMgr.ListLiveSessions(ctx)
	latency := time.Since(start).Milliseconds()
	s.logActivity(ctx, "", "colab_list_all_sessions", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}
	return successResult(map[string]any{
		"sessions":       sessions,
		"total_sessions": len(sessions),
	}, latency, "")
}

// SwitchActiveSessionInput specifies arguments for changing the active notebook context.
type SwitchActiveSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"The ID of the session to set as active. Required."`
}

func (s *Server) colabSwitchActiveSession(ctx context.Context, req *mcp.CallToolRequest, input SwitchActiveSessionInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	if s.sessionMgr == nil {
		return errorResult(errors.New("session manager is not initialized"), 0)
	}
	if input.SessionID == "" {
		return errorResult(errors.New("session_id is required"), 0)
	}

	err := s.sessionMgr.Store().SetActiveSession(input.SessionID)
	latency := time.Since(start).Milliseconds()
	s.logActivity(ctx, input.SessionID, "colab_switch_active_session", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}

	sess, _ := s.sessionMgr.Store().GetSession(input.SessionID)
	return successResult(map[string]any{
		"message":        "Active session switched successfully",
		"active_session": sess,
	}, latency, "")
}

// CreateSessionInput specifies arguments for opening a new notebook session.
type CreateSessionInput struct {
	ProfileID   string `json:"profile_id" jsonschema:"Google account profile ID to launch the notebook under. Required."`
	NotebookURL string `json:"notebook_url" jsonschema:"Full Google Colab notebook URL (e.g. https://colab.research.google.com/drive/...). Required."`
	Name        string `json:"name,omitempty" jsonschema:"Optional display name for this notebook worktree."`
}

func (s *Server) colabCreateSession(ctx context.Context, req *mcp.CallToolRequest, input CreateSessionInput) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	if s.sessionMgr == nil {
		return errorResult(errors.New("session manager is not initialized"), 0)
	}
	if input.ProfileID == "" || input.NotebookURL == "" {
		return errorResult(errors.New("profile_id and notebook_url are required"), 0)
	}

	act, err := s.sessionMgr.OpenNotebook(ctx, input.ProfileID, input.NotebookURL, input.Name)
	latency := time.Since(start).Milliseconds()
	s.logActivity(ctx, "", "colab_create_session", latency, err)
	if err != nil {
		return errorResult(err, latency)
	}

	return successResult(map[string]any{
		"message": "Session created and attached successfully",
		"session": act.Session,
	}, latency, "")
}
