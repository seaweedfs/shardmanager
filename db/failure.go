package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// Failure operations
func (db *DB) ReportFailure(ctx context.Context, failureType string, entityID uuid.UUID, details json.RawMessage) error {
	query := `
		INSERT INTO failure_reports (type, entity_id, details)
		VALUES ($1, $2, $3)`

	_, err := db.ExecContext(ctx, query, failureType, entityID, details)
	return err
}
