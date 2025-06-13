package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Shard operations
func (db *DB) RegisterShard(ctx context.Context, shard *Shard) error {
	query := `
		INSERT INTO shards (id, type, size, node_id, status, version, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return db.QueryRowContext(ctx, query,
		shard.ID, shard.Type, shard.Size, shard.NodeID, shard.Status, shard.Version, shard.Metadata,
	).Scan(&shard.CreatedAt, &shard.UpdatedAt)
}

func (db *DB) ListShards(ctx context.Context) ([]*Shard, error) {
	query := `
		SELECT id, type, size, node_id, status, version, metadata, created_at, updated_at
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
			&shard.Status, &shard.Version, &shard.Metadata,
			&shard.CreatedAt, &shard.UpdatedAt,
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
		SELECT id, type, size, node_id, status, version, metadata, created_at, updated_at
		FROM shards
		WHERE id = $1`

	shard := &Shard{}
	err := db.QueryRowContext(ctx, query, shardID).Scan(
		&shard.ID, &shard.Type, &shard.Size, &shard.NodeID,
		&shard.Status, &shard.Version, &shard.Metadata,
		&shard.CreatedAt, &shard.UpdatedAt,
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
		SET node_id = $1, version = version + 1
		WHERE id = $2`

	_, err := db.ExecContext(ctx, query, nodeID, shardID)
	return err
}

func (db *DB) UpdateShardStatus(ctx context.Context, shardID uuid.UUID, status string) error {
	query := `
		UPDATE shards
		SET status = $1, version = version + 1
		WHERE id = $2`

	_, err := db.ExecContext(ctx, query, status, shardID)
	return err
}

// CreateShard inserts a new shard into the database
func CreateShard(db *sql.DB, shard *Shard) error {
	query := `
		INSERT INTO shards (id, node_id, type, size, status, version, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING created_at, updated_at
	`
	return db.QueryRow(query,
		shard.ID.String(), shard.NodeID.String(), shard.Type, shard.Size,
		shard.Status, shard.Version, shard.Metadata,
	).Scan(&shard.CreatedAt, &shard.UpdatedAt)
}

// GetShard retrieves a shard by ID
func GetShard(db *sql.DB, id uuid.UUID) (*Shard, error) {
	query := `
		SELECT id, node_id, type, size, status, version, metadata, created_at, updated_at
		FROM shards
		WHERE id = ?
	`
	var sid, nodeID, stype, status string
	var size int64
	var version int
	var metadata []byte
	var createdAt, updatedAt time.Time
	err := db.QueryRow(query, id.String()).Scan(
		&sid, &nodeID, &stype, &size, &status,
		&version, &metadata,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	parsedNodeID := uuid.MustParse(nodeID)
	return &Shard{
		ID:        uuid.MustParse(sid),
		NodeID:    &parsedNodeID,
		Type:      stype,
		Size:      size,
		Status:    status,
		Version:   version,
		Metadata:  metadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// ListShards retrieves all shards
func ListShards(db *sql.DB) ([]*Shard, error) {
	query := `
		SELECT id, node_id, type, size, status, version, metadata, created_at, updated_at
		FROM shards
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shards []*Shard
	for rows.Next() {
		var sid, nodeID, stype, status string
		var size int64
		var version int
		var metadata []byte
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&sid, &nodeID, &stype, &size, &status,
			&version, &metadata,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}
		parsedNodeID := uuid.MustParse(nodeID)
		shards = append(shards, &Shard{
			ID:        uuid.MustParse(sid),
			NodeID:    &parsedNodeID,
			Type:      stype,
			Size:      size,
			Status:    status,
			Version:   version,
			Metadata:  metadata,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	return shards, rows.Err()
}

// UpdateShard updates an existing shard
func UpdateShard(db *sql.DB, shard *Shard) error {
	query := `
		UPDATE shards
		SET node_id = ?, type = ?, size = ?, status = ?, version = version + 1, metadata = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query,
		shard.NodeID.String(), shard.Type, shard.Size,
		shard.Status, shard.Metadata, shard.ID.String(),
	)
	return err
}

// DeleteShard removes a shard by ID
func DeleteShard(db *sql.DB, id uuid.UUID) error {
	query := `DELETE FROM shards WHERE id = ?`
	_, err := db.Exec(query, id.String())
	return err
}

// AssignShard assigns a shard to a new node
func AssignShard(db *sql.DB, shardID uuid.UUID, nodeID uuid.UUID) error {
	query := `
		UPDATE shards
		SET node_id = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, nodeID.String(), shardID.String())
	return err
}
