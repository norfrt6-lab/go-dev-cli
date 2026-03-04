package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary = lipgloss.Color("#7C3AED")
	ColorSuccess = lipgloss.Color("#10B981")
	ColorWarning = lipgloss.Color("#F59E0B")
	ColorDanger  = lipgloss.Color("#EF4444")
	ColorMuted   = lipgloss.Color("#6B7280")
	ColorSurface = lipgloss.Color("#1F2937")
	ColorText    = lipgloss.Color("#F9FAFB")

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 1)

	StyleStatusUp = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	StyleStatusDown = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	StyleLogInfo = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleLogWarn = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleLogError = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	StyleLogDebug = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSelected = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
