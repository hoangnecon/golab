package sqlite

import (
	"database/sql"
	"fmt"
)

type columnDef struct {
	name string
	decl string
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

	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		notebook_url TEXT NOT NULL UNIQUE,
		owner_profile_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
		project_id TEXT,
		owner_profile_id TEXT,
		notebook_url TEXT NOT NULL,
		notebook_id TEXT,
		tab_id TEXT,
		is_active BOOLEAN DEFAULT 0,
		is_locked BOOLEAN DEFAULT 0,
		permission TEXT DEFAULT 'unknown',
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		request_json TEXT,
		request_bytes INTEGER,
		request_truncated BOOLEAN DEFAULT 0,
		response_json TEXT,
		response_bytes INTEGER,
		response_truncated BOOLEAN DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	if err := s.addMissingColumns("sessions", []columnDef{
		{"project_id", "TEXT"},
		{"owner_profile_id", "TEXT"},
		{"is_locked", "BOOLEAN DEFAULT 0"},
		{"permission", "TEXT DEFAULT 'unknown'"},
	}); err != nil {
		return err
	}

	return s.addMissingColumns("activity_logs", []columnDef{
		{"request_json", "TEXT"},
		{"request_bytes", "INTEGER"},
		{"request_truncated", "BOOLEAN DEFAULT 0"},
		{"response_json", "TEXT"},
		{"response_bytes", "INTEGER"},
		{"response_truncated", "BOOLEAN DEFAULT 0"},
	})
}

func (s *Store) addMissingColumns(table string, cols []columnDef) error {
	rows, err := s.db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return fmt.Errorf("read %s columns: %w", table, err)
	}
	have := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		have[name] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, c := range cols {
		if have[c.name] {
			continue
		}
		if _, err := s.db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, c.name, c.decl)); err != nil {
			return fmt.Errorf("add %s.%s: %w", table, c.name, err)
		}
	}
	return nil
}

func payloadColumns(p *domainPayload) (any, any, bool) {
	if p == nil || len(p.Body) == 0 {
		return nil, nil, false
	}
	return string(p.Body), p.Bytes, p.Truncated
}

func payloadFromColumns(body sql.NullString, size int, truncated bool) *domainPayload {
	if !body.Valid {
		return nil
	}
	return &domainPayload{Body: []byte(body.String), Bytes: size, Truncated: truncated}
}
