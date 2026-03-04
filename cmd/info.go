package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("System Information")
		fmt.Println("==================")
		fmt.Printf("OS:        %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Go:        %s\n", runtime.Version())
		fmt.Printf("CPUs:      %d\n", runtime.NumCPU())

		// Check common tools
		tools := []struct {
			name string
			cmd  string
			args []string
		}{
			{"Docker", "docker", []string{"--version"}},
			{"Node.js", "node", []string{"--version"}},
			{"Python", "python3", []string{"--version"}},
			{"Git", "git", []string{"--version"}},
		}

		fmt.Println()
		fmt.Println("Installed Tools")
		fmt.Println("---------------")
		for _, tool := range tools {
			ver := getToolVersion(tool.cmd, tool.args)
			if ver != "" {
				fmt.Printf("%-10s %s\n", tool.name+":", ver)
			} else {
				// Try alternate command names
				if tool.cmd == "python3" {
					ver = getToolVersion("python", tool.args)
					if ver != "" {
						fmt.Printf("%-10s %s\n", tool.name+":", ver)
						continue
					}
				}
				fmt.Printf("%-10s not found\n", tool.name+":")
			}
		}

		return nil
	},
}

var infoPortsCmd = &cobra.Command{
	Use:   "ports",
	Short: "List all listening ports with process info",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		scanner := service.NewScanner(1 * time.Second)
		results := scanner.ScanCommonPorts(host)

		if len(results) == 0 {
			fmt.Println("No open ports found on common development ports.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PORT\tSTATUS\tLATENCY")
		for _, r := range results {
			fmt.Fprintf(w, "%d\topen\t%s\n", r.Port, r.Latency.Round(time.Millisecond))
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoPortsCmd.Flags().String("host", "127.0.0.1", "host to scan")
	infoCmd.AddCommand(infoPortsCmd)
}

func getToolVersion(cmd string, args []string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	version := strings.TrimSpace(string(out))
	// Clean up common prefixes
	version = strings.TrimPrefix(version, "Docker version ")
	version = strings.TrimPrefix(version, "git version ")
	version = strings.TrimPrefix(version, "Python ")
	version = strings.TrimPrefix(version, "v")
	// Take first line only
	if idx := strings.IndexByte(version, '\n'); idx > 0 {
		version = version[:idx]
	}
	return version
}
