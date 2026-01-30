package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <command>")
		fmt.Println("Commands:")
		fmt.Println("  up       - Run all pending migrations")
		fmt.Println("  down     - Rollback last migration")
		fmt.Println("  status   - Show migration status")
		os.Exit(1)
	}

	// Read database config from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://talos_user:your_secure_password@localhost:5432/talos?sslmode=disable"
		fmt.Println("⚠️  Using default connection string. Set DATABASE_URL env var for production.")
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to PostgreSQL")

	command := os.Args[1]
	switch command {
	case "up":
		runMigrations(pool)
	case "status":
		showStatus(pool)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func runMigrations(pool *pgxpool.Pool) {
	ctx := context.Background()

	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	fmt.Println("🚀 Running migration: 001_initial_schema.sql")

	// Execute migration
	_, err = pool.Exec(ctx, string(migrationSQL))
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("\n📊 Database schema created:")
	fmt.Println("  - actions (idempotent ledger)")
	fmt.Println("  - ai_decisions (AI audit trail)")
	fmt.Println("  - token_usage (cost tracking)")
	fmt.Println("  - savings_events (ROI tracking)")
	fmt.Println("  - organizations (multi-tenancy)")
	fmt.Println("  - users (RBAC)")
	fmt.Println("  - resources (cloud inventory)")
	fmt.Println("  - audit_log (compliance)")
}

func showStatus(pool *pgxpool.Pool) {
	ctx := context.Background()

	// Check if tables exist
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		ORDER BY table_name
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	fmt.Println("📋 Database Status:")
	fmt.Println("\nExisting tables:")

	count := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("  ✓ %s\n", tableName)
		count++
	}

	if count == 0 {
		fmt.Println("  (No tables found - run 'migrate up' to create schema)")
	} else {
		fmt.Printf("\nTotal: %d tables\n", count)
	}
}
