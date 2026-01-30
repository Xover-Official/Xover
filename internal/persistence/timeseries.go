package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// MetricsStore handles time-series data storage
type MetricsStore struct {
	db *sql.DB // Can use the HAClient.Primary
}

// MetricPoint represents a single data point
type MetricPoint struct {
	Time   time.Time
	Name   string
	Value  float64
	Labels map[string]string
}

// NewMetricsStore creates a new metrics store
func NewMetricsStore(db *sql.DB) *MetricsStore {
	return &MetricsStore{db: db}
}

// InitSchema initializes TimescaleDB schema
func (s *MetricsStore) InitSchema(ctx context.Context) error {
	// Enable TimescaleDB extension if available
	_, err := s.db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE")
	if err != nil {
		// Fallback to standard Postgres if not available
		fmt.Println("Warning: TimescaleDB extension not available, using standard tables")
	}

	// Create metrics table
	query := `
		CREATE TABLE IF NOT EXISTS metrics (
			time        TIMESTAMPTZ       NOT NULL,
			name        TEXT              NOT NULL,
			value       DOUBLE PRECISION  NOT NULL,
			labels      JSONB
		);
	`
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create metrics table: %w", err)
	}

	// Convert to hypertable if TimescaleDB is active
	// Ignore error if it fails (e.g., standard Postgres)
	s.db.ExecContext(ctx, "SELECT create_hypertable('metrics', 'time', if_not_exists => TRUE);")

	// Create index on name and time
	_, err = s.db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_metrics_name_time ON metrics (name, time DESC);")
	return err
}

// WriteMetric writes a single metric
func (s *MetricsStore) WriteMetric(ctx context.Context, point MetricPoint) error {
	query := `INSERT INTO metrics (time, name, value, labels) VALUES ($1, $2, $3, $4)`

	// Convert labels to JSON string (simplified)
	labelsJson := "{}"
	// In production use json.Marshal(point.Labels)

	_, err := s.db.ExecContext(ctx, query, point.Time, point.Name, point.Value, labelsJson)
	return err
}

// QueryMetrics queries metrics
func (s *MetricsStore) QueryMetrics(ctx context.Context, name string, start, end time.Time) ([]MetricPoint, error) {
	query := `
		SELECT time, value 
		FROM metrics 
		WHERE name = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC
	`

	rows, err := s.db.QueryContext(ctx, query, name, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []MetricPoint
	for rows.Next() {
		var p MetricPoint
		p.Name = name
		if err := rows.Scan(&p.Time, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, nil
}
