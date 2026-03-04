package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment profiles",
	Long:  "List, show, switch, diff, and export .env profiles per project.",
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List environment profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getEnvManager()
		profiles, err := mgr.List()
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			fmt.Println("No environment profiles found.")
			fmt.Printf("Create profiles in .devx/envs/<name>.env\n")
			return nil
		}

		active := mgr.Active()
		for _, p := range profiles {
			marker := "  "
			if p == active {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, p)
		}
		return nil
	},
}

var envShowCmd = &cobra.Command{
	Use:   "show <profile>",
	Short: "Display variables in a profile (values masked)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reveal, _ := cmd.Flags().GetBool("reveal")
		mgr := getEnvManager()
		profile, err := mgr.Show(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Profile: %s\n", profile.Name)
		fmt.Printf("Path:    %s\n\n", profile.Path)

		for _, v := range profile.Vars {
			val := maskValue(v.Value)
			if reveal {
				val = v.Value
			}
			fmt.Printf("  %s=%s\n", v.Key, val)
		}
		return nil
	},
}

var envSwitchCmd = &cobra.Command{
	Use:   "switch <profile>",
	Short: "Copy selected profile as .env",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getEnvManager()
		if err := mgr.Switch(args[0]); err != nil {
			return err
		}
		fmt.Printf("Switched to profile %q. .env updated.\n", args[0])
		return nil
	},
}

var envDiffCmd = &cobra.Command{
	Use:   "diff <profile-a> <profile-b>",
	Short: "Diff two environment profiles",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getEnvManager()
		added, removed, changed, err := mgr.Diff(args[0], args[1])
		if err != nil {
			return err
		}

		if len(added) == 0 && len(removed) == 0 && len(changed) == 0 {
			fmt.Printf("Profiles %q and %q are identical.\n", args[0], args[1])
			return nil
		}

		fmt.Printf("Diff: %s → %s\n\n", args[0], args[1])

		if len(added) > 0 {
			fmt.Println("Added:")
			for _, k := range added {
				fmt.Printf("  + %s\n", k)
			}
		}
		if len(removed) > 0 {
			fmt.Println("Removed:")
			for _, k := range removed {
				fmt.Printf("  - %s\n", k)
			}
		}
		if len(changed) > 0 {
			fmt.Println("Changed:")
			for _, k := range changed {
				fmt.Printf("  ~ %s\n", k)
			}
		}
		return nil
	},
}

var envExportCmd = &cobra.Command{
	Use:   "export <profile>",
	Short: "Print export statements for a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := getEnvManager()
		output, err := mgr.Export(args[0])
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(envCmd)

	envShowCmd.Flags().Bool("reveal", false, "show actual values instead of masked")
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envShowCmd)
	envCmd.AddCommand(envSwitchCmd)
	envCmd.AddCommand(envDiffCmd)
	envCmd.AddCommand(envExportCmd)
}

func getEnvManager() *service.EnvManager {
	dir, _ := os.Getwd()
	return service.NewEnvManager(dir)
}

func maskValue(val string) string {
	if len(val) <= 4 {
		return strings.Repeat("*", len(val))
	}
	return val[:2] + strings.Repeat("*", len(val)-4) + val[len(val)-2:]
}
