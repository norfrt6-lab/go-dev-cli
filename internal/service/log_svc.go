package service

import (
	"context"
	"fmt"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

type LogService struct {
	repo *database.LogRepo
}

func NewLogService(db *database.DB) *LogService {
	return &LogService{
		repo: database.NewLogRepo(db),
	}
}

func (s *LogService) Store(entry *model.LogEntry) error {
	return s.repo.Insert(entry)
}

func (s *LogService) StoreBatch(entries []*model.LogEntry) error {
	return s.repo.InsertBatch(entries)
}

func (s *LogService) Query(q database.LogQuery) ([]*model.LogEntry, error) {
	return s.repo.Query(q)
}

func (s *LogService) Search(query string, limit int) ([]*model.LogEntry, error) {
	return s.repo.Search(query, limit)
}

func (s *LogService) Sources() ([]string, error) {
	return s.repo.Sources()
}

func (s *LogService) Count() (int64, error) {
	return s.repo.Count()
}

func (s *LogService) Clear(source string) (int64, error) {
	if source == "" {
		return s.repo.DeleteAll()
	}
	return s.repo.DeleteBySource(source)
}

func (s *LogService) Cleanup(retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return s.repo.DeleteOlderThan(retentionDays)
}

// TailFile reads existing lines and then follows a file for new entries.
// It calls onEntry for each parsed log entry. Cancel context to stop.
func (s *LogService) TailFile(ctx context.Context, path string, follow bool, onEntry func(*model.LogEntry)) error {
	tailer := NewTailer(path, path)

	// Read existing lines
	entries, err := tailer.ReadExisting()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		onEntry(entry)
	}

	if !follow {
		return nil
	}

	// Follow new lines
	ch := make(chan *model.LogEntry, 100)

	go func() {
		for entry := range ch {
			if err := s.repo.Insert(entry); err != nil {
				continue
			}
			onEntry(entry)
		}
	}()

	err = tailer.Follow(ctx, ch)
	close(ch)

	if err != nil {
		return fmt.Errorf("tail stopped: %w", err)
	}
	return nil
}
