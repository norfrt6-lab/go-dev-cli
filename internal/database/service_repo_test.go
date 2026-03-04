package database

import (
	"testing"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	svc := &model.Service{
		Name:       "api-server",
		Host:       "127.0.0.1",
		Port:       8080,
		HealthPath: "/health",
		Status:     model.StatusUnknown,
	}

	err := repo.Create(svc)
	require.NoError(t, err)
	assert.NotZero(t, svc.ID)
}

func TestServiceRepo_Create_DuplicateHostPort(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	s1 := &model.Service{Name: "svc-a", Host: "127.0.0.1", Port: 3000, Status: model.StatusUnknown}
	s2 := &model.Service{Name: "svc-b", Host: "127.0.0.1", Port: 3000, Status: model.StatusUnknown}

	require.NoError(t, repo.Create(s1))
	err := repo.Create(s2)
	assert.Error(t, err)
}

func TestServiceRepo_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	svc := &model.Service{
		Name:       "web-app",
		Host:       "localhost",
		Port:       3000,
		HealthPath: "/healthz",
		Status:     model.StatusUnknown,
	}
	require.NoError(t, repo.Create(svc))

	found, err := repo.GetByName("web-app")
	require.NoError(t, err)
	assert.Equal(t, "web-app", found.Name)
	assert.Equal(t, "localhost", found.Host)
	assert.Equal(t, 3000, found.Port)
	assert.Equal(t, "/healthz", found.HealthPath)
}

func TestServiceRepo_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	_, err := repo.GetByName("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	services := []model.Service{
		{Name: "alpha", Host: "127.0.0.1", Port: 3000, Status: model.StatusUnknown},
		{Name: "beta", Host: "127.0.0.1", Port: 4000, Status: model.StatusUnknown},
		{Name: "gamma", Host: "127.0.0.1", Port: 5000, Status: model.StatusUnknown},
	}

	for i := range services {
		require.NoError(t, repo.Create(&services[i]))
	}

	list, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)
	assert.Equal(t, "alpha", list[0].Name)
	assert.Equal(t, "beta", list[1].Name)
	assert.Equal(t, "gamma", list[2].Name)
}

func TestServiceRepo_List_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	list, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestServiceRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	svc := &model.Service{Name: "removable", Host: "127.0.0.1", Port: 9090, Status: model.StatusUnknown}
	require.NoError(t, repo.Create(svc))

	err := repo.Delete("removable")
	require.NoError(t, err)

	_, err = repo.GetByName("removable")
	assert.Error(t, err)
}

func TestServiceRepo_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	err := repo.Delete("ghost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceRepo_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	svc := &model.Service{Name: "status-test", Host: "127.0.0.1", Port: 7070, Status: model.StatusUnknown}
	require.NoError(t, repo.Create(svc))

	err := repo.UpdateStatus("status-test", model.StatusUp)
	require.NoError(t, err)

	found, err := repo.GetByName("status-test")
	require.NoError(t, err)
	assert.Equal(t, model.StatusUp, found.Status)
	assert.NotNil(t, found.LastCheck)
}

func TestServiceRepo_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepo(db)

	svc := &model.Service{Name: "exists-test", Host: "127.0.0.1", Port: 6060, Status: model.StatusUnknown}
	require.NoError(t, repo.Create(svc))

	exists, err := repo.Exists("127.0.0.1", 6060)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.Exists("127.0.0.1", 9999)
	require.NoError(t, err)
	assert.False(t, exists)
}
