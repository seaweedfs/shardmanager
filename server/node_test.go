package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/server/testutil"
)

func TestNodeService(t *testing.T) {
	mockDB := testutil.NewMockDB()
	server := &Server{db: mockDB}

	t.Run("RegisterNode", func(t *testing.T) {
		mockDB.Reset()
		nodeID := uuid.New()
		req := &shardmanagerpb.RegisterNodeRequest{
			Node: &shardmanagerpb.Node{
				Id:       nodeID.String(),
				Location: "localhost:8080",
				Capacity: 1000,
				Status:   "active",
			},
		}

		resp, err := server.RegisterNode(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("Heartbeat", func(t *testing.T) {
		mockDB.Reset()
		nodeID := uuid.New()
		req := &shardmanagerpb.HeartbeatRequest{
			NodeId: nodeID.String(),
			Status: "active",
			Load:   10,
		}

		resp, err := server.Heartbeat(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("ListNodes", func(t *testing.T) {
		mockDB.Reset()
		// Register a test node first
		nodeID := uuid.New()
		node := &db.Node{
			ID:          nodeID,
			Location:    "localhost:8080",
			Capacity:    1000,
			Status:      "active",
			CurrentLoad: 0,
		}
		err := mockDB.RegisterNode(context.Background(), node)
		require.NoError(t, err)

		req := &shardmanagerpb.ListNodesRequest{}
		resp, err := server.ListNodes(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Nodes, 1)
		assert.Equal(t, nodeID.String(), resp.Nodes[0].Id)
		assert.Equal(t, "localhost:8080", resp.Nodes[0].Location)
		assert.Equal(t, int64(1000), resp.Nodes[0].Capacity)
		assert.Equal(t, "active", resp.Nodes[0].Status)
	})
}
