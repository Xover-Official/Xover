package memory

import (
	"context"
	"database/sql"
	"time"
)

// Prediction represents a forecasted metric value
type Prediction struct {
	ResourceID    string
	PredictedAt   time.Time
	ForecastValue float64
	Confidence    float64
	ModelUsed     string
}

// ActionOutcome represents the result of an executed action
type ActionOutcome struct {
	ActionID string
	Type     string
	Success  bool
	Duration int64 // milliseconds
	Savings  float64
	Error    string
}

// MemoryStore handles storing predictions and outcomes for learning
type MemoryStore struct {
	db *sql.DB
}

func NewMemoryStore(db *sql.DB) *MemoryStore {
	return &MemoryStore{db: db}
}

// InitSchema creates the necessary tables for P-S-E memory
func (m *MemoryStore) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS memory_predictions (
			id SERIAL PRIMARY KEY,
			resource_id TEXT NOT NULL,
			predicted_at TIMESTAMPTZ NOT NULL,
			forecast_value DOUBLE PRECISION NOT NULL,
			actual_value DOUBLE PRECISION, -- Filled later for accuracy tracking
			confidence DOUBLE PRECISION,
			model_used TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS memory_actions (
			action_id UUID PRIMARY KEY,
			type TEXT NOT NULL,
			success BOOLEAN NOT NULL,
			duration_ms BIGINT,
			savings DOUBLE PRECISION,
			error TEXT,
			risk_score DOUBLE PRECISION,
			executed_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_pred_resource_time ON memory_predictions(resource_id, predicted_at);`,
		`CREATE INDEX IF NOT EXISTS idx_actions_type_success ON memory_actions(type, success);`,
	}

	for _, query := range queries {
		if _, err := m.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

// RecordPrediction stores a forecast
func (m *MemoryStore) RecordPrediction(ctx context.Context, p Prediction) error {
	query := `INSERT INTO memory_predictions (resource_id, predicted_at, forecast_value, confidence, model_used) 
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := m.db.ExecContext(ctx, query, p.ResourceID, p.PredictedAt, p.ForecastValue, p.Confidence, p.ModelUsed)
	return err
}

// UpdateActuals updates predictions with real values once they occur
func (m *MemoryStore) UpdateActuals(ctx context.Context, resourceID string, timestamp time.Time, actual float64) error {
	// Find predictions close to this timestamp (e.g., within 1 minute)
	query := `UPDATE memory_predictions 
			  SET actual_value = $1 
			  WHERE resource_id = $2 
			  AND predicted_at BETWEEN $3 AND $4
			  AND actual_value IS NULL`

	window := 1 * time.Minute
	_, err := m.db.ExecContext(ctx, query, actual, resourceID, timestamp.Add(-window), timestamp.Add(window))
	return err
}

// RecordOutcome stores the result of an action
func (m *MemoryStore) RecordOutcome(ctx context.Context, o ActionOutcome) error {
	query := `INSERT INTO memory_actions (action_id, type, success, duration_ms, savings, error) 
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := m.db.ExecContext(ctx, query, o.ActionID, o.Type, o.Success, o.Duration, o.Savings, o.Error)
	return err
}

// GetSuccessRate calculates the success rate of a specific action type
func (m *MemoryStore) GetSuccessRate(ctx context.Context, actionType string) (float64, error) {
	query := `SELECT 
				COUNT(*) FILTER (WHERE success) * 100.0 / COUNT(*) 
			  FROM memory_actions 
			  WHERE type = $1 AND executed_at > NOW() - INTERVAL '30 days'`

	var rate float64
	err := m.db.QueryRowContext(ctx, query, actionType).Scan(&rate)
	if err != nil {
		return 0, err
	}
	return rate, nil
}
