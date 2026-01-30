package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"

	_ "github.com/lib/pq" // Postgres driver
)

// HAClient manages High Availability Postgres connections
type HAClient struct {
	Primary      *sql.DB
	ReadReplicas []*sql.DB
	mu           sync.RWMutex
}

// Config holds HA configuration
type HAConfig struct {
	PrimaryDSN   string
	ReplicaDSNs  []string
	MaxOpenConns int
	MaxIdleConns int
}

// NewHAClient creates a new HA client
func NewHAClient(config HAConfig) (*HAClient, error) {
	primary, err := openDB(config.PrimaryDSN, config)
	if err != nil {
		return nil, fmt.Errorf("failed to open primary: %w", err)
	}

	replicas := make([]*sql.DB, 0, len(config.ReplicaDSNs))
	for _, dsn := range config.ReplicaDSNs {
		replica, err := openDB(dsn, config)
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to open replica: %v\n", err)
			continue
		}
		replicas = append(replicas, replica)
	}

	return &HAClient{
		Primary:      primary,
		ReadReplicas: replicas,
	}, nil
}

func openDB(dsn string, config HAConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(1 * time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

// Master returns the primary writer connection
func (c *HAClient) Master() *sql.DB {
	return c.Primary
}

// Replica returns a random read replica, or the primary if no replicas are available
func (c *HAClient) Replica() *sql.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ReadReplicas) == 0 {
		return c.Primary
	}

	// Simple round-robin or random selection
	idx := rand.Intn(len(c.ReadReplicas))
	return c.ReadReplicas[idx]
}

// Exec executes a query on the primary
func (c *HAClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.Primary.ExecContext(ctx, query, args...)
}

// Query executes a query on a replica
func (c *HAClient) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.Replica().QueryContext(ctx, query, args...)
}

// QueryRow executes a query on a replica
func (c *HAClient) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.Replica().QueryRowContext(ctx, query, args...)
}

// Close closes all connections
func (c *HAClient) Close() error {
	var errs []error
	if err := c.Primary.Close(); err != nil {
		errs = append(errs, err)
	}
	for _, r := range c.ReadReplicas {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing dbs: %v", errs)
	}
	return nil
}
