package storage

import (
	"context"
	"os"
	"testing"

	"github.com/seaweedfs/shardmanager/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) (*PersistentPolicyStore, func()) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "policy-*.db")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	// Create store
	store, err := NewPersistentPolicyStore(tmpfile.Name())
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		store.Close()
		os.Remove(tmpfile.Name())
	}

	return store, cleanup
}

func TestPersistentPolicyStore_CRUD(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create test policy
	p := &policy.Policy{
		Name:        "test-policy",
		Description: "Test policy for CRUD operations",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{
				{
					Metric:   "cpu_usage",
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
				Constraints: map[string]interface{}{
					"shard_id":    "shard1",
					"target_node": "node2",
				},
			},
		},
	}

	// Test Store
	err := store.Store(ctx, p)
	require.NoError(t, err)

	// Test Get
	retrieved, err := store.Get(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.Name, retrieved.Name)
	assert.Equal(t, p.Description, retrieved.Description)
	assert.Equal(t, p.Type, retrieved.Type)
	assert.Equal(t, p.Priority, retrieved.Priority)
	assert.Equal(t, p.Conditions, retrieved.Conditions)
	assert.Equal(t, p.Actions, retrieved.Actions)

	// Test List
	policies, err := store.List(ctx)
	require.NoError(t, err)
	assert.Len(t, policies, 1)
	assert.Equal(t, p.ID, policies[0].ID)

	// Test Delete
	err = store.Delete(ctx, p.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = store.Get(ctx, p.ID)
	assert.Error(t, err)
}

func TestPersistentPolicyStore_History(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create test policy
	p := &policy.Policy{
		Name:        "test-policy",
		Description: "Test policy for history tracking",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{
				{
					Metric:   "cpu_usage",
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
			},
		},
	}

	// Store policy
	err := store.Store(ctx, p)
	require.NoError(t, err)

	// Update policy
	p.Description = "Updated description"
	err = store.Store(ctx, p)
	require.NoError(t, err)

	// Delete policy
	err = store.Delete(ctx, p.ID)
	require.NoError(t, err)

	// Get history
	history, err := store.GetHistory(ctx, p.ID)
	require.NoError(t, err)
	assert.Len(t, history, 3)

	// Verify history entries (in reverse chronological order)
	assert.Equal(t, "delete", history[0].Action)
	assert.Equal(t, "store", history[1].Action)
	assert.Equal(t, "store", history[2].Action)
}

func TestPersistentPolicyStore_ConcurrentAccess(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create test policy
	p := &policy.Policy{
		Name:        "test-policy",
		Description: "Test policy for concurrent access",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{
				{
					Metric:   "cpu_usage",
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
			},
		},
	}

	// Store policy
	err := store.Store(ctx, p)
	require.NoError(t, err)

	// Simulate concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// Read
			_, err := store.Get(ctx, p.ID)
			assert.NoError(t, err)

			// Update
			p.Description = "Updated by goroutine"
			err = store.Store(ctx, p)
			assert.NoError(t, err)

			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	retrieved, err := store.Get(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated by goroutine", retrieved.Description)
}
