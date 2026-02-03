package main

import (
	"fmt"
	"log"

	"github.com/Xover-Official/Xover/internal/idempotency"
	"github.com/Xover-Official/Xover/internal/logger"
)

func main() {
	// Setup
	ledger, err := idempotency.NewLedger("atlas_ledger.db")
	if err != nil {
		log.Fatalf("Failed to create ledger: %v", err)
	}
	engine := idempotency.NewEngine(ledger)

	// Mock action payload
	payload := map[string]string{
		"action":   "resize_rds",
		"instance": "db-prod-01",
		"new_type": "db.m5.large",
	}

	// Action function
	mockAction := func() (string, error) {
		fmt.Println("Executing real cloud action...")
		return "res-12345", nil
	}

	fmt.Println("--- FIRST ATTEMPT ---")
	res1, err := engine.ExecuteGuarded(logger.Builder, "ResizeRDS", payload, mockAction)
	if err != nil {
		log.Fatalf("First attempt failed: %v", err)
	}
	fmt.Printf("Result 1: %s\n\n", res1)

	fmt.Println("--- SECOND ATTEMPT (Should be skipped) ---")
	res2, err := engine.ExecuteGuarded(logger.Builder, "ResizeRDS", payload, mockAction)
	if err != nil {
		log.Fatalf("Second attempt failed: %v", err)
	}
	fmt.Printf("Result 2: %s (Successful skip if no 'Executing real cloud action' print)\n", res2)

	// Cleanup for demo
	// os.Remove("atlas_ledger.db")
	// os.Remove("SESSION_LOG.json")
}
