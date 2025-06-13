package policy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PolicyStore defines the interface for storing and retrieving policies
type PolicyStore interface {
	// Store stores a policy
	Store(ctx context.Context, policy *Policy) error
	// Get retrieves a policy by ID
	Get(ctx context.Context, id string) (*Policy, error)
	// List retrieves all policies
	List(ctx context.Context) ([]*Policy, error)
	// ListByType retrieves policies of a specific type
	ListByType(ctx context.Context, policyType PolicyType) ([]*Policy, error)
	// Delete deletes a policy by ID
	Delete(ctx context.Context, id string) error
	// Update updates an existing policy
	Update(ctx context.Context, policy *Policy) error
}

// InMemoryPolicyStore implements PolicyStore with an in-memory storage
type InMemoryPolicyStore struct {
	policies map[string]*Policy
	mu       sync.RWMutex
}

// NewInMemoryPolicyStore creates a new InMemoryPolicyStore
func NewInMemoryPolicyStore() *InMemoryPolicyStore {
	return &InMemoryPolicyStore{
		policies: make(map[string]*Policy),
	}
}

// Store stores a policy
func (s *InMemoryPolicyStore) Store(ctx context.Context, policy *Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	s.policies[policy.ID.String()] = policy
	return nil
}

// Get retrieves a policy by ID
func (s *InMemoryPolicyStore) Get(ctx context.Context, id string) (*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[id]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", id)
	}

	return policy, nil
}

// List retrieves all policies
func (s *InMemoryPolicyStore) List(ctx context.Context) ([]*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*Policy, 0, len(s.policies))
	for _, policy := range s.policies {
		policies = append(policies, policy)
	}

	return policies, nil
}

// ListByType retrieves policies of a specific type
func (s *InMemoryPolicyStore) ListByType(ctx context.Context, policyType PolicyType) ([]*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var policies []*Policy
	for _, policy := range s.policies {
		if policy.Type == policyType {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// Delete deletes a policy by ID
func (s *InMemoryPolicyStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[id]; !exists {
		return fmt.Errorf("policy not found: %s", id)
	}

	delete(s.policies, id)
	return nil
}

// Update updates an existing policy
func (s *InMemoryPolicyStore) Update(ctx context.Context, policy *Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.ID.String()]; !exists {
		return fmt.Errorf("policy not found: %s", policy.ID)
	}

	policy.UpdatedAt = time.Now()
	s.policies[policy.ID.String()] = policy
	return nil
}
