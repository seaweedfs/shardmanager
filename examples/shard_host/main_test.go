package main

import (
	"testing"
	"time"

	"github.com/seaweedfs/shardmanager/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardHost(t *testing.T) {
	// Create two shard hosts
	host1 := NewShardHost("node1", 8080)
	host2 := NewShardHost("node2", 8081)
	host1.StartMetricsCollection()
	host2.StartMetricsCollection()

	// Test adding a shard to host1
	shardID := "test-shard"
	data := []byte("test data")
	metadata := map[string]interface{}{
		"created_at": time.Now(),
		"size":       len(data),
	}

	err := host1.AddShard(shardID, data, metadata)
	require.NoError(t, err)

	// Test getting the shard from host1
	shard, err := host1.GetShard(shardID)
	require.NoError(t, err)
	assert.Equal(t, shardID, shard.ID)
	assert.Equal(t, data, shard.Data)
	assert.Equal(t, metadata["size"], shard.Metadata["size"])

	// Wait for metrics to be updated
	time.Sleep(1200 * time.Millisecond)

	// Test metrics for host1
	metrics := host1.GetMetrics()
	assert.Greater(t, metrics["cpu_usage"], 0.0)
	assert.Greater(t, metrics["memory_usage"], 0.0)

	// Test handling migration action from host1 to host2
	action := policy.Action{
		Type: "migrate_shard",
		Constraints: map[string]interface{}{
			"source": "node1",
			"target": "node2",
			"shard":  shardID,
		},
	}
	// First, call HandleMigrateShard on host1 to remove the shard
	err = host1.HandleMigrateShard(action)
	require.NoError(t, err)

	// Then, call HandleMigrateShard on host2 to add the shard
	err = host2.HandleMigrateShard(action)
	require.NoError(t, err)

	// Verify the shard is now on host2
	shard, err = host2.GetShard(shardID)
	require.NoError(t, err)
	assert.Equal(t, shardID, shard.ID)

	// Test handling invalid migration action
	invalidAction := policy.Action{
		Type: "migrate_shard",
		Constraints: map[string]interface{}{
			"source": "node1",
			// Missing target
		},
	}
	err = host1.HandleMigrateShard(invalidAction)
	assert.Error(t, err)

	// Test removing the shard from host2
	err = host2.RemoveShard(shardID)
	require.NoError(t, err)

	// Test getting a non-existent shard from host2
	_, err = host2.GetShard(shardID)
	assert.Error(t, err)

	// Cleanup
	close(host1.stopChan)
	close(host2.stopChan)
}
