package service

import (
	"fmt"
	"time"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

type ServiceMonitor struct {
	repo    *database.ServiceRepo
	scanner *Scanner
}

func NewServiceMonitor(db *database.DB) *ServiceMonitor {
	return &ServiceMonitor{
		repo:    database.NewServiceRepo(db),
		scanner: NewScanner(2 * time.Second),
	}
}

func (m *ServiceMonitor) Add(name, host string, port int, healthPath string) (*model.Service, error) {
	exists, err := m.repo.Exists(host, port)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("service already registered on %s:%d", host, port)
	}

	if healthPath == "" {
		healthPath = "/health"
	}

	svc := &model.Service{
		Name:       name,
		Host:       host,
		Port:       port,
		HealthPath: healthPath,
		Status:     model.StatusUnknown,
	}

	if err := m.repo.Create(svc); err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	return svc, nil
}

func (m *ServiceMonitor) List() ([]*model.Service, error) {
	return m.repo.List()
}

func (m *ServiceMonitor) Get(name string) (*model.Service, error) {
	return m.repo.GetByName(name)
}

func (m *ServiceMonitor) Remove(name string) error {
	return m.repo.Delete(name)
}

func (m *ServiceMonitor) CheckHealth(name string) (*model.Service, bool, time.Duration, error) {
	svc, err := m.repo.GetByName(name)
	if err != nil {
		return nil, false, 0, err
	}

	healthy, latency := CheckHealth(svc.Host, svc.Port, svc.HealthPath, 5*time.Second)

	var status model.ServiceStatus
	if healthy {
		status = model.StatusUp
	} else {
		// Check if port is at least open
		result := m.scanner.ScanPort(svc.Host, svc.Port)
		if result.Open {
			status = model.StatusUp
		} else {
			status = model.StatusDown
		}
	}

	if err := m.repo.UpdateStatus(name, status); err != nil {
		return svc, healthy, latency, fmt.Errorf("health checked but failed to update status: %w", err)
	}

	svc.Status = status
	return svc, healthy, latency, nil
}

func (m *ServiceMonitor) CheckAllHealth() ([]*model.Service, error) {
	services, err := m.repo.List()
	if err != nil {
		return nil, err
	}

	for _, svc := range services {
		result := m.scanner.ScanPort(svc.Host, svc.Port)
		var status model.ServiceStatus
		if result.Open {
			status = model.StatusUp
		} else {
			status = model.StatusDown
		}
		_ = m.repo.UpdateStatus(svc.Name, status)
		svc.Status = status
	}

	return services, nil
}

func (m *ServiceMonitor) Scan(host string, ports []int) []PortResult {
	if len(ports) == 0 {
		return m.scanner.ScanCommonPorts(host)
	}
	return m.scanner.ScanPorts(host, ports)
}
