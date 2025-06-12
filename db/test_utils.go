package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testDB *DB

func setupTestDB(t *testing.T) {
	// Use a test database
	dsn := "postgres://postgres:postgres@localhost:5432/shardmanager_test?sslmode=disable"
	var err error
	testDB, err = NewDB(dsn)
	require.NoError(t, err)

	// Create test tables
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id UUID PRIMARY KEY,
			location TEXT NOT NULL,
			capacity BIGINT NOT NULL,
			status TEXT NOT NULL,
			last_heartbeat TIMESTAMP WITH TIME ZONE,
			current_load BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS shards (
			id UUID PRIMARY KEY,
			type TEXT NOT NULL,
			size BIGINT NOT NULL,
			node_id UUID REFERENCES nodes(id),
			status TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS policies (
			id UUID PRIMARY KEY,
			policy_type TEXT NOT NULL,
			parameters JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS failure_reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type TEXT NOT NULL,
			entity_id UUID NOT NULL,
			details JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)
}

func teardownTestDB(t *testing.T) {
	// Drop test tables
	_, err := testDB.Exec(`
		DROP TABLE IF EXISTS failure_reports;
		DROP TABLE IF EXISTS policies;
		DROP TABLE IF EXISTS shards;
		DROP TABLE IF EXISTS nodes;
	`)
	require.NoError(t, err)
	testDB.Close()
}
