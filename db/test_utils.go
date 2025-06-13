package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

var testDBFile = "test_shardmanager.db"

// setupTestDB creates a new file-based SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	// Remove any existing test DB file
	_ = os.Remove(testDBFile)
	db, err := sql.Open("sqlite3", testDBFile)
	require.NoError(t, err)

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			location TEXT NOT NULL,
			capacity INTEGER NOT NULL,
			status TEXT NOT NULL,
			last_heartbeat TIMESTAMP,
			current_load INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS shards (
			id TEXT PRIMARY KEY,
			node_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			size INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (node_id) REFERENCES nodes(id)
		);

		CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			conditions TEXT NOT NULL,
			actions TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS failures (
			id TEXT PRIMARY KEY,
			node_id TEXT NOT NULL,
			shard_id TEXT NOT NULL,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (node_id) REFERENCES nodes(id),
			FOREIGN KEY (shard_id) REFERENCES shards(id)
		);
	`)
	require.NoError(t, err)

	return db
}

// cleanupTestDB closes the test database connection and removes the file
func cleanupTestDB(t *testing.T, db *sql.DB) {
	err := db.Close()
	require.NoError(t, err)
	_ = os.Remove(testDBFile)
}
