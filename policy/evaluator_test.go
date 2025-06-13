package policy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultEvaluator_Evaluate_Match(t *testing.T) {
	parser := NewDefaultParser()
	evaluator := NewDefaultEvaluator(parser)
	p := &Policy{
		Name: "cpu-high",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{{Type: "migrate_shard"}},
	}
	state := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 90.0, Timestamp: time.Now()},
		},
	}
	ctx := context.Background()
	result, err := evaluator.Evaluate(ctx, p, state)
	assert.NoError(t, err)
	assert.True(t, result.Matched)
	assert.True(t, result.Success)
	assert.Len(t, result.Actions, 1)
}

func TestDefaultEvaluator_Evaluate_NoMatch(t *testing.T) {
	parser := NewDefaultParser()
	evaluator := NewDefaultEvaluator(parser)
	p := &Policy{
		Name: "cpu-high",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{{Type: "migrate_shard"}},
	}
	state := &SystemState{
		Metrics: map[string]MetricValue{
			"cpu_usage": {Value: 50.0, Timestamp: time.Now()},
		},
	}
	ctx := context.Background()
	result, err := evaluator.Evaluate(ctx, p, state)
	assert.NoError(t, err)
	assert.False(t, result.Matched)
	assert.True(t, result.Success)
	assert.Len(t, result.Actions, 0)
}

func TestDefaultEvaluator_Evaluate_MissingMetric(t *testing.T) {
	parser := NewDefaultParser()
	evaluator := NewDefaultEvaluator(parser)
	p := &Policy{
		Name: "cpu-high",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{{Type: "migrate_shard"}},
	}
	state := &SystemState{
		Metrics: map[string]MetricValue{},
	}
	ctx := context.Background()
	result, err := evaluator.Evaluate(ctx, p, state)
	assert.Error(t, err)
	assert.False(t, result.Matched)
}
