package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeCRUD(t *testing.T) {
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

	// Test CreateNode
	err := CreateNode(db, node)
	require.NoError(t, err)
	assert.NotZero(t, node.CreatedAt)
	assert.NotZero(t, node.UpdatedAt)

	// Test GetNode
	retrievedNode, err := GetNode(db, nodeID)
	require.NoError(t, err)
	assert.Equal(t, node.ID, retrievedNode.ID)
	assert.Equal(t, node.Location, retrievedNode.Location)
	assert.Equal(t, node.Capacity, retrievedNode.Capacity)
	assert.Equal(t, node.Status, retrievedNode.Status)

	// Test UpdateNode
	node.Status = "inactive"
	err = UpdateNode(db, node)
	require.NoError(t, err)

	updatedNode, err := GetNode(db, nodeID)
	require.NoError(t, err)
	assert.Equal(t, "inactive", updatedNode.Status)

	// Test ListNodes
	nodes, err := ListNodes(db)
	require.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, nodeID, nodes[0].ID)

	// Test DeleteNode
	err = DeleteNode(db, nodeID)
	require.NoError(t, err)

	_, err = GetNode(db, nodeID)
	assert.Error(t, err)
}

func TestNodeConcurrentOperations(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create multiple nodes concurrently
	const numNodes = 10
	done := make(chan struct{})
	for i := 0; i < numNodes; i++ {
		go func() {
			node := &Node{
				ID:       uuid.New(),
				Location: "test-location",
				Capacity: 1000,
				Status:   "active",
			}
			err := CreateNode(db, node)
			require.NoError(t, err)
			done <- struct{}{}
		}()
	}

	// Wait for all nodes to be created
	for i := 0; i < numNodes; i++ {
		<-done
	}

	// Verify all nodes were created
	nodes, err := ListNodes(db)
	require.NoError(t, err)
	assert.Len(t, nodes, numNodes)
}
