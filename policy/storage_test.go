package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryPolicyStore_CRUD(t *testing.T) {
	store := NewInMemoryPolicyStore()
	ctx := context.Background()

	p := &Policy{
		Name: "test-policy",
		Type: PolicyTypeLoadBalancing,
		Conditions: Conditions{
			All: []Condition{{Metric: "cpu_usage", Operator: OperatorGreaterThan, Value: 80.0}},
		},
		Actions: []Action{{Type: "migrate_shard"}},
	}

	// Store
	err := store.Store(ctx, p)
	assert.NoError(t, err)
	assert.NotEqual(t, "", p.ID.String())

	// Get
	got, err := store.Get(ctx, p.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, p.Name, got.Name)

	// List
	policies, err := store.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, policies, 1)

	// ListByType
	byType, err := store.ListByType(ctx, PolicyTypeLoadBalancing)
	assert.NoError(t, err)
	assert.Len(t, byType, 1)

	// Update
	p.Name = "updated-policy"
	err = store.Update(ctx, p)
	assert.NoError(t, err)
	got, _ = store.Get(ctx, p.ID.String())
	assert.Equal(t, "updated-policy", got.Name)

	// Delete
	err = store.Delete(ctx, p.ID.String())
	assert.NoError(t, err)
	_, err = store.Get(ctx, p.ID.String())
	assert.Error(t, err)
}
