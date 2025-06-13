package policy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdvancedEvaluator_EvaluateWithHistory(t *testing.T) {
	parser := NewDefaultParser()
	evaluator := NewAdvancedEvaluator(parser, 1*time.Hour)
	ctx := context.Background()

	// Create a test policy
	policy := &Policy{
		Name: "test-policy",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{{Type: "migrate_shard"}},
	}

	// Create test states
	highLoadState := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 90.0, Timestamp: time.Now()},
		},
	}

	lowLoadState := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 50.0, Timestamp: time.Now()},
		},
	}

	// Evaluate multiple times
	result1, err := evaluator.EvaluateWithHistory(ctx, policy, highLoadState)
	assert.NoError(t, err)
	assert.True(t, result1.Matched)

	result2, err := evaluator.EvaluateWithHistory(ctx, policy, lowLoadState)
	assert.NoError(t, err)
	assert.False(t, result2.Matched)

	// Check history
	history := evaluator.getPolicyHistory(policy.ID.String())
	assert.Len(t, history, 2)
	assert.True(t, history[0].Matched)
	assert.False(t, history[1].Matched)
}

func TestAdvancedEvaluator_EvaluateWithTimeWindow(t *testing.T) {
	parser := NewDefaultParser()
	evaluator := NewAdvancedEvaluator(parser, 1*time.Hour)
	ctx := context.Background()

	// Create a test policy
	policy := &Policy{
		Name: "test-policy",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{{Type: "migrate_shard"}},
	}

	// Create test states with different timestamps
	now := time.Now()
	oldState := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 90.0, Timestamp: now.Add(-2 * time.Hour)},
		},
	}

	newState := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 85.0, Timestamp: now},
		},
	}

	// Evaluate with old state
	result1, err := evaluator.EvaluateWithTimeWindow(ctx, policy, oldState)
	assert.NoError(t, err)
	assert.True(t, result1.Matched)

	// Evaluate with new state
	result2, err := evaluator.EvaluateWithTimeWindow(ctx, policy, newState)
	assert.NoError(t, err)
	assert.True(t, result2.Matched)
	assert.NotNil(t, result2.Details) // Should have trend analysis
}

func TestAdvancedEvaluator_EvaluatePolicyChain(t *testing.T) {
	parser := NewDefaultParser()
	evaluator := NewAdvancedEvaluator(parser, 1*time.Hour)
	ctx := context.Background()

	// Create a chain of policies
	policies := []*Policy{
		{
			Name: "cpu-check",
			Type: PolicyTypeLoadBalancing,
			Conditions: Conditions{
				All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
			},
			Actions: []Action{{
				Type: "update_metrics",
				Constraints: map[string]interface{}{
					"metric": "load_high",
					"value":  1.0,
				},
			}},
		},
		{
			Name: "load-response",
			Type: PolicyTypeLoadBalancing,
			Conditions: Conditions{
				All: []Condition{{Metric: "load_high", Operator: OperatorEquals, Value: 1.0}},
			},
			Actions: []Action{{
				Type: "migrate_shard",
				Constraints: map[string]interface{}{
					"shard_id":    "shard1",
					"target_node": "node2",
				},
			}},
		},
	}

	// Initial state
	state := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 90.0, Timestamp: time.Now()},
		},
		Shards: map[string]ShardState{
			"shard1": {ID: "shard1", NodeID: "node1"},
		},
	}

	// Evaluate the chain
	results, err := evaluator.EvaluatePolicyChain(ctx, policies, state)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Matched)
	assert.True(t, results[1].Matched)

	// Check that the state was updated correctly
	assert.Equal(t, 1.0, state.Metrics["load_high"].Value)
	assert.Equal(t, "node2", state.Shards["shard1"].NodeID)
}
