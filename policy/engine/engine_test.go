package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	callback        func(ctx context.Context, action policy.Action) error
}

func newMockActionExecutor() *mockActionExecutor {
	return &mockActionExecutor{
		executedActions: make([]policy.Action, 0),
	}
}

func (m *mockActionExecutor) ExecuteAction(ctx context.Context, action policy.Action) error {
	if m.callback != nil {
		if err := m.callback(ctx, action); err != nil {
			return err
		}
	}
	m.executedActions = append(m.executedActions, action)
	return nil
}

func (m *mockActionExecutor) GetExecutedActions() []policy.Action {
	return m.executedActions
}

// ErroringActionExecutor simulates an error on a specific action
type erroringActionExecutor struct {
	failOnType string
	executed   []policy.Action
}

func (e *erroringActionExecutor) ExecuteAction(ctx context.Context, action policy.Action) error {
	e.executed = append(e.executed, action)
	if action.Type == e.failOnType {
		return fmt.Errorf("simulated error for action: %s", action.Type)
	}
	return nil
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

	// Verify both actions were executed (in priority order)
	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 2)
	assert.Equal(t, "high_priority_action", actions[0].Type)
	assert.Equal(t, "low_priority_action", actions[1].Type)
}

func TestEngine_EvaluatePolicy_MultipleActions(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := newMockActionExecutor()
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	metricProvider.SetMetric("cpu_usage", 90.0)

	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "multi-action-policy",
		Description: "Policy with multiple actions",
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
			{Type: "migrate_shard"},
			{Type: "notify_admin"},
		},
	}

	executed, err := engine.EvaluatePolicy(ctx, p)
	require.NoError(t, err)
	assert.True(t, executed)
	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 2)
	assert.Equal(t, "migrate_shard", actions[0].Type)
	assert.Equal(t, "notify_admin", actions[1].Type)
}

func TestEngine_EvaluatePolicy_ActionError(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := &erroringActionExecutor{failOnType: "notify_admin"}
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	metricProvider.SetMetric("cpu_usage", 90.0)

	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "error-action-policy",
		Description: "Policy with erroring action",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{{
				Metric:   "cpu_usage",
				Operator: policy.OperatorGreaterThan,
				Value:    80.0,
			}},
		},
		Actions: []policy.Action{
			{Type: "migrate_shard"},
			{Type: "notify_admin"},
		},
	}

	executed, err := engine.EvaluatePolicy(ctx, p)
	assert.True(t, executed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated error for action: notify_admin")
	assert.Len(t, actionExecutor.executed, 2)
}

func TestEngine_EvaluatePolicies_ChainedPolicies(t *testing.T) {
	metricProvider := newMockMetricProvider()
	actionExecutor := newMockActionExecutor()
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	// Simulate a scenario where the first policy's action triggers a metric change that enables the second policy
	metricProvider.SetMetric("cpu_usage", 90.0)
	metricProvider.SetMetric("disk_usage", 50.0)

	firstPolicy := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "first-policy",
		Description: "First policy",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    2,
		Conditions: policy.Conditions{
			All: []policy.Condition{{
				Metric:   "cpu_usage",
				Operator: policy.OperatorGreaterThan,
				Value:    80.0,
			}},
		},
		Actions: []policy.Action{{Type: "reduce_cpu"}},
	}

	secondPolicy := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "second-policy",
		Description: "Second policy",
		Type:        policy.PolicyTypeLoadBalancing,
		Priority:    1,
		Conditions: policy.Conditions{
			All: []policy.Condition{{
				Metric:   "disk_usage",
				Operator: policy.OperatorGreaterThan,
				Value:    80.0,
			}},
		},
		Actions: []policy.Action{{Type: "cleanup_disk"}},
	}

	// Simulate chaining: after first policy, update disk_usage to trigger second
	actionExecutor.callback = func(ctx context.Context, action policy.Action) error {
		if action.Type == "reduce_cpu" {
			metricProvider.SetMetric("disk_usage", 85.0)
		}
		return nil
	}

	err := engine.EvaluatePolicies(ctx, []*policy.Policy{firstPolicy, secondPolicy})
	require.NoError(t, err)
	actions := actionExecutor.GetExecutedActions()
	assert.Equal(t, []string{"reduce_cpu", "cleanup_disk"}, []string{actions[0].Type, actions[1].Type})
}

func TestEngine_Integration_RealSystem(t *testing.T) {
	metricProvider := &RealMetricProvider{Metrics: make(map[string]float64)}
	actionExecutor := &RealActionExecutor{ExecutedActions: make([]policy.Action, 0)}
	engine := NewEngine(metricProvider, actionExecutor)
	ctx := context.Background()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 90.0)
	metricProvider.SetMetric("memory_usage", 85.0)

	// Create a test policy
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "integration-policy",
		Description: "Integration test policy",
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
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
			},
			{
				Type: "notify_admin",
			},
		},
	}

	// Test policy evaluation
	executed, err := engine.EvaluatePolicy(ctx, p)
	require.NoError(t, err)
	assert.True(t, executed)

	// Verify actions were executed
	actions := actionExecutor.GetExecutedActions()
	require.Len(t, actions, 2)
	assert.Equal(t, "migrate_shard", actions[0].Type)
	assert.Equal(t, "notify_admin", actions[1].Type)
}

func TestEngine_Integration_PeriodicEvaluation(t *testing.T) {
	metricProvider := &RealMetricProvider{Metrics: make(map[string]float64)}
	actionExecutor := &RealActionExecutor{ExecutedActions: make([]policy.Action, 0)}
	engine := NewEngine(metricProvider, actionExecutor)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 90.0)
	metricProvider.SetMetric("memory_usage", 85.0)

	// Create a test policy
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "periodic-policy",
		Description: "Periodic evaluation test policy",
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
					Operator: policy.OperatorGreaterThan,
					Value:    80.0,
				},
			},
		},
		Actions: []policy.Action{
			{
				Type: "migrate_shard",
			},
			{
				Type: "notify_admin",
			},
		},
	}

	// Simulate periodic evaluation
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			executed, err := engine.EvaluatePolicy(ctx, p)
			require.NoError(t, err)
			assert.True(t, executed)
		case <-ctx.Done():
			return
		}
	}
}
