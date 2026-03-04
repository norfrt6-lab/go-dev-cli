package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

type LogViewerModel struct {
	db         *database.DB
	entries    []*model.LogEntry
	width      int
	height     int
	scroll     int
	filter     string
	filterMode bool
	quitting   bool
}

func NewLogViewer(db *database.DB) LogViewerModel {
	return LogViewerModel{db: db}
}

func (m LogViewerModel) Init() tea.Cmd {
	return m.loadLogs()
}

func (m LogViewerModel) loadLogs() tea.Cmd {
	return func() tea.Msg {
		svc := service.NewLogService(m.db)
		entries, _ := svc.Query(database.LogQuery{Limit: 500})
		return logsUpdatedMsg(entries)
	}
}

func (m LogViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterMode {
			switch msg.String() {
			case "enter", "esc":
				m.filterMode = false
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.filter += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "r":
			return m, m.loadLogs()
		case "/", "f":
			m.filterMode = true
		case "c":
			m.filter = ""
		case "j", "down":
			m.scroll++
		case "k", "up":
			if m.scroll > 0 {
				m.scroll--
			}
		case "g":
			m.scroll = 0
		case "G":
			m.scroll = len(m.filteredEntries()) - 1
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case logsUpdatedMsg:
		m.entries = msg
		m.scroll = 0
	}

	return m, nil
}

func (m LogViewerModel) filteredEntries() []*model.LogEntry {
	if m.filter == "" {
		return m.entries
	}

	lower := strings.ToLower(m.filter)
	var filtered []*model.LogEntry
	for _, e := range m.entries {
		if strings.Contains(strings.ToLower(e.Message), lower) ||
			strings.Contains(strings.ToLower(string(e.Level)), lower) ||
			strings.Contains(strings.ToLower(e.Source), lower) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func (m LogViewerModel) View() string {
	if m.quitting {
		return ""
	}

	header := StyleHeader.Width(m.width).Render(" Log Viewer")

	entries := m.filteredEntries()

	// Filter bar
	filterBar := ""
	if m.filterMode {
		filterBar = fmt.Sprintf("  Filter: %s█", m.filter)
	} else if m.filter != "" {
		filterBar = StyleMuted.Render(fmt.Sprintf("  Filter: %s  [c] clear", m.filter))
	}

	// Log content
	contentHeight := m.height - 4
	if filterBar != "" {
		contentHeight--
	}

	var lines []string
	start := m.scroll
	end := start + contentHeight
	if end > len(entries) {
		end = len(entries)
	}
	if start >= len(entries) {
		start = 0
		end = contentHeight
		if end > len(entries) {
			end = len(entries)
		}
	}

	for i := start; i < end; i++ {
		e := entries[i]
		ts := StyleMuted.Render(e.Timestamp.Format("15:04:05"))

		var levelStr string
		switch e.Level {
		case model.LogError:
			levelStr = StyleLogError.Render("ERR")
		case model.LogWarn:
			levelStr = StyleLogWarn.Render("WRN")
		case model.LogDebug:
			levelStr = StyleLogDebug.Render("DBG")
		default:
			levelStr = StyleLogInfo.Render("INF")
		}

		source := StyleMuted.Render(e.Source)
		msg := e.Message
		maxMsg := m.width - 40
		if maxMsg > 0 && len(msg) > maxMsg {
			msg = msg[:maxMsg-3] + "..."
		}

		lines = append(lines, fmt.Sprintf("  %s %s %s %s", ts, levelStr, source, msg))
	}

	if len(entries) == 0 {
		lines = append(lines, StyleMuted.Render("  No log entries found."))
	}

	content := strings.Join(lines, "\n")

	info := fmt.Sprintf("  %d entries", len(entries))
	if len(entries) > 0 {
		info += fmt.Sprintf("  |  showing %d-%d", start+1, end)
	}
	statusBar := StyleHelp.Render(info + "  |  [j/k] scroll  [/] filter  [r] refresh  [q] quit")

	parts := []string{header}
	if filterBar != "" {
		parts = append(parts, filterBar)
	}
	parts = append(parts, content, statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
