package policy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DefaultParser implements the PolicyParser interface
type DefaultParser struct{}

// NewDefaultParser creates a new DefaultParser
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse parses a policy from JSON
func (p *DefaultParser) Parse(policyJSON []byte) (*Policy, error) {
	var policy Policy
	if err := json.Unmarshal(policyJSON, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	// Set default values if not provided
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}
	if policy.UpdatedAt.IsZero() {
		policy.UpdatedAt = time.Now()
	}

	// Validate the policy
	if err := p.Validate(&policy); err != nil {
		return nil, err
	}

	return &policy, nil
}

// Validate validates a policy
func (p *DefaultParser) Validate(policy *Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Type == "" {
		return fmt.Errorf("policy type is required")
	}

	// Validate conditions
	if len(policy.Conditions.All) == 0 && len(policy.Conditions.Any) == 0 {
		return fmt.Errorf("policy must have at least one condition")
	}

	// Validate actions
	if len(policy.Actions) == 0 {
		return fmt.Errorf("policy must have at least one action")
	}

	// Validate each condition
	for _, condition := range policy.Conditions.All {
		if err := validateCondition(condition); err != nil {
			return fmt.Errorf("invalid condition in 'all': %w", err)
		}
	}

	for _, condition := range policy.Conditions.Any {
		if err := validateCondition(condition); err != nil {
			return fmt.Errorf("invalid condition in 'any': %w", err)
		}
	}

	// Validate each action
	for i, action := range policy.Actions {
		if action.Type == "" {
			return fmt.Errorf("action type is required for action at index %d", i)
		}
	}

	return nil
}

// validateCondition validates a single condition
func validateCondition(condition Condition) error {
	if condition.Metric == "" {
		return fmt.Errorf("metric name is required")
	}

	if condition.Operator == "" {
		return fmt.Errorf("operator is required")
	}

	// Validate operator
	switch condition.Operator {
	case OperatorLessThan, OperatorGreaterThan, OperatorEquals, OperatorNotEquals:
		// Valid operators
	default:
		return fmt.Errorf("invalid operator: %s", condition.Operator)
	}

	// Validate value type based on operator
	switch condition.Operator {
	case OperatorLessThan, OperatorGreaterThan:
		// These operators require numeric values
		switch v := condition.Value.(type) {
		case float64, int, int64:
			// Valid numeric types
		default:
			return fmt.Errorf("operator %s requires numeric value, got %T", condition.Operator, v)
		}
	}

	return nil
}
