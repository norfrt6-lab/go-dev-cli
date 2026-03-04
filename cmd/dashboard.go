package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/norfrt6-lab/go-dev-cli/internal/tui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Interactive TUI dashboard",
	Long:  "Launch a fullscreen interactive dashboard with project, service, and log panels.",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := tui.NewDashboard(getDB())
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("dashboard error: %w", err)
		}
		return nil
	},
}

var serviceWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Fullscreen service monitor",
	Long:  "Watch registered services in real-time with auto-refresh.",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := tui.NewServiceWatch(getDB())
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("service watch error: %w", err)
		}
		return nil
	},
}

var logsViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Fullscreen log viewer",
	Long:  "Browse stored log entries with filtering and search.",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := tui.NewLogViewer(getDB())
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("log viewer error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	serviceCmd.AddCommand(serviceWatchCmd)
	logsCmd.AddCommand(logsViewCmd)
}
