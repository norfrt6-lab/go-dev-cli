package database

import "fmt"

var migrations = []string{
	// Migration 1: Core tables
	`CREATE TABLE IF NOT EXISTS projects (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT    NOT NULL UNIQUE,
		path        TEXT    NOT NULL,
		language    TEXT    DEFAULT 'unknown',
		description TEXT    DEFAULT '',
		created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
		updated_at  TEXT    NOT NULL DEFAULT (datetime('now')),
		last_opened TEXT
	);

	CREATE TABLE IF NOT EXISTS services (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id  INTEGER REFERENCES projects(id) ON DELETE SET NULL,
		name        TEXT    NOT NULL,
		host        TEXT    NOT NULL DEFAULT '127.0.0.1',
		port        INTEGER NOT NULL,
		health_path TEXT    DEFAULT '/health',
		status      TEXT    NOT NULL DEFAULT 'unknown',
		last_check  TEXT,
		created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
		UNIQUE(host, port)
	);

	CREATE TABLE IF NOT EXISTS log_entries (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		source     TEXT    NOT NULL,
		level      TEXT    DEFAULT 'info',
		message    TEXT    NOT NULL,
		timestamp  TEXT    NOT NULL DEFAULT (datetime('now')),
		metadata   TEXT    DEFAULT '{}'
	);

	CREATE INDEX IF NOT EXISTS idx_log_entries_source    ON log_entries(source);
	CREATE INDEX IF NOT EXISTS idx_log_entries_level     ON log_entries(level);
	CREATE INDEX IF NOT EXISTS idx_log_entries_timestamp ON log_entries(timestamp DESC);

	CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT    NOT NULL DEFAULT (datetime('now'))
	);`,

	// Migration 2: FTS5 full-text search for log entries
	`CREATE VIRTUAL TABLE IF NOT EXISTS log_entries_fts USING fts5(
		source,
		level,
		message,
		content='log_entries',
		content_rowid='id'
	);

	CREATE TRIGGER IF NOT EXISTS log_entries_ai AFTER INSERT ON log_entries BEGIN
		INSERT INTO log_entries_fts(rowid, source, level, message)
		VALUES (new.id, new.source, new.level, new.message);
	END;

	CREATE TRIGGER IF NOT EXISTS log_entries_ad AFTER DELETE ON log_entries BEGIN
		INSERT INTO log_entries_fts(log_entries_fts, rowid, source, level, message)
		VALUES ('delete', old.id, old.source, old.level, old.message);
	END;`,
}

func (d *DB) Migrate() error {
	currentVersion := d.getCurrentVersion()

	for i := currentVersion; i < len(migrations); i++ {
		tx, err := d.conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", i+1, err)
		}

		if _, err := tx.Exec(migrations[i]); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", i+1); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", i+1, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", i+1, err)
		}
	}

	return nil
}

func (d *DB) getCurrentVersion() int {
	// Create schema_migrations if it doesn't exist (bootstrap)
	_, _ = d.conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT    NOT NULL DEFAULT (datetime('now'))
	)`)

	var version int
	err := d.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0
	}
	return version
}
