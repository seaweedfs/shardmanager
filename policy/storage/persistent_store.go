package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seaweedfs/shardmanager/policy"
)

// PersistentPolicyStore implements policy storage using SQLite
type PersistentPolicyStore struct {
	db   *sql.DB
	mu   sync.RWMutex
	path string
}

// NewPersistentPolicyStore creates a new persistent policy store
func NewPersistentPolicyStore(dbPath string) (*PersistentPolicyStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &PersistentPolicyStore{
		db:   db,
		path: dbPath,
	}

	if err := store.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

// initialize sets up the database schema
func (s *PersistentPolicyStore) initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create policies table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			priority INTEGER NOT NULL,
			conditions TEXT NOT NULL,
			actions TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create policies table: %w", err)
	}

	// Create policy_history table for tracking changes
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS policy_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			policy_id TEXT NOT NULL,
			action TEXT NOT NULL,
			details TEXT NULL,
			timestamp TIMESTAMP NOT NULL,
			FOREIGN KEY (policy_id) REFERENCES policies(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create policy_history table: %w", err)
	}

	return nil
}

// Store saves a policy to the database
func (s *PersistentPolicyStore) Store(ctx context.Context, p *policy.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conditionsJSON, err := json.Marshal(p.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(p.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	now := time.Now().UTC()

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or update policy
	_, err = tx.ExecContext(ctx, `
		INSERT INTO policies (id, name, description, type, priority, conditions, actions, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			type = excluded.type,
			priority = excluded.priority,
			conditions = excluded.conditions,
			actions = excluded.actions,
			updated_at = excluded.updated_at
	`, p.ID.String(), p.Name, p.Description, p.Type, p.Priority, conditionsJSON, actionsJSON, now, now)
	if err != nil {
		return fmt.Errorf("failed to store policy: %w", err)
	}

	// Record history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO policy_history (policy_id, action, details, timestamp)
		VALUES (?, ?, ?, ?)
	`, p.ID.String(), "store", string(conditionsJSON), now)
	if err != nil {
		return fmt.Errorf("failed to record history: %w", err)
	}

	return tx.Commit()
}

// Get retrieves a policy by ID
func (s *PersistentPolicyStore) Get(ctx context.Context, id uuid.UUID) (*policy.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var p policy.Policy
	var conditionsJSON, actionsJSON string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, type, priority, conditions, actions
		FROM policies
		WHERE id = ?
	`, id.String()).Scan(&p.ID, &p.Name, &p.Description, &p.Type, &p.Priority, &conditionsJSON, &actionsJSON)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if err := json.Unmarshal([]byte(conditionsJSON), &p.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	if err := json.Unmarshal([]byte(actionsJSON), &p.Actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
	}

	return &p, nil
}

// List returns all policies
func (s *PersistentPolicyStore) List(ctx context.Context) ([]*policy.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, type, priority, conditions, actions
		FROM policies
		ORDER BY priority DESC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []*policy.Policy
	for rows.Next() {
		var p policy.Policy
		var conditionsJSON, actionsJSON string
		var idStr string

		if err := rows.Scan(&idStr, &p.Name, &p.Description, &p.Type, &p.Priority, &conditionsJSON, &actionsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		p.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse policy ID: %w", err)
		}

		if err := json.Unmarshal([]byte(conditionsJSON), &p.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		if err := json.Unmarshal([]byte(actionsJSON), &p.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}

		policies = append(policies, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating policies: %w", err)
	}

	return policies, nil
}

// Delete removes a policy
func (s *PersistentPolicyStore) Delete(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete policy
	result, err := tx.ExecContext(ctx, "DELETE FROM policies WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("policy not found: %s", id)
	}

	// Record history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO policy_history (policy_id, action, timestamp)
		VALUES (?, ?, ?)
	`, id.String(), "delete", time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to record history: %w", err)
	}

	return tx.Commit()
}

// GetHistory returns the change history for a policy
func (s *PersistentPolicyStore) GetHistory(ctx context.Context, id uuid.UUID) ([]PolicyHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT action, details, timestamp
		FROM policy_history
		WHERE policy_id = ?
		ORDER BY timestamp DESC
	`, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var history []PolicyHistory
	for rows.Next() {
		var h PolicyHistory
		var details sql.NullString
		if err := rows.Scan(&h.Action, &details, &h.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		if details.Valid {
			h.Details = details.String
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}

// Close closes the database connection
func (s *PersistentPolicyStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

// PolicyHistory represents a change to a policy
type PolicyHistory struct {
	Action    string
	Details   string
	Timestamp time.Time
}
