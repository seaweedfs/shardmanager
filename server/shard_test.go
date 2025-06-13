package server

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/server/testutil"
)

func TestShardServiceOperations(t *testing.T) {
	mockDB := testutil.NewMockDB()
	server := NewServer(mockDB)

	const testShardID = "11111111-1111-1111-1111-111111111111"
	const testNodeID = "22222222-2222-2222-2222-222222222222" // Node ID must also be a valid UUID

	// Register a node before registering a shard
	nodeUUID, _ := uuid.Parse(testNodeID)
	mockDB.RegisterNode(context.Background(), &db.Node{
		ID:       nodeUUID,
		Location: "localhost:1234",
		Capacity: 100,
		Status:   "active",
	})

	t.Run("RegisterShard", func(t *testing.T) {
		req := &shardmanagerpb.RegisterShardRequest{
			Shard: &shardmanagerpb.Shard{
				Id:     testShardID,
				Type:   "test-type",
				Size:   100,
				Status: "active",
			},
		}
		_, err := server.RegisterShard(context.Background(), req)
		if err != nil {
			t.Errorf("RegisterShard failed: %v", err)
		}
		log.Printf("Registered shard: %v", req.Shard)
	})

	t.Run("ListShards", func(t *testing.T) {
		req := &shardmanagerpb.ListShardsRequest{}
		_, err := server.ListShards(context.Background(), req)
		if err != nil {
			t.Errorf("ListShards failed: %v", err)
		}
	})

	t.Run("GetShardInfo", func(t *testing.T) {
		// Create a shard before the test
		req := &shardmanagerpb.RegisterShardRequest{
			Shard: &shardmanagerpb.Shard{
				Id:     testShardID,
				Type:   "test-type",
				Size:   100,
				Status: "active",
			},
		}
		_, err := server.RegisterShard(context.Background(), req)
		if err != nil {
			t.Errorf("RegisterShard failed: %v", err)
		}
		log.Printf("Registered shard: %v", req.Shard)

		// Retrieve the shard
		reqGet := &shardmanagerpb.GetShardInfoRequest{
			ShardId: testShardID,
		}
		_, err = server.GetShardInfo(context.Background(), reqGet)
		if err != nil {
			t.Errorf("GetShardInfo failed: %v", err)
		}

		// Convert string to uuid.UUID
		shardID, err := uuid.Parse(testShardID)
		if err != nil {
			t.Errorf("Failed to parse shard ID: %v", err)
		}

		// Assert that the shard was retrieved correctly
		shard, err := mockDB.GetShardInfo(context.Background(), shardID)
		if err != nil {
			t.Errorf("GetShardInfo failed: %v", err)
		}
		if shard == nil {
			t.Errorf("Shard not found")
		}
		log.Printf("Retrieved shard: %v", shard)
	})

	t.Run("AssignShard", func(t *testing.T) {
		// Use a valid UUID for the node ID
		req := &shardmanagerpb.AssignShardRequest{
			ShardId: testShardID,
			NodeId:  testNodeID,
		}
		_, err := server.AssignShard(context.Background(), req)
		if err != nil {
			t.Errorf("AssignShard failed: %v", err)
		}
	})

	t.Run("UpdateShardStatus", func(t *testing.T) {
		req := &shardmanagerpb.UpdateShardStatusRequest{
			ShardId: testShardID,
			Status:  "inactive",
		}
		_, err := server.UpdateShardStatus(context.Background(), req)
		if err != nil {
			t.Errorf("UpdateShardStatus failed: %v", err)
		}
	})
}
