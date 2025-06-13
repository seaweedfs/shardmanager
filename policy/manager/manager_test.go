package manager

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/seaweedfs/shardmanager/policy/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPolicyStore implements policy.PolicyStore for testing
type mockPolicyStore struct {
	policies []*policy.Policy
}

func (m *mockPolicyStore) Store(ctx context.Context, p *policy.Policy) error {
	m.policies = append(m.policies, p)
	return nil
}

func (m *mockPolicyStore) Get(ctx context.Context, id string) (*policy.Policy, error) {
	for _, p := range m.policies {
		if p.ID.String() == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockPolicyStore) List(ctx context.Context) ([]*policy.Policy, error) {
	return m.policies, nil
}

func (m *mockPolicyStore) ListByType(ctx context.Context, policyType policy.PolicyType) ([]*policy.Policy, error) {
	var filtered []*policy.Policy
	for _, p := range m.policies {
		if p.Type == policyType {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (m *mockPolicyStore) Update(ctx context.Context, p *policy.Policy) error {
	for i, existing := range m.policies {
		if existing.ID == p.ID {
			m.policies[i] = p
			return nil
		}
	}
	return nil
}

func (m *mockPolicyStore) Delete(ctx context.Context, id string) error {
	for i, p := range m.policies {
		if p.ID.String() == id {
			m.policies = append(m.policies[:i], m.policies[i+1:]...)
			return nil
		}
	}
	return nil
}

func TestPolicyManager_PeriodicEvaluation(t *testing.T) {
	metricProvider := &engine.RealMetricProvider{Metrics: make(map[string]float64)}
	actionExecutor := &engine.RealActionExecutor{ExecutedActions: make([]policy.Action, 0)}
	store := &mockPolicyStore{}

	// Create a policy manager with a short evaluation interval
	manager := NewPolicyManager(metricProvider, actionExecutor, store, 100*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 90.0)

	// Create and store a test policy
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "test-policy",
		Description: "Test policy",
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
				Type: "test_action",
			},
		},
	}
	require.NoError(t, store.Store(ctx, p))

	// Start the policy manager
	require.NoError(t, manager.Start(ctx))

	// Wait for context cancellation
	<-ctx.Done()

	// Stop the policy manager
	manager.Stop()

	// Verify that the policy was evaluated multiple times
	actions := actionExecutor.GetExecutedActions()
	assert.Greater(t, len(actions), 1)
	for _, action := range actions {
		assert.Equal(t, "test_action", action.Type)
	}
}

func TestPolicyManager_EventDrivenEvaluation(t *testing.T) {
	metricProvider := &engine.RealMetricProvider{Metrics: make(map[string]float64)}
	actionExecutor := &engine.RealActionExecutor{ExecutedActions: make([]policy.Action, 0)}
	store := &mockPolicyStore{}

	// Create a policy manager with a long evaluation interval
	manager := NewPolicyManager(metricProvider, actionExecutor, store, 1*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Set up test metrics
	metricProvider.SetMetric("cpu_usage", 90.0)

	// Create and store a test policy
	p := &policy.Policy{
		ID:          uuid.New(),
		Version:     "v1",
		Name:        "test-policy",
		Description: "Test policy",
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
				Type: "test_action",
			},
		},
	}
	require.NoError(t, store.Store(ctx, p))

	// Start the policy manager
	require.NoError(t, manager.Start(ctx))

	// Trigger multiple evaluations
	manager.TriggerEvaluation()
	manager.TriggerEvaluation()
	manager.TriggerEvaluation()

	// Wait for context cancellation
	<-ctx.Done()

	// Stop the policy manager
	manager.Stop()

	// Verify that the policy was evaluated multiple times
	actions := actionExecutor.GetExecutedActions()
	assert.Greater(t, len(actions), 1)
	for _, action := range actions {
		assert.Equal(t, "test_action", action.Type)
	}
}
