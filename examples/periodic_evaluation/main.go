package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/seaweedfs/shardmanager/policy/engine"
	"github.com/seaweedfs/shardmanager/policy/manager"
)

// mockPolicyStore implements policy.PolicyStore for this example
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

func main() {
	// Create components
	metricProvider := &engine.RealMetricProvider{
		Metrics: make(map[string]float64),
	}
	actionExecutor := &engine.RealActionExecutor{
		ExecutedActions: make([]policy.Action, 0),
	}
	store := &mockPolicyStore{}

	// Create the policy manager with a 5-second evaluation interval
	pm := manager.NewPolicyManager(metricProvider, actionExecutor, store, 5*time.Second)

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
	ctx := context.Background()
	if err := store.Store(ctx, policy); err != nil {
		log.Fatalf("Failed to store policy: %v", err)
	}

	// Start the policy manager
	if err := pm.Start(ctx); err != nil {
		log.Fatalf("Failed to start policy manager: %v", err)
	}

	// Simulate metric changes
	go func() {
		for {
			// Simulate CPU usage fluctuating between 70% and 90%
			cpuUsage := 70.0 + float64(time.Now().Unix()%20)
			metricProvider.SetMetric("cpu_usage", cpuUsage)
			fmt.Printf("Current CPU usage: %.1f%%\n", cpuUsage)
			time.Sleep(2 * time.Second)
		}
	}()

	// Print executed actions
	go func() {
		for {
			actions := actionExecutor.GetExecutedActions()
			if len(actions) > 0 {
				fmt.Println("\nExecuted actions:")
				for _, action := range actions {
					fmt.Printf("- %s: %v\n", action.Type, action.Constraints)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop the policy manager
	pm.Stop()
	fmt.Println("\nPolicy manager stopped")
}
