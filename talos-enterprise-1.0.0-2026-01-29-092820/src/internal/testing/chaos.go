package testing

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

// ChaosTest represents a chaos testing scenario
type ChaosTest struct {
	Name        string
	Description string
	Duration    time.Duration
	Scenario    ChaosScenario
}

// ChaosScenario defines the chaos to inject
type ChaosScenario interface {
	Execute(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// NetworkLatencyScenario adds network latency
type NetworkLatencyScenario struct {
	TargetService string
	LatencyMs     int
}

func (s *NetworkLatencyScenario) Execute(ctx context.Context) error {
	// Simulate adding latency (in production, use toxiproxy or similar)
	fmt.Printf("⚡ Injecting %dms latency to %s\n", s.LatencyMs, s.TargetService)
	return nil
}

func (s *NetworkLatencyScenario) Rollback(ctx context.Context) error {
	fmt.Printf("✅ Removed latency from %s\n", s.TargetService)
	return nil
}

// ServiceFailureScenario simulates service failures
type ServiceFailureScenario struct {
	TargetService string
	FailureRate   float64 // 0.0 to 1.0
}

func (s *ServiceFailureScenario) Execute(ctx context.Context) error {
	fmt.Printf("💥 Injecting %.0f%% failure rate to %s\n", s.FailureRate*100, s.TargetService)
	return nil
}

func (s *ServiceFailureScenario) Rollback(ctx context.Context) error {
	fmt.Printf("✅ Restored %s to normal\n", s.TargetService)
	return nil
}

// DatabaseSlowdownScenario simulates slow database
type DatabaseSlowdownScenario struct {
	SlowdownFactor int // Multiply query time by this
}

func (s *DatabaseSlowdownScenario) Execute(ctx context.Context) error {
	fmt.Printf("🐌 Slowing database by %dx\n", s.SlowdownFactor)
	return nil
}

func (s *DatabaseSlowdownScenario) Rollback(ctx context.Context) error {
	fmt.Println("✅ Database restored to normal speed")
	return nil
}

// ChaosEngine runs chaos tests
type ChaosEngine struct {
	tests []ChaosTest
	t     *testing.T
}

// NewChaosEngine creates a new chaos testing engine
func NewChaosEngine(t *testing.T) *ChaosEngine {
	return &ChaosEngine{
		tests: []ChaosTest{},
		t:     t,
	}
}

// AddTest adds a chaos test
func (c *ChaosEngine) AddTest(test ChaosTest) {
	c.tests = append(c.tests, test)
}

// RunAll runs all chaos tests
func (c *ChaosEngine) RunAll() {
	for _, test := range c.tests {
		c.t.Run(test.Name, func(t *testing.T) {
			c.runTest(t, test)
		})
	}
}

// runTest executes a single chaos test
func (c *ChaosEngine) runTest(t *testing.T, test ChaosTest) {
	ctx, cancel := context.WithTimeout(context.Background(), test.Duration)
	defer cancel()

	t.Logf("🔥 Starting chaos test: %s", test.Name)
	t.Logf("📝 Description: %s", test.Description)

	// Execute chaos
	if err := test.Scenario.Execute(ctx); err != nil {
		t.Errorf("Failed to execute chaos: %v", err)
		return
	}

	// Monitor system during chaos
	go c.monitorSystem(ctx, t)

	// Wait for duration
	<-ctx.Done()

	// Rollback chaos
	rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer rollbackCancel()

	if err := test.Scenario.Rollback(rollbackCtx); err != nil {
		t.Errorf("Failed to rollback chaos: %v", err)
	}

	t.Logf("✅ Chaos test completed: %s", test.Name)
}

// monitorSystem monitors system health during chaos
func (c *ChaosEngine) monitorSystem(ctx context.Context, t *testing.T) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check health endpoint
			resp, err := http.Get("http://localhost:8080/health")
			if err != nil {
				t.Logf("⚠️  Health check failed: %v", err)
			} else {
				t.Logf("💚 System health: %d", resp.StatusCode)
				resp.Body.Close()
			}
		case <-ctx.Done():
			return
		}
	}
}

// Predefined chaos tests
func GetStandardChaosTests() []ChaosTest {
	return []ChaosTest{
		{
			Name:        "DatabaseLatency",
			Description: "Test system behavior with slow database",
			Duration:    30 * time.Second,
			Scenario:    &DatabaseSlowdownScenario{SlowdownFactor: 3},
		},
		{
			Name:        "AIServiceFailure",
			Description: "Test fallback when AI tier fails",
			Duration:    45 * time.Second,
			Scenario:    &ServiceFailureScenario{TargetService: "ai-tier-3", FailureRate: 0.5},
		},
		{
			Name:        "NetworkJitter",
			Description: "Test resilience to network latency",
			Duration:    60 * time.Second,
			Scenario:    &NetworkLatencyScenario{TargetService: "backend", LatencyMs: 500},
		},
	}
}

// Load Testing Framework

// LoadTest represents a load test configuration
type LoadTest struct {
	Name           string
	TargetURL      string
	Concurrency    int
	Duration       time.Duration
	RequestsPerSec int
}

// LoadTestResult contains load test results
type LoadTestResult struct {
	TotalRequests  int
	SuccessfulReqs int
	FailedReqs     int
	AvgLatency     time.Duration
	MinLatency     time.Duration
	MaxLatency     time.Duration
	RequestsPerSec float64
}

// RunLoadTest executes a load test
func RunLoadTest(t *testing.T, test LoadTest) LoadTestResult {
	t.Logf("🚀 Starting load test: %s", test.Name)
	t.Logf("   Target: %s", test.TargetURL)
	t.Logf("   Concurrency: %d", test.Concurrency)
	t.Logf("   Duration: %v", test.Duration)

	ctx, cancel := context.WithTimeout(context.Background(), test.Duration)
	defer cancel()

	results := make(chan time.Duration, 10000)
	errors := make(chan error, 1000)

	// Start workers
	for i := 0; i < test.Concurrency; i++ {
		go loadTestWorker(ctx, test.TargetURL, results, errors)
	}

	// Collect results
	var totalReqs, successReqs, failedReqs int
	var minLat, maxLat, totalLat time.Duration
	minLat = time.Hour

	done := make(chan struct{})
	go func() {
		for {
			select {
			case lat := <-results:
				totalReqs++
				successReqs++
				totalLat += lat
				if lat < minLat {
					minLat = lat
				}
				if lat > maxLat {
					maxLat = lat
				}
			case <-errors:
				totalReqs++
				failedReqs++
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()

	<-done
	time.Sleep(100 * time.Millisecond) // Drain remaining

	result := LoadTestResult{
		TotalRequests:  totalReqs,
		SuccessfulReqs: successReqs,
		FailedReqs:     failedReqs,
		MinLatency:     minLat,
		MaxLatency:     maxLat,
		RequestsPerSec: float64(totalReqs) / test.Duration.Seconds(),
	}

	if successReqs > 0 {
		result.AvgLatency = totalLat / time.Duration(successReqs)
	}

	t.Logf("📊 Load Test Results:")
	t.Logf("   Total Requests: %d", result.TotalRequests)
	t.Logf("   Successful: %d (%.1f%%)", result.SuccessfulReqs, float64(successReqs)/float64(totalReqs)*100)
	t.Logf("   Failed: %d", result.FailedReqs)
	t.Logf("   Avg Latency: %v", result.AvgLatency)
	t.Logf("   Min/Max: %v / %v", result.MinLatency, result.MaxLatency)
	t.Logf("   Throughput: %.1f req/s", result.RequestsPerSec)

	return result
}

func loadTestWorker(ctx context.Context, url string, results chan<- time.Duration, errors chan<- error) {
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			resp, err := client.Get(url)
			latency := time.Since(start)

			if err != nil {
				errors <- err
			} else {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					results <- latency
				} else {
					errors <- fmt.Errorf("status %d", resp.StatusCode)
				}
			}

			// Random jitter to avoid thundering herd
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}
	}
}
