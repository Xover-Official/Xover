package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/project-atlas/atlas/internal/loop"
	"github.com/project-atlas/atlas/internal/persistence"
)

func main() {
	printBanner()

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Validate API keys
	if cfg.AI.OpenRouterKey == "" {
		log.Fatal("❌ OPENROUTER_API_KEY not set! See docs/API_KEYS_SETUP.md")
	}

	// Initialize persistence layer
	var ledger persistence.Ledger
	if cfg.Database.Type == "postgres" {
		log.Println("📊 Connecting to PostgreSQL...")
		ledger, err = persistence.NewPostgresLedger(cfg.Database.DSN())
		if err != nil {
			log.Fatalf("❌ PostgreSQL connection failed: %v", err)
		}
	} else {
		log.Println("📊 Using SQLite (development mode)...")
		os.MkdirAll(cfg.Storage.DataPath, 0755)
		ledger, err = persistence.NewSQLiteLedger(cfg.Storage.DataPath + "/talos.db")
		if err != nil {
			log.Fatalf("❌ SQLite initialization failed: %v", err)
		}
	}
	defer ledger.Close()

	// Initialize token tracker
	log.Println("💰 Initializing token tracker...")
	tokenTracker := analytics.NewTokenTracker(cfg.Storage.DataPath + "/tokens.json")

	// Initialize AI Orchestrator (OpenRouter + Devin)
	log.Println("🤖 Initializing AI Swarm...")
	log.Println("   • Tier 1 (Sentinel): Gemini Flash via OpenRouter")
	log.Println("   • Tier 2 (Strategist): Gemini Pro via OpenRouter")
	log.Println("   • Tier 3 (Arbiter): Claude 3.5 via OpenRouter")
	log.Println("   • Tier 4 (Reasoning): GPT-4o Mini via OpenRouter")

	if cfg.AI.DevinKey != "" {
		log.Println("   • Tier 5 (Oracle): Devin (direct API)")
	} else {
		log.Println("   • Tier 5 (Oracle): ⚠️  Disabled (no DEVIN_API_KEY)")
	}

	// Determine cache address
	cacheAddr := ""
	if cfg.Database.Type == "postgres" {
		cacheAddr = "redis:6379" // Enable cache in production
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(
		cfg.AI.OpenRouterKey,
		cfg.AI.DevinKey,
		cacheAddr,
		tokenTracker,
		log.Default(),
	)
	if err != nil {
		log.Fatalf("❌ AI orchestrator initialization failed: %v", err)
	}
	defer orchestrator.Close()

	// Health check all AI tiers
	log.Println("\n🏥 Running AI health checks...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	healthResults := orchestrator.HealthCheckAll(ctx)
	cancel()

	healthyCount := 0
	for tier, err := range healthResults {
		if err == nil {
			healthyCount++
			log.Printf("  ✅ %s: OPERATIONAL", tier)
		} else {
			log.Printf("  ⚠️  %s: %v", tier, err)
		}
	}

	if healthyCount == 0 {
		log.Fatal("\n❌ No AI tiers operational. Check your API keys!")
	}

	log.Printf("\n✅ %d AI tiers ready\n", healthyCount)

	// Initialize OODA loop
	log.Println("🔄 Starting OODA loop...")
	oodaLoop := loop.NewOODALoop(cfg, ledger, orchestrator, tokenTracker)

	// Start loop in background
	go func() {
		if err := oodaLoop.Start(); err != nil {
			log.Fatalf("❌ OODA loop failed: %v", err)
		}
	}()

	// Print startup summary
	printStartupSummary(cfg, healthyCount)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\n🛑 Shutting down gracefully...")

	oodaLoop.Stop()

	// Print final stats
	stats := tokenTracker.GetStats()
	log.Println("\n" + strings.Repeat("═", 60))
	log.Println("📊 FINAL SESSION STATS")
	log.Println(strings.Repeat("═", 60))
	log.Printf("  AI Cost:        $%.2f", stats["total_cost_usd"])
	log.Printf("  Cloud Savings:  $%.2f", stats["total_savings_usd"])
	log.Printf("  Net Profit:     $%.2f", stats["net_profit_usd"])
	log.Printf("  ROI:           %.1f%%", stats["net_roi"])
	log.Println(strings.Repeat("═", 60))

	log.Println("\n👋 Talos shutdown complete. Stay optimized!")
}

func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║   ████████╗ █████╗ ██╗      ██████╗ ███████╗                 ║
║   ╚══██╔══╝██╔══██╗██║     ██╔═══██╗██╔════╝                 ║
║      ██║   ███████║██║     ██║   ██║███████╗                 ║
║      ██║   ██╔══██║██║     ██║   ██║╚════██║                 ║
║      ██║   ██║  ██║███████╗╚██████╔╝███████║                 ║
║      ╚═╝   ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚══════╝                 ║
║                                                                ║
║          THE AUTONOMOUS CLOUD GUARDIAN                         ║
║                   v1.0.0-beta                                  ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printStartupSummary(cfg *config.Config, healthyTiers int) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("⚙️  CONFIGURATION")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("  Mode:           %s\n", cfg.Guardian.Mode)
	fmt.Printf("  Risk Threshold: %.1f/10\n", cfg.Guardian.RiskThreshold)
	fmt.Printf("  Dry Run:        %v\n", cfg.Guardian.DryRun)
	fmt.Printf("  AI Tiers:       %d active\n", healthyTiers)
	fmt.Printf("  Database:       %s\n", cfg.Database.Type)
	fmt.Printf("  Dashboard:      http://localhost:%s\n", cfg.Network.DashboardPort)
	fmt.Println(strings.Repeat("═", 60))

	if cfg.Guardian.DryRun {
		fmt.Println("\n⚠️  DRY RUN MODE ENABLED")
		fmt.Println("   AI will analyze but NOT apply changes")
		fmt.Println("   Set dry_run: false in config.yaml to enable automation")
	} else {
		fmt.Println("\n🚀 AUTONOMOUS MODE ACTIVE")
		fmt.Println("   AI will automatically apply safe optimizations")
		fmt.Println("   Monitor dashboard: http://localhost:8080")
	}

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📡 MONITORING")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("  • Dashboard:   http://localhost:8080")
	fmt.Println("  • Prometheus:  http://localhost:9090")
	fmt.Println("  • Grafana:     http://localhost:3000")
	fmt.Println(strings.Repeat("═", 60))

	fmt.Println("\n🎯 Press Ctrl+C to stop\n")
}
