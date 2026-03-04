package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/norfrt6-lab/go-dev-cli/internal/scaffold"
	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage development projects",
	Long:  "Register, list, scaffold, and manage your local development projects.",
}

var projectAddCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Register an existing project directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath := "."
		if len(args) > 0 {
			projectPath = args[0]
		}

		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("path not found: %s\n  Ensure the directory exists before running 'devx project add'", absPath)
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s\n  Provide a project directory, not a file", absPath)
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = filepath.Base(absPath)
		}

		lang, _ := cmd.Flags().GetString("lang")
		if lang == "" {
			lang = detectLanguage(absPath)
		}

		desc, _ := cmd.Flags().GetString("description")

		svc := service.NewProjectService(getDB())
		project, err := svc.Add(name, absPath, lang, desc)
		if err != nil {
			return err
		}

		fmt.Printf("Project '%s' registered at %s\n", project.Name, project.Path)
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all registered projects",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := service.NewProjectService(getDB())
		projects, err := svc.List()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Println("No projects registered. Use 'devx project add' to register one.")
			return nil
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(projects)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tLANGUAGE\tPATH\tLAST OPENED")
		for _, p := range projects {
			lastOpened := "never"
			if p.LastOpened != nil {
				lastOpened = timeAgo(*p.LastOpened)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Language, p.Path, lastOpened)
		}
		return w.Flush()
	},
}

var projectRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Unregister a project (does not delete files)",
	Aliases: []string{"rm"},
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := service.NewProjectService(getDB())
		if err := svc.Remove(args[0]); err != nil {
			return err
		}
		fmt.Printf("Project '%s' unregistered.\n", args[0])
		return nil
	},
}

var projectInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := service.NewProjectService(getDB())
		project, err := svc.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Name:        %s\n", project.Name)
		fmt.Printf("Path:        %s\n", project.Path)
		fmt.Printf("Language:    %s\n", project.Language)
		if project.Description != "" {
			fmt.Printf("Description: %s\n", project.Description)
		}
		fmt.Printf("Created:     %s\n", project.CreatedAt.Format(time.RFC3339))
		if project.LastOpened != nil {
			fmt.Printf("Last Opened: %s (%s)\n", project.LastOpened.Format(time.RFC3339), timeAgo(*project.LastOpened))
		}

		exists := "yes"
		if _, err := os.Stat(project.Path); os.IsNotExist(err) {
			exists = "no (path missing)"
		}
		fmt.Printf("Exists:      %s\n", exists)

		return nil
	},
}

var projectOpenCmd = &cobra.Command{
	Use:   "open <name>",
	Short: "Open project in editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := service.NewProjectService(getDB())
		return svc.Open(args[0])
	},
}

var projectInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Scaffold a new project from template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		tmpl, _ := cmd.Flags().GetString("template")
		outputDir, _ := cmd.Flags().GetString("output")

		if tmpl == "" {
			return fmt.Errorf("template is required. Available: %s", strings.Join(scaffold.AvailableTemplates(), ", "))
		}

		if outputDir == "" {
			outputDir = filepath.Join(".", name)
		}

		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("failed to resolve output path: %w", err)
		}

		s := scaffold.New()
		if err := s.Generate(tmpl, name, absPath); err != nil {
			return err
		}

		// Register the new project
		svc := service.NewProjectService(getDB())
		lang := scaffold.TemplateLanguage(tmpl)
		if _, err := svc.Add(name, absPath, lang, ""); err != nil {
			fmt.Printf("Warning: project created but not registered: %v\n", err)
		}

		fmt.Printf("Project '%s' scaffolded at %s (template: %s)\n", name, absPath, tmpl)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)

	projectAddCmd.Flags().StringP("name", "n", "", "project name (default: directory name)")
	projectAddCmd.Flags().StringP("lang", "l", "", "language (auto-detected if omitted)")
	projectAddCmd.Flags().StringP("description", "d", "", "project description")
	projectCmd.AddCommand(projectAddCmd)

	projectListCmd.Flags().String("output", "table", "output format (table, json)")
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectRemoveCmd)
	projectCmd.AddCommand(projectInfoCmd)
	projectCmd.AddCommand(projectOpenCmd)

	projectInitCmd.Flags().StringP("template", "t", "", "template name (go-cli, go-api, node-api, react-app, python-api)")
	projectInitCmd.Flags().StringP("output", "o", "", "output directory (default: ./<name>)")
	projectCmd.AddCommand(projectInitCmd)
}

func detectLanguage(path string) string {
	indicators := map[string]string{
		"go.mod":         "go",
		"package.json":   "node",
		"requirements.txt": "python",
		"pyproject.toml": "python",
		"Cargo.toml":     "rust",
		"pom.xml":        "java",
		"build.gradle":   "java",
		"Gemfile":        "ruby",
		"composer.json":  "php",
	}

	for file, lang := range indicators {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			return lang
		}
	}
	return "unknown"
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
