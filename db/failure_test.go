package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFailureDBReporting(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	t.Run("ReportFailure", func(t *testing.T) {
		entityID := uuid.New()
		details := []byte(`{"error": "test error"}`)

		err := testDB.ReportFailure(ctx, "test-failure", entityID, details)
		require.NoError(t, err)
	})

	t.Run("ReportMultipleFailures", func(t *testing.T) {
		// Report failures for different entities
		entity1ID := uuid.New()
		entity2ID := uuid.New()

		err := testDB.ReportFailure(ctx, "failure-type-1", entity1ID, []byte(`{"error": "error1"}`))
		require.NoError(t, err)

		err = testDB.ReportFailure(ctx, "failure-type-2", entity2ID, []byte(`{"error": "error2"}`))
		require.NoError(t, err)
	})

	t.Run("ReportFailureWithComplexDetails", func(t *testing.T) {
		entityID := uuid.New()
		details := []byte(`{
			"error": "complex error",
			"timestamp": "2024-02-14T12:00:00Z",
			"context": {
				"operation": "shard-migration",
				"source": "node-1",
				"destination": "node-2"
			},
			"metrics": {
				"duration_ms": 1500,
				"retry_count": 3
			}
		}`)

		err := testDB.ReportFailure(ctx, "complex-failure", entityID, details)
		require.NoError(t, err)
	})
}
