package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailableTemplates(t *testing.T) {
	templates := AvailableTemplates()
	assert.Contains(t, templates, "go-cli")
	assert.Contains(t, templates, "go-api")
	assert.Contains(t, templates, "node-api")
	assert.Contains(t, templates, "react-app")
	assert.Contains(t, templates, "python-api")
	assert.Len(t, templates, 5)
}

func TestTemplateLanguage(t *testing.T) {
	tests := []struct {
		tmpl     string
		expected string
	}{
		{"go-cli", "go"},
		{"go-api", "go"},
		{"node-api", "node"},
		{"react-app", "node"},
		{"python-api", "python"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.tmpl, func(t *testing.T) {
			assert.Equal(t, tt.expected, TemplateLanguage(tt.tmpl))
		})
	}
}

func TestScaffolder_Generate_GoCLI(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "my-tool")

	s := New()
	err := s.Generate("go-cli", "my-tool", outputDir)
	require.NoError(t, err)

	// Verify files exist
	expectedFiles := []string{"main.go", "go.mod", "Makefile", ".gitignore"}
	for _, f := range expectedFiles {
		path := filepath.Join(outputDir, f)
		_, err := os.Stat(path)
		assert.NoError(t, err, "file %s should exist", f)
	}

	// Verify template variable substitution
	content, err := os.ReadFile(filepath.Join(outputDir, "main.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "my-tool")
}

func TestScaffolder_Generate_NodeAPI(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "my-api")

	s := New()
	err := s.Generate("node-api", "my-api", outputDir)
	require.NoError(t, err)

	// Verify package.json was created with the project name
	content, err := os.ReadFile(filepath.Join(outputDir, "package.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"name": "my-api"`)
}

func TestScaffolder_Generate_InvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	s := New()
	err := s.Generate("invalid-template", "test", dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

func TestScaffolder_Generate_NonEmptyDir(t *testing.T) {
	dir := t.TempDir()
	// Create a file in the dir to make it non-empty
	os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("data"), 0o644)

	s := New()
	err := s.Generate("go-cli", "test", dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not empty")
}
