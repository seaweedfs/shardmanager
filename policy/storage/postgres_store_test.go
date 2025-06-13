package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSQLiteTestDB(t *testing.T) *db.DB {
	dbConn, err := db.NewDBWithDriver("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	_, err = dbConn.Exec(`
		CREATE TABLE policies (
			id TEXT PRIMARY KEY,
			policy_type TEXT,
			parameters BLOB,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	return dbConn
}

func TestPostgresPolicyStore_CRUD(t *testing.T) {
	dbConn := setupSQLiteTestDB(t)
	store := NewPostgresPolicyStore(dbConn)
	ctx := context.Background()

	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "test-policy",
		Description: "A test policy",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{{
				Metric:   "cpu_usage",
				Operator: policy.OperatorGreaterThan,
				Value:    80.0,
			}},
		},
		Actions: []policy.Action{{
			Type: "migrate_shard",
		}},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := store.Store(ctx, p)
	require.NoError(t, err)

	// Get by ID
	got, err := store.Get(ctx, p.ID.String())
	require.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)
	assert.Equal(t, p.Name, got.Name)
	assert.Equal(t, p.Type, got.Type)

	// List
	plist, err := store.List(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, plist)
	found := false
	for _, pol := range plist {
		if pol.ID == p.ID {
			found = true
		}
	}
	assert.True(t, found)

	// ListByType
	byType, err := store.ListByType(ctx, policy.PolicyTypeLoadBalancing)
	require.NoError(t, err)
	assert.NotEmpty(t, byType)
	assert.Equal(t, p.ID, byType[0].ID)

	// Update
	p.Description = "Updated description"
	err = store.Update(ctx, p)
	require.NoError(t, err)
	got, err = store.Get(ctx, p.ID.String())
	require.NoError(t, err)
	assert.Equal(t, "Updated description", got.Description)
}

func TestPostgresPolicyStore_GetNotFound(t *testing.T) {
	dbConn := setupSQLiteTestDB(t)
	store := NewPostgresPolicyStore(dbConn)
	ctx := context.Background()

	_, err := store.Get(ctx, uuid.New().String())
	assert.Error(t, err)
}

func TestPostgresPolicyStore_ListByType_Empty(t *testing.T) {
	dbConn := setupSQLiteTestDB(t)
	store := NewPostgresPolicyStore(dbConn)
	ctx := context.Background()

	byType, err := store.ListByType(ctx, policy.PolicyTypeReplication)
	require.NoError(t, err)
	assert.Empty(t, byType)
}
