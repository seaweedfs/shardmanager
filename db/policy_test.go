package db

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Test CreatePolicy
	policyID := uuid.New()
	params, _ := json.Marshal(map[string]interface{}{"foo": "bar"})
	policy := &Policy{
		ID:         policyID,
		PolicyType: "test-type",
		Parameters: params,
	}
	err := CreatePolicy(db, policy)
	require.NoError(t, err)

	// Test GetPolicy
	retrieved, err := GetPolicy(db, policyID)
	require.NoError(t, err)
	assert.Equal(t, policy.ID, retrieved.ID)
	assert.Equal(t, policy.PolicyType, retrieved.PolicyType)
	assert.JSONEq(t, string(policy.Parameters), string(retrieved.Parameters))

	// Test ListPolicies
	policies, err := ListPolicies(db)
	require.NoError(t, err)
	assert.Len(t, policies, 1)

	// Test DeletePolicy
	err = DeletePolicy(db, policyID)
	require.NoError(t, err)

	_, err = GetPolicy(db, policyID)
	assert.Error(t, err)
}
