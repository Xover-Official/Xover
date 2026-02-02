package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteLedger implements the Ledger interface using SQLite
type SQLiteLedger struct {
	db *sql.DB
}

// NewSQLiteLedger creates a new SQLite-backed ledger
func NewSQLiteLedger(dbPath string) (*SQLiteLedger, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_id TEXT NOT NULL,
		action_type TEXT NOT NULL,
		status TEXT NOT NULL,
		checksum TEXT NOT NULL,
		payload TEXT,
		reasoning TEXT,
		risk_score REAL,
		estimated_savings REAL,
		created_at DATETIME NOT NULL,
		started_at DATETIME,
		completed_at DATETIME,
		error_message TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_status ON actions(status);
	CREATE INDEX IF NOT EXISTS idx_checksum ON actions(checksum);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &SQLiteLedger{db: db}, nil
}

// RecordAction records a new action in the ledger
func (s *SQLiteLedger) RecordAction(ctx context.Context, action *Action) error {

	payloadJSON, err := json.Marshal(action.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO actions (
			resource_id, action_type, status, checksum, payload, reasoning,
			risk_score, estimated_savings, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		action.ResourceID,
		action.ActionType,
		action.Status,
		action.Checksum,
		string(payloadJSON),
		action.Reasoning,
		action.RiskScore,
		action.EstimatedSavings,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert action: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	action.ID = fmt.Sprintf("%d", id)
	return nil
}

// GetPendingActions retrieves all pending actions for recovery
func (s *SQLiteLedger) GetPendingActions(ctx context.Context) ([]Action, error) {

	query := `
		SELECT id, resource_id, action_type, status, checksum, payload, reasoning,
		       risk_score, estimated_savings, created_at
		FROM actions
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending actions: %w", err)
	}
	defer rows.Close()

	var actions []Action
	for rows.Next() {
		var action Action
		var payloadJSON []byte

		err := rows.Scan(
			&action.ID,
			&action.ResourceID,
			&action.ActionType,
			&action.Status,
			&action.Checksum,
			&payloadJSON,
			&action.Reasoning,
			&action.RiskScore,
			&action.EstimatedSavings,
			&action.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &action.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// MarkComplete marks an action as completed
func (s *SQLiteLedger) MarkComplete(ctx context.Context, actionID string) error {

	query := `UPDATE actions SET status = 'completed', completed_at = ? WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, time.Now(), actionID)
	if err != nil {
		return fmt.Errorf("failed to mark action complete: %w", err)
	}

	return nil
}

// MarkFailed marks an action as failed with an error message
func (s *SQLiteLedger) MarkFailed(ctx context.Context, actionID string, errorMsg string) error {

	query := `UPDATE actions SET status = 'failed', completed_at = ?, error_message = ? WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, time.Now(), errorMsg, actionID)
	if err != nil {
		return fmt.Errorf("failed to mark action failed: %w", err)
	}

	return nil
}

// GetActionByChecksum retrieves an action by its checksum (for idempotency)
func (s *SQLiteLedger) GetActionByChecksum(ctx context.Context, checksum string) (*Action, error) {

	query := `
		SELECT id, resource_id, action_type, status, checksum, payload, reasoning,
		       risk_score, estimated_savings, created_at, completed_at
		FROM actions
		WHERE checksum = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var action Action
	var payloadJSON []byte
	var completedAt *time.Time

	err := s.db.QueryRowContext(ctx, query, checksum).Scan(
		&action.ID,
		&action.ResourceID,
		&action.ActionType,
		&action.Status,
		&action.Checksum,
		&payloadJSON,
		&action.Reasoning,
		&action.RiskScore,
		&action.EstimatedSavings,
		&action.CreatedAt,
		&completedAt,
	)

	if err != nil {
		return nil, err // Not found or other error
	}

	if err := json.Unmarshal(payloadJSON, &action.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	action.CompletedAt = completedAt
	return &action, nil
}

// GetStats returns statistics about the ledger
func (s *SQLiteLedger) GetStats(ctx context.Context) (map[string]int, error) {

	query := `
		SELECT 
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			COUNT(*) as total
		FROM actions
	`

	var stats struct {
		Pending   int
		Completed int
		Failed    int
		Total     int
	}

	err := s.db.QueryRowContext(ctx, query).Scan(
		&stats.Pending,
		&stats.Completed,
		&stats.Failed,
		&stats.Total,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return map[string]int{
		"pending":   stats.Pending,
		"completed": stats.Completed,
		"failed":    stats.Failed,
		"total":     stats.Total,
	}, nil
}

// Close closes the database connection
func (s *SQLiteLedger) Close() {
	s.db.Close()
}
