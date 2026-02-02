// Copyright (c) 2026 Project Atlas (Talos)
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/auth"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/cloud/aws"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// server represents the dependency container for the application
type server struct {
	tracker       *analytics.TokenTracker
	orchestrator  *ai.UnifiedOrchestrator
	adapter       cloud.CloudAdapter
	redisClient   *redis.Client
	logger        *zap.Logger
	config        *config.Config
	jwtManager    *auth.JWTManager
	mode          string
	resourceCache struct {
		sync.RWMutex
		resources    []*cloud.ResourceV2
		fetchedAt    time.Time
		isRefreshing bool
		refreshMu    sync.Mutex
	}
}

func main() {
	// 1. Setup zap logging
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := zapCfg.Build()
	defer logger.Sync()

	// 2. Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("failed to load configuration", zap.Error(err))
		os.Exit(1)
	}

	// 3. Parse command-line flags
	runLoadTest := flag.Bool("run-load-test", false, "Run load test simulation")
	flag.Parse()

	// 4. Initialize dependencies
	ctx := context.Background()

	cloudCfg := cloud.CloudConfig{
		Region: cfg.Cloud.Region,
		DryRun: cfg.Cloud.DryRun,
	}

	awsAdapter, err := aws.New(ctx, cloudCfg)
	if err != nil {
		logger.Error("could not create AWS adapter", zap.Error(err))
		os.Exit(1)
	}

	// Type assert to cloud.CloudAdapter
	var adapter cloud.CloudAdapter = awsAdapter

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})
	defer rdb.Close()

	tracker := analytics.NewTokenTracker(cfg.Analytics.PersistPath)

	aiConfig := &ai.Config{
		GeminiAPIKey: cfg.AI.GeminiAPIKey,
		ClaudeAPIKey: cfg.AI.ClaudeAPIKey,
		GPT5APIKey:   cfg.AI.GPT5MiniAPIKey, // Mapped to GPT5Mini based on available config
		DevinAPIKey:  cfg.AI.DevinKey,
		CacheEnabled: true,
		CacheAddr:    cfg.Redis.Address,
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(aiConfig, tracker, logger)
	if err != nil {
		logger.Error("could not create AI orchestrator", zap.Error(err))
		os.Exit(1)
	}

	jwtMgr := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration)

	srv := &server{
		tracker:      tracker,
		orchestrator: orchestrator,
		adapter:      adapter, // Use the cloud.CloudAdapter interface
		redisClient:  rdb,
		logger:       logger,
		config:       cfg,
		jwtManager:   jwtMgr,
	}

	if *runLoadTest {
		runSimulation(srv)
		return
	}

	// 5. Router Setup
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	// Auth routes (Implemented in auth_handlers.go)
	mux.HandleFunc("/auth/login/", srv.handleLogin)
	mux.HandleFunc("/auth/callback/", srv.handleCallback)
	mux.HandleFunc("/auth/logout", srv.handleLogout)

	// API routes
	api := http.NewServeMux()
	api.HandleFunc("/roi", srv.handleROI)
	api.HandleFunc("/token-breakdown", srv.handleTokenBreakdown)
	api.HandleFunc("/system/status", srv.handleSystemStatus)
	api.HandleFunc("/resources", srv.handleResources)
	api.HandleFunc("/healthz", srv.handleHealthz)

	// New handlers from token_handlers.go
	api.HandleFunc("/token-stats", srv.handleTokenStats)
	api.HandleFunc("/resource-metrics", srv.handleResourceMetrics)
	api.HandleFunc("/optimization-suggestions", srv.handleOptimizationSuggestions)

	// Final routing with middleware
	mainRouter := http.NewServeMux()
	// Public probes
	mainRouter.HandleFunc("/healthz", srv.handleHealthz)
	mainRouter.Handle("/", mux)
	mainRouter.Handle("/api/", srv.authMiddleware(api))

	httpServer := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: mainRouter,
	}

	// 6. Start Server
	go func() {
		logger.Info("starting dashboard", zap.String("port", cfg.Server.Port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", zap.Error(err))
		}
	}()

	// 7. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)
	logger.Info("server stopped")
}

func runSimulation(s *server) {
	s.logger.Info("simulation mode active")
}

// --- API Handlers ---
// Note: handleLogin, handleCallback, handleLogout, and authMiddleware
// are removed from here because they are defined in auth_handlers.go

func (s *server) handleROI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roi": map[string]interface{}{
			"total_savings":  1250.50,
			"total_costs":    342.75,
			"net_roi":        907.75,
			"roi_percentage": 264.8,
		},
		"period": "30 days",
	})
}
func (s *server) handleTokenBreakdown(w http.ResponseWriter, r *http.Request) {
	stats := s.tracker.GetStats()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_cost_usd":    stats["total_cost_usd"],
		"total_tokens":      stats["total_tokens"],
		"total_savings_usd": stats["total_savings_usd"],
		"net_profit_usd":    stats["net_profit_usd"],
		"breakdown": map[string]interface{}{
			"sentinel":   map[string]interface{}{"tokens": 1500, "cost": 0.75},
			"strategist": map[string]interface{}{"tokens": 800, "cost": 1.20},
			"arbiter":    map[string]interface{}{"tokens": 400, "cost": 0.80},
			"reasoning":  map[string]interface{}{"tokens": 200, "cost": 0.60},
		},
	})
}
func (s *server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
		"uptime":  "2h 34m",
		"services": map[string]interface{}{
			"ai_orchestrator": "online",
			"cloud_adapter":   "online",
			"redis":           "online",
			"database":        "online",
		},
		"metrics": map[string]interface{}{
			"active_optimizations": 3,
			"resources_monitored":  127,
			"cost_savings_today":   45.75,
		},
	})
}
func (s *server) handleResources(w http.ResponseWriter, r *http.Request) {
	// Fetch fresh resources if cache is stale
	s.resourceCache.RLock()
	if time.Since(s.resourceCache.fetchedAt) > 5*time.Minute {
		s.resourceCache.RUnlock()
		s.refreshResourceCache()
	} else {
		s.resourceCache.RUnlock()
	}

	s.resourceCache.RLock()
	defer s.resourceCache.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resources":    s.resourceCache.resources,
		"total_count":  len(s.resourceCache.resources),
		"last_updated": s.resourceCache.fetchedAt,
	})
}
func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}
