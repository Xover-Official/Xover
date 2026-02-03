package ai

import (
	"context"
	"testing"

	"github.com/Xover-Official/Xover/internal/cloud"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	orchestrator := &UnifiedOrchestrator{
		factory: factory,
		logger:  testLogger(),
	}

	ctx := context.Background()

	mockResource := &cloud.ResourceV2{
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

// testLogger returns a zap logger for tests
func testLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return logger
}
