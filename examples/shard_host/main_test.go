package main

import (
	"testing"
	"time"

	"github.com/seaweedfs/shardmanager/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardHost(t *testing.T) {
	// Create a new shard host
	host := NewShardHost("node1")
	host.StartMetricsCollection()

	// Test adding a shard
	shardID := "test-shard"
	data := []byte("test data")
	metadata := map[string]interface{}{
		"created_at": time.Now(),
		"size":       len(data),
	}

	err := host.AddShard(shardID, data, metadata)
	require.NoError(t, err)

	// Test getting the shard
	shard, err := host.GetShard(shardID)
	require.NoError(t, err)
	assert.Equal(t, shardID, shard.ID)
	assert.Equal(t, data, shard.Data)
	assert.Equal(t, metadata["size"], shard.Metadata["size"])

	// Wait for metrics to be updated
	time.Sleep(1200 * time.Millisecond)

	// Test metrics
	metrics := host.GetMetrics()
	assert.Greater(t, metrics["cpu_usage"], 0.0)
	assert.Greater(t, metrics["memory_usage"], 0.0)

	// Test removing the shard
	err = host.RemoveShard(shardID)
	require.NoError(t, err)

	// Test getting a non-existent shard
	_, err = host.GetShard(shardID)
	assert.Error(t, err)

	// Test handling migration action
	action := policy.Action{
		Type: "migrate_shard",
		Constraints: map[string]interface{}{
			"source": "node1",
			"target": "node2",
		},
	}

	err = host.HandleMigrateShard(action)
	require.NoError(t, err)

	// Test handling invalid migration action
	invalidAction := policy.Action{
		Type: "migrate_shard",
		Constraints: map[string]interface{}{
			"source": "node1",
			// Missing target
		},
	}

	err = host.HandleMigrateShard(invalidAction)
	assert.Error(t, err)

	// Cleanup
	close(host.stopChan)
}
