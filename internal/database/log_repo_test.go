package database

import (
	"testing"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLogEntry(source, level, message string) *model.LogEntry {
	return &model.LogEntry{
		Source:    source,
		Level:    model.LogLevel(level),
		Message:  message,
		Timestamp: time.Now(),
		Metadata:  "{}",
	}
}

func TestLogRepo_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	entry := newLogEntry("app.log", "info", "server started")
	err := repo.Insert(entry)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)
}

func TestLogRepo_InsertBatch(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	entries := []*model.LogEntry{
		newLogEntry("app.log", "info", "line 1"),
		newLogEntry("app.log", "warn", "line 2"),
		newLogEntry("app.log", "error", "line 3"),
	}

	err := repo.InsertBatch(entries)
	require.NoError(t, err)

	for _, e := range entries {
		assert.NotZero(t, e.ID)
	}

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestLogRepo_Query_All(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("a", "info", "msg 1")))
	require.NoError(t, repo.Insert(newLogEntry("b", "error", "msg 2")))

	entries, err := repo.Query(LogQuery{})
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestLogRepo_Query_BySource(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("api", "info", "request")))
	require.NoError(t, repo.Insert(newLogEntry("worker", "info", "job done")))

	entries, err := repo.Query(LogQuery{Source: "api"})
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "request", entries[0].Message)
}

func TestLogRepo_Query_ByLevel(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("app", "info", "ok")))
	require.NoError(t, repo.Insert(newLogEntry("app", "error", "fail")))
	require.NoError(t, repo.Insert(newLogEntry("app", "error", "crash")))

	entries, err := repo.Query(LogQuery{Level: model.LogError})
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestLogRepo_Query_WithLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Insert(newLogEntry("app", "info", "msg")))
	}

	entries, err := repo.Query(LogQuery{Limit: 3})
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestLogRepo_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("api", "error", "database connection timeout")))
	require.NoError(t, repo.Insert(newLogEntry("api", "info", "request handled successfully")))
	require.NoError(t, repo.Insert(newLogEntry("worker", "error", "database deadlock detected")))

	entries, err := repo.Search("database", 10)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestLogRepo_Search_NoResults(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("app", "info", "hello world")))

	entries, err := repo.Search("nonexistent", 10)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestLogRepo_DeleteBySource(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("keep", "info", "keep this")))
	require.NoError(t, repo.Insert(newLogEntry("remove", "info", "remove this")))

	deleted, err := repo.DeleteBySource("remove")
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestLogRepo_DeleteAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("a", "info", "1")))
	require.NoError(t, repo.Insert(newLogEntry("b", "info", "2")))

	deleted, err := repo.DeleteAll()
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted)

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestLogRepo_Sources(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	require.NoError(t, repo.Insert(newLogEntry("api", "info", "msg")))
	require.NoError(t, repo.Insert(newLogEntry("worker", "info", "msg")))
	require.NoError(t, repo.Insert(newLogEntry("api", "error", "msg")))

	sources, err := repo.Sources()
	require.NoError(t, err)
	assert.Len(t, sources, 2)
	assert.Contains(t, sources, "api")
	assert.Contains(t, sources, "worker")
}

func TestLogRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLogRepo(db)

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	require.NoError(t, repo.Insert(newLogEntry("app", "info", "msg")))

	count, err = repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
