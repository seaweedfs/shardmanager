package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardVersioning(t *testing.T) {
	ctx := context.Background()
	sqldb := setupTestDB(t)
	defer cleanupTestDB(t, sqldb)
	db := &DB{DB: sqldb, DriverName: "sqlite3"}

	// Create test node
	nodeID := uuid.New()
	node := &Node{
		ID:            nodeID,
		Location:      "test-location",
		Capacity:      1000,
		Status:        "active",
		LastHeartbeat: time.Now(),
		CurrentLoad:   0,
	}
	err := db.RegisterNode(ctx, node)
	require.NoError(t, err)

	// Create test shard
	shardID := uuid.New()
	shard := &Shard{
		ID:       shardID,
		Type:     "test-type",
		Size:     100,
		NodeID:   &nodeID,
		Status:   "active",
		Version:  1,
		Metadata: json.RawMessage(`{"key": "value"}`),
	}
	err = db.RegisterShard(ctx, shard)
	require.NoError(t, err)

	// After registration, there should be no version record in shard_versions
	version, err := db.GetShardVersion(ctx, shardID, 1)
	require.NoError(t, err)
	assert.Nil(t, version)

	// The current state should be in the shards table
	current, err := db.GetShardInfo(ctx, shardID)
	require.NoError(t, err)
	assert.NotNil(t, current)
	assert.Equal(t, shardID, current.ID)
	assert.Equal(t, 1, current.Version)
	assert.Equal(t, "test-type", current.Type)
	assert.Equal(t, int64(100), current.Size)
	assert.Equal(t, nodeID, *current.NodeID)
	assert.Equal(t, "active", current.Status)
	assert.Equal(t, `{"key": "value"}`, string(current.Metadata))

	// Test UpdateShardVersion
	shard.Type = "updated-type"
	shard.Size = 200
	shard.Metadata = json.RawMessage(`{"key": "updated-value"}`)
	err = db.UpdateShardVersion(ctx, shard)
	require.NoError(t, err)
	assert.Equal(t, 2, shard.Version)

	// Now there should be a version record for version 1
	version, err = db.GetShardVersion(ctx, shardID, 1)
	require.NoError(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, 1, version.Version)
	assert.Equal(t, "test-type", version.Type)
	assert.Equal(t, int64(100), version.Size)
	assert.Equal(t, nodeID, *version.NodeID)
	assert.Equal(t, "active", version.Status)
	assert.Equal(t, `{"key": "value"}`, string(version.Metadata))

	// Verify version history
	versions, err := db.ListShardVersions(ctx, shardID)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, 1, versions[0].Version)
	assert.Equal(t, "test-type", versions[0].Type)
	assert.Equal(t, int64(100), versions[0].Size)
	assert.Equal(t, `{"key": "value"}`, string(versions[0].Metadata))

	// Test RollbackShardVersion
	err = db.RollbackShardVersion(ctx, shardID, 1)
	require.NoError(t, err)

	// Verify rollback
	shard, err = db.GetShardInfo(ctx, shardID)
	require.NoError(t, err)
	assert.Equal(t, 3, shard.Version)
	assert.Equal(t, "test-type", shard.Type)
	assert.Equal(t, int64(100), shard.Size)
	assert.Equal(t, `{"key": "value"}`, string(shard.Metadata))

	// Verify version history after rollback
	versions, err = db.ListShardVersions(ctx, shardID)
	require.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, 2, versions[0].Version)
	assert.Equal(t, "updated-type", versions[0].Type)
	assert.Equal(t, int64(200), versions[0].Size)
	assert.Equal(t, `{"key": "updated-value"}`, string(versions[0].Metadata))
	assert.Equal(t, 1, versions[1].Version)
	assert.Equal(t, "test-type", versions[1].Type)
	assert.Equal(t, int64(100), versions[1].Size)
	assert.Equal(t, `{"key": "value"}`, string(versions[1].Metadata))
}
