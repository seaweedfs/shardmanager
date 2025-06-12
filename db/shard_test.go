package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardOperations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	// Create a test node
	nodeID := uuid.New()
	node := &Node{
		ID:       nodeID,
		Location: "test-location",
		Capacity: 1000,
		Status:   "active",
	}
	err := testDB.RegisterNode(ctx, node)
	require.NoError(t, err)

	t.Run("RegisterShard", func(t *testing.T) {
		shard := &Shard{
			ID:     uuid.New(),
			Type:   "test-type",
			Size:   100,
			NodeID: &nodeID,
			Status: "active",
		}

		err := testDB.RegisterShard(ctx, shard)
		require.NoError(t, err)
		assert.NotZero(t, shard.CreatedAt)
		assert.NotZero(t, shard.UpdatedAt)
	})

	t.Run("ListShards", func(t *testing.T) {
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

		err := testDB.RegisterShard(ctx, shard1)
		require.NoError(t, err)
		err = testDB.RegisterShard(ctx, shard2)
		require.NoError(t, err)

		shards, err := testDB.ListShards(ctx)
		require.NoError(t, err)
		assert.Len(t, shards, 2)

		// Verify shard details
		found := make(map[string]bool)
		for _, shard := range shards {
			if shard.Type == "type-1" {
				assert.Equal(t, int64(100), shard.Size)
				assert.Equal(t, "active", shard.Status)
				assert.Equal(t, nodeID, *shard.NodeID)
				found["type-1"] = true
			} else if shard.Type == "type-2" {
				assert.Equal(t, int64(200), shard.Size)
				assert.Equal(t, "inactive", shard.Status)
				assert.Equal(t, nodeID, *shard.NodeID)
				found["type-2"] = true
			}
		}
		assert.True(t, found["type-1"])
		assert.True(t, found["type-2"])
	})

	t.Run("GetShardInfo", func(t *testing.T) {
		shardID := uuid.New()
		shard := &Shard{
			ID:     shardID,
			Type:   "test-type-2",
			Size:   200,
			NodeID: &nodeID,
			Status: "active",
		}

		err := testDB.RegisterShard(ctx, shard)
		require.NoError(t, err)

		info, err := testDB.GetShardInfo(ctx, shardID)
		require.NoError(t, err)
		assert.Equal(t, shardID, info.ID)
		assert.Equal(t, "test-type-2", info.Type)
		assert.Equal(t, int64(200), info.Size)
		assert.Equal(t, nodeID, *info.NodeID)
	})

	t.Run("AssignShard", func(t *testing.T) {
		// Create a new node
		newNodeID := uuid.New()
		newNode := &Node{
			ID:       newNodeID,
			Location: "new-location",
			Capacity: 2000,
			Status:   "active",
		}
		err := testDB.RegisterNode(ctx, newNode)
		require.NoError(t, err)

		shardID := uuid.New()
		shard := &Shard{
			ID:     shardID,
			Type:   "test-type-3",
			Size:   300,
			Status: "active",
		}

		err = testDB.RegisterShard(ctx, shard)
		require.NoError(t, err)

		err = testDB.AssignShard(ctx, shardID, newNodeID)
		require.NoError(t, err)

		info, err := testDB.GetShardInfo(ctx, shardID)
		require.NoError(t, err)
		assert.Equal(t, newNodeID, *info.NodeID)
	})

	t.Run("UpdateShardStatus", func(t *testing.T) {
		shardID := uuid.New()
		shard := &Shard{
			ID:     shardID,
			Type:   "test-type-4",
			Size:   400,
			NodeID: &nodeID,
			Status: "active",
		}

		err := testDB.RegisterShard(ctx, shard)
		require.NoError(t, err)

		err = testDB.UpdateShardStatus(ctx, shardID, "migrating")
		require.NoError(t, err)

		info, err := testDB.GetShardInfo(ctx, shardID)
		require.NoError(t, err)
		assert.Equal(t, "migrating", info.Status)
	})
}
