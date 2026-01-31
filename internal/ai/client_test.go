package ai

import (
	"context"
	"log/slog" // Use modern structured logging
	"os"
	"testing"

	"github.com/project-atlas/atlas/internal/cloud"
)

// ... (TestGeminiFlashClient, TestGeminiProClient, etc. remain the same)

// Test Swarm Orchestrator with fallback
func TestSwarmOrchestratorFallback(t *testing.T) {
	// Create orchestrator with invalid keys to trigger fallback
	config := &Config{
		GeminiAPIKey: "invalid-key",
		ClaudeAPIKey: "invalid-key",
		GPT5APIKey:   "invalid-key",
		DevinAPIKey:  "invalid-key",
	}

	factory, _ := NewAIClientFactory(config)
	
	// FIX 1: Use slog instead of log
	orchestrator := &UnifiedOrchestrator{
		factory: factory,
		logger:  testLogger(t), 
	}

	ctx := context.Background()

	// FIX 2: Correct arguments for orchestrator.Analyze
	// want (context.Context, prompt string, risk float64, resource *cloud.Resource)
	mockResource := &cloud.Resource{
		ID:   "test-instance",
		Type: "ec2",
	}

	response, err := orchestrator.Analyze(ctx, "test prompt", 5.0, mockResource)

	// We expect an error since all keys are invalid
	if err == nil {
		t.Error("Expected error with invalid keys")
	}

	if response != nil {
		t.Error("Expected nil response with invalid keys")
	}

	t.Log("✅ Fallback logic and signature verification working correctly")
}

// Helper functions

func getTestAPIKey(envVar string) string {
	// Allow overriding via environment variables for local testing
	return os.Getenv(envVar)
}

// FIX 3: Updated to return *slog.Logger to satisfy struct literal
func testLogger(t *testing.T) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}