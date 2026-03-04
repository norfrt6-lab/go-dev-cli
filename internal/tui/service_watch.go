package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

type ServiceWatchModel struct {
	db       *database.DB
	services []*model.Service
	width    int
	height   int
	quitting bool
}

func NewServiceWatch(db *database.DB) ServiceWatchModel {
	return ServiceWatchModel{db: db}
}

func (m ServiceWatchModel) Init() tea.Cmd {
	return tea.Batch(m.refresh(), tickCmd())
}

func (m ServiceWatchModel) refresh() tea.Cmd {
	return func() tea.Msg {
		monitor := service.NewServiceMonitor(m.db)
		services, _ := monitor.CheckAllHealth()
		return servicesUpdatedMsg(services)
	}
}

func (m ServiceWatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "r":
			return m, m.refresh()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tea.Batch(m.refresh(), tickCmd())

	case servicesUpdatedMsg:
		m.services = msg
	}

	return m, nil
}

func (m ServiceWatchModel) View() string {
	if m.quitting {
		return ""
	}

	header := StyleHeader.Width(m.width).Render(" Service Monitor")

	rows := make([]string, 0, len(m.services)+4)
	rows = append(rows, "")
	rows = append(rows, fmt.Sprintf("  %-20s %-15s %-8s %s", "NAME", "HOST:PORT", "STATUS", "LAST CHECK"))
	rows = append(rows, "  "+strings.Repeat("─", 65))

	for _, s := range m.services {
		var statusStyle lipgloss.Style
		var icon string
		if s.Status == model.StatusUp {
			statusStyle = StyleStatusUp
			icon = "[UP]"
		} else {
			statusStyle = StyleStatusDown
			icon = "[DN]"
		}

		lastCheck := "never"
		if s.LastCheck != nil {
			lastCheck = time.Since(*s.LastCheck).Round(time.Second).String() + " ago"
		}

		row := fmt.Sprintf("  %s %-17s %s:%-8d %s %s",
			statusStyle.Render(icon),
			s.Name,
			s.Host, s.Port,
			statusStyle.Render(string(s.Status)),
			StyleMuted.Render(lastCheck),
		)
		rows = append(rows, row)
	}

	if len(m.services) == 0 {
		rows = append(rows, "")
		rows = append(rows, StyleMuted.Render("  No services registered. Use 'devx service add' to register one."))
	}

	statusBar := StyleHelp.Render("  [r] refresh  [q] quit  |  Auto-refreshes every 5s")

	content := strings.Join(rows, "\n")

	// Pad to fill screen
	usedLines := len(rows) + 3 // header + status
	for i := usedLines; i < m.height; i++ {
		content += "\n"
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}
