package testing

import (
	"testing"
	"time"
)

// TestChaosScenarios runs all chaos tests
func TestChaosScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos tests in short mode")
	}

	engine := NewChaosEngine(t)

	for _, test := range GetStandardChaosTests() {
		engine.AddTest(test)
	}

	engine.RunAll()
}

// TestLoadBaseline tests system under normal load
func TestLoadBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	test := LoadTest{
		Name:           "Baseline Load",
		TargetURL:      "http://localhost:8080/health",
		Concurrency:    10,
		Duration:       30 * time.Second,
		RequestsPerSec: 100,
	}

	result := RunLoadTest(t, test)

	// Assert performance criteria
	if result.AvgLatency > 500*time.Millisecond {
		t.Errorf("Average latency too high: %v", result.AvgLatency)
	}

	successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
	if successRate < 0.95 {
		t.Errorf("Success rate too low: %.2f%%", successRate*100)
	}
}

// TestLoadSpike tests system under sudden load spike
func TestLoadSpike(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	test := LoadTest{
		Name:           "Load Spike",
		TargetURL:      "http://localhost:8080/api/optimize",
		Concurrency:    50,
		Duration:       20 * time.Second,
		RequestsPerSec: 500,
	}

	result := RunLoadTest(t, test)

	// More lenient criteria for spike test
	successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
	if successRate < 0.80 {
		t.Errorf("Success rate too low during spike: %.2f%%", successRate*100)
	}
}
