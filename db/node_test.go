package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeDBOperations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	t.Run("RegisterNode", func(t *testing.T) {
		node := &Node{
			ID:       uuid.New(),
			Location: "test-location",
			Capacity: 1000,
			Status:   "active",
		}

		err := testDB.RegisterNode(ctx, node)
		require.NoError(t, err)
		assert.NotZero(t, node.CreatedAt)
		assert.NotZero(t, node.UpdatedAt)
	})

	t.Run("UpdateNodeHeartbeat", func(t *testing.T) {
		nodeID := uuid.New()
		node := &Node{
			ID:       nodeID,
			Location: "test-location",
			Capacity: 1000,
			Status:   "active",
		}

		err := testDB.RegisterNode(ctx, node)
		require.NoError(t, err)

		err = testDB.UpdateNodeHeartbeat(ctx, nodeID, "active", 500)
		require.NoError(t, err)

		nodes, err := testDB.ListNodes(ctx)
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		assert.Equal(t, int64(500), nodes[0].CurrentLoad)
		assert.NotZero(t, nodes[0].LastHeartbeat)
	})

	t.Run("ListNodes", func(t *testing.T) {
		// Create multiple nodes
		node1 := &Node{
			ID:       uuid.New(),
			Location: "location-1",
			Capacity: 1000,
			Status:   "active",
		}
		node2 := &Node{
			ID:       uuid.New(),
			Location: "location-2",
			Capacity: 2000,
			Status:   "inactive",
		}

		err := testDB.RegisterNode(ctx, node1)
		require.NoError(t, err)
		err = testDB.RegisterNode(ctx, node2)
		require.NoError(t, err)

		nodes, err := testDB.ListNodes(ctx)
		require.NoError(t, err)
		assert.Len(t, nodes, 2)

		// Verify node details
		found := make(map[string]bool)
		for _, node := range nodes {
			if node.Location == "location-1" {
				assert.Equal(t, int64(1000), node.Capacity)
				assert.Equal(t, "active", node.Status)
				found["location-1"] = true
			} else if node.Location == "location-2" {
				assert.Equal(t, int64(2000), node.Capacity)
				assert.Equal(t, "inactive", node.Status)
				found["location-2"] = true
			}
		}
		assert.True(t, found["location-1"])
		assert.True(t, found["location-2"])
	})
}
