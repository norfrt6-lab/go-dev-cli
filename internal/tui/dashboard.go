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

// Messages
type tickMsg time.Time
type servicesUpdatedMsg []*model.Service
type logsUpdatedMsg []*model.LogEntry

// DashboardModel is the main TUI model.
type DashboardModel struct {
	db          *database.DB
	width       int
	height      int
	activePanel int // 0=projects, 1=services, 2=logs
	showHelp    bool
	quitting    bool

	projects    []*model.Project
	projCursor  int

	services    []*model.Service

	logEntries  []*model.LogEntry
	logScroll   int
}

func NewDashboard(db *database.DB) DashboardModel {
	return DashboardModel{
		db:          db,
		activePanel: 1,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadData(),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m DashboardModel) loadData() tea.Cmd {
	return func() tea.Msg {
		projSvc := service.NewProjectService(m.db)
		projects, _ := projSvc.List()

		monitor := service.NewServiceMonitor(m.db)
		services, _ := monitor.CheckAllHealth()

		logSvc := service.NewLogService(m.db)
		logs, _ := logSvc.Query(database.LogQuery{Limit: 50})

		return struct {
			projects []*model.Project
			services []*model.Service
			logs     []*model.LogEntry
		}{projects, services, logs}
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.activePanel = (m.activePanel + 1) % 3
		case "1":
			m.activePanel = 0
		case "2":
			m.activePanel = 1
		case "3":
			m.activePanel = 2
		case "?":
			m.showHelp = !m.showHelp
		case "r":
			return m, m.loadData()
		case "j", "down":
			m.moveCursor(1)
		case "k", "up":
			m.moveCursor(-1)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tea.Batch(m.refreshServices(), tickCmd())

	case struct {
		projects []*model.Project
		services []*model.Service
		logs     []*model.LogEntry
	}:
		m.projects = msg.projects
		m.services = msg.services
		m.logEntries = msg.logs

	case servicesUpdatedMsg:
		m.services = msg
	}

	return m, nil
}

func (m *DashboardModel) moveCursor(delta int) {
	switch m.activePanel {
	case 0: // projects
		m.projCursor += delta
		if m.projCursor < 0 {
			m.projCursor = 0
		}
		if m.projCursor >= len(m.projects) {
			m.projCursor = len(m.projects) - 1
		}
	case 2: // logs
		m.logScroll += delta
		if m.logScroll < 0 {
			m.logScroll = 0
		}
		maxScroll := len(m.logEntries) - 1
		if m.logScroll > maxScroll {
			m.logScroll = maxScroll
		}
	}
}

func (m DashboardModel) refreshServices() tea.Cmd {
	return func() tea.Msg {
		monitor := service.NewServiceMonitor(m.db)
		services, _ := monitor.CheckAllHealth()
		return servicesUpdatedMsg(services)
	}
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	header := StyleHeader.Width(m.width).Render(" devx dashboard")

	// Calculate panel sizes
	contentHeight := m.height - 4 // header + status bar
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 3 // border
	topHeight := contentHeight * 2 / 3
	bottomHeight := contentHeight - topHeight - 1

	// Render panels
	projectPanel := m.renderProjects(topHeight)
	servicePanel := m.renderServices(topHeight)
	logPanel := m.renderLogs(m.width-2, bottomHeight)

	// Borders for active panel
	projBorder := StyleBorder.Width(leftWidth)
	svcBorder := StyleBorder.Width(rightWidth)
	logBorder := StyleBorder.Width(m.width - 2)

	switch m.activePanel {
	case 0:
		projBorder = projBorder.BorderForeground(ColorPrimary)
	case 1:
		svcBorder = svcBorder.BorderForeground(ColorPrimary)
	default:
		logBorder = logBorder.BorderForeground(ColorPrimary)
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		projBorder.Height(topHeight).Render(projectPanel),
		svcBorder.Height(topHeight).Render(servicePanel),
	)

	bottom := logBorder.Height(bottomHeight).Render(logPanel)

	statusBar := StyleHelp.Render("  [Tab] switch panel  [j/k] navigate  [r] refresh  [?] help  [q] quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		topRow,
		bottom,
		statusBar,
	)
}

func (m DashboardModel) renderProjects(height int) string {
	title := StyleTitle.Render("PROJECTS")
	if len(m.projects) == 0 {
		return title + "\n" + StyleMuted.Render("No projects registered")
	}

	lines := make([]string, 0, len(m.projects)+1)
	lines = append(lines, title)
	for i, p := range m.projects {
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.projCursor && m.activePanel == 0 {
			prefix = "> "
			style = StyleSelected
		}
		line := fmt.Sprintf("%s%s (%s)", prefix, p.Name, p.Language)
		lines = append(lines, style.Render(line))
		if len(lines) >= height {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func (m DashboardModel) renderServices(height int) string {
	title := StyleTitle.Render("SERVICE MONITOR")
	if len(m.services) == 0 {
		return title + "\n" + StyleMuted.Render("No services registered")
	}

	lines := make([]string, 0, len(m.services)+1)
	lines = append(lines, title)
	for _, s := range m.services {
		icon := StyleStatusDown.Render("[DN]")
		if s.Status == model.StatusUp {
			icon = StyleStatusUp.Render("[UP]")
		}
		line := fmt.Sprintf("%s %-15s :%d  %s", icon, s.Name, s.Port, s.Status)
		lines = append(lines, line)
		if len(lines) >= height {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func (m DashboardModel) renderLogs(width, height int) string {
	title := StyleTitle.Render("LOGS")
	if len(m.logEntries) == 0 {
		return title + "\n" + StyleMuted.Render("No log entries")
	}

	var lines []string
	lines = append(lines, title)

	start := m.logScroll
	end := start + height - 1
	if end > len(m.logEntries) {
		end = len(m.logEntries)
	}

	for i := start; i < end; i++ {
		e := m.logEntries[i]
		ts := e.Timestamp.Format("15:04:05")
		var levelStyle lipgloss.Style
		var levelTag string
		switch e.Level {
		case model.LogError:
			levelStyle = StyleLogError
			levelTag = "ERR"
		case model.LogWarn:
			levelStyle = StyleLogWarn
			levelTag = "WRN"
		case model.LogDebug:
			levelStyle = StyleLogDebug
			levelTag = "DBG"
		default:
			levelStyle = StyleLogInfo
			levelTag = "INF"
		}

		msg := e.Message
		maxMsg := width - 20
		if maxMsg > 0 && len(msg) > maxMsg {
			msg = msg[:maxMsg-3] + "..."
		}
		line := fmt.Sprintf("%s %s %s", StyleMuted.Render(ts), levelStyle.Render(levelTag), msg)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m DashboardModel) renderHelp() string {
	help := `
  KEYBOARD SHORTCUTS
  ==================

  Tab          Switch panel
  1 / 2 / 3   Jump to Projects / Services / Logs
  j / k        Navigate up / down
  r            Refresh data
  ?            Toggle this help
  q / Ctrl+C   Quit

  Press any key to close help...
`
	return StyleBorder.
		Width(50).
		Align(lipgloss.Center).
		Render(help)
}
