package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

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

// CreatePolicy inserts a new policy into the database
func CreatePolicy(db *sql.DB, policy *Policy) error {
	query := `
		INSERT INTO policies (id, name, description, type, conditions, actions, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := db.Exec(query, policy.ID.String(), "", "", policy.PolicyType, string(policy.Parameters), "[]")
	return err
}

// GetPolicy retrieves a policy by ID
func GetPolicy(db *sql.DB, id uuid.UUID) (*Policy, error) {
	query := `
		SELECT id, type, conditions, created_at, updated_at
		FROM policies
		WHERE id = ?
	`
	var pid, ptype, params string
	var createdAt, updatedAt time.Time
	err := db.QueryRow(query, id.String()).Scan(&pid, &ptype, &params, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return &Policy{
		ID:         uuid.MustParse(pid),
		PolicyType: ptype,
		Parameters: json.RawMessage(params),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// ListPolicies retrieves all policies
func ListPolicies(db *sql.DB) ([]*Policy, error) {
	query := `
		SELECT id, type, conditions, created_at, updated_at
		FROM policies
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*Policy
	for rows.Next() {
		var pid, ptype, params string
		var createdAt, updatedAt time.Time
		err := rows.Scan(&pid, &ptype, &params, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		policies = append(policies, &Policy{
			ID:         uuid.MustParse(pid),
			PolicyType: ptype,
			Parameters: json.RawMessage(params),
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		})
	}
	return policies, rows.Err()
}

// DeletePolicy removes a policy by ID
func DeletePolicy(db *sql.DB, id uuid.UUID) error {
	query := `DELETE FROM policies WHERE id = ?`
	_, err := db.Exec(query, id.String())
	return err
}
