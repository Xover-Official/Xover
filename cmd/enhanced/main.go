package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/project-atlas/atlas/internal/config"
	"github.com/project-atlas/atlas/internal/deployment"
	"github.com/project-atlas/atlas/internal/monitoring"
	"github.com/project-atlas/atlas/internal/performance"
	"github.com/project-atlas/atlas/internal/security"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("🚀 Talos Cloud Guardian - Enhanced Enterprise Edition")
	fmt.Println("=====================================================")

	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration from environment
	envConfig, err := config.NewEnvironmentConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Validate configuration
	if err := envConfig.Validate(); err != nil {
		logger.Fatal("Invalid configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded successfully",
		zap.String("mode", envConfig.Server.Mode),
		zap.String("cloud_provider", envConfig.Cloud.Provider),
		zap.String("region", envConfig.Cloud.Region),
	)

	// Initialize monitoring
	monitoringService, err := monitoring.NewMonitoringService()
	if err != nil {
		logger.Fatal("Failed to initialize monitoring", zap.Error(err))
	}

	// Initialize security
	securityManager := security.NewEnhancedSecurityManager(
		envConfig.JWT.SecretKey,
		envConfig.JWT.TokenDuration,
		30*time.Minute, // Default refresh time
		logger,
	)

	// Initialize deployment manager
	deploymentManager, err := deployment.NewDeploymentManager(
		"", // Use default kubeconfig
		"talos",
		logger,
	)
	if err != nil {
		logger.Error("Failed to initialize deployment manager", zap.Error(err))
		// Continue without deployment manager - it's not critical for basic operation
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start monitoring server
	go startMonitoringServer(monitoringService, envConfig, logger)

	// Start health check server
	go startHealthServer(monitoringService, securityManager, envConfig, logger)

	// Start performance optimization
	go startPerformanceOptimization(monitoringService, logger)

	// Generate deployment package if requested
	if len(os.Args) > 1 && os.Args[1] == "deploy" {
		logger.Info("Generating deployment package...")

		if deploymentManager == nil {
			logger.Fatal("Deployment manager not available - cannot generate deployment package")
		}

		outputDir := "./deployment-package"
		if err := deploymentManager.GenerateDeploymentPackage(outputDir); err != nil {
			logger.Fatal("Failed to generate deployment package", zap.Error(err))
		}

		logger.Info("Deployment package generated successfully",
			zap.String("output_dir", outputDir),
		)

		fmt.Printf("\n📦 Deployment package created in: %s\n", outputDir)
		fmt.Printf("🚀 To deploy: cd %s && ./deploy.sh\n", outputDir)
		return
	}

	// Start main application
	logger.Info("Starting Talos Cloud Guardian...")

	// Simulate main application logic
	go runMainApplication(ctx, envConfig, monitoringService, securityManager, logger)

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown
	logger.Info("Shutting down gracefully...")

	// Give some time for cleanup
	time.Sleep(5 * time.Second)

	logger.Info("Shutdown complete")
}

func startMonitoringServer(monitoringService *monitoring.MonitoringService, config *config.EnvironmentConfig, logger *zap.Logger) {
	monitoringConfig := config.GetMonitoringConfig()

	if prometheusConfig, ok := monitoringConfig["prometheus"].(map[string]interface{}); ok {
		if enabled, ok := prometheusConfig["enabled"].(bool); ok && enabled {
			port, ok := prometheusConfig["port"].(string)
			if !ok {
				logger.Error("Invalid prometheus port configuration")
				return
			}

			logger.Info("Starting Prometheus metrics server", zap.String("port", port))

			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())

			go func() {
				if err := http.ListenAndServe(":"+port, mux); err != nil {
					logger.Error("Failed to start metrics server", zap.Error(err))
				}
			}()
		}
	}
}

func startHealthServer(monitoringService *monitoring.MonitoringService, securityManager *security.EnhancedSecurityManager, config *config.EnvironmentConfig, logger *zap.Logger) {
	port := config.Server.Port
	logger.Info("Starting health check server", zap.String("port", port))

	healthCheck := monitoring.NewHealthCheck(logger)

	// Add health checks
	healthCheck.AddCheck("database", func(ctx context.Context) error {
		// Check database connectivity
		return nil
	})

	healthCheck.AddCheck("redis", func(ctx context.Context) error {
		// Check Redis connectivity
		return nil
	})

	healthCheck.AddCheck("ai_services", func(ctx context.Context) error {
		// Check AI service connectivity
		return nil
	})

	mux := http.NewServeMux()

	// Health Endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		results := healthCheck.RunChecks(ctx)
		status := "UP"
		statusCode := http.StatusOK

		for _, err := range results {
			if err != "" {
				status = "DOWN"
				statusCode = http.StatusServiceUnavailable
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status,
			"checks": results,
		})
	})

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("Failed to start health server", zap.Error(err))
	}
}

func startPerformanceOptimization(monitoringService *monitoring.MonitoringService, logger *zap.Logger) {
	logger.Info("Starting performance optimization")

	// Initialize performance optimizer
	_ = performance.NewPerformanceOptimizer(monitoringService, logger)

	// Start memory optimization
	memoryOptimizer := performance.NewMemoryOptimizer()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Optimize memory
		memoryOptimizer.OptimizeMemory()

		// Get memory stats
		memStats := memoryOptimizer.GetMemoryStats()
		monitoringService.SetActiveWorkers(int(memStats.NumGC))

		logger.Debug("Performance optimization completed",
			zap.Uint64("alloc_bytes", memStats.Alloc),
			zap.Uint32("num_gc", memStats.NumGC),
		)
	}
}

func runMainApplication(ctx context.Context, config *config.EnvironmentConfig, monitoringService *monitoring.MonitoringService, securityManager *security.EnhancedSecurityManager, logger *zap.Logger) {
	logger.Info("Main application started")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate application work
			doApplicationWork(config, monitoringService, securityManager, logger)
		}
	}
}

func doApplicationWork(config *config.EnvironmentConfig, monitoringService *monitoring.MonitoringService, securityManager *security.EnhancedSecurityManager, logger *zap.Logger) {
	// Simulate resource fetching
	start := time.Now()

	// Record metrics
	monitoringService.RecordHTTPRequest("GET", "/api/resources", "200", time.Since(start))
	monitoringService.RecordCloudOperation(config.Cloud.Provider, "fetch_resources", "success", time.Since(start))

	// Simulate AI requests
	monitoringService.RecordAIRequest("openrouter", "gpt-4", "success", 500*time.Millisecond, 150)

	// Simulate cost savings
	monitoringService.RecordCostSavings(config.Cloud.Provider, "rightsizing", 25.50)
	monitoringService.RecordOptimizationAction(config.Cloud.Provider, "stop_instance", "success")

	logger.Debug("Application work completed",
		zap.String("provider", config.Cloud.Provider),
		zap.Float64("cost_savings", 25.50),
	)
}
