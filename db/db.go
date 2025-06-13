package db

import (
	"context"
	"database/sql"
	"encoding/json"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
)

type DB struct {
	*sql.DB
	DriverName string
}

// DBOperations defines the interface for database operations (production, no Reset)
type DBOperations interface {
	RegisterNode(ctx context.Context, node *Node) error
	UpdateNodeHeartbeat(ctx context.Context, nodeID uuid.UUID, status string, currentLoad int64) error
	ListNodes(ctx context.Context) ([]*Node, error)
	RegisterShard(ctx context.Context, shard *Shard) error
	ListShards(ctx context.Context) ([]*Shard, error)
	GetShardInfo(ctx context.Context, shardID uuid.UUID) (*Shard, error)
	AssignShard(ctx context.Context, shardID, nodeID uuid.UUID) error
	UpdateShardStatus(ctx context.Context, shardID uuid.UUID, status string) error
	SetPolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, policyType string) (*Policy, error)
	ReportFailure(ctx context.Context, failureType string, entityID uuid.UUID, details json.RawMessage) error

	// Version-related operations
	GetShardVersion(ctx context.Context, shardID uuid.UUID, version int) (*ShardVersion, error)
	ListShardVersions(ctx context.Context, shardID uuid.UUID) ([]*ShardVersion, error)
	UpdateShardVersion(ctx context.Context, shard *Shard) error
	RollbackShardVersion(ctx context.Context, shardID uuid.UUID, version int) error
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db, DriverName: "postgres"}, nil
}

// NewDBWithDriver allows specifying the SQL driver (e.g., "sqlite3" or "postgres")
func NewDBWithDriver(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db, DriverName: driver}, nil
}

func (db *DB) Driver() string {
	return db.DriverName
}
