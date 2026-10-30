package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hoangnecon/golab/internal/domain"
	_ "modernc.org/sqlite"
)

type domainPayload = domain.Payload

// Store implements the persistence layer for Golab on SQLite.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// New opens or creates an SQLite database at dbPath and initializes the schema.
func New(dbPath string) (*Store, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve user home dir: %w", err)
		}
		dir := filepath.Join(home, ".golab")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create golab dir: %w", err)
		}
		dbPath = filepath.Join(dir, "state.db")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return s, nil
}

// Close terminates the underlying database connection.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

func parseSQLiteTime(val any) time.Time {
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v.UTC()
	case string:
		layouts := []string{
			"2006-01-02 15:04:05.999999999 -0700 MST",
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05",
			time.RFC3339Nano,
			time.RFC3339,
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, v); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Time{}
}

// --- Profile Repository Implementation ---

func (s *Store) CreateProfile(ctx context.Context, p *domain.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	query := `INSERT INTO profiles (id, name, account_email, tier, user_data_dir, cdp_port, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, p.ID, p.Name, p.AccountEmail, p.Tier, p.UserDataDir, p.CDPPort, p.CreatedAt)
	return err
}

func (s *Store) GetProfile(ctx context.Context, id string) (*domain.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, COALESCE(account_email, ''), tier, user_data_dir, cdp_port, created_at
	          FROM profiles WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var p domain.Profile
	var rawCreated any
	if err := row.Scan(&p.ID, &p.Name, &p.AccountEmail, &p.Tier, &p.UserDataDir, &p.CDPPort, &rawCreated); err != nil {
		return nil, err
	}
	p.CreatedAt = parseSQLiteTime(rawCreated)
	return &p, nil
}

func (s *Store) ListProfiles(ctx context.Context) ([]*domain.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, COALESCE(account_email, ''), tier, user_data_dir, cdp_port, created_at
	          FROM profiles ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Profile
	for rows.Next() {
		var p domain.Profile
		var rawCreated any
		if err := rows.Scan(&p.ID, &p.Name, &p.AccountEmail, &p.Tier, &p.UserDataDir, &p.CDPPort, &rawCreated); err != nil {
			return nil, err
		}
		p.CreatedAt = parseSQLiteTime(rawCreated)
		list = append(list, &p)
	}
	return list, nil
}

func (s *Store) UpdateProfile(ctx context.Context, p *domain.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE profiles SET name = ?, account_email = ?, tier = ?, user_data_dir = ?, cdp_port = ?
	          WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, p.Name, p.AccountEmail, p.Tier, p.UserDataDir, p.CDPPort, p.ID)
	return err
}

func (s *Store) DeleteProfile(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM profiles WHERE id = ?`, id)
	return err
}

func (s *Store) GetNextAvailableCDPPort(ctx context.Context, startPort int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, `SELECT cdp_port FROM profiles ORDER BY cdp_port ASC`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	used := make(map[int]bool)
	for rows.Next() {
		var port int
		if err := rows.Scan(&port); err == nil {
			used[port] = true
		}
	}

	port := startPort
	for {
		if !used[port] {
			return port, nil
		}
		port++
	}
}

// --- Project Repository Implementation ---

func (s *Store) CreateProject(ctx context.Context, p *domain.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	query := `INSERT INTO projects (id, name, notebook_url, owner_profile_id, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, p.ID, p.Name, p.NotebookURL, p.OwnerProfileID, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *Store) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, notebook_url, owner_profile_id, created_at, updated_at FROM projects WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var p domain.Project
	var rawCreated, rawUpdated any
	if err := row.Scan(&p.ID, &p.Name, &p.NotebookURL, &p.OwnerProfileID, &rawCreated, &rawUpdated); err != nil {
		return nil, err
	}
	p.CreatedAt = parseSQLiteTime(rawCreated)
	p.UpdatedAt = parseSQLiteTime(rawUpdated)
	return &p, nil
}

func (s *Store) GetProjectByURL(ctx context.Context, notebookURL string) (*domain.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, notebook_url, owner_profile_id, created_at, updated_at FROM projects WHERE notebook_url = ?`
	row := s.db.QueryRowContext(ctx, query, notebookURL)

	var p domain.Project
	var rawCreated, rawUpdated any
	if err := row.Scan(&p.ID, &p.Name, &p.NotebookURL, &p.OwnerProfileID, &rawCreated, &rawUpdated); err != nil {
		return nil, err
	}
	p.CreatedAt = parseSQLiteTime(rawCreated)
	p.UpdatedAt = parseSQLiteTime(rawUpdated)
	return &p, nil
}

func (s *Store) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, notebook_url, owner_profile_id, created_at, updated_at FROM projects ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Project
	for rows.Next() {
		var p domain.Project
		var rawCreated, rawUpdated any
		if err := rows.Scan(&p.ID, &p.Name, &p.NotebookURL, &p.OwnerProfileID, &rawCreated, &rawUpdated); err != nil {
			return nil, err
		}
		p.CreatedAt = parseSQLiteTime(rawCreated)
		p.UpdatedAt = parseSQLiteTime(rawUpdated)
		list = append(list, &p)
	}
	return list, nil
}

func (s *Store) UpdateProject(ctx context.Context, p *domain.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.UpdatedAt = time.Now().UTC()
	query := `UPDATE projects SET name = ?, notebook_url = ?, owner_profile_id = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, p.Name, p.NotebookURL, p.OwnerProfileID, p.UpdatedAt, p.ID)
	return err
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	return err
}

// --- Session Repository Implementation ---

const sessionColumns = `id, name, profile_id, COALESCE(project_id, ''), COALESCE(owner_profile_id, ''),
	notebook_url, COALESCE(notebook_id, ''), COALESCE(tab_id, ''), is_active, is_locked,
	COALESCE(permission, 'unknown'), created_at, last_attached_at`

func (s *Store) SaveSession(ctx context.Context, sess *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = time.Now().UTC()
	}
	sess.LastAttachedAt = time.Now().UTC()

	query := `INSERT INTO sessions (id, name, profile_id, project_id, owner_profile_id, notebook_url, notebook_id, tab_id, is_active, is_locked, permission, created_at, last_attached_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	          ON CONFLICT(id) DO UPDATE SET
	            name = excluded.name,
	            profile_id = excluded.profile_id,
	            project_id = excluded.project_id,
	            owner_profile_id = excluded.owner_profile_id,
	            notebook_url = excluded.notebook_url,
	            notebook_id = excluded.notebook_id,
	            tab_id = excluded.tab_id,
	            is_active = excluded.is_active,
	            is_locked = excluded.is_locked,
	            permission = excluded.permission,
	            last_attached_at = excluded.last_attached_at`
	_, err := s.db.ExecContext(ctx, query, sess.ID, sess.Name, sess.ProfileID, sess.ProjectID, sess.OwnerProfileID, sess.NotebookURL, sess.NotebookID, sess.TabID, sess.IsActive, sess.IsLocked, sess.Permission, sess.CreatedAt, sess.LastAttachedAt)
	return err
}

func (s *Store) SetSessionLocked(ctx context.Context, id string, locked bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET is_locked = ? WHERE id = ?`, locked, id)
	return err
}

func (s *Store) SetSessionPermission(ctx context.Context, id string, perm domain.PermissionLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET permission = ? WHERE id = ?`, perm, id)
	return err
}

func (s *Store) GetLockedSession(ctx context.Context) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT ` + sessionColumns + ` FROM sessions WHERE is_locked = 1 LIMIT 1`
	row := s.db.QueryRowContext(ctx, query)

	var sess domain.Session
	var rawCreated, rawAttached any
	if err := row.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.ProjectID, &sess.OwnerProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &sess.IsLocked, &sess.Permission, &rawCreated, &rawAttached); err != nil {
		return nil, err
	}
	sess.CreatedAt = parseSQLiteTime(rawCreated)
	sess.LastAttachedAt = parseSQLiteTime(rawAttached)
	return &sess, nil
}

func (s *Store) SetActiveSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET is_active = 0`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET is_active = 1, last_attached_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT ` + sessionColumns + ` FROM sessions WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var sess domain.Session
	var rawCreated, rawAttached any
	if err := row.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.ProjectID, &sess.OwnerProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &sess.IsLocked, &sess.Permission, &rawCreated, &rawAttached); err != nil {
		return nil, err
	}
	sess.CreatedAt = parseSQLiteTime(rawCreated)
	sess.LastAttachedAt = parseSQLiteTime(rawAttached)
	return &sess, nil
}

func (s *Store) GetActiveSession(ctx context.Context) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT ` + sessionColumns + ` FROM sessions WHERE is_active = 1 LIMIT 1`
	row := s.db.QueryRowContext(ctx, query)

	var sess domain.Session
	var rawCreated, rawAttached any
	if err := row.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.ProjectID, &sess.OwnerProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &sess.IsLocked, &sess.Permission, &rawCreated, &rawAttached); err != nil {
		queryFallback := `SELECT ` + sessionColumns + ` FROM sessions ORDER BY last_attached_at DESC LIMIT 1`
		rowFallback := s.db.QueryRowContext(ctx, queryFallback)
		if errFall := rowFallback.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.ProjectID, &sess.OwnerProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &sess.IsLocked, &sess.Permission, &rawCreated, &rawAttached); errFall != nil {
			return nil, errFall
		}
	}
	sess.CreatedAt = parseSQLiteTime(rawCreated)
	sess.LastAttachedAt = parseSQLiteTime(rawAttached)
	return &sess, nil
}

func (s *Store) ListSessions(ctx context.Context) ([]*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT ` + sessionColumns + ` FROM sessions ORDER BY is_active DESC, last_attached_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Session
	for rows.Next() {
		var sess domain.Session
		var rawCreated, rawAttached any
		if err := rows.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.ProjectID, &sess.OwnerProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &sess.IsLocked, &sess.Permission, &rawCreated, &rawAttached); err != nil {
			return nil, err
		}
		sess.CreatedAt = parseSQLiteTime(rawCreated)
		sess.LastAttachedAt = parseSQLiteTime(rawAttached)
		list = append(list, &sess)
	}
	return list, nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// --- Activity Repository Implementation ---

const activityColumns = `id, COALESCE(session_id, ''), tool_name, latency_ms, status, COALESCE(error_message, ''), created_at,
	(request_json IS NOT NULL OR response_json IS NOT NULL)`

func (s *Store) LogActivity(ctx context.Context, sessionID, toolName string, latencyMs int64, status, errorMessage string) error {
	_, err := s.RecordActivity(ctx, &domain.ActivityRecord{
		SessionID:    sessionID,
		ToolName:     toolName,
		LatencyMs:    latencyMs,
		Status:       status,
		ErrorMessage: errorMessage,
	})
	return err
}

func (s *Store) RecordActivity(ctx context.Context, rec *domain.ActivityRecord) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqBody, reqBytes, reqTrunc := payloadColumns(rec.Request)
	respBody, respBytes, respTrunc := payloadColumns(rec.Response)
	query := `INSERT INTO activity_logs (session_id, tool_name, latency_ms, status, error_message, created_at,
	              request_json, request_bytes, request_truncated, response_json, response_bytes, response_truncated)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := s.db.ExecContext(ctx, query, rec.SessionID, rec.ToolName, rec.LatencyMs, rec.Status, rec.ErrorMessage, time.Now().UTC(),
		reqBody, reqBytes, reqTrunc, respBody, respBytes, respTrunc)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListRecentActivity(ctx context.Context, limit int) ([]*domain.ActivityLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	query := `SELECT ` + activityColumns + ` FROM activity_logs ORDER BY created_at DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ActivityLog
	for rows.Next() {
		var logEntry domain.ActivityLog
		var rawCreated any
		if err := rows.Scan(&logEntry.ID, &logEntry.SessionID, &logEntry.ToolName, &logEntry.LatencyMs, &logEntry.Status, &logEntry.ErrorMessage, &rawCreated, &logEntry.HasPayload); err != nil {
			return nil, err
		}
		logEntry.CreatedAt = parseSQLiteTime(rawCreated)
		list = append(list, &logEntry)
	}
	return list, rows.Err()
}

func (s *Store) GetActivity(ctx context.Context, id int64) (*domain.ActivityDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT ` + activityColumns + `,
	              request_json, COALESCE(request_bytes, 0), COALESCE(request_truncated, 0),
	              response_json, COALESCE(response_bytes, 0), COALESCE(response_truncated, 0)
	          FROM activity_logs WHERE id = ?`
	var d domain.ActivityDetail
	var rawCreated any
	var reqBody, respBody sql.NullString
	var reqBytes, respBytes int
	var reqTrunc, respTrunc bool
	err := s.db.QueryRowContext(ctx, query, id).Scan(&d.ID, &d.SessionID, &d.ToolName, &d.LatencyMs, &d.Status, &d.ErrorMessage, &rawCreated, &d.HasPayload,
		&reqBody, &reqBytes, &reqTrunc, &respBody, &respBytes, &respTrunc)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d.CreatedAt = parseSQLiteTime(rawCreated)
	d.Request = payloadFromColumns(reqBody, reqBytes, reqTrunc)
	d.Response = payloadFromColumns(respBody, respBytes, respTrunc)
	return &d, nil
}
