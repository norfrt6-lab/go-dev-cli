package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devx %s\n", version)
		fmt.Printf("  commit:    %s\n", commit)
		fmt.Printf("  built:     %s\n", buildTime)
		fmt.Printf("  go:        %s\n", runtime.Version())
		fmt.Printf("  os/arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
