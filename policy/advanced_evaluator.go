package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AdvancedEvaluator extends the basic evaluator with more sophisticated features
type AdvancedEvaluator struct {
	*DefaultEvaluator
	timeWindow time.Duration
	history    map[string][]EvaluationResult
}

// NewAdvancedEvaluator creates a new AdvancedEvaluator
func NewAdvancedEvaluator(parser *DefaultParser, timeWindow time.Duration) *AdvancedEvaluator {
	return &AdvancedEvaluator{
		DefaultEvaluator: NewDefaultEvaluator(parser),
		timeWindow:       timeWindow,
		history:          make(map[string][]EvaluationResult),
	}
}

// EvaluateWithHistory evaluates a policy and maintains evaluation history
func (e *AdvancedEvaluator) EvaluateWithHistory(ctx context.Context, policy *Policy, state *SystemState) (*EvaluationResult, error) {
	result, err := e.Evaluate(ctx, policy, state)
	if err != nil {
		return result, err
	}

	// Store evaluation result in history
	e.storeEvaluationResult(policy.ID.String(), result)

	// Clean up old history entries
	e.cleanupHistory()

	return result, nil
}

// EvaluateWithTimeWindow evaluates a policy considering historical data within a time window
func (e *AdvancedEvaluator) EvaluateWithTimeWindow(ctx context.Context, policy *Policy, state *SystemState) (*EvaluationResult, error) {
	// Get historical results for this policy
	history := e.getPolicyHistory(policy.ID.String())

	// Evaluate current state
	currentResult, err := e.Evaluate(ctx, policy, state)
	if err != nil {
		return currentResult, err
	}

	// Store the result in history
	e.storeEvaluationResult(policy.ID.String(), currentResult)

	// Analyze historical trends
	trend := e.analyzeTrend(append(history, *currentResult))
	if trend != nil {
		currentResult.Details = trend
	}

	return currentResult, nil
}

// EvaluatePolicyChain evaluates a chain of policies in sequence
func (e *AdvancedEvaluator) EvaluatePolicyChain(ctx context.Context, policies []*Policy, state *SystemState) ([]*EvaluationResult, error) {
	var results []*EvaluationResult
	currentState := state

	for _, policy := range policies {
		result, err := e.Evaluate(ctx, policy, currentState)
		if err != nil {
			return results, fmt.Errorf("error evaluating policy %s: %w", policy.ID, err)
		}

		results = append(results, result)

		// If policy matched and has actions, apply them to the state
		if result.Matched && len(result.Actions) > 0 {
			currentState = e.applyActionsToState(currentState, result.Actions)
		}
	}

	// Update the original state with the final state
	*state = *currentState

	return results, nil
}

// storeEvaluationResult stores an evaluation result in the history
func (e *AdvancedEvaluator) storeEvaluationResult(policyID string, result *EvaluationResult) {
	e.history[policyID] = append(e.history[policyID], *result)
}

// cleanupHistory removes old evaluation results from history
func (e *AdvancedEvaluator) cleanupHistory() {
	cutoff := time.Now().Add(-e.timeWindow)
	for policyID, results := range e.history {
		var validResults []EvaluationResult
		for _, result := range results {
			if result.EvaluatedAt.After(cutoff) {
				validResults = append(validResults, result)
			}
		}
		e.history[policyID] = validResults
	}
}

// getPolicyHistory returns the evaluation history for a policy
func (e *AdvancedEvaluator) getPolicyHistory(policyID string) []EvaluationResult {
	return e.history[policyID]
}

// analyzeTrend analyzes the trend of evaluation results
func (e *AdvancedEvaluator) analyzeTrend(history []EvaluationResult) []byte {
	if len(history) < 2 {
		return nil
	}

	// Count matches and total evaluations
	var matches, total int
	for _, result := range history {
		total++
		if result.Matched {
			matches++
		}
	}

	// Calculate trend data
	trend := struct {
		MatchRate   float64 `json:"match_rate"`
		TotalEvals  int     `json:"total_evaluations"`
		TimeWindow  string  `json:"time_window"`
		LastMatched bool    `json:"last_matched"`
		TrendStable bool    `json:"trend_stable"`
	}{
		MatchRate:   float64(matches) / float64(total),
		TotalEvals:  total,
		TimeWindow:  e.timeWindow.String(),
		LastMatched: history[len(history)-1].Matched,
		TrendStable: matches > total/2, // Consider trend stable if more than 50% matches
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(trend)
	return jsonData
}

// applyActionsToState applies policy actions to the system state
func (e *AdvancedEvaluator) applyActionsToState(state *SystemState, actions []Action) *SystemState {
	// Create a copy of the state to modify
	newState := &SystemState{
		Nodes:   make(map[string]NodeState),
		Shards:  make(map[string]ShardState),
		Metrics: make(map[string]MetricValue),
	}

	// Copy existing state
	for k, v := range state.Nodes {
		newState.Nodes[k] = v
	}
	for k, v := range state.Shards {
		newState.Shards[k] = v
	}
	for k, v := range state.Metrics {
		newState.Metrics[k] = v
	}

	// Apply each action to the state
	for _, action := range actions {
		switch action.Type {
		case "migrate_shard":
			// Example: Update shard location
			if shardID, ok := action.Constraints["shard_id"].(string); ok {
				if nodeID, ok := action.Constraints["target_node"].(string); ok {
					if shard, exists := newState.Shards[shardID]; exists {
						shard.NodeID = nodeID
						newState.Shards[shardID] = shard
					}
				}
			}
		case "update_metrics":
			// Example: Update system metrics
			if metricName, ok := action.Constraints["metric"].(string); ok {
				if value, ok := action.Constraints["value"].(float64); ok {
					newState.Metrics[metricName] = MetricValue{
						Value:     value,
						Timestamp: time.Now(),
					}
				}
			}
		}
	}

	return newState
}
