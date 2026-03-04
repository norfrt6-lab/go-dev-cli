package service

import (
	"testing"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestService(t *testing.T) *ProjectService {
	t.Helper()
	db, err := database.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewProjectService(db)
}

func TestProjectService_AddAndGet(t *testing.T) {
	svc := setupTestService(t)

	project, err := svc.Add("my-app", "/tmp/my-app", "go", "A Go application")
	require.NoError(t, err)
	assert.Equal(t, "my-app", project.Name)
	assert.Equal(t, "/tmp/my-app", project.Path)

	found, err := svc.Get("my-app")
	require.NoError(t, err)
	assert.Equal(t, project.Name, found.Name)
	assert.Equal(t, "A Go application", found.Description)
}

func TestProjectService_List(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Add("proj-a", "/tmp/a", "go", "")
	require.NoError(t, err)
	_, err = svc.Add("proj-b", "/tmp/b", "node", "")
	require.NoError(t, err)

	projects, err := svc.List()
	require.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestProjectService_Remove(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Add("removable", "/tmp/rm", "python", "")
	require.NoError(t, err)

	err = svc.Remove("removable")
	require.NoError(t, err)

	_, err = svc.Get("removable")
	assert.Error(t, err)
}

func TestProjectService_Remove_NotFound(t *testing.T) {
	svc := setupTestService(t)

	err := svc.Remove("nonexistent")
	assert.Error(t, err)
}

func TestProjectService_Add_Duplicate(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Add("dup", "/tmp/dup1", "go", "")
	require.NoError(t, err)

	_, err = svc.Add("dup", "/tmp/dup2", "go", "")
	assert.Error(t, err)
}
