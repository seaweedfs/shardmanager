package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/policy"
	"github.com/seaweedfs/shardmanager/policy/engine"
)

func main() {
	// Create a metric provider and action executor
	metricProvider := &engine.RealMetricProvider{
		Metrics: make(map[string]float64),
	}
	actionExecutor := &engine.RealActionExecutor{
		ExecutedActions: make([]policy.Action, 0),
	}

	// Create the policy engine
	engine := engine.NewEngine(metricProvider, actionExecutor)

	// Set up some test metrics
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

	// Evaluate the policy
	ctx := context.Background()
	executed, err := engine.EvaluatePolicy(ctx, policy)
	if err != nil {
		log.Fatalf("Failed to evaluate policy: %v", err)
	}

	fmt.Printf("Policy executed: %v\n", executed)
	fmt.Println("Executed actions:")
	for _, action := range actionExecutor.GetExecutedActions() {
		fmt.Printf("- %s: %v\n", action.Type, action.Constraints)
	}
}
