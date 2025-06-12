package server

import (
	"context"
	"testing"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/seaweedfs/shardmanager/server/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyServiceOperations(t *testing.T) {
	server := &Server{db: testutil.NewMockDB()}
	ctx := context.Background()

	t.Run("SetPolicy", func(t *testing.T) {
		req := &shardmanagerpb.SetPolicyRequest{
			PolicyType: "test-policy",
			Parameters: `{"key": "value"}`,
		}

		resp, err := server.SetPolicy(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("GetPolicy", func(t *testing.T) {
		req := &shardmanagerpb.GetPolicyRequest{}

		resp, err := server.GetPolicy(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
