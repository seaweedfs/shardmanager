package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// Shard operations
func (db *DB) RegisterShard(ctx context.Context, shard *Shard) error {
	query := `
		INSERT INTO shards (id, type, size, node_id, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	return db.QueryRowContext(ctx, query,
		shard.ID, shard.Type, shard.Size, shard.NodeID, shard.Status,
	).Scan(&shard.CreatedAt, &shard.UpdatedAt)
}

func (db *DB) ListShards(ctx context.Context) ([]*Shard, error) {
	query := `
		SELECT id, type, size, node_id, status, created_at, updated_at
		FROM shards`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shards []*Shard
	for rows.Next() {
		shard := &Shard{}
		err := rows.Scan(
			&shard.ID, &shard.Type, &shard.Size, &shard.NodeID,
			&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		shards = append(shards, shard)
	}
	return shards, rows.Err()
}

func (db *DB) GetShardInfo(ctx context.Context, shardID uuid.UUID) (*Shard, error) {
	query := `
		SELECT id, type, size, node_id, status, created_at, updated_at
		FROM shards
		WHERE id = $1`

	shard := &Shard{}
	err := db.QueryRowContext(ctx, query, shardID).Scan(
		&shard.ID, &shard.Type, &shard.Size, &shard.NodeID,
		&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return shard, nil
}

func (db *DB) AssignShard(ctx context.Context, shardID, nodeID uuid.UUID) error {
	query := `
		UPDATE shards
		SET node_id = $1
		WHERE id = $2`

	_, err := db.ExecContext(ctx, query, nodeID, shardID)
	return err
}

func (db *DB) UpdateShardStatus(ctx context.Context, shardID uuid.UUID, status string) error {
	query := `
		UPDATE shards
		SET status = $1
		WHERE id = $2`

	_, err := db.ExecContext(ctx, query, status, shardID)
	return err
}
