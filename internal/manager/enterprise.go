package manager

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

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/Xover-Official/Xover/internal/ai"
	"github.com/Xover-Official/Xover/internal/analytics"
	"github.com/Xover-Official/Xover/internal/config"
	"github.com/Xover-Official/Xover/internal/persistence"
)

// Task represents a distributed work item
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "scan", "analyze", "optimize"
	Priority    int                    `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
}

// EnterpriseManager manages the distributed Talos system
type EnterpriseManager struct {
	id           string
	redis        *redis.Client
	db           persistence.Ledger
	orchestrator *ai.UnifiedOrchestrator
	tokenTracker *analytics.TokenTracker
	config       *config.Config

	// HTTP server
	server *http.Server

	// Manager state
	isRunning    bool
	shutdownChan chan struct{}
}

// NewEnterpriseManager creates a new enterprise manager
func NewEnterpriseManager(cfg *config.Config, db persistence.Ledger, orchestrator *ai.UnifiedOrchestrator, tracker *analytics.TokenTracker) (*EnterpriseManager, error) {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	manager := &EnterpriseManager{
		id:           fmt.Sprintf("manager-%d", time.Now().Unix()),
		redis:        rdb,
		db:           db,
		orchestrator: orchestrator,
		tokenTracker: tracker,
		config:       cfg,
		shutdownChan: make(chan struct{}),
	}

	return manager, nil
}

// Start begins the enterprise manager
func (m *EnterpriseManager) Start(ctx context.Context) error {
	m.isRunning = true
	log.Printf("🚀 Starting Enterprise Manager: %s", m.id)

	// Start HTTP API server
	if err := m.startAPIServer(); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	// Start background tasks
	go m.taskScheduler(ctx)
	go m.workerMonitor(ctx)
	go m.metricsCollector(ctx)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("🛑 Received shutdown signal")
	case <-ctx.Done():
		log.Println("🛑 Context cancelled")
	case <-m.shutdownChan:
		log.Println("🛑 Shutdown requested")
	}

	return m.Shutdown()
}

// Shutdown gracefully stops the manager
func (m *EnterpriseManager) Shutdown() error {
	if !m.isRunning {
		return nil
	}

	log.Println("🛑 Shutting down Enterprise Manager...")
	m.isRunning = false
	close(m.shutdownChan)

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Error shutting down server: %v", err)
	}

	// Close Redis connection
	if err := m.redis.Close(); err != nil {
		log.Printf("⚠️  Error closing Redis: %v", err)
	}

	log.Println("✅ Enterprise Manager shutdown complete")
	return nil
}

// startAPIServer starts the HTTP API server
func (m *EnterpriseManager) startAPIServer() error {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Task management
	api.HandleFunc("/tasks", m.createTaskHandler).Methods("POST")
	api.HandleFunc("/tasks", m.listTasksHandler).Methods("GET")
	api.HandleFunc("/tasks/{id}", m.getTaskHandler).Methods("GET")

	// Worker management
	api.HandleFunc("/workers", m.listWorkersHandler).Methods("GET")
	api.HandleFunc("/workers/{id}", m.getWorkerHandler).Methods("GET")

	// Metrics
	api.HandleFunc("/metrics", m.metricsHandler).Methods("GET")

	// Health check
	router.HandleFunc("/health", m.healthHandler).Methods("GET")

	m.server = &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("🌐 API server starting on :8080")
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Server error: %v", err)
		}
	}()

	return nil
}

// taskScheduler schedules periodic tasks
func (m *EnterpriseManager) taskScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.shutdownChan:
			return
		case <-ticker.C:
			m.scheduleScanTasks(ctx)
		}
	}
}

// scheduleScanTasks creates scanning tasks for all organizations
func (m *EnterpriseManager) scheduleScanTasks(ctx context.Context) {
	log.Println("📋 Scheduling periodic scan tasks")

	// Create scan tasks for different providers
	providers := []string{"aws", "azure", "gcp"}
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}

	for _, provider := range providers {
		for _, region := range regions {
			task := Task{
				ID:       fmt.Sprintf("scan-%s-%s-%d", provider, region, time.Now().Unix()),
				Type:     "scan",
				Priority: 3,
				Payload: map[string]interface{}{
					"org_id":   "default",
					"provider": provider,
					"region":   region,
				},
				CreatedAt:   time.Now(),
				Attempts:    0,
				MaxAttempts: 3,
			}

			if err := m.enqueueTask(ctx, task); err != nil {
				log.Printf("⚠️  Failed to enqueue scan task: %v", err)
			}
		}
	}
}

// enqueueTask adds a task to the Redis queue
func (m *EnterpriseManager) enqueueTask(ctx context.Context, task Task) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	queue := "tasks:normal"
	if task.Priority > 5 {
		queue = "tasks:high_priority"
	}

	return m.redis.LPush(ctx, queue, taskData).Err()
}

// workerMonitor monitors active workers
func (m *EnterpriseManager) workerMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.shutdownChan:
			return
		case <-ticker.C:
			m.checkWorkerHealth(ctx)
		}
	}
}

// checkWorkerHealth verifies worker status
func (m *EnterpriseManager) checkWorkerHealth(ctx context.Context) {
	workers, err := m.redis.SMembers(ctx, "workers:active").Result()
	if err != nil {
		log.Printf("⚠️  Failed to get active workers: %v", err)
		return
	}

	log.Printf("📊 Active workers: %d", len(workers))

	for _, workerID := range workers {
		key := fmt.Sprintf("workers:%s", workerID)
		data, err := m.redis.Get(ctx, key).Result()
		if err == redis.Nil {
			// Worker heartbeat expired, remove from active set
			m.redis.SRem(ctx, "workers:active", workerID)
			log.Printf("⚠️  Worker %s heartbeat expired", workerID)
		} else if err != nil {
			log.Printf("⚠️  Error checking worker %s: %v", workerID, err)
		} else {
			var heartbeat map[string]interface{}
			if err := json.Unmarshal([]byte(data), &heartbeat); err == nil {
				log.Printf("💓 Worker %s: processed=%v, failed=%v",
					workerID, heartbeat["tasks_processed"], heartbeat["tasks_failed"])
			}
		}
	}
}

// metricsCollector collects and stores metrics
func (m *EnterpriseManager) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.shutdownChan:
			return
		case <-ticker.C:
			m.collectMetrics(ctx)
		}
	}
}

// collectMetrics gathers system metrics
func (m *EnterpriseManager) collectMetrics(ctx context.Context) {
	// Get worker count
	workers, _ := m.redis.SMembers(ctx, "workers:active").Result()
	workerCount := len(workers)

	// Get queue sizes
	highPriorityQueue, _ := m.redis.LLen(ctx, "tasks:high_priority").Result()
	normalQueue, _ := m.redis.LLen(ctx, "tasks:normal").Result()

	// Get token tracker stats
	stats := m.tokenTracker.GetStats()

	metrics := map[string]interface{}{
		"timestamp":           time.Now().Unix(),
		"worker_count":        workerCount,
		"high_priority_queue": highPriorityQueue,
		"normal_queue":        normalQueue,
		"total_tokens":        stats["total_tokens"],
		"total_cost":          stats["total_cost_usd"],
		"total_savings":       stats["total_savings_usd"],
	}

	// Store metrics in Redis
	metricsData, _ := json.Marshal(metrics)
	m.redis.LPush(ctx, "metrics:timeline", metricsData)
	m.redis.LTrim(ctx, "metrics:timeline", 0, 1000) // Keep last 1000 data points

	log.Printf("📈 Metrics: workers=%d, queues=%d/%d, cost=$%.2f, savings=$%.2f",
		workerCount, highPriorityQueue, normalQueue,
		stats["total_cost_usd"], stats["total_savings_usd"])
}

// HTTP Handlers

func (m *EnterpriseManager) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "healthy",
		"manager_id": m.id,
		"version":    "2.0.0",
	})
}

func (m *EnterpriseManager) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.ID = fmt.Sprintf("task-%d", time.Now().Unix())
	task.CreatedAt = time.Now()
	task.Attempts = 0
	task.MaxAttempts = 3

	if err := m.enqueueTask(r.Context(), task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (m *EnterpriseManager) listTasksHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement task listing from Redis
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]Task{})
}

func (m *EnterpriseManager) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement task retrieval
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": mux.Vars(r)["id"]})
}

func (m *EnterpriseManager) listWorkersHandler(w http.ResponseWriter, r *http.Request) {
	workers, err := m.redis.SMembers(r.Context(), "workers:active").Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

func (m *EnterpriseManager) getWorkerHandler(w http.ResponseWriter, r *http.Request) {
	workerID := mux.Vars(r)["id"]
	key := fmt.Sprintf("workers:%s", workerID)

	data, err := m.redis.Get(r.Context(), key).Result()
	if err == redis.Nil {
		http.Error(w, "Worker not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (m *EnterpriseManager) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := m.redis.LRange(r.Context(), "metrics:timeline", 0, 100).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
