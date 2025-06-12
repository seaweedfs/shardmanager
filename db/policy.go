package db

import (
	"context"
	"database/sql"
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
