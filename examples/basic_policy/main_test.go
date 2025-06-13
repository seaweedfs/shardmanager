package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/seaweedfs/shardmanager/policy/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicPolicyExample(t *testing.T) {
	// Create components
	metricProvider := &engine.RealMetricProvider{
		Metrics: make(map[string]float64),
	}
	actionExecutor := &engine.RealActionExecutor{
		ExecutedActions: make([]policy.Action, 0),
	}
	engine := engine.NewEngine(metricProvider, actionExecutor)

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 85.0)
	metricProvider.SetMetric("memory_usage", 75.0)

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
				{
					Metric:   "memory_usage",
					Operator: policy.OperatorLessThan,
					Value:    90.0,
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
			{
				Type: "notify_admin",
				Constraints: map[string]interface{}{
					"message": "High CPU usage detected",
				},
			},
		},
	}

	// Test case 1: Conditions met
	executed, err := engine.EvaluatePolicy(context.Background(), policy)
	require.NoError(t, err)
	assert.True(t, executed)

	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 2)
	assert.Equal(t, "migrate_shard", actions[0].Type)
	assert.Equal(t, "notify_admin", actions[1].Type)

	// Test case 2: Conditions not met (CPU too low)
	metricProvider.SetMetric("cpu_usage", 75.0)
	actionExecutor.ExecutedActions = nil // Clear previous actions

	executed, err = engine.EvaluatePolicy(context.Background(), policy)
	require.NoError(t, err)
	assert.False(t, executed)
	assert.Empty(t, actionExecutor.GetExecutedActions())

	// Test case 3: Conditions not met (Memory too high)
	metricProvider.SetMetric("cpu_usage", 85.0)
	metricProvider.SetMetric("memory_usage", 95.0)
	actionExecutor.ExecutedActions = nil // Clear previous actions

	executed, err = engine.EvaluatePolicy(context.Background(), policy)
	require.NoError(t, err)
	assert.False(t, executed)
	assert.Empty(t, actionExecutor.GetExecutedActions())
}
