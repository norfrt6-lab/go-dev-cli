package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate_CreatesTablesSuccessfully(t *testing.T) {
	db, err := OpenMemory()
	require.NoError(t, err)
	defer db.Close()

	// Verify tables exist by querying them
	tables := []string{"projects", "services", "log_entries", "schema_migrations"}
	for _, table := range tables {
		var name string
		err := db.conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		require.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := OpenMemory()
	require.NoError(t, err)
	defer db.Close()

	// Running migrate again should not error
	err = db.Migrate()
	assert.NoError(t, err)
}

func TestMigrate_RecordsVersion(t *testing.T) {
	db, err := OpenMemory()
	require.NoError(t, err)
	defer db.Close()

	var version int
	err = db.conn.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), version)
}
