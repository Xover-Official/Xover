package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	Database          string        `yaml:"database"`
	Username          string        `yaml:"username"`
	Password          string        `yaml:"password"`
	SSLMode           string        `yaml:"ssl_mode"`
	MaxConnections    int           `yaml:"max_connections"`
	MinConnections    int           `yaml:"min_connections"`
	MaxIdleTime       time.Duration `yaml:"max_idle_time"`
	MaxLifetime       time.Duration `yaml:"max_lifetime"`
	ConnectTimeout    time.Duration `yaml:"connect_timeout"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period"`
}

// DatabaseManager manages database connections and operations
type DatabaseManager struct {
	pool   *pgxpool.Pool
	config DatabaseConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// NewDatabaseManager creates a new database manager with zap logging
func NewDatabaseManager(config DatabaseConfig, logger *zap.Logger, tracer trace.Tracer) (*DatabaseManager, error) {
	// Use url.URL to safely encode parameters, especially the password
	u := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", config.Host, config.Port),
		Path:   config.Database,
	}
	q := u.Query()
	q.Set("sslmode", config.SSLMode)
	u.User = url.UserPassword(config.Username, config.Password)
	u.RawQuery = q.Encode()

	poolConfig, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(config.MaxConnections)
	poolConfig.MinConns = int32(config.MinConnections)
	poolConfig.MaxConnIdleTime = config.MaxIdleTime
	poolConfig.MaxConnLifetime = config.MaxLifetime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Configure connection before use
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		// Health check before acquiring connection
		err := conn.Ping(ctx)
		if err != nil {
			logger.Warn("Database connection health check failed", zap.Error(err))
			return false
		}
		return true
	}

	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Set connection parameters
		_, err := conn.Exec(ctx, "SET application_name = 'talos-atlas'")
		if err != nil {
			logger.Warn("Failed to set application name", zap.Error(err))
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection pool established",
		zap.String("host", config.Host),
		zap.String("database", config.Database),
		zap.Int("max_connections", config.MaxConnections),
		zap.Int("min_connections", config.MinConnections),
	)

	return &DatabaseManager{
		pool:   pool,
		config: config,
		logger: logger,
		tracer: tracer,
	}, nil
}

// GetPool returns the underlying connection pool
func (dm *DatabaseManager) GetPool() *pgxpool.Pool {
	return dm.pool
}

// GetSQLDB returns a *sql.DB for compatibility with libraries that need it
func (dm *DatabaseManager) GetSQLDB() (*sql.DB, error) {
	return stdlib.OpenDBFromPool(dm.pool), nil
}

// HealthCheck performs a health check on the database
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
	ctx, span := dm.tracer.Start(ctx, "database.health_check")
	defer span.End()

	return dm.pool.Ping(ctx)
}

// Close closes the database connection pool
func (dm *DatabaseManager) Close() {
	if dm.pool != nil {
		dm.pool.Close()
		dm.logger.Info("Database connection pool closed")
	}
}

// Stats returns connection pool statistics
func (dm *DatabaseManager) Stats() *pgxpool.Stat {
	return dm.pool.Stat()
}

// Query executes a query that returns rows
func (dm *DatabaseManager) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	ctx, span := dm.tracer.Start(ctx, "database.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.query", query),
		attribute.Int("db.args_count", len(args)),
	)

	rows, err := dm.pool.Query(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return rows, nil
}

// QueryRow executes a query that returns a single row
func (dm *DatabaseManager) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	ctx, span := dm.tracer.Start(ctx, "database.query_row")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.query", query),
		attribute.Int("db.args_count", len(args)),
	)

	return dm.pool.QueryRow(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (dm *DatabaseManager) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	ctx, span := dm.tracer.Start(ctx, "database.exec")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.query", query),
		attribute.Int("db.args_count", len(args)),
	)

	result, err := dm.pool.Exec(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return pgconn.CommandTag{}, fmt.Errorf("exec failed: %w", err)
	}

	span.SetAttributes(
		attribute.Int64("db.rows_affected", result.RowsAffected()),
	)

	return result, nil
}

// Transaction executes a function within a database transaction
func (dm *DatabaseManager) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	ctx, span := dm.tracer.Start(ctx, "database.transaction")
	defer span.End()

	tx, err := dm.pool.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			dm.logger.Error("Failed to rollback transaction", zap.Error(rbErr))
		}
		span.RecordError(err)
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DefaultDatabaseConfig returns a default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		Database:          "talos",
		Username:          "talos_user",
		Password:          "talos_password",
		SSLMode:           "disable",
		MaxConnections:    20,
		MinConnections:    5,
		MaxIdleTime:       30 * time.Minute,
		MaxLifetime:       2 * time.Hour,
		ConnectTimeout:    10 * time.Second,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

// ProductionDatabaseConfig returns a production database configuration
func ProductionDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		Database:          "talos_prod",
		Username:          "talos_prod_user",
		Password:          "", // Should come from environment/secrets
		SSLMode:           "require",
		MaxConnections:    50,
		MinConnections:    10,
		MaxIdleTime:       15 * time.Minute,
		MaxLifetime:       1 * time.Hour,
		ConnectTimeout:    5 * time.Second,
		HealthCheckPeriod: 30 * time.Second,
	}
}
