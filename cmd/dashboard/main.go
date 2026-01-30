package main

import (
	"context"
	"flag"
	"log/slog"
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
)

// server represents the dependency container for the application
type server struct {
	tracker       *analytics.TokenTracker
	orchestrator  *ai.UnifiedOrchestrator
	adapter       cloud.CloudAdapter
	redisClient   *redis.Client
	logger        *slog.Logger
	config        *config.Config
	jwtManager    *auth.JWTManager
	mode          string
	resourceCache struct {
		sync.RWMutex
		resources []*cloud.ResourceV2
		fetchedAt time.Time
	}
}

func main() {
	// 1. Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
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
		logger.Error("could not create AWS adapter", "error", err)
		os.Exit(1)
	}

	// Type assert to cloud.CloudAdapter
	var adapter cloud.CloudAdapter = awsAdapter

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})
	defer rdb.Close()

	tracker := analytics.NewTokenTracker(cfg.Analytics.PersistPath)

	// FIX: Use the ai.Config struct with correct field names
	aiConfig := &ai.Config{
		GeminiAPIKey: cfg.AI.OpenRouterKey, // Map OpenRouter to Gemini
		DevinAPIKey:  cfg.AI.DevinKey,
		CacheEnabled: true,
		CacheAddr:    cfg.Redis.Address,
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(aiConfig, tracker, logger)
	if err != nil {
		logger.Error("could not create AI orchestrator", "error", err)
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

	// Final routing with middleware
	mainRouter := http.NewServeMux()
	mainRouter.Handle("/", mux)
	mainRouter.Handle("/api/", srv.authMiddleware(api))

	httpServer := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: mainRouter,
	}

	// 6. Start Server
	go func() {
		logger.Info("starting dashboard", "port", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
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

func (s *server) handleROI(w http.ResponseWriter, r *http.Request)            {}
func (s *server) handleTokenBreakdown(w http.ResponseWriter, r *http.Request) {}
func (s *server) handleSystemStatus(w http.ResponseWriter, r *http.Request)   {}
func (s *server) handleResources(w http.ResponseWriter, r *http.Request)      {}
func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request)        {}
