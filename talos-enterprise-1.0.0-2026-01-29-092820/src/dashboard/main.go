package main

	"github.com/project-atlas/atlas/internal/auth"
	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/cloud/aws"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/redis/go-redis/v9"
)

type server struct {
	tracker       *analytics.TokenTracker
	orchestrator  *ai.UnifiedOrchestrator
	adapter       cloud.CloudAdapter
	redisClient   *redis.Client
	logger        *slog.Logger
	config        *config.Config
	jwtManager    *auth.JWTManager
	// In-memory cache for cloud resources
	resourceCache struct {
		sync.RWMutex
		resources []*cloud.ResourceV2
		fetchedAt time.Time
	}
}

func main() {
	// --- Production-Grade Setup ---
	// 1. Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("configuration loaded", "mode", cfg.Server.Mode, "dry_run", cfg.Cloud.DryRun)

	// 3. Parse command-line flags
	runLoadTest := flag.Bool("run-load-test", false, "Set to true to run the load test simulation.")
	flag.Parse()

	// 4. Initialize dependencies
	ctx := context.Background()
	awsAdapter, err := aws.New(ctx, cloud.CloudConfig(cfg.Cloud))
	if err != nil {
		logger.Error("could not create AWS adapter", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})
	defer rdb.Close()

	tracker := analytics.NewTokenTracker(cfg.Analytics.PersistPath)
	aiCfg := ai.Config{
		OpenRouterKey: cfg.AI.OpenRouterKey,
		DevinAPIKey:   cfg.AI.DevinKey,
		CacheEnabled:  cfg.AI.CacheEnabled,
		CacheAddr:     cfg.Redis.Address,
	}
	orchestrator, err := ai.NewUnifiedOrchestrator(&aiCfg, tracker, logger)
	if err != nil {
		logger.Error("could not create AI orchestrator", "error", err)
		os.Exit(1)
	}

	jwtMgr := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration)

	srv := &server{
		tracker:      tracker,
		orchestrator: orchestrator,
		adapter:      awsAdapter,
		redisClient:  rdb,
		logger:       logger,
		config:       cfg,
		jwtManager:   jwtMgr,
	}

	if *runLoadTest {
		runSimulation(srv)
		return
	}

	// 5. Start the HTTP server
	mux := http.NewServeMux()

	// Static file serving for UI
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	// Auth routes
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

	// Main router with auth middleware
	router := http.NewServeMux()
	router.Handle("/", mux) // Let the root handler manage login redirection
	router.Handle("/api/", srv.authMiddleware(api))


	httpServer := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: srv.authMiddleware(router),
	}

	go func() {
		logger.Info("starting dashboard API server", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("could not start server", "error", err)
			os.Exit(1)
		}
	}()

	// 6. Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	// 7. Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
	logger.Info("server exited gracefully")
}

// runSimulation is the load test logic, separated from main server logic.
func runSimulation(s *server) {
	s.logger.Info("starting load test simulation")
	// ... load test logic from previous version would go here ...
	s.logger.Info("load test simulation complete")
}
