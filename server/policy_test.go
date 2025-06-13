package server

import (
	"context"
	"testing"

	"github.com/seaweedfs/shardmanager/server/testutil"
	"github.com/seaweedfs/shardmanager/shardmanagerpb"
)

func TestPolicyServiceOperations(t *testing.T) {
	mockDB := testutil.NewMockDB()
	server := NewServer(mockDB)

	t.Run("SetPolicy", func(t *testing.T) {
		req := &shardmanagerpb.SetPolicyRequest{
			PolicyType: "test-policy",
			Parameters: `{"key": "value"}`,
		}
		_, err := server.SetPolicy(context.Background(), req)
		if err != nil {
			t.Errorf("SetPolicy failed: %v", err)
		}
	})

	t.Run("GetPolicy", func(t *testing.T) {
		// Create a policy before the test
		req := &shardmanagerpb.SetPolicyRequest{
			PolicyType: "test-policy",
			Parameters: `{"key": "value"}`,
		}
		_, err := server.SetPolicy(context.Background(), req)
		if err != nil {
			t.Errorf("SetPolicy failed: %v", err)
		}

		// Retrieve the policy
		reqGet := &shardmanagerpb.GetPolicyRequest{
			PolicyType: "test-policy",
		}
		_, err = server.GetPolicy(context.Background(), reqGet)
		if err != nil {
			t.Errorf("GetPolicy failed: %v", err)
		}

		// Assert that the policy was retrieved correctly
		policy, err := mockDB.GetPolicy(context.Background(), "test-policy")
		if err != nil {
			t.Errorf("GetPolicy failed: %v", err)
		}
		if policy == nil {
			t.Errorf("Policy not found")
		}
	})
}
