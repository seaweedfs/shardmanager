package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Node represents a storage node
type Node struct {
	ID            uuid.UUID
	Location      string
	Capacity      int64
	Status        string
	LastHeartbeat time.Time
	CurrentLoad   int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Shard represents a data shard
type Shard struct {
	ID        uuid.UUID
	Type      string
	Size      int64
	NodeID    *uuid.UUID
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Policy represents a shard management policy
type Policy struct {
	ID         uuid.UUID
	PolicyType string
	Parameters json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
