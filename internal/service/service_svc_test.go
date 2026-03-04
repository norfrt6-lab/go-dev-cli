package service

import (
	"testing"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestMonitor(t *testing.T) *ServiceMonitor {
	t.Helper()
	db, err := database.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewServiceMonitor(db)
}

func TestServiceMonitor_Add(t *testing.T) {
	m := setupTestMonitor(t)

	svc, err := m.Add("api", "127.0.0.1", 8080, "/health")
	require.NoError(t, err)
	assert.Equal(t, "api", svc.Name)
	assert.Equal(t, "127.0.0.1", svc.Host)
	assert.Equal(t, 8080, svc.Port)
	assert.Equal(t, "/health", svc.HealthPath)
}

func TestServiceMonitor_Add_DefaultHealthPath(t *testing.T) {
	m := setupTestMonitor(t)

	svc, err := m.Add("api", "127.0.0.1", 8080, "")
	require.NoError(t, err)
	assert.Equal(t, "/health", svc.HealthPath)
}

func TestServiceMonitor_Add_DuplicateHostPort(t *testing.T) {
	m := setupTestMonitor(t)

	_, err := m.Add("svc-a", "127.0.0.1", 3000, "/health")
	require.NoError(t, err)

	_, err = m.Add("svc-b", "127.0.0.1", 3000, "/health")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestServiceMonitor_List(t *testing.T) {
	m := setupTestMonitor(t)

	_, err := m.Add("svc-1", "127.0.0.1", 3000, "")
	require.NoError(t, err)
	_, err = m.Add("svc-2", "127.0.0.1", 4000, "")
	require.NoError(t, err)

	services, err := m.List()
	require.NoError(t, err)
	assert.Len(t, services, 2)
}

func TestServiceMonitor_Get(t *testing.T) {
	m := setupTestMonitor(t)

	_, err := m.Add("my-svc", "localhost", 5000, "/healthz")
	require.NoError(t, err)

	svc, err := m.Get("my-svc")
	require.NoError(t, err)
	assert.Equal(t, "my-svc", svc.Name)
	assert.Equal(t, 5000, svc.Port)
}

func TestServiceMonitor_Get_NotFound(t *testing.T) {
	m := setupTestMonitor(t)

	_, err := m.Get("nonexistent")
	assert.Error(t, err)
}

func TestServiceMonitor_Remove(t *testing.T) {
	m := setupTestMonitor(t)

	_, err := m.Add("removable", "127.0.0.1", 9090, "")
	require.NoError(t, err)

	err = m.Remove("removable")
	require.NoError(t, err)

	_, err = m.Get("removable")
	assert.Error(t, err)
}

func TestServiceMonitor_Remove_NotFound(t *testing.T) {
	m := setupTestMonitor(t)

	err := m.Remove("ghost")
	assert.Error(t, err)
}

func TestServiceMonitor_Scan(t *testing.T) {
	m := setupTestMonitor(t)

	// Scan with empty ports should use common ports (returns slice, possibly empty)
	results := m.Scan("127.0.0.1", nil)
	_ = results // may be nil if no ports open

	// Scan with specific closed ports
	results = m.Scan("127.0.0.1", []int{59990, 59991})
	assert.Empty(t, results)
}
