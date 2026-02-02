package tests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/cloud/aws"
	"github.com/project-atlas/atlas/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// EdgeCaseTestSuite tests edge cases and error conditions
type EdgeCaseTestSuite struct {
	orchestrator *ai.UnifiedOrchestrator
	adapter      cloud.CloudAdapter
	tracker      *analytics.TokenTracker
	secManager   *security.SecurityManager
	logger       *zap.Logger
}

func NewEdgeCaseTestSuite(tb testing.TB) *EdgeCaseTestSuite {
	tb.Helper()

	logger := zap.NewNop()
	tracker := analytics.NewTokenTracker("test-token-tracker")

	config := &ai.Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		ClaudeAPIKey: os.Getenv("CLAUDE_API_KEY"),
		CacheEnabled: true,
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(config, tracker, logger)
	if err != nil {
		tb.Logf("Warning: Failed to create orchestrator: %v", err)
		// Continue with nil orchestrator for edge case testing
	}

	ctx := context.Background()
	awsConfig := cloud.CloudConfig{
		Region: "us-east-1",
		DryRun: true, // Use dry run for edge case testing
	}
	adapter, err := aws.New(ctx, awsConfig)
	if err != nil {
		tb.Logf("Warning: Failed to create AWS adapter: %v", err)
		// Continue with nil adapter for edge case testing
	}

	secManager := security.NewSecurityManager(
		"test-jwt-secret-key-for-edge-case-testing-32-chars",
		time.Hour,
		24*time.Hour,
		logger,
	)

	return &EdgeCaseTestSuite{
		orchestrator: orchestrator,
		adapter:      adapter,
		tracker:      tracker,
		secManager:   secManager,
		logger:       logger,
	}
}

// TestEdgeCaseNilInputs tests behavior with nil inputs
func TestEdgeCaseNilInputs(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("NilResourceAnalysis", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		// Test with nil resource
		_, err := suite.orchestrator.Analyze(context.Background(), "test prompt", 5.0, nil)
		assert.Error(t, err, "Should return error for nil resource")
		assert.Contains(t, err.Error(), "resource", "Error should mention resource")
	})

	t.Run("NilContext", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
		}

		// Test with nil context
		_, err := suite.orchestrator.Analyze(nil, "test prompt", 5.0, resource)
		assert.Error(t, err, "Should return error for nil context")
	})

	t.Run("EmptyPrompt", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
		}

		// Test with empty prompt
		_, err := suite.orchestrator.Analyze(context.Background(), "", 5.0, resource)
		assert.Error(t, err, "Should return error for empty prompt")
	})
}

// TestEdgeCaseExtremeValues tests behavior with extreme values
func TestEdgeCaseExtremeValues(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("ExtremeRiskScores", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
		}

		testCases := []float64{-1000, -100, -1, 0, 1000, 10000}

		for _, riskScore := range testCases {
			t.Run(fmt.Sprintf("RiskScore_%.1f", riskScore), func(t *testing.T) {
				_, err := suite.orchestrator.Analyze(context.Background(), "test prompt", riskScore, resource)
				// Should handle extreme values gracefully
				if err != nil {
					t.Logf("Risk score %.1f resulted in error: %v", riskScore, err)
				}
			})
		}
	})

	t.Run("ExtremeResourceValues", func(t *testing.T) {
		if suite.adapter == nil {
			t.Skip("Adapter not available")
		}

		testCases := []struct {
			name     string
			resource *cloud.ResourceV2
		}{
			{
				name: "ZeroValues",
				resource: &cloud.ResourceV2{
					ID:           "zero-resource",
					Type:         "test-type",
					CPUUsage:     0,
					MemoryUsage:  0,
					CostPerMonth: 0,
				},
			},
			{
				name: "NegativeValues",
				resource: &cloud.ResourceV2{
					ID:           "negative-resource",
					Type:         "test-type",
					CPUUsage:     -10,
					MemoryUsage:  -20,
					CostPerMonth: -100,
				},
			},
			{
				name: "MaximumValues",
				resource: &cloud.ResourceV2{
					ID:           "max-resource",
					Type:         "test-type",
					CPUUsage:     1000,
					MemoryUsage:  1000,
					CostPerMonth: 1000000,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test resource processing with extreme values
				_, err := suite.adapter.ApplyOptimization(context.Background(), tc.resource, "test-action")
				// Should handle extreme values gracefully
				if err != nil {
					t.Logf("Extreme resource %s resulted in error: %v", tc.name, err)
				}
			})
		}
	})
}

// TestEdgeCaseLongInputs tests behavior with very long inputs
func TestEdgeCaseLongInputs(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("VeryLongPrompt", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		// Generate a very long prompt (1MB)
		longPrompt := string(make([]byte, 1024*1024))
		for i := range longPrompt {
			longPrompt = longPrompt[:i] + "a" + longPrompt[i+1:]
		}

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
		}

		// Test with very long prompt
		_, err := suite.orchestrator.Analyze(context.Background(), longPrompt, 5.0, resource)
		// Should handle long inputs gracefully or return appropriate error
		if err != nil {
			t.Logf("Long prompt resulted in error: %v", err)
		}
	})

	t.Run("VeryLongResourceID", func(t *testing.T) {
		if suite.adapter == nil {
			t.Skip("Adapter not available")
		}

		// Generate a very long resource ID (10KB)
		longID := string(make([]byte, 10240))
		for i := range longID {
			longID = longID[:i] + "a" + longID[i+1:]
		}

		resource := &cloud.ResourceV2{
			ID:   longID,
			Type: "test-type",
		}

		// Test with very long resource ID
		_, err := suite.adapter.ApplyOptimization(context.Background(), resource, "test-action")
		// Should handle long IDs gracefully or return appropriate error
		if err != nil {
			t.Logf("Long resource ID resulted in error: %v", err)
		}
	})
}

// TestEdgeCaseConcurrentAccess tests concurrent access scenarios
func TestEdgeCaseConcurrentAccess(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("ConcurrentAnalysis", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		const numGoroutines = 100
		const numRequests = 10

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numRequests)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numRequests; j++ {
					resource := &cloud.ResourceV2{
						ID:   fmt.Sprintf("resource-%d-%d", id, j),
						Type: "test-type",
					}

					_, err := suite.orchestrator.Analyze(
						context.Background(),
						fmt.Sprintf("test prompt %d-%d", id, j),
						5.0,
						resource,
					)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d, request %d: %w", id, j, err)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Logf("Concurrent access error: %v", err)
			errorCount++
		}

		// Allow some errors due to rate limiting or API limits
		maxAllowedErrors := numGoroutines * numRequests / 10 // Allow up to 10% error rate
		if errorCount > maxAllowedErrors {
			t.Errorf("Too many errors in concurrent access: %d > %d", errorCount, maxAllowedErrors)
		}
	})
}

// TestEdgeCaseSecurityScenarios tests security-related edge cases
func TestEdgeCaseSecurityScenarios(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("InvalidJWTToken", func(t *testing.T) {
		// Test with invalid JWT tokens
		invalidTokens := []string{
			"",
			"invalid.token.here",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			"malformed_token",
			string(make([]byte, 1000)), // Very long invalid token
		}

		for _, token := range invalidTokens {
			t.Run(fmt.Sprintf("Token_%s", token[:min(len(token), 20)]), func(t *testing.T) {
				_, err := suite.secManager.ValidateToken(token)
				assert.Error(t, err, "Should return error for invalid token")
			})
		}
	})

	t.Run("PasswordEdgeCases", func(t *testing.T) {
		passwordTestCases := []struct {
			name     string
			password string
		}{
			{"EmptyPassword", ""},
			{"VeryLongPassword", string(make([]byte, 10000))},
			{"UnicodePassword", "🔐🔑🔒💻🛡️"},
			{"SQLInjection", "'; DROP TABLE users; --"},
			{"XSSPayload", "<script>alert('xss')</script>"},
		}

		for _, tc := range passwordTestCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test password hashing
				hash, err := suite.secManager.HashPassword(tc.password)
				if err != nil {
					t.Logf("Password hashing failed for %s: %v", tc.name, err)
					return
				}

				// Test password verification
				ok := suite.secManager.CheckPassword(tc.password, hash)
				if !ok {
					t.Logf("Password check failed for %s", tc.name)
				}
			})
		}
	})

	t.Run("TokenGenerationEdgeCases", func(t *testing.T) {
		tokenTestCases := []struct {
			name     string
			userID   string
			username string
			roles    []string
		}{
			{"EmptyUser", "", "", []string{}},
			{"VeryLongUser", string(make([]byte, 1000)), string(make([]byte, 1000)), []string{"admin"}},
			{"ManyRoles", "test-user", "test-user", make([]string, 100)},
			{"SpecialChars", "test@user#123", "user!@#$%", []string{"role@1", "role#2"}},
		}

		for _, tc := range tokenTestCases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := suite.secManager.GenerateTokenPair(tc.userID, tc.username, tc.roles)
				if err != nil {
					t.Logf("Token generation failed for %s: %v", tc.name, err)
				}
			})
		}
	})
}

// TestEdgeCaseResourceScenarios tests resource-related edge cases
func TestEdgeCaseResourceScenarios(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("InvalidResourceTypes", func(t *testing.T) {
		if suite.adapter == nil {
			t.Skip("Adapter not available")
		}

		invalidResources := []*cloud.ResourceV2{
			{
				ID:   "",
				Type: "",
			},
			{
				ID:   "valid-id",
				Type: "",
			},
			{
				ID:   "",
				Type: "valid-type",
			},
			{
				ID:   string(make([]byte, 0)), // Empty
				Type: "test-type",
			},
		}

		for i, resource := range invalidResources {
			t.Run(fmt.Sprintf("InvalidResource_%d", i), func(t *testing.T) {
				_, err := suite.adapter.ApplyOptimization(context.Background(), resource, "test-action")
				// Should handle invalid resources gracefully
				if err != nil {
					t.Logf("Invalid resource %d resulted in error: %v", i, err)
				}
			})
		}
	})

	t.Run("ExtremeResourceTags", func(t *testing.T) {
		if suite.adapter == nil {
			t.Skip("Adapter not available")
		}

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
			Tags: map[string]string{
				"":                         "empty-key",
				"valid-key":                "",
				string(make([]byte, 1000)): "very-long-key",
				"valid-key-duplicate":      string(make([]byte, 1000)),
				"unicode-tag":              "🏷️🔖📋",
				"sql-injection":            "'; DROP TABLE resources; --",
				"new-tag":                  "new-tag-value",
			},
		}

		_, err := suite.adapter.ApplyOptimization(context.Background(), resource, "test-action")
		// Should handle extreme tags gracefully
		if err != nil {
			t.Logf("Extreme tags resulted in error: %v", err)
		}
	})
}

// TestEdgeCaseNetworkScenarios tests network-related edge cases
func TestEdgeCaseNetworkScenarios(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("TimeoutScenarios", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		// Test with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
		}

		_, err := suite.orchestrator.Analyze(ctx, "test prompt", 5.0, resource)
		// Should handle timeout gracefully
		if err != nil {
			t.Logf("Timeout scenario resulted in error: %v", err)
		}
	})

	t.Run("CancelledContext", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		resource := &cloud.ResourceV2{
			ID:   "test-resource",
			Type: "test-type",
		}

		_, err := suite.orchestrator.Analyze(ctx, "test prompt", 5.0, resource)
		assert.Error(t, err, "Should return error for cancelled context")
		assert.Contains(t, err.Error(), "context", "Error should mention context")
	})
}

// TestEdgeCaseMemoryScenarios tests memory-related edge cases
func TestEdgeCaseMemoryScenarios(t *testing.T) {
	suite := NewEdgeCaseTestSuite(t)

	t.Run("LargeResourceSlice", func(t *testing.T) {
		if suite.adapter == nil {
			t.Skip("Adapter not available")
		}

		// Create a very large slice of resources
		resources := make([]*cloud.ResourceV2, 100000)
		for i := range resources {
			resources[i] = &cloud.ResourceV2{
				ID:           fmt.Sprintf("resource-%d", i),
				Type:         "test-type",
				CPUUsage:     float64(i % 100),
				MemoryUsage:  float64(i % 100),
				CostPerMonth: float64(i) * 0.01,
				Tags: map[string]string{
					"index": fmt.Sprintf("%d", i),
				},
			}
		}

		// Test processing large resource slice
		for i, resource := range resources[:100] { // Test first 100 to avoid timeout
			_, err := suite.adapter.ApplyOptimization(context.Background(), resource, "test-action")
			if err != nil {
				t.Logf("Large resource slice processing failed at index %d: %v", i, err)
			}
		}
	})
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BenchmarkEdgeCasePerformance benchmarks edge case scenarios
func BenchmarkEdgeCasePerformance(b *testing.B) {
	suite := NewEdgeCaseTestSuite(b)

	b.Run("TokenGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			suite.secManager.GenerateTokenPair(
				fmt.Sprintf("user-%d", i),
				fmt.Sprintf("username-%d", i),
				[]string{"user"},
			)
		}
	})

	b.Run("PasswordHashing", func(b *testing.B) {
		password := "test-password-123"
		for i := 0; i < b.N; i++ {
			suite.secManager.HashPassword(password)
		}
	})

	b.Run("TokenValidation", func(b *testing.B) {
		// Generate a token once
		token, _, err := suite.secManager.GenerateTokenPair("test-user", "test-user", []string{"user"})
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			suite.secManager.ValidateToken(token)
		}
	})
}
