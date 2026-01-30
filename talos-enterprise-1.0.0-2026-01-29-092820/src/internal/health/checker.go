package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// CheckResult represents a health check result
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Latency   time.Duration          `json:"latency"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Checker represents a health check function
type Checker interface {
	Check(ctx context.Context) CheckResult
	Name() string
}

// HealthManager manages all health checks
type HealthManager struct {
	mu       sync.RWMutex
	checkers map[string]Checker
}

// NewHealthManager creates a new health manager
func NewHealthManager() *HealthManager {
	return &HealthManager{
		checkers: make(map[string]Checker),
	}
}

// Register adds a health checker
func (h *HealthManager) Register(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[checker.Name()] = checker
}

// CheckAll runs all health checks
func (h *HealthManager) CheckAll(ctx context.Context) map[string]CheckResult {
	h.mu.RLock()
	checkers := make(map[string]Checker, len(h.checkers))
	for k, v := range h.checkers {
		checkers[k] = v
	}
	h.mu.RUnlock()

	results := make(map[string]CheckResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range checkers {
		wg.Add(1)
		go func(n string, c Checker) {
			defer wg.Done()

			result := c.Check(ctx)
			mu.Lock()
			results[n] = result
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	return results
}

// GetOverallStatus determines overall system health
func (h *HealthManager) GetOverallStatus(ctx context.Context) Status {
	results := h.CheckAll(ctx)

	unhealthyCount := 0
	degradedCount := 0

	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	// If any critical component is unhealthy
	if unhealthyCount > 0 {
		return StatusUnhealthy
	}

	// If any component is degraded
	if degradedCount > 0 {
		return StatusDegraded
	}

	return StatusHealthy
}

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	name   string
	pingFn func(context.Context) error
}

func NewDatabaseChecker(name string, pingFn func(context.Context) error) *DatabaseChecker {
	return &DatabaseChecker{name: name, pingFn: pingFn}
}

func (d *DatabaseChecker) Name() string {
	return d.name
}

func (d *DatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	err := d.pingFn(ctx)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:      d.name,
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("database connection failed: %v", err),
			Latency:   latency,
			Timestamp: time.Now(),
		}
	}

	status := StatusHealthy
	if latency > 100*time.Millisecond {
		status = StatusDegraded
	}

	return CheckResult{
		Name:      d.name,
		Status:    status,
		Latency:   latency,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"latency_ms": latency.Milliseconds(),
		},
	}
}

// AIChecker checks AI service availability
type AIChecker struct {
	name     string
	healthFn func(context.Context) error
}

func NewAIChecker(name string, healthFn func(context.Context) error) *AIChecker {
	return &AIChecker{name: name, healthFn: healthFn}
}

func (a *AIChecker) Name() string {
	return a.name
}

func (a *AIChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	err := a.healthFn(ctx)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:      a.name,
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("AI service unavailable: %v", err),
			Latency:   latency,
			Timestamp: time.Now(),
		}
	}

	return CheckResult{
		Name:      a.name,
		Status:    StatusHealthy,
		Latency:   latency,
		Timestamp: time.Now(),
	}
}

// CacheChecker checks cache connectivity
type CacheChecker struct {
	name   string
	pingFn func(context.Context) error
}

func NewCacheChecker(name string, pingFn func(context.Context) error) *CacheChecker {
	return &CacheChecker{name: name, pingFn: pingFn}
}

func (c *CacheChecker) Name() string {
	return c.name
}

func (c *CacheChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	err := c.pingFn(ctx)
	latency := time.Since(start)

	if err != nil {
		// Cache is non-critical, so degraded instead of unhealthy
		return CheckResult{
			Name:      c.name,
			Status:    StatusDegraded,
			Message:   fmt.Sprintf("cache unavailable (non-critical): %v", err),
			Latency:   latency,
			Timestamp: time.Now(),
		}
	}

	return CheckResult{
		Name:      c.name,
		Status:    StatusHealthy,
		Latency:   latency,
		Timestamp: time.Now(),
	}
}
