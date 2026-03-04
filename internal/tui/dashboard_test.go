package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewDashboard(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	assert.Equal(t, 1, m.activePanel)
	assert.False(t, m.quitting)
	assert.False(t, m.showHelp)
}

func TestDashboard_KeyboardNavigation(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	m.width = 120
	m.height = 40

	// Tab switches panel (use tea.KeyTab, not rune)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm := updated.(DashboardModel)
	assert.Equal(t, 2, dm.activePanel)

	// Number keys jump to panel
	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	dm = updated.(DashboardModel)
	assert.Equal(t, 0, dm.activePanel)

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	dm = updated.(DashboardModel)
	assert.Equal(t, 2, dm.activePanel)
}

func TestDashboard_QuitKey(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	dm := updated.(DashboardModel)
	assert.True(t, dm.quitting)
	assert.NotNil(t, cmd) // should return tea.Quit
}

func TestDashboard_ToggleHelp(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	m.width = 80
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	dm := updated.(DashboardModel)
	assert.True(t, dm.showHelp)

	// Help view should contain keyboard shortcuts
	view := dm.View()
	assert.Contains(t, view, "KEYBOARD SHORTCUTS")
}

func TestDashboard_WindowSizeMsg(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	dm := updated.(DashboardModel)
	assert.Equal(t, 100, dm.width)
	assert.Equal(t, 50, dm.height)
}

func TestDashboard_DataUpdate(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	m.width = 120
	m.height = 40

	// Simulate data arriving
	projects := []*model.Project{{Name: "test-app", Language: "go"}}
	services := []*model.Service{{Name: "api", Port: 3000, Status: model.StatusUp}}

	updated, _ := m.Update(struct {
		projects []*model.Project
		services []*model.Service
		logs     []*model.LogEntry
	}{projects, services, nil})

	dm := updated.(DashboardModel)
	assert.Len(t, dm.projects, 1)
	assert.Len(t, dm.services, 1)
}

func TestDashboard_ViewLoading(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	// width=0 should show loading
	assert.Equal(t, "Loading...", m.View())
}

func TestDashboard_ViewQuitting(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	m.quitting = true
	assert.Equal(t, "", m.View())
}

func TestDashboard_ServicesUpdatedMsg(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)

	services := []*model.Service{
		{Name: "web", Port: 8080, Status: model.StatusUp},
		{Name: "db", Port: 5432, Status: model.StatusDown},
	}

	updated, _ := m.Update(servicesUpdatedMsg(services))
	dm := updated.(DashboardModel)
	assert.Len(t, dm.services, 2)
}

func TestDashboard_MoveCursor(t *testing.T) {
	db := setupTestDB(t)
	m := NewDashboard(db)
	m.activePanel = 0
	m.projects = []*model.Project{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}

	m.moveCursor(1)
	assert.Equal(t, 1, m.projCursor)

	m.moveCursor(1)
	assert.Equal(t, 2, m.projCursor)

	// Should not go past end
	m.moveCursor(1)
	assert.Equal(t, 2, m.projCursor)

	// Should not go below 0
	m.moveCursor(-1)
	m.moveCursor(-1)
	m.moveCursor(-1)
	assert.Equal(t, 0, m.projCursor)
}
