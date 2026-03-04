package database

import (
	"testing"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestProjectRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	project := &model.Project{
		Name:     "test-project",
		Path:     "/tmp/test-project",
		Language: "go",
	}

	err := repo.Create(project)
	require.NoError(t, err)
	assert.NotZero(t, project.ID)
}

func TestProjectRepo_Create_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	p1 := &model.Project{Name: "dup", Path: "/tmp/a", Language: "go"}
	p2 := &model.Project{Name: "dup", Path: "/tmp/b", Language: "go"}

	require.NoError(t, repo.Create(p1))
	err := repo.Create(p2)
	assert.Error(t, err)
}

func TestProjectRepo_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	original := &model.Project{
		Name:        "my-app",
		Path:        "/home/user/my-app",
		Language:    "node",
		Description: "A test app",
	}
	require.NoError(t, repo.Create(original))

	found, err := repo.GetByName("my-app")
	require.NoError(t, err)
	assert.Equal(t, "my-app", found.Name)
	assert.Equal(t, "/home/user/my-app", found.Path)
	assert.Equal(t, "node", found.Language)
	assert.Equal(t, "A test app", found.Description)
}

func TestProjectRepo_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	_, err := repo.GetByName("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProjectRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	projects := []model.Project{
		{Name: "alpha", Path: "/tmp/alpha", Language: "go"},
		{Name: "beta", Path: "/tmp/beta", Language: "python"},
		{Name: "gamma", Path: "/tmp/gamma", Language: "rust"},
	}

	for i := range projects {
		require.NoError(t, repo.Create(&projects[i]))
	}

	list, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)
	// Should be ordered by name
	assert.Equal(t, "alpha", list[0].Name)
	assert.Equal(t, "beta", list[1].Name)
	assert.Equal(t, "gamma", list[2].Name)
}

func TestProjectRepo_List_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	list, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestProjectRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	p := &model.Project{Name: "to-delete", Path: "/tmp/del", Language: "go"}
	require.NoError(t, repo.Create(p))

	err := repo.Delete("to-delete")
	require.NoError(t, err)

	_, err = repo.GetByName("to-delete")
	assert.Error(t, err)
}

func TestProjectRepo_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	err := repo.Delete("ghost")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProjectRepo_UpdateLastOpened(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProjectRepo(db)

	p := &model.Project{Name: "opener", Path: "/tmp/open", Language: "go"}
	require.NoError(t, repo.Create(p))

	err := repo.UpdateLastOpened("opener")
	require.NoError(t, err)

	found, err := repo.GetByName("opener")
	require.NoError(t, err)
	assert.NotNil(t, found.LastOpened)
}
