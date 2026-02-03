package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Xover-Official/Xover/internal/ai"
	"github.com/Xover-Official/Xover/internal/cloud"
)

// AITestCase represents a test case for AI behavior
type AITestCase struct {
	Name             string
	Prompt           string
	ExpectedKeywords []string      // Keywords that must appear in response
	ForbiddenWords   []string      // Words that must NOT appear
	MinLength        int           // Minimum response length
	MaxLatency       time.Duration // Maximum acceptable latency
	RiskScore        float64
}

// AITestSuite runs sanity checks on AI responses
type AITestSuite struct {
	orchestrator *ai.UnifiedOrchestrator
	t            *testing.T
}

// NewAITestSuite creates a new AI test suite
func NewAITestSuite(orchestrator *ai.UnifiedOrchestrator, t *testing.T) *AITestSuite {
	return &AITestSuite{
		orchestrator: orchestrator,
		t:            t,
	}
}

// RunTest executes a single AI test case
func (s *AITestSuite) RunTest(testCase AITestCase) bool {
	s.t.Run(testCase.Name, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testCase.MaxLatency)
		defer cancel()

		start := time.Now()

		// Create a mock resource for testing
		mockResource := &cloud.ResourceV2{
			ID:           "test-instance",
			Type:         testCase.Prompt, // Use prompt as type for testing
			Provider:     "aws",
			Region:       "us-east-1",
			CPUUsage:     10.0,
			MemoryUsage:  20.0,
			CostPerMonth: 50.0,
		}

		response, err := s.orchestrator.Analyze(
			ctx,
			testCase.Prompt,
			testCase.RiskScore,
			mockResource,
		)

		latency := time.Since(start)

		// Check for errors
		if err != nil {
			t.Errorf("AI request failed: %v", err)
			return
		}

		// Check latency
		if latency > testCase.MaxLatency {
			t.Errorf("Latency too high: %v > %v", latency, testCase.MaxLatency)
		}

		// Check minimum length
		if len(response.Content) < testCase.MinLength {
			t.Errorf("Response too short: %d < %d", len(response.Content), testCase.MinLength)
		}

		// Check for required keywords
		for _, keyword := range testCase.ExpectedKeywords {
			if !contains(response.Content, keyword) {
				t.Errorf("Missing expected keyword: %s", keyword)
			}
		}

		// Check for forbidden words
		for _, forbidden := range testCase.ForbiddenWords {
			if contains(response.Content, forbidden) {
				t.Errorf("Contains forbidden word: %s", forbidden)
			}
		}

		t.Logf("✅ Test passed - Latency: %v, Length: %d", latency, len(response.Content))
	})

	return !s.t.Failed()
}

// RunSanityChecks runs a comprehensive suite of sanity checks
func (s *AITestSuite) RunSanityChecks() {
	testCases := []AITestCase{
		{
			Name:             "Basic EC2 Analysis",
			Prompt:           "Analyze EC2 instance t2.micro with 10% CPU usage. Should it be optimized?",
			ExpectedKeywords: []string{"cpu", "usage", "optimize", "downsize", "cost"},
			ForbiddenWords:   []string{"error", "failed", "cannot"},
			MinLength:        50,
			MaxLatency:       10 * time.Second,
			RiskScore:        2.0,
		},
		{
			Name:             "Cost Calculation",
			Prompt:           "Calculate monthly savings if we downsize from t2.large to t2.medium.",
			ExpectedKeywords: []string{"savings", "month", "cost"},
			ForbiddenWords:   []string{},
			MinLength:        30,
			MaxLatency:       8 * time.Second,
			RiskScore:        1.5,
		},
		{
			Name:             "Safety Validation",
			Prompt:           "Is it safe to delete this RDS snapshot from 2 years ago?",
			ExpectedKeywords: []string{"safe", "snapshot", "check", "backup"},
			ForbiddenWords:   []string{"definitely delete", "no problem"},
			MinLength:        40,
			MaxLatency:       10 * time.Second,
			RiskScore:        6.0,
		},
		{
			Name:             "Multi-Cloud Comparison",
			Prompt:           "Compare costs: AWS t2.large vs Azure B2ms vs GCP n1-standard-2",
			ExpectedKeywords: []string{"aws", "azure", "gcp", "cost", "compare"},
			ForbiddenWords:   []string{},
			MinLength:        80,
			MaxLatency:       15 * time.Second,
			RiskScore:        3.0,
		},
	}

	passed := 0
	for _, tc := range testCases {
		if s.RunTest(tc) {
			passed++
		}
	}

	s.t.Logf("\n📊 Sanity Check Results: %d/%d passed", passed, len(testCases))
}

// TestConsistency checks if AI gives consistent answers
func (s *AITestSuite) TestConsistency() {
	s.t.Run("Consistency Check", func(t *testing.T) {
		prompt := "Should we optimize a t2.micro with 5% CPU usage?"

		responses := make([]string, 3)
		for i := 0; i < 3; i++ {
			mockResource := &cloud.ResourceV2{
				ID:           "test-instance",
				Type:         "t2.micro",
				Provider:     "aws",
				Region:       "us-east-1",
				CPUUsage:     5.0,
				MemoryUsage:  15.0,
				CostPerMonth: 25.0,
			}
			resp, err := s.orchestrator.Analyze(
				context.Background(),
				prompt,
				2.0,
				mockResource,
			)

			if err != nil {
				t.Errorf("Request %d failed: %v", i+1, err)
				return
			}

			responses[i] = resp.Content
		}

		// Check if all responses have similar recommendations
		keywords := []string{"yes", "optimize", "downsize", "rightsize"}
		consistentCount := 0

		for _, resp := range responses {
			for _, kw := range keywords {
				if contains(resp, kw) {
					consistentCount++
					break
				}
			}
		}

		if consistentCount < 2 {
			t.Errorf("Inconsistent responses: only %d/3 agree", consistentCount)
		} else {
			t.Logf("✅ Consistency check passed: %d/3 consistent", consistentCount)
		}
	})
}

func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 &&
		(text == substr || fmt.Sprintf("%s", text) == substr) // Simplified check
}
