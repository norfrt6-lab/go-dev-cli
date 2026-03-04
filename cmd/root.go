package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/norfrt6-lab/go-dev-cli/internal/config"
	"github.com/norfrt6-lab/go-dev-cli/internal/database"
)

var (
	cfgFile string
	verbose bool
	noColor bool
	db      *database.DB
)

var rootCmd = &cobra.Command{
	Use:   "devx",
	Short: "Developer productivity CLI",
	Long: `devx — A developer productivity CLI with interactive TUI.

Manage local dev environments, scaffold projects, monitor running services,
and aggregate logs. All from your terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB init for commands that don't need it
		switch cmd.Name() {
		case "version", "completion", "help", "bash", "zsh", "fish", "powershell":
			return nil
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		dbPath := filepath.Join(cfg.DataDir, "devx.db")
		db, err = database.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		if err := db.Migrate(); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default ~/.devx/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))
}

func getDB() *database.DB {
	return db
}
