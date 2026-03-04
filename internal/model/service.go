package model

import "time"

// ServiceStatus represents the health state of a monitored service.
type ServiceStatus string

const (
	StatusUp      ServiceStatus = "up"
	StatusDown    ServiceStatus = "down"
	StatusUnknown ServiceStatus = "unknown"
)

// Service represents a registered network service to monitor.
type Service struct {
	ID         int64         `json:"id"`
	ProjectID  *int64        `json:"project_id,omitempty"`
	Name       string        `json:"name"`
	Host       string        `json:"host"`
	Port       int           `json:"port"`
	HealthPath string        `json:"health_path"`
	Status     ServiceStatus `json:"status"`
	LastCheck  *time.Time    `json:"last_check,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}
