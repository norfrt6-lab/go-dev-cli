package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Monitor local services",
	Long:  "Track and monitor running local services — check ports, health endpoints, and status.",
}

var serviceScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Auto-detect running services on common ports",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		monitor := service.NewServiceMonitor(getDB())
		results := monitor.Scan(host, nil)

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

var serviceAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Register a service to monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		healthPath, _ := cmd.Flags().GetString("health-path")

		if port == 0 {
			return fmt.Errorf("--port is required")
		}

		monitor := service.NewServiceMonitor(getDB())
		svc, err := monitor.Add(name, host, port, healthPath)
		if err != nil {
			return err
		}

		fmt.Printf("Service '%s' registered at %s:%d\n", svc.Name, svc.Host, svc.Port)
		return nil
	},
}

var serviceListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show all services with status",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		monitor := service.NewServiceMonitor(getDB())

		check, _ := cmd.Flags().GetBool("check")
		var services []*model.Service
		var err error

		if check {
			services, err = monitor.CheckAllHealth()
		} else {
			services, err = monitor.List()
		}
		if err != nil {
			return err
		}

		if len(services) == 0 {
			fmt.Println("No services registered. Use 'devx service add' to register one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tHOST:PORT\tSTATUS\tHEALTH PATH\tLAST CHECK")
		for _, s := range services {
			statusIcon := statusSymbol(s.Status)
			lastCheck := "never"
			if s.LastCheck != nil {
				lastCheck = timeAgo(*s.LastCheck)
			}
			fmt.Fprintf(w, "%s %s\t%s:%d\t%s\t%s\t%s\n",
				statusIcon, s.Name, s.Host, s.Port, s.Status, s.HealthPath, lastCheck)
		}
		return w.Flush()
	},
}

var serviceHealthCmd = &cobra.Command{
	Use:   "health [name]",
	Short: "Check health endpoint for a service",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		monitor := service.NewServiceMonitor(getDB())

		if len(args) == 0 {
			// Check all
			services, err := monitor.CheckAllHealth()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tSTATUS\tHOST:PORT")
			for _, s := range services {
				fmt.Fprintf(w, "%s %s\t%s\t%s:%d\n", statusSymbol(s.Status), s.Name, s.Status, s.Host, s.Port)
			}
			return w.Flush()
		}

		svc, healthy, latency, err := monitor.CheckHealth(args[0])
		if err != nil {
			return err
		}

		icon := statusSymbol(svc.Status)
		healthStatus := "unhealthy"
		if healthy {
			healthStatus = "healthy"
		}
		fmt.Printf("%s %s (%s:%d)\n", icon, svc.Name, svc.Host, svc.Port)
		fmt.Printf("  Status:  %s\n", svc.Status)
		fmt.Printf("  Health:  %s\n", healthStatus)
		fmt.Printf("  Latency: %s\n", latency.Round(time.Millisecond))
		return nil
	},
}

var serviceRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Unregister a service",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		monitor := service.NewServiceMonitor(getDB())
		if err := monitor.Remove(args[0]); err != nil {
			return err
		}
		fmt.Printf("Service '%s' unregistered.\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	serviceScanCmd.Flags().String("host", "127.0.0.1", "host to scan")
	serviceCmd.AddCommand(serviceScanCmd)

	serviceAddCmd.Flags().String("host", "127.0.0.1", "service host")
	serviceAddCmd.Flags().IntP("port", "p", 0, "service port (required)")
	serviceAddCmd.Flags().String("health-path", "/health", "health check endpoint path")
	serviceCmd.AddCommand(serviceAddCmd)

	serviceListCmd.Flags().BoolP("check", "c", false, "run health checks before listing")
	serviceCmd.AddCommand(serviceListCmd)

	serviceCmd.AddCommand(serviceHealthCmd)
	serviceCmd.AddCommand(serviceRemoveCmd)
}

func statusSymbol(status model.ServiceStatus) string {
	switch status {
	case model.StatusUp:
		return "[UP]"
	case model.StatusDown:
		return "[DN]"
	default:
		return "[??]"
	}
}
