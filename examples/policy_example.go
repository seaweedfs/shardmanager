package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/seaweedfs/shardmanager/policy"
)

func main() {
	// Create policy engine components
	parser := policy.NewDefaultParser()
	evaluator := policy.NewDefaultEvaluator(parser)
	store := policy.NewInMemoryPolicyStore()

	// Create a sample policy
	policyJSON := []byte(`{
		"name": "high-cpu-load-balancing",
		"description": "Balance load when CPU usage is high",
		"type": "load_balancing",
		"priority": 1,
		"conditions": {
			"all": [
				{
					"metric": "cpu_usage",
					"operator": "gt",
					"value": 80.0
				}
			]
		},
		"actions": [
			{
				"type": "migrate_shard",
				"strategy": "least_loaded",
				"constraints": {
					"max_nodes": 3
				}
			}
		]
	}`)

	// Parse and store the policy
	p, err := parser.Parse(policyJSON)
	if err != nil {
		log.Fatalf("Failed to parse policy: %v", err)
	}

	ctx := context.Background()
	if err := store.Store(ctx, p); err != nil {
		log.Fatalf("Failed to store policy: %v", err)
	}

	// Create a sample system state
	state := &policy.SystemState{
		Metrics: map[string]policy.MetricValue{
			"cpu_usage": {
				Value:     85.0,
				Timestamp: time.Now(),
			},
		},
	}

	// Evaluate the policy
	result, err := evaluator.Evaluate(ctx, p, state)
	if err != nil {
		log.Fatalf("Failed to evaluate policy: %v", err)
	}

	// Print the evaluation result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Policy evaluation result:\n%s\n", resultJSON)

	// List all policies
	policies, err := store.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list policies: %v", err)
	}

	fmt.Printf("\nStored policies:\n")
	for _, p := range policies {
		policyJSON, _ := json.MarshalIndent(p, "", "  ")
		fmt.Printf("%s\n", policyJSON)
	}
}
