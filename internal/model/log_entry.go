// Package model defines the core domain types used across the application.
package model

import "time"

// LogLevel represents the severity of a log entry.
type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// LogEntry represents a single parsed log line with source, level, and timestamp.
type LogEntry struct {
	ID        int64    `json:"id"`
	Source    string   `json:"source"`
	Level    LogLevel `json:"level"`
	Message  string   `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metadata string   `json:"metadata"`
}
