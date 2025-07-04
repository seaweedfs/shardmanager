package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// CreateNode creates a new node in the database
func CreateNode(db *sql.DB, node *Node) error {
	query := `
		INSERT INTO nodes (id, location, capacity, status)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at
	`
	return db.QueryRow(query, node.ID, node.Location, node.Capacity, node.Status).Scan(&node.CreatedAt, &node.UpdatedAt)
}

// GetNode retrieves a node by ID
func GetNode(db *sql.DB, id uuid.UUID) (*Node, error) {
	query := `
		SELECT id, location, capacity, status, last_heartbeat, current_load, created_at, updated_at
		FROM nodes
		WHERE id = ?
	`
	node := &Node{}
	var lastHeartbeat sql.NullTime
	err := db.QueryRow(query, id).Scan(
		&node.ID,
		&node.Location,
		&node.Capacity,
		&node.Status,
		&lastHeartbeat,
		&node.CurrentLoad,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastHeartbeat.Valid {
		node.LastHeartbeat = lastHeartbeat.Time
	}
	return node, nil
}

// UpdateNode updates an existing node
func UpdateNode(db *sql.DB, node *Node) error {
	query := `
		UPDATE nodes
		SET location = ?, capacity = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, node.Location, node.Capacity, node.Status, node.ID)
	return err
}

// ListNodes retrieves all nodes
func ListNodes(db *sql.DB) ([]*Node, error) {
	query := `
		SELECT id, location, capacity, status, last_heartbeat, current_load, created_at, updated_at
		FROM nodes
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		var lastHeartbeat sql.NullTime
		err := rows.Scan(
			&node.ID,
			&node.Location,
			&node.Capacity,
			&node.Status,
			&lastHeartbeat,
			&node.CurrentLoad,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if lastHeartbeat.Valid {
			node.LastHeartbeat = lastHeartbeat.Time
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

// DeleteNode removes a node by ID
func DeleteNode(db *sql.DB, id uuid.UUID) error {
	query := `DELETE FROM nodes WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
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

func (db *DB) GetNodeInfo(ctx context.Context, nodeID uuid.UUID) (*Node, error) {
	query := `
		SELECT id, location, capacity, status, last_heartbeat, current_load, created_at, updated_at
		FROM nodes
		WHERE id = $1`
	node := &Node{}
	var lastHeartbeat sql.NullTime
	err := db.QueryRowContext(ctx, query, nodeID).Scan(
		&node.ID,
		&node.Location,
		&node.Capacity,
		&node.Status,
		&lastHeartbeat,
		&node.CurrentLoad,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastHeartbeat.Valid {
		node.LastHeartbeat = lastHeartbeat.Time
	}
	return node, nil
}
