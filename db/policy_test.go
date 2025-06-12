package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyDBOperations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	t.Run("SetPolicy", func(t *testing.T) {
		policy := &Policy{
			ID:         uuid.New(),
			PolicyType: "test-policy",
			Parameters: []byte(`{"key": "value"}`),
		}

		err := testDB.SetPolicy(ctx, policy)
		require.NoError(t, err)
		assert.NotZero(t, policy.CreatedAt)
		assert.NotZero(t, policy.UpdatedAt)
	})

	t.Run("GetPolicy", func(t *testing.T) {
		// Create multiple policies of the same type
		policy1 := &Policy{
			ID:         uuid.New(),
			PolicyType: "test-policy-2",
			Parameters: []byte(`{"key": "value1"}`),
		}
		policy2 := &Policy{
			ID:         uuid.New(),
			PolicyType: "test-policy-2",
			Parameters: []byte(`{"key": "value2"}`),
		}

		err := testDB.SetPolicy(ctx, policy1)
		require.NoError(t, err)
		err = testDB.SetPolicy(ctx, policy2)
		require.NoError(t, err)

		// Should get the most recent policy
		policy, err := testDB.GetPolicy(ctx, "test-policy-2")
		require.NoError(t, err)
		assert.Equal(t, "test-policy-2", policy.PolicyType)
		assert.Equal(t, `{"key": "value2"}`, string(policy.Parameters))
	})

	t.Run("GetPolicyNotFound", func(t *testing.T) {
		policy, err := testDB.GetPolicy(ctx, "non-existent-policy")
		require.NoError(t, err)
		assert.Nil(t, policy)
	})
}
