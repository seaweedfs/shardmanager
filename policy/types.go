package policy

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PolicyType represents the type of policy
type PolicyType string

const (
	PolicyTypePlacement        PolicyType = "placement"
	PolicyTypeMigration        PolicyType = "migration"
	PolicyTypeReplication      PolicyType = "replication"
	PolicyTypeLoadBalancing    PolicyType = "load_balancing"
	PolicyTypeCostOptimization PolicyType = "cost_optimization"
)

// Operator represents comparison operators
type Operator string

const (
	OperatorLessThan    Operator = "lt"
	OperatorGreaterThan Operator = "gt"
	OperatorEquals      Operator = "eq"
	OperatorNotEquals   Operator = "ne"
)

// Condition represents a single condition in a policy
type Condition struct {
	Metric   string      `json:"metric"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// Conditions represents a group of conditions
type Conditions struct {
	All []Condition `json:"all,omitempty"`
	Any []Condition `json:"any,omitempty"`
}

// Action represents a policy action
type Action struct {
	Type        string                 `json:"type"`
	Strategy    string                 `json:"strategy,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// Policy represents a complete policy definition
type Policy struct {
	ID          uuid.UUID  `json:"id"`
	Version     string     `json:"version"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        PolicyType `json:"type"`
	Priority    int        `json:"priority"`
	Conditions  Conditions `json:"conditions"`
	Actions     []Action   `json:"actions"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// EvaluationResult represents the result of a policy evaluation
type EvaluationResult struct {
	PolicyID    uuid.UUID       `json:"policy_id"`
	Success     bool            `json:"success"`
	Matched     bool            `json:"matched"`
	Actions     []Action        `json:"actions,omitempty"`
	Error       string          `json:"error,omitempty"`
	EvaluatedAt time.Time       `json:"evaluated_at"`
	Details     json.RawMessage `json:"details,omitempty"`
}

// SystemState represents the current state of the system
type SystemState struct {
	Nodes   map[string]NodeState   `json:"nodes"`
	Shards  map[string]ShardState  `json:"shards"`
	Metrics map[string]MetricValue `json:"metrics"`
}

// NodeState represents the state of a node
type NodeState struct {
	ID       string                 `json:"id"`
	Status   string                 `json:"status"`
	Metrics  map[string]MetricValue `json:"metrics"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ShardState represents the state of a shard
type ShardState struct {
	ID       string                 `json:"id"`
	NodeID   string                 `json:"node_id"`
	Status   string                 `json:"status"`
	Metrics  map[string]MetricValue `json:"metrics"`
	Metadata map[string]interface{} `json:"metadata"`
}

// MetricValue represents a metric value with timestamp
type MetricValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// PolicyParser defines the interface for parsing policies
type PolicyParser interface {
	Parse(policyJSON []byte) (*Policy, error)
	Validate(policy *Policy) error
}

// PolicyEvaluator defines the interface for evaluating policies
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, policy *Policy, state *SystemState) (*EvaluationResult, error)
	GetApplicablePolicies(ctx context.Context, operationType PolicyType) ([]*Policy, error)
}

// ActionExecutor defines the interface for executing policy actions
type ActionExecutor interface {
	Execute(ctx context.Context, action *Action) error
	Rollback(ctx context.Context, action *Action) error
}
