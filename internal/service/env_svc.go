package service

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/norfrt6-lab/go-dev-cli/internal/model"
)

const envDir = ".devx/envs"

// EnvManager handles environment profile operations.
type EnvManager struct {
	projectDir string
}

// NewEnvManager creates a new EnvManager for the given project directory.
func NewEnvManager(projectDir string) *EnvManager {
	return &EnvManager{projectDir: projectDir}
}

// envsDir returns the path to the envs directory.
func (m *EnvManager) envsDir() string {
	return filepath.Join(m.projectDir, envDir)
}

// List returns all available environment profile names.
func (m *EnvManager) List() ([]string, error) {
	dir := m.envsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read envs dir: %w", err)
	}

	var profiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".env") {
			profiles = append(profiles, strings.TrimSuffix(name, ".env"))
		}
	}
	sort.Strings(profiles)
	return profiles, nil
}

// Show reads and returns the env profile with the given name.
func (m *EnvManager) Show(name string) (*model.EnvProfile, error) {
	path := filepath.Join(m.envsDir(), name+".env")
	vars, err := parseEnvFile(path)
	if err != nil {
		return nil, fmt.Errorf("read profile %q: %w", name, err)
	}
	return &model.EnvProfile{
		Name: name,
		Vars: vars,
		Path: path,
	}, nil
}

// Switch symlinks .env in the project root to the selected profile.
func (m *EnvManager) Switch(name string) error {
	src := filepath.Join(m.envsDir(), name+".env")
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("profile %q not found: %w", name, err)
	}

	dst := filepath.Join(m.projectDir, ".env")

	// Remove existing .env (file or symlink)
	_ = os.Remove(dst)

	// Copy the profile file as .env (symlinks not reliable on Windows)
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read profile: %w", err)
	}
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}
	return nil
}

// Diff compares two profiles and returns added, removed, and changed variables.
func (m *EnvManager) Diff(nameA, nameB string) (added, removed, changed []string, err error) {
	profileA, err := m.Show(nameA)
	if err != nil {
		return nil, nil, nil, err
	}
	profileB, err := m.Show(nameB)
	if err != nil {
		return nil, nil, nil, err
	}

	mapA := varsToMap(profileA.Vars)
	mapB := varsToMap(profileB.Vars)

	// Keys in B but not in A = added
	for k := range mapB {
		if _, ok := mapA[k]; !ok {
			added = append(added, k)
		}
	}

	// Keys in A but not in B = removed
	for k := range mapA {
		if _, ok := mapB[k]; !ok {
			removed = append(removed, k)
		}
	}

	// Keys in both with different values = changed
	for k, vA := range mapA {
		if vB, ok := mapB[k]; ok && vA != vB {
			changed = append(changed, k)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)

	return added, removed, changed, nil
}

// Export returns export statements for the given profile.
func (m *EnvManager) Export(name string) (string, error) {
	profile, err := m.Show(name)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, v := range profile.Vars {
		fmt.Fprintf(&sb, "export %s=%q\n", v.Key, v.Value)
	}
	return sb.String(), nil
}

// Active returns the name of the currently active profile by comparing
// .env content with available profiles.
func (m *EnvManager) Active() string {
	dotenv := filepath.Join(m.projectDir, ".env")
	currentData, err := os.ReadFile(dotenv)
	if err != nil {
		return ""
	}

	profiles, err := m.List()
	if err != nil {
		return ""
	}

	for _, name := range profiles {
		profilePath := filepath.Join(m.envsDir(), name+".env")
		profileData, err := os.ReadFile(profilePath)
		if err != nil {
			continue
		}
		if string(currentData) == string(profileData) {
			return name
		}
	}
	return ""
}

// parseEnvFile reads a .env file and returns key-value pairs.
func parseEnvFile(path string) ([]model.EnvVar, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var vars []model.EnvVar
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		// Strip surrounding quotes
		value = stripQuotes(value)
		vars = append(vars, model.EnvVar{Key: key, Value: value})
	}
	return vars, scanner.Err()
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func varsToMap(vars []model.EnvVar) map[string]string {
	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}
	return m
}
