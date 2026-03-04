// Package database provides SQLite-based persistence for projects, services, and log entries.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database connection with migration support.
type DB struct {
	conn *sql.DB
	path string
}

// Open creates or opens a SQLite database at the given path with WAL mode enabled.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn, path: path}, nil
}

// OpenMemory creates an in-memory SQLite database with migrations applied. Useful for testing.
func OpenMemory() (*DB, error) {
	conn, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn, path: ":memory:"}
	if err := db.Migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// Conn returns the underlying sql.DB connection.
func (d *DB) Conn() *sql.DB {
	return d.conn
}
