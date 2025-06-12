package server

import (
	"context"
	"testing"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/server/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailureServiceOperations(t *testing.T) {
	server := &Server{db: testutil.NewMockDB()}
	ctx := context.Background()

	t.Run("ReportFailure", func(t *testing.T) {
		entityID := uuid.New()
		req := &shardmanagerpb.ReportFailureRequest{
			Type:    "test-failure",
			Id:      entityID.String(),
			Details: `{"error": "test error"}`,
		}

		resp, err := server.ReportFailure(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
