package server

import (
	"context"
	"testing"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/seaweedfs/shardmanager/server/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitoringService(t *testing.T) {
	server := &Server{db: testutil.NewMockDB()}
	ctx := context.Background()

	t.Run("GetDistribution", func(t *testing.T) {
		resp, err := server.GetDistribution(ctx, &shardmanagerpb.GetDistributionRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("GetHealth", func(t *testing.T) {
		resp, err := server.GetHealth(ctx, &shardmanagerpb.GetHealthRequest{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
