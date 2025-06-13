package engine

import (
	"context"
	"fmt"

	"github.com/seaweedfs/shardmanager/policy"
)

// MetricProvider defines the interface for getting metric values
type MetricProvider interface {
	GetMetric(ctx context.Context, metricName string) (float64, error)
}

// ActionExecutor defines the interface for executing policy actions
type ActionExecutor interface {
	ExecuteAction(ctx context.Context, action policy.Action) error
}

// Engine evaluates policies and executes their actions
type Engine struct {
	metricProvider MetricProvider
	actionExecutor ActionExecutor
}

// NewEngine creates a new policy engine
func NewEngine(metricProvider MetricProvider, actionExecutor ActionExecutor) *Engine {
	return &Engine{
		metricProvider: metricProvider,
		actionExecutor: actionExecutor,
	}
}

// EvaluatePolicy evaluates a policy's conditions and executes its actions if conditions are met
func (e *Engine) EvaluatePolicy(ctx context.Context, p *policy.Policy) (bool, error) {
	// Evaluate conditions
	conditionsMet, err := e.evaluateConditions(ctx, p.Conditions)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate conditions: %w", err)
	}

	if !conditionsMet {
		return false, nil
	}

	// Execute actions
	for _, action := range p.Actions {
		if err := e.actionExecutor.ExecuteAction(ctx, action); err != nil {
			return true, fmt.Errorf("failed to execute action %s: %w", action.Type, err)
		}
	}

	return true, nil
}

// evaluateConditions evaluates all conditions in a policy
func (e *Engine) evaluateConditions(ctx context.Context, conditions policy.Conditions) (bool, error) {
	// Evaluate ALL conditions
	for _, condition := range conditions.All {
		met, err := e.evaluateCondition(ctx, condition)
		if err != nil {
			return false, err
		}
		if !met {
			return false, nil
		}
	}

	// Evaluate ANY conditions
	if len(conditions.Any) > 0 {
		anyMet := false
		for _, condition := range conditions.Any {
			met, err := e.evaluateCondition(ctx, condition)
			if err != nil {
				return false, err
			}
			if met {
				anyMet = true
				break
			}
		}
		if !anyMet {
			return false, nil
		}
	}

	return true, nil
}

// evaluateCondition evaluates a single condition
func (e *Engine) evaluateCondition(ctx context.Context, condition policy.Condition) (bool, error) {
	value, err := e.metricProvider.GetMetric(ctx, condition.Metric)
	if err != nil {
		return false, fmt.Errorf("failed to get metric %s: %w", condition.Metric, err)
	}

	// Convert condition.Value to float64 for comparison
	conditionValue, ok := condition.Value.(float64)
	if !ok {
		return false, fmt.Errorf("condition value must be a number, got %T", condition.Value)
	}

	switch condition.Operator {
	case policy.OperatorGreaterThan:
		return value > conditionValue, nil
	case policy.OperatorLessThan:
		return value < conditionValue, nil
	case policy.OperatorEquals:
		return value == conditionValue, nil
	case policy.OperatorNotEquals:
		return value != conditionValue, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", condition.Operator)
	}
}

// EvaluatePolicies evaluates multiple policies in order of priority
func (e *Engine) EvaluatePolicies(ctx context.Context, policies []*policy.Policy) error {
	// Sort policies by priority (highest first)
	sortedPolicies := make([]*policy.Policy, len(policies))
	copy(sortedPolicies, policies)

	// Sort by priority (highest first)
	for i := 0; i < len(sortedPolicies)-1; i++ {
		for j := i + 1; j < len(sortedPolicies); j++ {
			if sortedPolicies[i].Priority < sortedPolicies[j].Priority {
				sortedPolicies[i], sortedPolicies[j] = sortedPolicies[j], sortedPolicies[i]
			}
		}
	}

	var firstErr error
	for _, p := range sortedPolicies {
		_, err := e.EvaluatePolicy(ctx, p)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to evaluate policy %s: %w", p.ID, err)
			}
		}
		// No break: evaluate all policies
	}

	return firstErr
}
