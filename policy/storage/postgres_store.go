package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/policy"
)

// PostgresPolicyStore implements policy.PolicyStore using *db.DB (which wraps *sql.DB)
type PostgresPolicyStore struct {
	db *db.DB
}

// NewPostgresPolicyStore creates a new PostgresPolicyStore
func NewPostgresPolicyStore(dbConn *db.DB) *PostgresPolicyStore {
	return &PostgresPolicyStore{db: dbConn}
}

// Store implements policy.PolicyStore
func (s *PostgresPolicyStore) Store(ctx context.Context, p *policy.Policy) error {
	params, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	query := `
		INSERT INTO policies (id, policy_type, parameters, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	if s.db.Driver() == "postgres" {
		query = `
			INSERT INTO policies (id, policy_type, parameters, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`
	}
	_, err = s.db.ExecContext(ctx, query, p.ID.String(), string(p.Type), params, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert policy: %w", err)
	}
	return nil
}

// Get implements policy.PolicyStore
func (s *PostgresPolicyStore) Get(ctx context.Context, id string) (*policy.Policy, error) {
	query := `SELECT parameters FROM policies WHERE id = ?`
	if s.db.Driver() == "postgres" {
		query = `SELECT parameters FROM policies WHERE id = $1`
	}
	row := s.db.QueryRowContext(ctx, query, id)
	var params []byte
	if err := row.Scan(&params); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("policy not found: %s", id)
		}
		return nil, fmt.Errorf("failed to scan policy: %w", err)
	}
	var p policy.Policy
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy: %w", err)
	}
	return &p, nil
}

// List implements policy.PolicyStore
func (s *PostgresPolicyStore) List(ctx context.Context) ([]*policy.Policy, error) {
	query := `SELECT parameters FROM policies ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()
	var policies []*policy.Policy
	for rows.Next() {
		var params []byte
		if err := rows.Scan(&params); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		var p policy.Policy
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policy: %w", err)
		}
		policies = append(policies, &p)
	}
	return policies, nil
}

// ListByType implements policy.PolicyStore
func (s *PostgresPolicyStore) ListByType(ctx context.Context, policyType policy.PolicyType) ([]*policy.Policy, error) {
	query := `SELECT parameters FROM policies WHERE policy_type = ? ORDER BY created_at DESC`
	if s.db.Driver() == "postgres" {
		query = `SELECT parameters FROM policies WHERE policy_type = $1 ORDER BY created_at DESC`
	}
	rows, err := s.db.QueryContext(ctx, query, string(policyType))
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()
	var policies []*policy.Policy
	for rows.Next() {
		var params []byte
		if err := rows.Scan(&params); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		var p policy.Policy
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policy: %w", err)
		}
		policies = append(policies, &p)
	}
	return policies, nil
}

// Update implements policy.PolicyStore
func (s *PostgresPolicyStore) Update(ctx context.Context, p *policy.Policy) error {
	params, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}
	query := `UPDATE policies SET parameters = ?, updated_at = ? WHERE id = ?`
	if s.db.Driver() == "postgres" {
		query = `UPDATE policies SET parameters = $1, updated_at = $2 WHERE id = $3`
	}
	_, err = s.db.ExecContext(ctx, query, params, time.Now().UTC(), p.ID.String())
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}
	return nil
}

// Delete implements policy.PolicyStore
func (s *PostgresPolicyStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM policies WHERE id = ?`
	if s.db.Driver() == "postgres" {
		query = `DELETE FROM policies WHERE id = $1`
	}
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	return nil
}
