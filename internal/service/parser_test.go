package service

import (
	"testing"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLine_Empty(t *testing.T) {
	entry := ParseLogLine("", "test")
	assert.Nil(t, entry)
}

func TestParseLogLine_PlainText(t *testing.T) {
	entry := ParseLogLine("Server started successfully", "app.log")
	require.NotNil(t, entry)
	assert.Equal(t, "app.log", entry.Source)
	assert.Equal(t, model.LogInfo, entry.Level)
	assert.Equal(t, "Server started successfully", entry.Message)
}

func TestParseLogLine_PlainTextWithLevel(t *testing.T) {
	tests := []struct {
		line     string
		expected model.LogLevel
	}{
		{"ERROR: connection refused", model.LogError},
		{"fatal: something crashed", model.LogError},
		{"warning: deprecated function", model.LogWarn},
		{"debug: variable x = 5", model.LogDebug},
		{"just a normal message", model.LogInfo},
	}

	for _, tt := range tests {
		entry := ParseLogLine(tt.line, "test")
		require.NotNil(t, entry)
		assert.Equal(t, tt.expected, entry.Level, "line: %s", tt.line)
	}
}

func TestParseLogLine_JSON(t *testing.T) {
	line := `{"level":"error","msg":"connection failed","time":"2024-01-15T10:30:45Z","host":"db-1"}`
	entry := ParseLogLine(line, "api")
	require.NotNil(t, entry)
	assert.Equal(t, "api", entry.Source)
	assert.Equal(t, model.LogError, entry.Level)
	assert.Equal(t, "connection failed", entry.Message)
	assert.Contains(t, entry.Metadata, "db-1")
}

func TestParseLogLine_JSONWithMessageKey(t *testing.T) {
	line := `{"level":"info","message":"request handled","status":200}`
	entry := ParseLogLine(line, "svc")
	require.NotNil(t, entry)
	assert.Equal(t, "request handled", entry.Message)
	assert.Equal(t, model.LogInfo, entry.Level)
}

func TestParseLogLine_JSONNoMessage(t *testing.T) {
	line := `{"level":"debug","key":"value"}`
	entry := ParseLogLine(line, "test")
	require.NotNil(t, entry)
	assert.Equal(t, model.LogDebug, entry.Level)
	// Should use original line when no message field
	assert.Equal(t, line, entry.Message)
}

func TestParseLogLine_CommonFormat(t *testing.T) {
	line := "2024-01-15 10:30:45 [INFO] Server listening on port 8080"
	entry := ParseLogLine(line, "server.log")
	require.NotNil(t, entry)
	assert.Equal(t, model.LogInfo, entry.Level)
	assert.Equal(t, "Server listening on port 8080", entry.Message)
	assert.Equal(t, 2024, entry.Timestamp.Year())
}

func TestParseLogLine_CommonFormatWarning(t *testing.T) {
	line := "2024-03-10T14:22:33 WARNING Disk space low"
	entry := ParseLogLine(line, "system.log")
	require.NotNil(t, entry)
	assert.Equal(t, model.LogWarn, entry.Level)
	assert.Equal(t, "Disk space low", entry.Message)
}

func TestNormalizeLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected model.LogLevel
	}{
		{"debug", model.LogDebug},
		{"DEBUG", model.LogDebug},
		{"trace", model.LogDebug},
		{"info", model.LogInfo},
		{"INFO", model.LogInfo},
		{"information", model.LogInfo},
		{"warn", model.LogWarn},
		{"WARNING", model.LogWarn},
		{"error", model.LogError},
		{"ERR", model.LogError},
		{"fatal", model.LogError},
		{"CRITICAL", model.LogError},
		{"panic", model.LogError},
		{"unknown", model.LogInfo},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeLevel(tt.input), "input: %s", tt.input)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		input string
		year  int
	}{
		{"2024-01-15T10:30:45Z", 2024},
		{"2024-01-15 10:30:45", 2024},
		{"2024/01/15 10:30:45", 2024},
		{"invalid", 0},
	}

	for _, tt := range tests {
		ts := parseTimestamp(tt.input)
		if tt.year > 0 {
			assert.Equal(t, tt.year, ts.Year(), "input: %s", tt.input)
		} else {
			assert.True(t, ts.IsZero(), "input: %s should be zero", tt.input)
		}
	}
}
