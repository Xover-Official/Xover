package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/project-atlas/atlas/internal/persistence"
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

// DistributedWorker represents a scalable worker node
type DistributedWorker struct {
	id           string
	redis        *redis.Client
	db           persistence.Ledger
	orchestrator *ai.UnifiedOrchestrator
	tokenTracker *analytics.TokenTracker
	config       *config.Config

	// Worker state
	isRunning    bool
	taskQueue    chan Task
	errorQueue   chan error
	shutdownChan chan struct{}
	wg           sync.WaitGroup

	// Metrics
	tasksProcessed int64
	tasksFailed    int64
	lastHeartbeat  time.Time
}

// NewDistributedWorker creates a new enterprise worker
func NewDistributedWorker(workerID string, cfg *config.Config, db persistence.Ledger, orchestrator *ai.UnifiedOrchestrator, tracker *analytics.TokenTracker) (*DistributedWorker, error) {
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

	worker := &DistributedWorker{
		id:            workerID,
		redis:         rdb,
		db:            db,
		orchestrator:  orchestrator,
		tokenTracker:  tracker,
		config:        cfg,
		taskQueue:     make(chan Task, 100),
		errorQueue:    make(chan error, 10),
		shutdownChan:  make(chan struct{}),
		lastHeartbeat: time.Now(),
	}

	return worker, nil
}

// Start begins the worker's distributed processing loop
func (w *DistributedWorker) Start(ctx context.Context) error {
	w.isRunning = true
	log.Printf("🚀 Starting distributed worker: %s", w.id)

	// Create healthy file for K8s probes
	if err := os.WriteFile("/tmp/healthy", []byte("ok"), 0644); err != nil {
		log.Printf("⚠️  Failed to create health file: %v", err)
	}
	defer os.Remove("/tmp/healthy")

	// Start heartbeat goroutine
	w.wg.Add(1)
	go w.heartbeatLoop(ctx)

	// Start task processor goroutines
	concurrency := 10 // Default concurrency
	for i := 0; i < concurrency; i++ {
		w.wg.Add(1)
		go w.taskProcessor(ctx, i)
	}

	// Start error handler
	w.wg.Add(1)
	go w.errorHandler(ctx)

	// Start task fetcher
	w.wg.Add(1)
	go w.taskFetcher(ctx)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("🛑 Received shutdown signal")
	case <-ctx.Done():
		log.Println("🛑 Context cancelled")
	case <-w.shutdownChan:
		log.Println("🛑 Shutdown requested")
	}

	return w.Shutdown()
}

// Shutdown gracefully stops the worker
func (w *DistributedWorker) Shutdown() error {
	if !w.isRunning {
		return nil
	}

	log.Println("🛑 Shutting down distributed worker...")
	w.isRunning = false
	close(w.shutdownChan)

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ Worker shutdown complete")
	case <-time.After(30 * time.Second):
		log.Println("⚠️  Worker shutdown timeout")
	}

	// Close Redis connection
	if err := w.redis.Close(); err != nil {
		log.Printf("⚠️  Error closing Redis: %v", err)
	}

	return nil
}

// heartbeatLoop sends periodic heartbeats to Redis
func (w *DistributedWorker) heartbeatLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdownChan:
			return
		case <-ticker.C:
			w.sendHeartbeat(ctx)
		}
	}
}

// sendHeartbeat updates worker status in Redis
func (w *DistributedWorker) sendHeartbeat(ctx context.Context) {
	heartbeat := map[string]interface{}{
		"id":              w.id,
		"status":          "active",
		"last_seen":       time.Now().Unix(),
		"tasks_processed": w.tasksProcessed,
		"tasks_failed":    w.tasksFailed,
		"version":         "2.0.0",
	}

	data, _ := json.Marshal(heartbeat)
	key := fmt.Sprintf("workers:%s", w.id)

	w.redis.Set(ctx, key, data, 30*time.Second)
	w.redis.SAdd(ctx, "workers:active", w.id)
}

// taskFetcher pulls tasks from Redis queue
func (w *DistributedWorker) taskFetcher(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdownChan:
			return
		default:
			task, err := w.fetchTask(ctx)
			if err != nil {
				w.errorQueue <- fmt.Errorf("task fetch error: %w", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if task != nil {
				select {
				case w.taskQueue <- *task:
				case <-ctx.Done():
					return
				case <-w.shutdownChan:
					return
				}
			} else {
				// No tasks available, wait briefly
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// fetchTask retrieves a task from Redis queue
func (w *DistributedWorker) fetchTask(ctx context.Context) (*Task, error) {
	// Try to get a task from high priority queue first
	result, err := w.redis.BRPop(ctx, 1*time.Second, "tasks:high_priority").Result()
	if err == redis.Nil {
		// Try normal priority queue
		result, err = w.redis.BRPop(ctx, 1*time.Second, "tasks:normal").Result()
	}

	if err == redis.Nil {
		return nil, nil // No tasks available
	}
	if err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid task result")
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// taskProcessor processes tasks from the queue
func (w *DistributedWorker) taskProcessor(ctx context.Context, workerID int) {
	defer w.wg.Done()

	log.Printf("🔄 Starting task processor %d", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdownChan:
			return
		case task := <-w.taskQueue:
			w.processTask(ctx, task)
		}
	}
}

// processTask handles a single task
func (w *DistributedWorker) processTask(ctx context.Context, task Task) {
	log.Printf("📋 Processing task %s (type: %s, worker: %s)", task.ID, task.Type, w.id)

	start := time.Now()

	// Update task status to processing
	w.updateTaskStatus(ctx, task.ID, "processing")

	var err error
	switch task.Type {
	case "scan":
		err = w.handleScanTask(ctx, task)
	case "analyze":
		err = w.handleAnalyzeTask(ctx, task)
	case "optimize":
		err = w.handleOptimizeTask(ctx, task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	duration := time.Since(start)

	if err != nil {
		w.tasksFailed++
		w.errorQueue <- fmt.Errorf("task %s failed: %w", task.ID, err)
		w.updateTaskStatus(ctx, task.ID, "failed")

		// Retry logic
		if task.Attempts < task.MaxAttempts {
			task.Attempts++
			w.retryTask(ctx, task)
		}
	} else {
		w.tasksProcessed++
		w.updateTaskStatus(ctx, task.ID, "completed")
		log.Printf("✅ Task %s completed in %v", task.ID, duration)
	}
}

// handleScanTask processes cloud resource scanning
func (w *DistributedWorker) handleScanTask(ctx context.Context, task Task) error {
	// Extract scan parameters
	orgID, ok := task.Payload["org_id"].(string)
	if !ok {
		return fmt.Errorf("missing org_id in scan task")
	}

	provider, ok := task.Payload["provider"].(string)
	if !ok {
		return fmt.Errorf("missing provider in scan task")
	}

	region, _ := task.Payload["region"].(string)

	log.Printf("🔍 Scanning resources for org %s, provider %s, region %s", orgID, provider, region)

	// TODO: Implement actual cloud scanning logic
	// This would integrate with cloud provider APIs
	time.Sleep(2 * time.Second) // Simulate work

	return nil
}

// handleAnalyzeTask processes AI analysis
func (w *DistributedWorker) handleAnalyzeTask(ctx context.Context, task Task) error {
	resourceID, ok := task.Payload["resource_id"].(string)
	if !ok {
		return fmt.Errorf("missing resource_id in analyze task")
	}

	prompt, ok := task.Payload["prompt"].(string)
	if !ok {
		return fmt.Errorf("missing prompt in analyze task")
	}

	riskScore, _ := task.Payload["risk_score"].(float64)
	if riskScore == 0 {
		riskScore = 5.0 // Default
	}

	// Call AI orchestrator
	response, err := w.orchestrator.Analyze(ctx, prompt, riskScore, nil)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	// Store AI decision
	// TODO: Store in database
	log.Printf("🤖 AI analysis completed for %s: %s", resourceID, response.Content[:100])

	return nil
}

// handleOptimizeTask processes optimization actions
func (w *DistributedWorker) handleOptimizeTask(ctx context.Context, task Task) error {
	actionID, ok := task.Payload["action_id"].(string)
	if !ok {
		return fmt.Errorf("missing action_id in optimize task")
	}

	// Execute the optimization
	// TODO: Implement actual cloud optimization logic
	time.Sleep(1 * time.Second) // Simulate work

	log.Printf("⚡ Optimization completed for action %s", actionID)
	return nil
}

// updateTaskStatus updates task status in Redis
func (w *DistributedWorker) updateTaskStatus(ctx context.Context, taskID, status string) {
	key := fmt.Sprintf("tasks:%s:status", taskID)
	w.redis.Set(ctx, key, status, 24*time.Hour)
}

// retryTask requeues a failed task for retry
func (w *DistributedWorker) retryTask(ctx context.Context, task Task) {
	taskData, _ := json.Marshal(task)

	queue := "tasks:normal"
	if task.Priority > 5 {
		queue = "tasks:high_priority"
	}

	w.redis.LPush(ctx, queue, taskData)
	log.Printf("🔄 Retrying task %s (attempt %d/%d)", task.ID, task.Attempts, task.MaxAttempts)
}

// errorHandler processes errors from the error queue
func (w *DistributedWorker) errorHandler(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdownChan:
			return
		case err := <-w.errorQueue:
			log.Printf("❌ Worker error: %v", err)
			// TODO: Send to monitoring system
		}
	}
}
