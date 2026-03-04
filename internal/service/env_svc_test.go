package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEnvDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	envsDir := filepath.Join(dir, ".devx", "envs")
	require.NoError(t, os.MkdirAll(envsDir, 0o700))
	return dir
}

func writeProfile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, ".devx", "envs", name+".env")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func TestEnvManager_List_Empty(t *testing.T) {
	dir := t.TempDir()
	mgr := NewEnvManager(dir)
	profiles, err := mgr.List()
	require.NoError(t, err)
	assert.Empty(t, profiles)
}

func TestEnvManager_List(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "development", "DB=localhost\n")
	writeProfile(t, dir, "production", "DB=prod.example.com\n")
	writeProfile(t, dir, "staging", "DB=staging.example.com\n")

	mgr := NewEnvManager(dir)
	profiles, err := mgr.List()
	require.NoError(t, err)
	assert.Equal(t, []string{"development", "production", "staging"}, profiles)
}

func TestEnvManager_Show(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "dev", "DB_HOST=localhost\nDB_PORT=5432\nAPI_KEY=\"secret123\"\n")

	mgr := NewEnvManager(dir)
	profile, err := mgr.Show("dev")
	require.NoError(t, err)
	assert.Equal(t, "dev", profile.Name)
	assert.Len(t, profile.Vars, 3)
	assert.Equal(t, "DB_HOST", profile.Vars[0].Key)
	assert.Equal(t, "localhost", profile.Vars[0].Value)
	assert.Equal(t, "5432", profile.Vars[1].Value)
	assert.Equal(t, "secret123", profile.Vars[2].Value) // quotes stripped
}

func TestEnvManager_Show_NotFound(t *testing.T) {
	dir := setupEnvDir(t)
	mgr := NewEnvManager(dir)
	_, err := mgr.Show("nonexistent")
	assert.Error(t, err)
}

func TestEnvManager_Switch(t *testing.T) {
	dir := setupEnvDir(t)
	content := "DB_HOST=production.example.com\nAPI_KEY=prodkey\n"
	writeProfile(t, dir, "prod", content)

	mgr := NewEnvManager(dir)
	err := mgr.Switch("prod")
	require.NoError(t, err)

	// Verify .env was created with profile content
	dotenv, err := os.ReadFile(filepath.Join(dir, ".env"))
	require.NoError(t, err)
	assert.Equal(t, content, string(dotenv))
}

func TestEnvManager_Switch_NotFound(t *testing.T) {
	dir := setupEnvDir(t)
	mgr := NewEnvManager(dir)
	err := mgr.Switch("nonexistent")
	assert.Error(t, err)
}

func TestEnvManager_Switch_OverwritesExisting(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "dev", "MODE=development\n")
	writeProfile(t, dir, "prod", "MODE=production\n")

	// Create existing .env
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("old"), 0o600))

	mgr := NewEnvManager(dir)
	require.NoError(t, mgr.Switch("prod"))

	dotenv, err := os.ReadFile(filepath.Join(dir, ".env"))
	require.NoError(t, err)
	assert.Equal(t, "MODE=production\n", string(dotenv))
}

func TestEnvManager_Diff(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "dev", "DB=localhost\nDEBUG=true\nPORT=3000\n")
	writeProfile(t, dir, "prod", "DB=prod.db.com\nPORT=8080\nCDN=cdn.example.com\n")

	mgr := NewEnvManager(dir)
	added, removed, changed, err := mgr.Diff("dev", "prod")
	require.NoError(t, err)
	assert.Equal(t, []string{"CDN"}, added)
	assert.Equal(t, []string{"DEBUG"}, removed)
	assert.Equal(t, []string{"DB", "PORT"}, changed)
}

func TestEnvManager_Diff_Identical(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "a", "KEY=value\n")
	writeProfile(t, dir, "b", "KEY=value\n")

	mgr := NewEnvManager(dir)
	added, removed, changed, err := mgr.Diff("a", "b")
	require.NoError(t, err)
	assert.Empty(t, added)
	assert.Empty(t, removed)
	assert.Empty(t, changed)
}

func TestEnvManager_Export(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "dev", "DB_HOST=localhost\nAPI_KEY=secret\n")

	mgr := NewEnvManager(dir)
	output, err := mgr.Export("dev")
	require.NoError(t, err)
	assert.Contains(t, output, `export DB_HOST="localhost"`)
	assert.Contains(t, output, `export API_KEY="secret"`)
}

func TestEnvManager_Active(t *testing.T) {
	dir := setupEnvDir(t)
	content := "DB=localhost\n"
	writeProfile(t, dir, "dev", content)
	writeProfile(t, dir, "prod", "DB=prod.example.com\n")

	// Write .env matching dev profile
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600))

	mgr := NewEnvManager(dir)
	assert.Equal(t, "dev", mgr.Active())
}

func TestEnvManager_Active_NoMatch(t *testing.T) {
	dir := setupEnvDir(t)
	writeProfile(t, dir, "dev", "DB=localhost\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("CUSTOM=true\n"), 0o600))

	mgr := NewEnvManager(dir)
	assert.Equal(t, "", mgr.Active())
}

func TestParseEnvFile_CommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.env")
	content := `# This is a comment
DB_HOST=localhost

# Another comment
DB_PORT=5432
EMPTY_LINE_ABOVE=true
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	vars, err := parseEnvFile(path)
	require.NoError(t, err)
	assert.Len(t, vars, 3)
	assert.Equal(t, "DB_HOST", vars[0].Key)
	assert.Equal(t, "DB_PORT", vars[1].Key)
	assert.Equal(t, "EMPTY_LINE_ABOVE", vars[2].Key)
}

func TestParseEnvFile_QuotedValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.env")
	content := `DOUBLE="hello world"
SINGLE='another value'
NONE=plain
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	vars, err := parseEnvFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello world", vars[0].Value)
	assert.Equal(t, "another value", vars[1].Value)
	assert.Equal(t, "plain", vars[2].Value)
}

func TestStripQuotes(t *testing.T) {
	assert.Equal(t, "hello", stripQuotes(`"hello"`))
	assert.Equal(t, "hello", stripQuotes(`'hello'`))
	assert.Equal(t, "hello", stripQuotes("hello"))
	assert.Equal(t, "", stripQuotes(`""`))
	assert.Equal(t, "x", stripQuotes("x"))
}

func TestVarsToMap(t *testing.T) {
	m := varsToMap(nil)
	assert.Empty(t, m)
}

func TestValidateProfileName(t *testing.T) {
	// Valid names
	assert.NoError(t, validateProfileName("dev"))
	assert.NoError(t, validateProfileName("production"))
	assert.NoError(t, validateProfileName("staging-us-east"))
	assert.NoError(t, validateProfileName("v1.0"))
	assert.NoError(t, validateProfileName("my_profile"))

	// Invalid names
	assert.Error(t, validateProfileName(""))
	assert.Error(t, validateProfileName("../etc/passwd"))
	assert.Error(t, validateProfileName(".."))
	assert.Error(t, validateProfileName("/root"))
	assert.Error(t, validateProfileName("has space"))
	assert.Error(t, validateProfileName(".hidden"))
}
