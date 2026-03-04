package service

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

// ParseLogLine attempts to parse a log line into a structured LogEntry.
// It tries JSON format first, then common log patterns, then falls back to plain text.
func ParseLogLine(line, source string) *model.LogEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Try JSON log format
	if entry := parseJSON(line, source); entry != nil {
		return entry
	}

	// Try common log patterns
	if entry := parseCommonFormat(line, source); entry != nil {
		return entry
	}

	// Fallback: plain text
	return &model.LogEntry{
		Source:    source,
		Level:    detectLevel(line),
		Message:   line,
		Timestamp: time.Now(),
		Metadata:  "{}",
	}
}

func parseJSON(line, source string) *model.LogEntry {
	if len(line) == 0 || line[0] != '{' {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return nil
	}

	entry := &model.LogEntry{
		Source:    source,
		Level:    model.LogInfo,
		Timestamp: time.Now(),
		Metadata:  "{}",
	}

	// Extract message
	for _, key := range []string{"msg", "message", "text"} {
		if v, ok := data[key]; ok {
			entry.Message = toString(v)
			delete(data, key)
			break
		}
	}

	// Extract level
	for _, key := range []string{"level", "lvl", "severity"} {
		if v, ok := data[key]; ok {
			entry.Level = normalizeLevel(toString(v))
			delete(data, key)
			break
		}
	}

	// Extract timestamp
	for _, key := range []string{"time", "timestamp", "ts", "@timestamp"} {
		if v, ok := data[key]; ok {
			if t := parseTimestamp(toString(v)); !t.IsZero() {
				entry.Timestamp = t
			}
			delete(data, key)
			break
		}
	}

	// Remaining fields go to metadata
	if len(data) > 0 {
		if meta, err := json.Marshal(data); err == nil {
			entry.Metadata = string(meta)
		}
	}

	if entry.Message == "" {
		entry.Message = line
	}

	return entry
}

// Common log format: 2024-01-15 10:30:45 [INFO] Some message here
var commonLogPattern = regexp.MustCompile(
	`^(\d{4}[-/]\d{2}[-/]\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)\s+\[?(DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL|TRACE)\]?\s+(.+)`,
)

func parseCommonFormat(line, source string) *model.LogEntry {
	matches := commonLogPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	entry := &model.LogEntry{
		Source:    source,
		Level:    normalizeLevel(matches[2]),
		Message:  matches[3],
		Timestamp: time.Now(),
		Metadata:  "{}",
	}

	if t := parseTimestamp(matches[1]); !t.IsZero() {
		entry.Timestamp = t
	}

	return entry
}

func normalizeLevel(level string) model.LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "trace":
		return model.LogDebug
	case "info", "information":
		return model.LogInfo
	case "warn", "warning":
		return model.LogWarn
	case "error", "err", "fatal", "critical", "panic":
		return model.LogError
	default:
		return model.LogInfo
	}
}

func detectLevel(line string) model.LogLevel {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal") || strings.Contains(lower, "panic"):
		return model.LogError
	case strings.Contains(lower, "warn"):
		return model.LogWarn
	case strings.Contains(lower, "debug") || strings.Contains(lower, "trace"):
		return model.LogDebug
	default:
		return model.LogInfo
	}
}

var timestampFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02 15:04:05.000",
}

func parseTimestamp(s string) time.Time {
	for _, format := range timestampFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), `"`)
	}
}
