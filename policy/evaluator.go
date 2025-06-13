package policy

import (
	"context"
	"fmt"
	"time"
)

// DefaultEvaluator implements the PolicyEvaluator interface
type DefaultEvaluator struct {
	parser *DefaultParser
}

// NewDefaultEvaluator creates a new DefaultEvaluator
func NewDefaultEvaluator(parser *DefaultParser) *DefaultEvaluator {
	return &DefaultEvaluator{
		parser: parser,
	}
}

// Evaluate evaluates a policy against the current system state
func (e *DefaultEvaluator) Evaluate(ctx context.Context, policy *Policy, state *SystemState) (*EvaluationResult, error) {
	result := &EvaluationResult{
		PolicyID:    policy.ID,
		EvaluatedAt: time.Now(),
	}

	// Evaluate all conditions
	allConditionsMet := true
	if len(policy.Conditions.All) > 0 {
		for _, condition := range policy.Conditions.All {
			met, err := e.evaluateCondition(condition, state)
			if err != nil {
				result.Error = fmt.Sprintf("error evaluating condition: %v", err)
				return result, err
			}
			if !met {
				allConditionsMet = false
				break
			}
		}
	}

	// Evaluate any conditions
	anyConditionsMet := true
	if len(policy.Conditions.Any) > 0 {
		anyConditionsMet = false
		for _, condition := range policy.Conditions.Any {
			met, err := e.evaluateCondition(condition, state)
			if err != nil {
				result.Error = fmt.Sprintf("error evaluating condition: %v", err)
				return result, err
			}
			if met {
				anyConditionsMet = true
				break
			}
		}
	}

	// Policy matches if all conditions are met
	result.Matched = allConditionsMet && anyConditionsMet
	if result.Matched {
		result.Actions = policy.Actions
	}
	result.Success = true

	return result, nil
}

// GetApplicablePolicies returns policies applicable to the given operation type
func (e *DefaultEvaluator) GetApplicablePolicies(ctx context.Context, operationType PolicyType) ([]*Policy, error) {
	// This is a placeholder implementation
	// In a real implementation, this would query a policy store
	return nil, fmt.Errorf("not implemented")
}

// evaluateCondition evaluates a single condition against the system state
func (e *DefaultEvaluator) evaluateCondition(condition Condition, state *SystemState) (bool, error) {
	// Get the metric value from the system state
	metricValue, exists := state.Metrics[condition.Metric]
	if !exists {
		return false, fmt.Errorf("metric %s not found in system state", condition.Metric)
	}

	// Compare the metric value with the condition value
	switch condition.Operator {
	case OperatorLessThan:
		value, ok := condition.Value.(float64)
		if !ok {
			return false, fmt.Errorf("invalid value type for less than operator")
		}
		return metricValue.Value < value, nil

	case OperatorGreaterThan:
		value, ok := condition.Value.(float64)
		if !ok {
			return false, fmt.Errorf("invalid value type for greater than operator")
		}
		return metricValue.Value > value, nil

	case OperatorEquals:
		return metricValue.Value == condition.Value, nil

	case OperatorNotEquals:
		return metricValue.Value != condition.Value, nil

	default:
		return false, fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}
