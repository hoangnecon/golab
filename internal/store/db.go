package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides thread-safe persistence for profiles, sessions, and activity logs.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// DefaultDBPath returns ~/.golab/state.db.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".golab")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.db"), nil
}

// New creates or opens the SQLite database.
func New(dbPath string) (*Store, error) {
	if dbPath == "" {
		var err error
		dbPath, err = DefaultDBPath()
		if err != nil {
			return nil, fmt.Errorf("resolve default db path: %w", err)
		}
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return s, nil
}

func parseSQLiteTime(val any) time.Time {
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS profiles (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		account_email TEXT,
		tier TEXT DEFAULT 'free',
		user_data_dir TEXT NOT NULL,
		cdp_port INTEGER NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
		notebook_url TEXT NOT NULL,
		notebook_id TEXT,
		tab_id TEXT,
		is_active BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_attached_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT,
		tool_name TEXT NOT NULL,
		latency_ms INTEGER,
		status TEXT NOT NULL,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
	`
	_, err := s.db.Exec(schema)
	return err
}

// --- Profile Operations ---

func (s *Store) CreateProfile(p *Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	query := `INSERT INTO profiles (id, name, account_email, tier, user_data_dir, cdp_port, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, p.ID, p.Name, p.AccountEmail, p.Tier, p.UserDataDir, p.CDPPort, p.CreatedAt)
	return err
}

func (s *Store) GetProfile(id string) (*Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, COALESCE(account_email, ''), tier, user_data_dir, cdp_port, created_at
	          FROM profiles WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var p Profile
	var rawCreated any
	if err := row.Scan(&p.ID, &p.Name, &p.AccountEmail, &p.Tier, &p.UserDataDir, &p.CDPPort, &rawCreated); err != nil {
		return nil, err
	}
	p.CreatedAt = parseSQLiteTime(rawCreated)
	return &p, nil
}

func (s *Store) ListProfiles() ([]*Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, COALESCE(account_email, ''), tier, user_data_dir, cdp_port, created_at
	          FROM profiles ORDER BY created_at ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Profile
	for rows.Next() {
		var p Profile
		var rawCreated any
		if err := rows.Scan(&p.ID, &p.Name, &p.AccountEmail, &p.Tier, &p.UserDataDir, &p.CDPPort, &rawCreated); err != nil {
			return nil, err
		}
		p.CreatedAt = parseSQLiteTime(rawCreated)
		list = append(list, &p)
	}
	return list, nil
}

func (s *Store) DeleteProfile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM profiles WHERE id = ?`, id)
	return err
}

func (s *Store) GetNextAvailableCDPPort(startPort int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT cdp_port FROM profiles ORDER BY cdp_port ASC`)
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

// --- Session Operations ---

func (s *Store) SaveSession(sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = time.Now().UTC()
	}
	sess.LastAttachedAt = time.Now().UTC()

	query := `INSERT INTO sessions (id, name, profile_id, notebook_url, notebook_id, tab_id, is_active, created_at, last_attached_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	          ON CONFLICT(id) DO UPDATE SET
	            name = excluded.name,
	            profile_id = excluded.profile_id,
	            notebook_url = excluded.notebook_url,
	            notebook_id = excluded.notebook_id,
	            tab_id = excluded.tab_id,
	            is_active = excluded.is_active,
	            last_attached_at = excluded.last_attached_at`
	_, err := s.db.Exec(query, sess.ID, sess.Name, sess.ProfileID, sess.NotebookURL, sess.NotebookID, sess.TabID, sess.IsActive, sess.CreatedAt, sess.LastAttachedAt)
	return err
}

func (s *Store) SetActiveSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE sessions SET is_active = 0`); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE sessions SET is_active = 1, last_attached_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetSession(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, profile_id, notebook_url, COALESCE(notebook_id, ''), COALESCE(tab_id, ''), is_active, created_at, last_attached_at
	          FROM sessions WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var sess Session
	var rawCreated, rawAttached any
	if err := row.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &rawCreated, &rawAttached); err != nil {
		return nil, err
	}
	sess.CreatedAt = parseSQLiteTime(rawCreated)
	sess.LastAttachedAt = parseSQLiteTime(rawAttached)
	return &sess, nil
}

func (s *Store) GetActiveSession() (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, profile_id, notebook_url, COALESCE(notebook_id, ''), COALESCE(tab_id, ''), is_active, created_at, last_attached_at
	          FROM sessions WHERE is_active = 1 LIMIT 1`
	row := s.db.QueryRow(query)

	var sess Session
	var rawCreated, rawAttached any
	if err := row.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &rawCreated, &rawAttached); err != nil {
		queryFallback := `SELECT id, name, profile_id, notebook_url, COALESCE(notebook_id, ''), COALESCE(tab_id, ''), is_active, created_at, last_attached_at
		                  FROM sessions ORDER BY last_attached_at DESC LIMIT 1`
		rowFallback := s.db.QueryRow(queryFallback)
		if errFall := rowFallback.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &rawCreated, &rawAttached); errFall != nil {
			return nil, errFall
		}
	}
	sess.CreatedAt = parseSQLiteTime(rawCreated)
	sess.LastAttachedAt = parseSQLiteTime(rawAttached)
	return &sess, nil
}

func (s *Store) ListSessions() ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, profile_id, notebook_url, COALESCE(notebook_id, ''), COALESCE(tab_id, ''), is_active, created_at, last_attached_at
	          FROM sessions ORDER BY is_active DESC, last_attached_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Session
	for rows.Next() {
		var sess Session
		var rawCreated, rawAttached any
		if err := rows.Scan(&sess.ID, &sess.Name, &sess.ProfileID, &sess.NotebookURL, &sess.NotebookID, &sess.TabID, &sess.IsActive, &rawCreated, &rawAttached); err != nil {
			return nil, err
		}
		sess.CreatedAt = parseSQLiteTime(rawCreated)
		sess.LastAttachedAt = parseSQLiteTime(rawAttached)
		list = append(list, &sess)
	}
	return list, nil
}

func (s *Store) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// --- Activity Log Operations ---

func (s *Store) LogActivity(sessionID, toolName string, latencyMs int64, status, errorMessage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO activity_logs (session_id, tool_name, latency_ms, status, error_message, created_at)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, sessionID, toolName, latencyMs, status, errorMessage, time.Now().UTC())
	return err
}

func (s *Store) ListRecentActivity(limit int) ([]*ActivityLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, COALESCE(session_id, ''), tool_name, latency_ms, status, COALESCE(error_message, ''), created_at
	          FROM activity_logs ORDER BY created_at DESC LIMIT ?`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ActivityLog
	for rows.Next() {
		var logEntry ActivityLog
		var rawCreated any
		if err := rows.Scan(&logEntry.ID, &logEntry.SessionID, &logEntry.ToolName, &logEntry.LatencyMs, &logEntry.Status, &logEntry.ErrorMessage, &rawCreated); err != nil {
			return nil, err
		}
		logEntry.CreatedAt = parseSQLiteTime(rawCreated)
		list = append(list, &logEntry)
	}
	return list, nil
}
