package db

import (
	"context"

	"github.com/google/uuid"
)

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
