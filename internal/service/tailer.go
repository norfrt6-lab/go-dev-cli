package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

// Tailer follows a file and emits parsed log entries.
type Tailer struct {
	path     string
	source   string
	pollRate time.Duration
}

func NewTailer(path, source string) *Tailer {
	if source == "" {
		source = path
	}
	return &Tailer{
		path:     path,
		source:   source,
		pollRate: 250 * time.Millisecond,
	}
}

// ReadExisting reads all existing lines from the file and returns them as log entries.
func (t *Tailer) ReadExisting() ([]*model.LogEntry, error) {
	f, err := os.Open(t.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var entries []*model.LogEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		entry := ParseLogLine(scanner.Text(), t.source)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

// Follow tails the file, sending new log entries to the channel.
// It starts from the end of the file and polls for new lines.
// Cancel the context to stop following.
func (t *Tailer) Follow(ctx context.Context, entries chan<- *model.LogEntry) error {
	f, err := os.Open(t.path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Seek to end
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(t.pollRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				entry := ParseLogLine(line, t.source)
				if entry != nil {
					select {
					case entries <- entry:
					case <-ctx.Done():
						return nil
					}
				}
			}
		}
	}
}
