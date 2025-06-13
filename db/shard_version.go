package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// GetShardVersion retrieves a specific version of a shard
func (db *DB) GetShardVersion(ctx context.Context, shardID uuid.UUID, version int) (*ShardVersion, error) {
	query := `
		SELECT id, shard_id, version, type, size, node_id, status, metadata, created_at
		FROM shard_versions
		WHERE shard_id = $1 AND version = $2
	`
	var sv ShardVersion
	var nodeID sql.NullString
	err := db.QueryRowContext(ctx, query, shardID, version).Scan(
		&sv.ID,
		&sv.ShardID,
		&sv.Version,
		&sv.Type,
		&sv.Size,
		&nodeID,
		&sv.Status,
		&sv.Metadata,
		&sv.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if nodeID.Valid {
		parsedID, err := uuid.Parse(nodeID.String)
		if err != nil {
			return nil, err
		}
		sv.NodeID = &parsedID
	}
	return &sv, nil
}

// ListShardVersions retrieves all versions of a shard
func (db *DB) ListShardVersions(ctx context.Context, shardID uuid.UUID) ([]*ShardVersion, error) {
	query := `
		SELECT id, shard_id, version, type, size, node_id, status, metadata, created_at
		FROM shard_versions
		WHERE shard_id = $1
		ORDER BY version DESC
	`
	rows, err := db.QueryContext(ctx, query, shardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*ShardVersion
	for rows.Next() {
		var sv ShardVersion
		var nodeID sql.NullString
		err := rows.Scan(
			&sv.ID,
			&sv.ShardID,
			&sv.Version,
			&sv.Type,
			&sv.Size,
			&nodeID,
			&sv.Status,
			&sv.Metadata,
			&sv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if nodeID.Valid {
			parsedID, err := uuid.Parse(nodeID.String)
			if err != nil {
				return nil, err
			}
			sv.NodeID = &parsedID
		}
		versions = append(versions, &sv)
	}
	return versions, rows.Err()
}

// UpdateShardVersion updates a shard and creates a new version
func (db *DB) UpdateShardVersion(ctx context.Context, shard *Shard) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current version
	var currentVersion int
	err = tx.QueryRowContext(ctx, "SELECT version FROM shards WHERE id = $1", shard.ID).Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Create new version record
	versionQuery := `
		INSERT INTO shard_versions (shard_id, version, type, size, node_id, status, metadata)
		SELECT id, version, type, size, node_id, status, metadata
		FROM shards
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, versionQuery, shard.ID)
	if err != nil {
		return err
	}

	// Update shard with new version
	updateQuery := `
		UPDATE shards
		SET type = $1, size = $2, node_id = $3, status = $4, metadata = $5, version = version + 1
		WHERE id = $6
		RETURNING version
	`
	err = tx.QueryRowContext(ctx, updateQuery,
		shard.Type,
		shard.Size,
		shard.NodeID,
		shard.Status,
		shard.Metadata,
		shard.ID,
	).Scan(&shard.Version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RollbackShardVersion rolls back a shard to a specific version
func (db *DB) RollbackShardVersion(ctx context.Context, shardID uuid.UUID, version int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the version to rollback to
	var sv ShardVersion
	var nodeID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT id, shard_id, version, type, size, node_id, status, metadata, created_at
		FROM shard_versions
		WHERE shard_id = $1 AND version = $2
	`, shardID, version).Scan(
		&sv.ID,
		&sv.ShardID,
		&sv.Version,
		&sv.Type,
		&sv.Size,
		&nodeID,
		&sv.Status,
		&sv.Metadata,
		&sv.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Create new version record of current state
	versionQuery := `
		INSERT INTO shard_versions (shard_id, version, type, size, node_id, status, metadata)
		SELECT id, version, type, size, node_id, status, metadata
		FROM shards
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, versionQuery, shardID)
	if err != nil {
		return err
	}

	// Update shard with rolled back state
	updateQuery := `
		UPDATE shards
		SET type = $1, size = $2, node_id = $3, status = $4, metadata = $5, version = version + 1
		WHERE id = $6
	`
	_, err = tx.ExecContext(ctx, updateQuery,
		sv.Type,
		sv.Size,
		sv.NodeID,
		sv.Status,
		sv.Metadata,
		shardID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
