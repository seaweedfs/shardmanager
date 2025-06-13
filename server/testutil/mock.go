package testutil

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/db"
)

// DBOperations defines the interface for database operations
type DBOperations interface {
	RegisterNode(ctx context.Context, node *db.Node) error
	UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID, status string, currentLoad int64) error
	ListNodes(ctx context.Context) ([]*db.Node, error)
	RegisterShard(ctx context.Context, shard *db.Shard) error
	ListShards(ctx context.Context) ([]*db.Shard, error)
	GetShardInfo(ctx context.Context, shardID uuid.UUID) (*db.Shard, error)
	AssignShard(ctx context.Context, shardID, nodeID uuid.UUID) error
	UpdateShardStatus(ctx context.Context, shardID uuid.UUID, status string) error
	SetPolicy(ctx context.Context, policy *db.Policy) error
	GetPolicy(ctx context.Context, policyType string) (*db.Policy, error)
	ReportFailure(ctx context.Context, failureType string, entityID uuid.UUID, details json.RawMessage) error
	Reset()
	GetShardVersion(ctx context.Context, shardID uuid.UUID, version int) (*db.ShardVersion, error)
	ListShardVersions(ctx context.Context, shardID uuid.UUID) ([]*db.ShardVersion, error)
	UpdateShardVersion(ctx context.Context, shard *db.Shard) error
	RollbackShardVersion(ctx context.Context, shardID uuid.UUID, version int) error
	GetNodeInfo(ctx context.Context, nodeID uuid.UUID) (*db.Node, error)
}

// MockDB implements DBOperations for testing
type MockDB struct {
	nodes    map[uuid.UUID]*db.Node
	shards   map[uuid.UUID]*db.Shard
	policies map[string]*db.Policy
}

// NewMockDB creates a new mock database instance
func NewMockDB() DBOperations {
	return &MockDB{
		nodes:    make(map[uuid.UUID]*db.Node),
		shards:   make(map[uuid.UUID]*db.Shard),
		policies: make(map[string]*db.Policy),
	}
}

// Reset clears the mock database state
func (m *MockDB) Reset() {
	m.nodes = make(map[uuid.UUID]*db.Node)
	m.shards = make(map[uuid.UUID]*db.Shard)
	m.policies = make(map[string]*db.Policy)
}

// RegisterNode mocks the RegisterNode operation
func (m *MockDB) RegisterNode(ctx context.Context, node *db.Node) error {
	m.nodes[node.ID] = node
	log.Printf("Registered node: %v", node)
	return nil
}

// UpdateNodeHeartbeat mocks the UpdateNodeHeartbeat operation
func (m *MockDB) UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID, status string, currentLoad int64) error {
	if node, ok := m.nodes[nodeID]; ok {
		node.Status = status
		node.CurrentLoad = currentLoad
		log.Printf("Updated node heartbeat: %v", node)
		return nil
	}
	return nil
}

// ListNodes mocks the ListNodes operation
func (m *MockDB) ListNodes(ctx context.Context) ([]*db.Node, error) {
	nodes := make([]*db.Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	log.Printf("Listed nodes: %v", nodes)
	return nodes, nil
}

// RegisterShard mocks the RegisterShard operation
func (m *MockDB) RegisterShard(ctx context.Context, shard *db.Shard) error {
	m.shards[shard.ID] = shard
	log.Printf("Registered shard: %v", shard)
	return nil
}

// ListShards mocks the ListShards operation
func (m *MockDB) ListShards(ctx context.Context) ([]*db.Shard, error) {
	shards := make([]*db.Shard, 0, len(m.shards))
	for _, shard := range m.shards {
		shards = append(shards, shard)
	}
	log.Printf("Listed shards: %v", shards)
	return shards, nil
}

// GetShardInfo mocks the GetShardInfo operation
func (m *MockDB) GetShardInfo(ctx context.Context, shardID uuid.UUID) (*db.Shard, error) {
	if shard, ok := m.shards[shardID]; ok {
		log.Printf("Retrieved shard: %v", shard)
		return shard, nil
	}
	log.Printf("Shard not found: %v", shardID)
	return nil, nil
}

// AssignShard mocks the AssignShard operation
func (m *MockDB) AssignShard(ctx context.Context, shardID, nodeID uuid.UUID) error {
	if shard, ok := m.shards[shardID]; ok {
		shard.NodeID = &nodeID
		log.Printf("Assigned shard: %v", shard)
		return nil
	}
	return nil
}

// UpdateShardStatus mocks the UpdateShardStatus operation
func (m *MockDB) UpdateShardStatus(ctx context.Context, shardID uuid.UUID, status string) error {
	if shard, ok := m.shards[shardID]; ok {
		shard.Status = status
		log.Printf("Updated shard status: %v", shard)
		return nil
	}
	return nil
}

// SetPolicy mocks the SetPolicy operation
func (m *MockDB) SetPolicy(ctx context.Context, policy *db.Policy) error {
	m.policies[policy.PolicyType] = policy
	log.Printf("Set policy: %v", policy)
	return nil
}

// GetPolicy mocks the GetPolicy operation
func (m *MockDB) GetPolicy(ctx context.Context, policyType string) (*db.Policy, error) {
	if policy, ok := m.policies[policyType]; ok {
		log.Printf("Retrieved policy: %v", policy)
		return policy, nil
	}
	log.Printf("Policy not found: %v", policyType)
	return nil, nil
}

// ReportFailure mocks the ReportFailure operation
func (m *MockDB) ReportFailure(ctx context.Context, failureType string, entityID uuid.UUID, details json.RawMessage) error {
	return nil
}

// GetShardVersion mocks the GetShardVersion operation
func (m *MockDB) GetShardVersion(ctx context.Context, shardID uuid.UUID, version int) (*db.ShardVersion, error) {
	return nil, nil
}

// ListShardVersions mocks the ListShardVersions operation
func (m *MockDB) ListShardVersions(ctx context.Context, shardID uuid.UUID) ([]*db.ShardVersion, error) {
	return nil, nil
}

// UpdateShardVersion mocks the UpdateShardVersion operation
func (m *MockDB) UpdateShardVersion(ctx context.Context, shard *db.Shard) error {
	return nil
}

// RollbackShardVersion mocks the RollbackShardVersion operation
func (m *MockDB) RollbackShardVersion(ctx context.Context, shardID uuid.UUID, version int) error {
	return nil
}

// GetNodeInfo mocks the GetNodeInfo operation
func (m *MockDB) GetNodeInfo(ctx context.Context, nodeID uuid.UUID) (*db.Node, error) {
	node, ok := m.nodes[nodeID]
	if !ok {
		return nil, nil
	}
	return node, nil
}
