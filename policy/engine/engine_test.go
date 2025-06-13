package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMetricProvider implements MetricProvider for testing
type mockMetricProvider struct {
	metrics map[string]float64
}

func newMockMetricProvider() *mockMetricProvider {
	return &mockMetricProvider{
		metrics: make(map[string]float64),
	}
}

func (m *mockMetricProvider) GetMetric(ctx context.Context, metricName string) (float64, error) {
	value, ok := m.metrics[metricName]
	if !ok {
		return 0, nil
	}
	return value, nil
}

func (m *mockMetricProvider) SetMetric(metricName string, value float64) {
	m.metrics[metricName] = value
}

// mockActionExecutor implements ActionExecutor for testing
type mockActionExecutor struct {
	executedActions []policy.Action
}

func newMockActionExecutor() *mockActionExecutor {
	return &mockActionExecutor{
		executedActions: make([]policy.Action, 0),
	}
}

func (m *mockActionExecutor) ExecuteAction(ctx context.Context, action policy.Action) error {
	m.executedActions = append(m.executedActions, action)
	return nil
}

func (m *mockActionExecutor) GetExecutedActions() []policy.Action {
	return m.executedActions
}

func TestEngine_EvaluatePolicy(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := newMockActionExecutor()
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 85.0)
	metricProvider.SetMetric("memory_usage", 60.0)

	// Create a test policy
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "test-policy",
		Description: "A test policy",
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
					Value:    70.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
			},
		},
	}

	// Test policy evaluation
	executed, err := engine.EvaluatePolicy(ctx, p)
	require.NoError(t, err)
	assert.True(t, executed)

	// Verify actions were executed
	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "migrate_shard", actions[0].Type)
}

func TestEngine_EvaluatePolicy_ConditionsNotMet(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := newMockActionExecutor()
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 75.0) // Below threshold

	// Create a test policy
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "test-policy",
		Description: "A test policy",
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
			},
		},
	}

	// Test policy evaluation
	executed, err := engine.EvaluatePolicy(ctx, p)
	require.NoError(t, err)
	assert.False(t, executed)

	// Verify no actions were executed
	actions := actionExecutor.GetExecutedActions()
	assert.Empty(t, actions)
}

func TestEngine_EvaluatePolicy_AnyConditions(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := newMockActionExecutor()
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 75.0)
	metricProvider.SetMetric("memory_usage", 85.0)

	// Create a test policy with ANY conditions
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "test-policy",
		Description: "A test policy",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			Any: []policy.Condition{
				{
					Metric:   "cpu_usage",
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
				{
					Metric:   "memory_usage",
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
			},
		},
	}

	// Test policy evaluation
	executed, err := engine.EvaluatePolicy(ctx, p)
	require.NoError(t, err)
	assert.True(t, executed)

	// Verify actions were executed
	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "migrate_shard", actions[0].Type)
}

func TestEngine_EvaluatePolicies_Priority(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := newMockActionExecutor()
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 85.0)

	// Create two policies with different priorities
	highPriorityPolicy := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "high-priority",
		Description: "High priority policy",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    2,
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
				Type: "high_priority_action",
			},
		},
	}

	lowPriorityPolicy := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "low-priority",
		Description: "Low priority policy",
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
				Type: "low_priority_action",
			},
		},
	}

	// Test policy evaluation
	err := engine.EvaluatePolicies(ctx, []*policy.Policy{lowPriorityPolicy, highPriorityPolicy})
	require.NoError(t, err)

	// Verify only the high priority action was executed
	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "high_priority_action", actions[0].Type)
}
