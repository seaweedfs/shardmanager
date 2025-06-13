package db

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearShardsTable(t *testing.T, db *sql.DB) {
	_, _ = db.Exec("DELETE FROM shards")
}

func TestShardOperations(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create a test node
	nodeID := uuid.New()
	node := &Node{
		ID:       nodeID,
		Location: "test-location",
		Capacity: 1000,
		Status:   "active",
	}
	err := CreateNode(db, node)
	require.NoError(t, err)

	t.Run("RegisterShard", func(t *testing.T) {
		clearShardsTable(t, db)
		shard := &Shard{
			ID:     uuid.New(),
			Type:   "test-type",
			Size:   100,
			NodeID: &nodeID,
			Status: "active",
		}

		err := CreateShard(db, shard)
		require.NoError(t, err)
		assert.NotZero(t, shard.CreatedAt)
		assert.NotZero(t, shard.UpdatedAt)
	})

	t.Run("ListShards", func(t *testing.T) {
		clearShardsTable(t, db)
		// Create multiple shards
		shard1 := &Shard{
			ID:     uuid.New(),
			Type:   "type-1",
			Size:   100,
			NodeID: &nodeID,
			Status: "active",
		}
		shard2 := &Shard{
			ID:     uuid.New(),
			Type:   "type-2",
			Size:   200,
			NodeID: &nodeID,
			Status: "inactive",
		}

		err := CreateShard(db, shard1)
		require.NoError(t, err)
		err = CreateShard(db, shard2)
		require.NoError(t, err)

		shards, err := ListShards(db)
		require.NoError(t, err)
		assert.Len(t, shards, 2)

		// Verify shard details
		found := make(map[string]bool)
		for _, shard := range shards {
			if shard.Type == "type-1" {
				assert.Equal(t, int64(100), shard.Size)
				assert.Equal(t, "active", shard.Status)
				if shard.NodeID != nil {
					assert.Equal(t, nodeID, *shard.NodeID)
				}
				found["type-1"] = true
			} else if shard.Type == "type-2" {
				assert.Equal(t, int64(200), shard.Size)
				assert.Equal(t, "inactive", shard.Status)
				if shard.NodeID != nil {
					assert.Equal(t, nodeID, *shard.NodeID)
				}
				found["type-2"] = true
			}
		}
		assert.True(t, found["type-1"])
		assert.True(t, found["type-2"])
	})

	t.Run("GetShardInfo", func(t *testing.T) {
		clearShardsTable(t, db)
		shardID := uuid.New()
		shard := &Shard{
			ID:     shardID,
			Type:   "test-type-2",
			Size:   200,
			NodeID: &nodeID,
			Status: "active",
		}

		err := CreateShard(db, shard)
		require.NoError(t, err)

		retrieved, err := GetShard(db, shardID)
		require.NoError(t, err)
		assert.Equal(t, shardID, retrieved.ID)
		assert.Equal(t, "test-type-2", retrieved.Type)
		assert.Equal(t, int64(200), retrieved.Size)
		if retrieved.NodeID != nil {
			assert.Equal(t, nodeID, *retrieved.NodeID)
		}
	})

	t.Run("AssignShard", func(t *testing.T) {
		clearShardsTable(t, db)
		// Create a new node
		newNodeID := uuid.New()
		newNode := &Node{
			ID:       newNodeID,
			Location: "new-location",
			Capacity: 2000,
			Status:   "active",
		}
		err := CreateNode(db, newNode)
		require.NoError(t, err)

		shardID := uuid.New()
		shard := &Shard{
			ID:     shardID,
			Type:   "test-type-3",
			Size:   300,
			NodeID: &nodeID,
			Status: "active",
		}

		err = CreateShard(db, shard)
		require.NoError(t, err)

		err = AssignShard(db, shardID, newNodeID)
		require.NoError(t, err)

		retrieved, err := GetShard(db, shardID)
		require.NoError(t, err)
		if retrieved.NodeID != nil {
			assert.Equal(t, newNodeID, *retrieved.NodeID)
		}
	})

	t.Run("UpdateShardStatus", func(t *testing.T) {
		clearShardsTable(t, db)
		shardID := uuid.New()
		shard := &Shard{
			ID:     shardID,
			Type:   "test-type-4",
			Size:   400,
			NodeID: &nodeID,
			Status: "active",
		}

		err := CreateShard(db, shard)
		require.NoError(t, err)

		shard.Status = "migrating"
		err = UpdateShard(db, shard)
		require.NoError(t, err)

		retrieved, err := GetShard(db, shardID)
		require.NoError(t, err)
		assert.Equal(t, "migrating", retrieved.Status)
	})
}

func TestShardCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create a test node for foreign key constraint
	nodeID := uuid.New()
	node := &Node{
		ID:       nodeID,
		Location: "test-location",
		Capacity: 1000,
		Status:   "active",
	}
	err := CreateNode(db, node)
	require.NoError(t, err)

	// Test CreateShard
	shardID := uuid.New()
	shard := &Shard{
		ID:     shardID,
		NodeID: &nodeID,
		Status: "active",
		Size:   100,
	}
	err = CreateShard(db, shard)
	require.NoError(t, err)

	// Test GetShard
	retrieved, err := GetShard(db, shardID)
	require.NoError(t, err)
	assert.Equal(t, shard.ID, retrieved.ID)
	assert.Equal(t, shard.Status, retrieved.Status)
	assert.Equal(t, shard.Size, retrieved.Size)
	assert.Equal(t, *shard.NodeID, *retrieved.NodeID)

	// Test ListShards
	shards, err := ListShards(db)
	require.NoError(t, err)
	assert.Len(t, shards, 1)

	// Test UpdateShard
	shard.Status = "migrating"
	err = UpdateShard(db, shard)
	require.NoError(t, err)

	updated, err := GetShard(db, shardID)
	require.NoError(t, err)
	assert.Equal(t, "migrating", updated.Status)

	// Test DeleteShard
	err = DeleteShard(db, shardID)
	require.NoError(t, err)

	_, err = GetShard(db, shardID)
	assert.Error(t, err)
}
