package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLedger implements the Ledger interface using PostgreSQL
type PostgresLedger struct {
	pool *pgxpool.Pool
}

// NewPostgresLedger creates a new PostgreSQL-backed ledger
func NewPostgresLedger(connString string) (*PostgresLedger, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresLedger{pool: pool}, nil
}

// RecordAction records a new action in the ledger
func (p *PostgresLedger) RecordAction(ctx context.Context, action *Action) error {

	payloadJSON, err := json.Marshal(action.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO actions (
			resource_id, action_type, status, checksum, payload, 
			risk_score, estimated_savings, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id string
	err = p.pool.QueryRow(ctx, query,
		action.ResourceID,
		action.ActionType,
		action.Status,
		action.Checksum,
		payloadJSON,
		action.RiskScore,
		action.EstimatedSavings,
		time.Now(),
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to insert action: %w", err)
	}

	action.ID = id
	return nil
}

// GetPendingActions retrieves all pending actions for recovery
func (p *PostgresLedger) GetPendingActions(ctx context.Context) ([]Action, error) {

	query := `
		SELECT id, resource_id, action_type, status, checksum, payload,
		       risk_score, estimated_savings, created_at
		FROM actions
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
	`

	rows, err := p.pool.Query(ctx, query)
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
func (p *PostgresLedger) MarkComplete(ctx context.Context, actionID string) error {

	query := `
		UPDATE actions
		SET status = 'COMPLETED', completed_at = $1
		WHERE id = $2
	`

	_, err := p.pool.Exec(ctx, query, time.Now(), actionID)
	if err != nil {
		return fmt.Errorf("failed to mark action complete: %w", err)
	}

	return nil
}

// MarkFailed marks an action as failed with an error message
func (p *PostgresLedger) MarkFailed(ctx context.Context, actionID string, errorMsg string) error {

	query := `
		UPDATE actions
		SET status = 'FAILED', completed_at = $1, error_message = $2
		WHERE id = $3
	`

	_, err := p.pool.Exec(ctx, query, time.Now(), errorMsg, actionID)
	if err != nil {
		return fmt.Errorf("failed to mark action failed: %w", err)
	}

	return nil
}

// GetActionByChecksum retrieves an action by its checksum (for idempotency)
func (p *PostgresLedger) GetActionByChecksum(ctx context.Context, checksum string) (*Action, error) {

	query := `
		SELECT id, resource_id, action_type, status, checksum, payload,
		       risk_score, estimated_savings, created_at, completed_at
		FROM actions
		WHERE checksum = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var action Action
	var payloadJSON []byte
	var completedAt *time.Time

	err := p.pool.QueryRow(ctx, query, checksum).Scan(
		&action.ID,
		&action.ResourceID,
		&action.ActionType,
		&action.Status,
		&action.Checksum,
		&payloadJSON,
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

// Close closes the database connection pool
func (p *PostgresLedger) Close() {
	p.pool.Close()
}

// GetStats returns statistics about the ledger
func (p *PostgresLedger) GetStats(ctx context.Context) (map[string]int, error) {

	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'PENDING') as pending,
			COUNT(*) FILTER (WHERE status = 'COMPLETED') as completed,
			COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
			COUNT(*) as total
		FROM actions
	`

	var stats struct {
		Pending   int
		Completed int
		Failed    int
		Total     int
	}

	err := p.pool.QueryRow(ctx, query).Scan(
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
