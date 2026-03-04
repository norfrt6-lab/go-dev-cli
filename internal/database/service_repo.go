package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

// ServiceRepo provides CRUD operations for monitored services in SQLite.
type ServiceRepo struct {
	db *DB
}

// NewServiceRepo creates a new ServiceRepo using the given database connection.
func NewServiceRepo(db *DB) *ServiceRepo {
	return &ServiceRepo{db: db}
}

func (r *ServiceRepo) Create(svc *model.Service) error {
	result, err := r.db.conn.Exec(
		`INSERT INTO services (project_id, name, host, port, health_path, status) VALUES (?, ?, ?, ?, ?, ?)`,
		svc.ProjectID, svc.Name, svc.Host, svc.Port, svc.HealthPath, svc.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	svc.ID = id
	svc.CreatedAt = time.Now()

	return nil
}

func (r *ServiceRepo) GetByName(name string) (*model.Service, error) {
	row := r.db.conn.QueryRow(
		`SELECT id, project_id, name, host, port, health_path, status, last_check, created_at FROM services WHERE name = ?`,
		name,
	)
	return r.scanService(row)
}

func (r *ServiceRepo) List() ([]*model.Service, error) {
	rows, err := r.db.conn.Query(
		`SELECT id, project_id, name, host, port, health_path, status, last_check, created_at FROM services ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	var services []*model.Service
	for rows.Next() {
		s, err := r.scanServiceRows(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return services, rows.Err()
}

func (r *ServiceRepo) Delete(name string) error {
	result, err := r.db.conn.Exec("DELETE FROM services WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("service '%s' not found", name)
	}

	return nil
}

func (r *ServiceRepo) UpdateStatus(name string, status model.ServiceStatus) error {
	_, err := r.db.conn.Exec(
		"UPDATE services SET status = ?, last_check = datetime('now') WHERE name = ?",
		status, name,
	)
	return err
}

func (r *ServiceRepo) Exists(host string, port int) (bool, error) {
	var count int
	err := r.db.conn.QueryRow("SELECT COUNT(*) FROM services WHERE host = ? AND port = ?", host, port).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ServiceRepo) scanService(row *sql.Row) (*model.Service, error) {
	s := &model.Service{}
	var projectID sql.NullInt64
	var lastCheck sql.NullString
	var createdAt string

	err := row.Scan(&s.ID, &projectID, &s.Name, &s.Host, &s.Port, &s.HealthPath, &s.Status, &lastCheck, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service not found; run 'devx service list' to see registered services")
		}
		return nil, fmt.Errorf("failed to scan service: %w", err)
	}

	if projectID.Valid {
		pid := projectID.Int64
		s.ProjectID = &pid
	}
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if lastCheck.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastCheck.String)
		s.LastCheck = &t
	}

	return s, nil
}

func (r *ServiceRepo) scanServiceRows(rows *sql.Rows) (*model.Service, error) {
	s := &model.Service{}
	var projectID sql.NullInt64
	var lastCheck sql.NullString
	var createdAt string

	err := rows.Scan(&s.ID, &projectID, &s.Name, &s.Host, &s.Port, &s.HealthPath, &s.Status, &lastCheck, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan service: %w", err)
	}

	if projectID.Valid {
		pid := projectID.Int64
		s.ProjectID = &pid
	}
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if lastCheck.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastCheck.String)
		s.LastCheck = &t
	}

	return s, nil
}
