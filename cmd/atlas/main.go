package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
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

	// 1. Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Load configuration from YAML and environment variables
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 3. Initialize persistence layer based on configuration
	var ledger persistence.Ledger
	if cfg.Server.Mode == "production" {
		logger.Info("📊 Connecting to Production Ledger (PostgreSQL)...")
		ledger, err = persistence.NewPostgresLedger(cfg.Database.DSN)
	} else {
		logger.Info("📊 Using development Ledger (SQLite)...")
		dataPath := "./data"
		os.MkdirAll(dataPath, 0755)
		ledger, err = persistence.NewSQLiteLedger(dataPath + "/talos.db")
	}
	if err != nil {
		logger.Error("Persistence initialization failed", "error", err)
		os.Exit(1)
	}
	defer ledger.Close()

	// 4. Initialize token tracker for monitoring AI costs
	tokenTracker := analytics.NewTokenTracker(cfg.Analytics.PersistPath)

	// 5. Initialize AI Orchestrator with different AI models
	aiCfg := &ai.Config{
		// The OpenRouterKey is used for all Gemini and Claude models via the OpenRouter API.
		GeminiAPIKey: cfg.AI.OpenRouterKey,
		ClaudeAPIKey: cfg.AI.OpenRouterKey,
		GPT5APIKey:   cfg.AI.OpenRouterKey,
		DevinAPIKey:  cfg.AI.DevinKey,
		CacheEnabled: cfg.AI.CacheEnabled,
		CacheAddr:    cfg.Redis.Address,
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(aiCfg, tokenTracker, logger)
	if err != nil {
		logger.Error("AI orchestrator initialization failed", "error", err)
		os.Exit(1)
	}
	defer orchestrator.Close()

	// 6. Health check logic for all registered AI tiers
	logger.Info("🏥 Running AI health checks...")
	healthResults := runHealthChecks(orchestrator.GetFactory())
	printStartupSummary(cfg, healthResults)

	// 7. Initialize and start the main OODA loop in a separate goroutine
	logger.Info("🔄 Starting OODA loop...")
	oodaLoop := loop.NewOODALoop(cfg, ledger, orchestrator, tokenTracker)

	go func() {
		if err := oodaLoop.Start(); err != nil {
			logger.Error("OODA loop failed", "error", err)
			os.Exit(1)
		}
	}()

	// 8. Graceful Shutdown on SIGINT or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("🛑 Shutting down gracefully...")

	oodaLoop.Stop()

	// Print final cost and savings statistics
	stats := tokenTracker.GetStats()
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📊 FINAL SESSION STATS")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("  AI Cost:         $%.4f\n", stats["total_cost_usd"])
	fmt.Printf("  Cloud Savings:   $%.2f\n", stats["total_savings_usd"])
	fmt.Printf("  Net Profit:      $%.2f\n", stats["net_profit_usd"])
	fmt.Println(strings.Repeat("═", 60))

	logger.Info("👋 Talos shutdown complete.")
}

// runHealthChecks performs parallel health checks on all available AI clients.
func runHealthChecks(factory *ai.AIClientFactory) map[string]bool {
	clients := map[string]ai.AIClient{
		"Sentinel (Gemini Flash)": factory.GetClientByName("sentinel"),
		"Strategist (Gemini Pro)": factory.GetClientByName("strategist"),
		"Arbiter (Claude)":        factory.GetClientByName("arbiter"),
		"Reasoning (GPT-5 Mini)":  factory.GetClientByName("reasoning"),
		"Oracle (Devin)":          factory.GetClientByName("oracle"),
	}

	var wg sync.WaitGroup
	results := make(map[string]bool)
	var mu sync.Mutex

	for name, client := range clients {
		wg.Add(1)
		go func(name string, client ai.AIClient) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			err := client.HealthCheck(ctx)
			mu.Lock()
			results[name] = err == nil
			mu.Unlock()
		}(name, client)
	}

	wg.Wait()
	return results
}

func printBanner() {
	banner := `
  ████████╗ █████╗ ██╗      ██████╗ ███████╗
  ╚══██╔══╝██╔══██╗██║     ██╔═══██╗██╔════╝
     ██║   ███████║██║     ██║   ██║███████╗
     ██║   ██╔══██║██║     ██║   ██║╚════██║
     ██║   ██║  ██║███████╗╚██████╔╝███████║
     ╚═╝   ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚══════╝
      THE AUTONOMOUS CLOUD GUARDIAN v1.0.0
`
	fmt.Println(banner)
}

func printStartupSummary(cfg *config.Config, health map[string]bool) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("⚙️  SYSTEM CONFIGURATION")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  Mode:            %s\n", cfg.Server.Mode)
	fmt.Printf("  Cloud Provider:  %s\n", cfg.Cloud.Provider)
	fmt.Printf("  Cloud Region:    %s\n", cfg.Cloud.Region)
	fmt.Printf("  Dry Run:         %v\n", cfg.Cloud.DryRun)
	fmt.Printf("  Dashboard Port:  %s\n", cfg.Server.Port)
	fmt.Println(strings.Repeat("═", 60))

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("🤖 AI TIER STATUS")
	fmt.Println(strings.Repeat("─", 60))

	// Sort keys for deterministic output
	keys := make([]string, 0, len(health))
	for k := range health {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		isHealthy := health[name]
		status := "❌ OFFLINE"
		if isHealthy {
			status = "✅ ONLINE"
		}
		fmt.Printf("  %-25s %s\n", name, status)
	}
	fmt.Println(strings.Repeat("═", 60))

	if cfg.Cloud.DryRun {
		fmt.Println("\n⚠️  DRY RUN MODE ENABLED - No infrastructure changes will be applied.")
	} else {
		fmt.Println("\n🚀 AUTONOMOUS MODE ACTIVE - Real-time infrastructure optimization is live.")
	}
	fmt.Println(strings.Repeat("═", 60) + "\n")
}
