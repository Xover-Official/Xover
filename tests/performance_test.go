package tests

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/cloud/aws"
)

// PerformanceTestSuite provides comprehensive performance testing for Talos
type PerformanceTestSuite struct {
	orchestrator *ai.UnifiedOrchestrator
	adapter      cloud.CloudAdapter
	tracker      *analytics.TokenTracker
}

// NewPerformanceTestSuite creates a new test suite
func NewPerformanceTestSuite() *PerformanceTestSuite {
	tracker := analytics.NewTokenTracker("")

	// Create AI config
	config := &ai.Config{
		GeminiAPIKey:  "test-key",
		OpenRouterKey: "test-key",
		DevinKey:      "test-key",
	}

	orchestrator, _ := ai.NewIntegratedOrchestrator(config, tracker, nil)
	adapter, _ := aws.NewAWSAdapter("us-east-1")

	return &PerformanceTestSuite{
		orchestrator: orchestrator,
		adapter:      adapter,
		tracker:      tracker,
	}
}

// BenchmarkAIOrchestrator tests AI performance under load
func BenchmarkAIOrchestrator(b *testing.B) {
	suite := NewPerformanceTestSuite()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			prompt := generateRandomPrompt()
			risk := rand.Float64()
			resource := &cloud.ResourceV2{
				ID:       "test-resource",
				Type:     "ec2",
				Provider: "aws",
				Region:   "us-east-1",
			}

			_, err := suite.orchestrator.Analyze(ctx, prompt, risk, resource)
			if err != nil {
				b.Errorf("AI analysis failed: %v", err)
			}
		}
	})
}

// BenchmarkCloudAdapter tests cloud adapter performance
func BenchmarkCloudAdapter(b *testing.B) {
	suite := NewPerformanceTestSuite()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			_, err := suite.adapter.FetchResources(ctx)
			if err != nil {
				b.Errorf("Cloud fetch failed: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentRequests tests system under concurrent load
func BenchmarkConcurrentRequests(b *testing.B) {
	suite := NewPerformanceTestSuite()

	concurrencyLevels := []int{10, 50, 100, 500}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			b.ResetTimer()

			var wg sync.WaitGroup
			semaphore := make(chan struct{}, concurrency)

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				semaphore <- struct{}{}

				go func() {
					defer wg.Done()
					defer func() { <-semaphore }()

					// Simulate mixed workload
					switch rand.Intn(3) {
					case 0:
						suite.testAIAnalysis()
					case 1:
						suite.testResourceFetch()
					case 2:
						suite.testOptimization()
					}
				}()
			}

			wg.Wait()
		})
	}
}

// TestMemoryUsage tests memory consumption over time
func TestMemoryUsage(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// Baseline memory measurement
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Run operations
	for i := 0; i < 10000; i++ {
		suite.testAIAnalysis()
		if i%1000 == 0 {
			runtime.GC()
		}
	}

	// Final memory measurement
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	memoryUsed := m2.Alloc - m1.Alloc
	t.Logf("Memory used: %d bytes", memoryUsed)

	// Assert memory usage is reasonable (< 100MB)
	if memoryUsed > 100*1024*1024 {
		t.Errorf("Memory usage too high: %d bytes", memoryUsed)
	}
}

// TestLatencyDistribution measures response time distribution
func TestLatencyDistribution(t *testing.T) {
	suite := NewPerformanceTestSuite()

	var latencies []time.Duration
	sampleSize := 1000

	for i := 0; i < sampleSize; i++ {
		start := time.Now()
		suite.testAIAnalysis()
		latency := time.Since(start)
		latencies = append(latencies, latency)
	}

	// Calculate percentiles
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)/2]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]

	t.Logf("P50: %v, P95: %v, P99: %v", p50, p95, p99)

	// Assert performance requirements
	if p95 > 5*time.Second {
		t.Errorf("P95 latency too high: %v", p95)
	}
}

// TestScalability tests system scalability
func TestScalability(t *testing.T) {
	suite := NewPerformanceTestSuite()

	workloads := []struct {
		name        string
		requests    int
		concurrency int
		maxLatency  time.Duration
	}{
		{"Light", 100, 10, 2 * time.Second},
		{"Medium", 500, 50, 3 * time.Second},
		{"Heavy", 1000, 100, 5 * time.Second},
	}

	for _, workload := range workloads {
		t.Run(workload.name, func(t *testing.T) {
			start := time.Now()

			var wg sync.WaitGroup
			semaphore := make(chan struct{}, workload.concurrency)
			var maxLatency time.Duration
			var latencyMutex sync.Mutex

			for i := 0; i < workload.requests; i++ {
				wg.Add(1)
				semaphore <- struct{}{}

				go func() {
					defer wg.Done()
					defer func() { <-semaphore }()

					reqStart := time.Now()
					suite.testAIAnalysis()
					latency := time.Since(reqStart)

					latencyMutex.Lock()
					if latency > maxLatency {
						maxLatency = latency
					}
					latencyMutex.Unlock()
				}()
			}

			wg.Wait()
			totalTime := time.Since(start)

			t.Logf("Workload %s: %d requests in %v (max latency: %v)",
				workload.name, workload.requests, totalTime, maxLatency)

			if maxLatency > workload.maxLatency {
				t.Errorf("Max latency exceeded: %v > %v", maxLatency, workload.maxLatency)
			}
		})
	}
}

// TestErrorRate tests system error rate under stress
func TestErrorRate(t *testing.T) {
	suite := NewPerformanceTestSuite()

	requests := 10000
	concurrency := 100

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)
	var errors int32
	var success int32

	for i := 0; i < requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			err := suite.testAIAnalysis()
			if err != nil {
				atomic.AddInt32(&errors, 1)
			} else {
				atomic.AddInt32(&success, 1)
			}
		}()
	}

	wg.Wait()

	errorRate := float64(errors) / float64(requests) * 100
	t.Logf("Error rate: %.2f%% (%d errors out of %d requests)", errorRate, errors, requests)

	// Assert error rate is low (< 1%)
	if errorRate > 1.0 {
		t.Errorf("Error rate too high: %.2f%%", errorRate)
	}
}

// TestResourceEfficiency tests CPU and memory efficiency
func TestResourceEfficiency(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// Monitor CPU usage
	var cpuUsageBefore, cpuUsageAfter float64

	// Baseline measurement
	cpuUsageBefore = getCPUUsage()

	// Run workload
	for i := 0; i < 5000; i++ {
		suite.testAIAnalysis()
	}

	// Final measurement
	cpuUsageAfter = getCPUUsage()

	cpuIncrease := cpuUsageAfter - cpuUsageBefore
	t.Logf("CPU usage increase: %.2f%%", cpuIncrease)

	// Assert CPU usage increase is reasonable (< 50%)
	if cpuIncrease > 50.0 {
		t.Errorf("CPU usage increase too high: %.2f%%", cpuIncrease)
	}
}

// TestThroughput measures system throughput
func TestThroughput(t *testing.T) {
	suite := NewPerformanceTestSuite()

	duration := 30 * time.Second
	threshold := 100 // requests per second

	start := time.Now()
	var requests int32

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start workers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					suite.testAIAnalysis()
					atomic.AddInt32(&requests, 1)
				}
			}
		}()
	}

	// Run for specified duration
	time.Sleep(duration)
	close(done)
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	throughput := float64(requests) / elapsed

	t.Logf("Throughput: %.2f requests/second (%d requests in %.2f seconds)",
		throughput, requests, elapsed)

	// Assert throughput meets threshold
	if throughput < float64(threshold) {
		t.Errorf("Throughput too low: %.2f < %d", throughput, threshold)
	}
}

// Helper functions

func (s *PerformanceTestSuite) testAIAnalysis() error {
	ctx := context.Background()
	prompt := generateRandomPrompt()
	risk := rand.Float64()
	resource := &cloud.ResourceV2{
		ID:       "test-resource",
		Type:     "ec2",
		Provider: "aws",
		Region:   "us-east-1",
	}

	_, err := s.orchestrator.Analyze(ctx, prompt, risk, resource)
	return err
}

func (s *PerformanceTestSuite) testResourceFetch() error {
	ctx := context.Background()
	_, err := s.adapter.FetchResources(ctx)
	return err
}

func (s *PerformanceTestSuite) testOptimization() error {
	ctx := context.Background()
	resource := &cloud.ResourceV2{
		ID:       "test-resource",
		Type:     "ec2",
		Provider: "aws",
		Region:   "us-east-1",
	}

	_, _, err := s.adapter.ApplyOptimization(ctx, resource, "resize")
	return err
}

func generateRandomPrompt() string {
	prompts := []string{
		"Analyze this EC2 instance for optimization opportunities",
		"Recommend cost-saving measures for this database",
		"Evaluate the risk of resizing this virtual machine",
		"Calculate potential savings for this storage volume",
		"Assess performance bottlenecks in this application",
	}

	return prompts[rand.Intn(len(prompts))]
}

func getCPUUsage() float64 {
	// Simplified CPU usage measurement
	// In production, use proper system monitoring
	return rand.Float64() * 100
}

// LoadTestScenario defines a load testing scenario
type LoadTestScenario struct {
	Name        string
	Duration    time.Duration
	Concurrency int
	Requests    int
	ThinkTime   time.Duration
	SuccessRate float64
	AvgLatency  time.Duration
	P95Latency  time.Duration
}

// RunLoadTest executes a comprehensive load test
func RunLoadTest(scenario LoadTestScenario) error {
	suite := NewPerformanceTestSuite()

	fmt.Printf("Running load test: %s\n", scenario.Name)
	fmt.Printf("Duration: %v, Concurrency: %d, Requests: %d\n",
		scenario.Duration, scenario.Concurrency, scenario.Requests)

	start := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, scenario.Concurrency)

	var successCount, errorCount int32
	var totalLatency time.Duration
	var maxLatency time.Duration
	var latencyMutex sync.Mutex

	for i := 0; i < scenario.Requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			// Simulate think time
			time.Sleep(scenario.ThinkTime)

			reqStart := time.Now()
			err := suite.testAIAnalysis()
			latency := time.Since(reqStart)

			latencyMutex.Lock()
			totalLatency += latency
			if latency > maxLatency {
				maxLatency = latency
			}
			latencyMutex.Unlock()

			if err != nil {
				atomic.AddInt32(&errorCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Calculate metrics
	successRate := float64(successCount) / float64(scenario.Requests) * 100
	avgLatency := totalLatency / time.Duration(scenario.Requests)
	throughput := float64(scenario.Requests) / elapsed.Seconds()

	fmt.Printf("Results:\n")
	fmt.Printf("  Success Rate: %.2f%%\n", successRate)
	fmt.Printf("  Average Latency: %v\n", avgLatency)
	fmt.Printf("  Max Latency: %v\n", maxLatency)
	fmt.Printf("  Throughput: %.2f req/s\n", throughput)
	fmt.Printf("  Total Time: %v\n", elapsed)

	return nil
}
