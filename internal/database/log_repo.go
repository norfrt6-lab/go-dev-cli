package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

// LogRepo provides CRUD and full-text search operations for log entries in SQLite.
type LogRepo struct {
	db *DB
}

// NewLogRepo creates a new LogRepo using the given database connection.
func NewLogRepo(db *DB) *LogRepo {
	return &LogRepo{db: db}
}

func (r *LogRepo) Insert(entry *model.LogEntry) error {
	result, err := r.db.conn.Exec(
		`INSERT INTO log_entries (source, level, message, timestamp, metadata) VALUES (?, ?, ?, ?, ?)`,
		entry.Source, entry.Level, entry.Message, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to insert log entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = id
	return nil
}

func (r *LogRepo) InsertBatch(entries []*model.LogEntry) error {
	tx, err := r.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO log_entries (source, level, message, timestamp, metadata) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		result, err := stmt.Exec(entry.Source, entry.Level, entry.Message, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Metadata)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to insert log entry: %w", err)
		}
		id, _ := result.LastInsertId()
		entry.ID = id
	}

	return tx.Commit()
}

// LogQuery specifies filters for querying log entries.
type LogQuery struct {
	Source string
	Level  model.LogLevel
	Since  *time.Time
	Until  *time.Time
	Limit  int
	Offset int
}

func (r *LogRepo) Query(q LogQuery) ([]*model.LogEntry, error) {
	var conditions []string
	var args []interface{}

	if q.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, q.Source)
	}
	if q.Level != "" {
		conditions = append(conditions, "level = ?")
		args = append(args, q.Level)
	}
	if q.Since != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, q.Since.Format("2006-01-02 15:04:05"))
	}
	if q.Until != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, q.Until.Format("2006-01-02 15:04:05"))
	}

	query := "SELECT id, source, level, message, timestamp, metadata FROM log_entries"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY timestamp DESC"

	if q.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, q.Limit)
	}
	if q.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, q.Offset)
	}

	rows, err := r.db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query log entries: %w", err)
	}
	defer rows.Close()

	var entries []*model.LogEntry
	for rows.Next() {
		e := &model.LogEntry{}
		var ts string
		if err := rows.Scan(&e.ID, &e.Source, &e.Level, &e.Message, &ts, &e.Metadata); err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}
		e.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (r *LogRepo) Search(query string, limit int) ([]*model.LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.conn.Query(
		`SELECT le.id, le.source, le.level, le.message, le.timestamp, le.metadata
		 FROM log_entries_fts fts
		 JOIN log_entries le ON le.id = fts.rowid
		 WHERE log_entries_fts MATCH ?
		 ORDER BY rank
		 LIMIT ?`,
		query, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search log entries: %w", err)
	}
	defer rows.Close()

	var entries []*model.LogEntry
	for rows.Next() {
		e := &model.LogEntry{}
		var ts string
		if err := rows.Scan(&e.ID, &e.Source, &e.Level, &e.Message, &ts, &e.Metadata); err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}
		e.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (r *LogRepo) DeleteOlderThan(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	result, err := r.db.conn.Exec("DELETE FROM log_entries WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old log entries: %w", err)
	}
	return result.RowsAffected()
}

func (r *LogRepo) DeleteBySource(source string) (int64, error) {
	result, err := r.db.conn.Exec("DELETE FROM log_entries WHERE source = ?", source)
	if err != nil {
		return 0, fmt.Errorf("failed to delete log entries: %w", err)
	}
	return result.RowsAffected()
}

func (r *LogRepo) DeleteAll() (int64, error) {
	result, err := r.db.conn.Exec("DELETE FROM log_entries")
	if err != nil {
		return 0, fmt.Errorf("failed to clear log entries: %w", err)
	}
	return result.RowsAffected()
}

func (r *LogRepo) Count() (int64, error) {
	var count int64
	err := r.db.conn.QueryRow("SELECT COUNT(*) FROM log_entries").Scan(&count)
	return count, err
}

func (r *LogRepo) Sources() ([]string, error) {
	rows, err := r.db.conn.Query("SELECT DISTINCT source FROM log_entries ORDER BY source")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}
