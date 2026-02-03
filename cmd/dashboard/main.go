// Copyright (c) 2026 Project Atlas (Talos)
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Xover-Official/Xover/internal/analytics"
	"github.com/Xover-Official/Xover/internal/auth"
	"github.com/Xover-Official/Xover/internal/cloud"
	"github.com/Xover-Official/Xover/internal/cloud/aws"
	"github.com/Xover-Official/Xover/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type resourceCache struct {
	sync.RWMutex
	resources []*cloud.ResourceV2
	fetchedAt time.Time
	refreshMu sync.Mutex
}

type metricsCache struct {
	sync.RWMutex
	metrics   *ResourceMetricsResponse
	fetchedAt time.Time
}

type suggestionsCache struct {
	sync.RWMutex
	suggestions *OptimizationSuggestionsResponse
	fetchedAt   time.Time
}

// server represents the dependency container for the application
type server struct {
	tracker      *analytics.TokenTracker
	orchestrator AIOrchestrator // Use interface for decoupling
	adapter      cloud.CloudAdapter
	redisClient  *redis.Client
	logger       *zap.Logger
	config       *config.Config
	jwtManager   *auth.JWTManager
	userStore    UserStore // Use interface for decoupling
	mode             string
	resourceCache    resourceCache
	metricsCache     metricsCache
	suggestionsCache suggestionsCache
}

func main() {
	// 1. Setup zap logging
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := zapCfg.Build()
	if err != nil {
		// Can't use logger here, so panic
		panic("failed to initialize zap logger: " + err.Error())
	}
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// In a real enterprise application, the actual AI orchestrator would be initialized here.
	// For this refactoring, we use a mock that implements the AIOrchestrator interface.
	// The real `ai.UnifiedOrchestrator` would need to implement this interface.
	orchestrator := NewMockOrchestrator()

	jwtMgr := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration)

	// Initialize the user store. In production, this would be a database-backed store.
	userStore := NewInMemoryUserStore()

	srv := &server{
		tracker:      tracker,
		orchestrator: orchestrator,
		adapter:      adapter, // Use the cloud.CloudAdapter interface
		userStore:    userStore,
		redisClient:  rdb,
		logger:       logger,
		config:       cfg,
		jwtManager:   jwtMgr,
	}

	if *runLoadTest {
		runSimulation(srv)
		return
	}

	// Start background tasks
	go srv.startResourceCacheRefresh(ctx)

	// 5. Router Setup
	httpServer := &http.Server{
		Addr:        ":" + cfg.Server.Port,
		Handler:     srv.routes(),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
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

	// Cancel the base context to abort all active handlers immediately
	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)
	logger.Info("server stopped")
}

func runSimulation(s *server) {
	s.logger.Info("simulation mode active")
}
