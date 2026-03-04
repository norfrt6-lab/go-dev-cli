package model

import "time"

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

type LogEntry struct {
	ID        int64    `json:"id"`
	Source    string   `json:"source"`
	Level    LogLevel `json:"level"`
	Message  string   `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metadata string   `json:"metadata"`
}
