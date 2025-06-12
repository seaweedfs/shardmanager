package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Node operations
func (db *DB) RegisterNode(ctx context.Context, node *Node) error {
	query := `
		INSERT INTO nodes (id, location, capacity, status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`

	return db.QueryRowContext(ctx, query,
		node.ID, node.Location, node.Capacity, node.Status,
	).Scan(&node.CreatedAt, &node.UpdatedAt)
}

func (db *DB) UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID, status string, load int64) error {
	query := `
		UPDATE nodes 
		SET last_heartbeat = CURRENT_TIMESTAMP, status = $1, current_load = $2
		WHERE id = $3`

	_, err := db.ExecContext(ctx, query, status, load, nodeID)
	return err
}

func (db *DB) ListNodes(ctx context.Context) ([]*Node, error) {
	query := `
		SELECT id, location, capacity, status, last_heartbeat, current_load, created_at, updated_at
		FROM nodes`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		err := rows.Scan(
			&node.ID, &node.Location, &node.Capacity, &node.Status,
			&node.LastHeartbeat, &node.CurrentLoad, &node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

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

// Policy operations
func (db *DB) SetPolicy(ctx context.Context, policy *Policy) error {
	query := `
		INSERT INTO policies (id, policy_type, parameters)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at`

	return db.QueryRowContext(ctx, query,
		policy.ID, policy.PolicyType, policy.Parameters,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)
}

func (db *DB) GetPolicy(ctx context.Context, policyType string) (*Policy, error) {
	query := `
		SELECT id, policy_type, parameters, created_at, updated_at
		FROM policies
		WHERE policy_type = $1
		ORDER BY created_at DESC
		LIMIT 1`

	policy := &Policy{}
	err := db.QueryRowContext(ctx, query, policyType).Scan(
		&policy.ID, &policy.PolicyType, &policy.Parameters,
		&policy.CreatedAt, &policy.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// Failure operations
func (db *DB) ReportFailure(ctx context.Context, failureType string, entityID uuid.UUID, details json.RawMessage) error {
	query := `
		INSERT INTO failures (type, entity_id, details)
		VALUES ($1, $2, $3)`

	_, err := db.ExecContext(ctx, query, failureType, entityID, details)
	return err
}
