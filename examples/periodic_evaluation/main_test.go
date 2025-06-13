package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/seaweedfs/shardmanager/policy/engine"
	"github.com/seaweedfs/shardmanager/policy/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeriodicEvaluationExample(t *testing.T) {
	// Create components
	metricProvider := &engine.RealMetricProvider{
		Metrics: make(map[string]float64),
	}
	actionExecutor := &engine.RealActionExecutor{
		ExecutedActions: make([]policy.Action, 0),
	}
	store := &mockPolicyStore{}

	// Create the policy manager with a short evaluation interval for testing
	pm := manager.NewPolicyManager(metricProvider, actionExecutor, store, 100*time.Millisecond)

	// Create a test policy
	policy := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "high-cpu-policy",
		Description: "Policy to handle high CPU usage",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{
				{
					Metric:   "cpu_usage",
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
				Constraints: map[string]interface{}{
					"source": "node1",
					"target": "node2",
				},
			},
		},
	}

	// Store the policy
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	require.NoError(t, store.Store(ctx, policy))

	// Start the policy manager
	require.NoError(t, pm.Start(ctx))

	// Test case 1: CPU usage below threshold
	metricProvider.SetMetric("cpu_usage", 75.0)
	time.Sleep(200 * time.Millisecond)
	assert.Empty(t, actionExecutor.GetExecutedActions())

	// Test case 2: CPU usage above threshold
	metricProvider.SetMetric("cpu_usage", 85.0)
	time.Sleep(200 * time.Millisecond)
	actions := actionExecutor.GetExecutedActions()
	require.NotEmpty(t, actions)
	assert.Equal(t, "migrate_shard", actions[0].Type)

	// Test case 3: CPU usage back below threshold
	metricProvider.SetMetric("cpu_usage", 75.0)
	time.Sleep(200 * time.Millisecond)
	// No new actions should be executed

	// Stop the policy manager
	pm.Stop()
}
