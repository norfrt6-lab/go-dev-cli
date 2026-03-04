package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestLogService(t *testing.T) *LogService {
	t.Helper()
	db, err := database.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewLogService(db)
}

func TestLogService_StoreAndQuery(t *testing.T) {
	svc := setupTestLogService(t)

	entry := &model.LogEntry{
		Source:    "test-app",
		Level:    model.LogInfo,
		Message:  "hello world",
		Timestamp: time.Now(),
		Metadata:  "{}",
	}

	err := svc.Store(entry)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)

	entries, err := svc.Query(database.LogQuery{Source: "test-app"})
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "hello world", entries[0].Message)
}

func TestLogService_StoreBatch(t *testing.T) {
	svc := setupTestLogService(t)

	entries := []*model.LogEntry{
		{Source: "app", Level: model.LogInfo, Message: "msg 1", Timestamp: time.Now(), Metadata: "{}"},
		{Source: "app", Level: model.LogWarn, Message: "msg 2", Timestamp: time.Now(), Metadata: "{}"},
	}

	err := svc.StoreBatch(entries)
	require.NoError(t, err)

	count, err := svc.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestLogService_Search(t *testing.T) {
	svc := setupTestLogService(t)

	err := svc.Store(&model.LogEntry{
		Source: "api", Level: model.LogError, Message: "database timeout error",
		Timestamp: time.Now(), Metadata: "{}",
	})
	require.NoError(t, err)

	err = svc.Store(&model.LogEntry{
		Source: "api", Level: model.LogInfo, Message: "request completed",
		Timestamp: time.Now(), Metadata: "{}",
	})
	require.NoError(t, err)

	entries, err := svc.Search("database", 10)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Message, "database")
}

func TestLogService_Clear(t *testing.T) {
	svc := setupTestLogService(t)

	err := svc.Store(&model.LogEntry{Source: "a", Level: model.LogInfo, Message: "m", Timestamp: time.Now(), Metadata: "{}"})
	require.NoError(t, err)
	err = svc.Store(&model.LogEntry{Source: "b", Level: model.LogInfo, Message: "m", Timestamp: time.Now(), Metadata: "{}"})
	require.NoError(t, err)

	// Clear specific source
	deleted, err := svc.Clear("a")
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// Clear all
	deleted, err = svc.Clear("")
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}

func TestLogService_Sources(t *testing.T) {
	svc := setupTestLogService(t)

	err := svc.Store(&model.LogEntry{Source: "api", Level: model.LogInfo, Message: "m", Timestamp: time.Now(), Metadata: "{}"})
	require.NoError(t, err)
	err = svc.Store(&model.LogEntry{Source: "worker", Level: model.LogInfo, Message: "m", Timestamp: time.Now(), Metadata: "{}"})
	require.NoError(t, err)

	sources, err := svc.Sources()
	require.NoError(t, err)
	assert.Len(t, sources, 2)
}

func TestTailer_ReadExisting(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	content := `2024-01-15 10:30:45 [INFO] Server started
2024-01-15 10:30:46 [ERROR] Connection refused
plain log line
`
	require.NoError(t, os.WriteFile(logFile, []byte(content), 0o600))

	tailer := NewTailer(logFile, "test.log")
	entries, err := tailer.ReadExisting()
	require.NoError(t, err)
	assert.Len(t, entries, 3)

	assert.Equal(t, model.LogInfo, entries[0].Level)
	assert.Equal(t, model.LogError, entries[1].Level)
	assert.Equal(t, model.LogInfo, entries[2].Level)
}

func TestTailer_ReadExisting_FileNotFound(t *testing.T) {
	tailer := NewTailer("/nonexistent/file.log", "test")
	_, err := tailer.ReadExisting()
	assert.Error(t, err)
}
