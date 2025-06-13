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

	// Node lookup
	GetNodeInfo(ctx context.Context, nodeID uuid.UUID) (*Node, error)
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

func InitSQLiteSchema(db *DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    location TEXT NOT NULL,
    capacity INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    last_heartbeat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_load INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS shards (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    size INTEGER NOT NULL,
    node_id TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    version INTEGER NOT NULL DEFAULT 1,
    metadata TEXT DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS shard_versions (
    id TEXT PRIMARY KEY,
    shard_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    type TEXT NOT NULL,
    size INTEGER NOT NULL,
    node_id TEXT,
    status TEXT NOT NULL,
    metadata TEXT DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(shard_id, version)
);
CREATE TABLE IF NOT EXISTS shard_migrations (
    id TEXT PRIMARY KEY,
    shard_id TEXT NOT NULL,
    from_node_id TEXT,
    to_node_id TEXT,
    status TEXT NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    error_message TEXT
);
CREATE TABLE IF NOT EXISTS policies (
    id TEXT PRIMARY KEY,
    policy_type TEXT NOT NULL,
    parameters TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS failure_reports (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    details TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_shards_node_id ON shards(node_id);
CREATE INDEX IF NOT EXISTS idx_shards_status ON shards(status);
CREATE INDEX IF NOT EXISTS idx_shards_version ON shards(version);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
CREATE INDEX IF NOT EXISTS idx_nodes_last_heartbeat ON nodes(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_shard_migrations_shard_id ON shard_migrations(shard_id);
CREATE INDEX IF NOT EXISTS idx_shard_versions_shard_id ON shard_versions(shard_id);
CREATE INDEX IF NOT EXISTS idx_failure_reports_entity_id ON failure_reports(entity_id);
`
	_, err := db.Exec(schema)
	return err
}
