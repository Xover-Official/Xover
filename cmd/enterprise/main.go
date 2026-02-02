package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/project-atlas/atlas/internal/manager"
	"github.com/project-atlas/atlas/internal/persistence"
	"github.com/project-atlas/atlas/internal/worker"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: talos <command> [options]")
	}

	command := os.Args[1]

	switch command {
	case "manager":
		runManager()
	case "worker":
		runWorker()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runManager() {
	log.Println("🚀 Starting Talos Enterprise Manager")

	// Load configuration
	cfg, err := config.Load("config.enterprise.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Initialize PostgreSQL ledger
	log.Println("📊 Connecting to PostgreSQL...")
	ledger, err := persistence.NewPostgresLedger(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("❌ PostgreSQL connection failed: %v", err)
	}
	defer ledger.Close()

	// Initialize token tracker
	log.Println("💰 Initializing token tracker...")
	tokenTracker := analytics.NewTokenTracker("./data/tokens.json")

	// Initialize AI orchestrator
	log.Println("🤖 Initializing AI orchestrator...")
	aiCfg := &ai.Config{
		GeminiAPIKey: cfg.AI.OpenRouterKey,
		DevinAPIKey:  cfg.AI.DevinKey,
		CacheEnabled: true,
		CacheAddr:    "redis:6379",
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(aiCfg, tokenTracker, zap.NewExample())
	if err != nil {
		log.Fatalf("❌ AI orchestrator initialization failed: %v", err)
	}
	defer orchestrator.Close()

	// Create enterprise manager
	mgr, err := manager.NewEnterpriseManager(cfg, ledger, orchestrator, tokenTracker)
	if err != nil {
		log.Fatalf("❌ Failed to create manager: %v", err)
	}

	// Start manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mgr.Start(ctx); err != nil {
		log.Fatalf("❌ Manager failed: %v", err)
	}

	log.Println("✅ Manager shutdown complete")
}

func runWorker() {
	log.Println("🚀 Starting Talos Enterprise Worker")

	// Get worker ID from environment or generate one
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		workerID = fmt.Sprintf("worker-%d", os.Getpid())
	}

	// Load configuration
	cfg, err := config.Load("config.enterprise.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Initialize PostgreSQL ledger
	log.Println("📊 Connecting to PostgreSQL...")
	ledger, err := persistence.NewPostgresLedger(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("❌ PostgreSQL connection failed: %v", err)
	}
	defer ledger.Close()

	// Initialize token tracker
	log.Println("💰 Initializing token tracker...")
	tokenTracker := analytics.NewTokenTracker("./data/tokens.json")

	// Initialize AI orchestrator
	log.Println("🤖 Initializing AI orchestrator...")
	aiCfg := &ai.Config{
		GeminiAPIKey: cfg.AI.OpenRouterKey,
		DevinAPIKey:  cfg.AI.DevinKey,
		CacheEnabled: true,
		CacheAddr:    "redis:6379",
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(aiCfg, tokenTracker, zap.NewExample())
	if err != nil {
		log.Fatalf("❌ AI orchestrator initialization failed: %v", err)
	}
	defer orchestrator.Close()

	// Create distributed worker
	w, err := worker.NewDistributedWorker(workerID, cfg, ledger, orchestrator, tokenTracker)
	if err != nil {
		log.Fatalf("❌ Failed to create worker: %v", err)
	}

	// Start worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		log.Fatalf("❌ Worker failed: %v", err)
	}

	log.Println("✅ Worker shutdown complete")
}
