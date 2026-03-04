package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates
var templateFS embed.FS

type Scaffolder struct{}

func New() *Scaffolder {
	return &Scaffolder{}
}

type TemplateData struct {
	Name      string
	Module    string
	Year      int
}

func AvailableTemplates() []string {
	return []string{"go-cli", "go-api", "node-api", "react-app", "python-api"}
}

func TemplateLanguage(tmpl string) string {
	switch {
	case strings.HasPrefix(tmpl, "go-"):
		return "go"
	case strings.HasPrefix(tmpl, "node-"):
		return "node"
	case strings.HasPrefix(tmpl, "react-"):
		return "node"
	case strings.HasPrefix(tmpl, "python-"):
		return "python"
	default:
		return "unknown"
	}
}

func (s *Scaffolder) Generate(tmpl, name, outputDir string) error {
	if !isValidTemplate(tmpl) {
		return fmt.Errorf("unknown template '%s'. Available: %s", tmpl, strings.Join(AvailableTemplates(), ", "))
	}

	if _, err := os.Stat(outputDir); err == nil {
		entries, _ := os.ReadDir(outputDir)
		if len(entries) > 0 {
			return fmt.Errorf("output directory '%s' is not empty", outputDir)
		}
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data := TemplateData{
		Name:   name,
		Module: fmt.Sprintf("github.com/%s/%s", "user", name),
		Year:   2026,
	}

	templateDir := fmt.Sprintf("templates/%s", tmpl)
	return s.processDir(templateDir, outputDir, data)
}

func (s *Scaffolder) processDir(srcDir, destDir string, data TemplateData) error {
	entries, err := templateFS.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read template directory '%s': %w", srcDir, err)
	}

	for _, entry := range entries {
		// Use path.Join (forward slashes) for embed.FS, filepath.Join for OS paths
		srcPath := path.Join(srcDir, entry.Name())
		destName := strings.TrimSuffix(entry.Name(), ".tmpl")
		destPath := filepath.Join(destDir, destName)

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
			if err := s.processDir(srcPath, destPath, data); err != nil {
				return err
			}
			continue
		}

		content, err := templateFS.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read template file '%s': %w", srcPath, err)
		}

		if strings.HasSuffix(entry.Name(), ".tmpl") {
			tmpl, err := template.New(entry.Name()).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template '%s': %w", srcPath, err)
			}

			f, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file '%s': %w", destPath, err)
			}
			defer f.Close()

			if err := tmpl.Execute(f, data); err != nil {
				return fmt.Errorf("failed to execute template '%s': %w", srcPath, err)
			}
		} else {
			if err := os.WriteFile(destPath, content, 0o644); err != nil {
				return fmt.Errorf("failed to write file '%s': %w", destPath, err)
			}
		}
	}

	return nil
}

func isValidTemplate(tmpl string) bool {
	for _, t := range AvailableTemplates() {
		if t == tmpl {
			return true
		}
	}
	return false
}
