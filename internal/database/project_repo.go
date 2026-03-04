package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

// ProjectRepo provides CRUD operations for projects in SQLite.
type ProjectRepo struct {
	db *DB
}

// NewProjectRepo creates a new ProjectRepo using the given database connection.
func NewProjectRepo(db *DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(project *model.Project) error {
	result, err := r.db.conn.Exec(
		`INSERT INTO projects (name, path, language, description) VALUES (?, ?, ?, ?)`,
		project.Name, project.Path, project.Language, project.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	project.ID = id
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	return nil
}

func (r *ProjectRepo) GetByName(name string) (*model.Project, error) {
	row := r.db.conn.QueryRow(
		`SELECT id, name, path, language, description, created_at, updated_at, last_opened FROM projects WHERE name = ?`,
		name,
	)
	return r.scanProject(row)
}

func (r *ProjectRepo) List() ([]*model.Project, error) {
	rows, err := r.db.conn.Query(
		`SELECT id, name, path, language, description, created_at, updated_at, last_opened FROM projects ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		p, err := r.scanProjectRows(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

func (r *ProjectRepo) Delete(name string) error {
	result, err := r.db.conn.Exec("DELETE FROM projects WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("project '%s' not found", name)
	}

	return nil
}

func (r *ProjectRepo) UpdateLastOpened(name string) error {
	_, err := r.db.conn.Exec(
		"UPDATE projects SET last_opened = datetime('now'), updated_at = datetime('now') WHERE name = ?",
		name,
	)
	return err
}

func (r *ProjectRepo) scanProject(row *sql.Row) (*model.Project, error) {
	p := &model.Project{}
	var createdAt, updatedAt string
	var lastOpened sql.NullString

	err := row.Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.Description, &createdAt, &updatedAt, &lastOpened)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found; run 'devx project list' to see registered projects")
		}
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if lastOpened.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastOpened.String)
		p.LastOpened = &t
	}

	return p, nil
}

func (r *ProjectRepo) scanProjectRows(rows *sql.Rows) (*model.Project, error) {
	p := &model.Project{}
	var createdAt, updatedAt string
	var lastOpened sql.NullString

	err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Language, &p.Description, &createdAt, &updatedAt, &lastOpened)
	if err != nil {
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if lastOpened.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastOpened.String)
		p.LastOpened = &t
	}

	return p, nil
}
