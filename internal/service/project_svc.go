package service

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/norfrt6-lab/go-dev-cli/internal/config"
	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

// ProjectService manages project registration, listing, and editor launching.
type ProjectService struct {
	repo *database.ProjectRepo
}

// NewProjectService creates a ProjectService backed by the given database.
func NewProjectService(db *database.DB) *ProjectService {
	return &ProjectService{
		repo: database.NewProjectRepo(db),
	}
}

func (s *ProjectService) Add(name, path, language, description string) (*model.Project, error) {
	project := &model.Project{
		Name:        name,
		Path:        path,
		Language:    language,
		Description: description,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, fmt.Errorf("failed to register project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) List() ([]*model.Project, error) {
	return s.repo.List()
}

func (s *ProjectService) Get(name string) (*model.Project, error) {
	return s.repo.GetByName(name)
}

func (s *ProjectService) Remove(name string) error {
	return s.repo.Delete(name)
}

func (s *ProjectService) Open(name string) error {
	project, err := s.repo.GetByName(name)
	if err != nil {
		return err
	}

	editor := config.GetEditor()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", editor, project.Path)
	default:
		cmd = exec.Command(editor, project.Path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open editor '%s': %w", editor, err)
	}

	if err := s.repo.UpdateLastOpened(name); err != nil {
		return fmt.Errorf("project opened but failed to update timestamp: %w", err)
	}

	fmt.Printf("Opening '%s' in %s...\n", project.Name, editor)
	return nil
}
