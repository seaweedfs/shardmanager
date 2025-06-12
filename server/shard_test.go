package server

import (
	"context"
	"testing"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/server/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardServiceOperations(t *testing.T) {
	server := &Server{db: testutil.NewMockDB()}
	ctx := context.Background()

	t.Run("RegisterShard", func(t *testing.T) {
		shardID := uuid.New()
		req := &shardmanagerpb.RegisterShardRequest{
			Shard: &shardmanagerpb.Shard{
				Id:     shardID.String(),
				Type:   "test-type",
				Size:   100,
				Status: "active",
			},
		}

		resp, err := server.RegisterShard(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("ListShards", func(t *testing.T) {
		resp, err := server.ListShards(ctx, &shardmanagerpb.ListShardsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("GetShardInfo", func(t *testing.T) {
		shardID := uuid.New()
		req := &shardmanagerpb.GetShardInfoRequest{
			ShardId: shardID.String(),
		}

		resp, err := server.GetShardInfo(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("AssignShard", func(t *testing.T) {
		shardID := uuid.New()
		nodeID := uuid.New()
		req := &shardmanagerpb.AssignShardRequest{
			ShardId: shardID.String(),
			NodeId:  nodeID.String(),
		}

		resp, err := server.AssignShard(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("UpdateShardStatus", func(t *testing.T) {
		shardID := uuid.New()
		req := &shardmanagerpb.UpdateShardStatusRequest{
			ShardId: shardID.String(),
			Status:  "migrating",
		}

		resp, err := server.UpdateShardStatus(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
