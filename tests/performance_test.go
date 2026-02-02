package tests

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/cloud/aws"
)

type PerformanceTestSuite struct {
	orchestrator *ai.UnifiedOrchestrator
	adapter      cloud.CloudAdapter
	tracker      *analytics.TokenTracker
}

func NewPerformanceTestSuite(tb testing.TB) *PerformanceTestSuite {
	tb.Helper()

	tracker := analytics.NewTokenTracker("")

	config := &ai.Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		ClaudeAPIKey: os.Getenv("CLAUDE_API_KEY"),
		GPT5APIKey:   os.Getenv("GPT5_API_KEY"),
	}

	orchestrator, err := ai.NewUnifiedOrchestrator(config, tracker, zap.NewNop())
	if err != nil {
		tb.Fatalf("failed to create orchestrator: %v", err)
	}

	adapter, err := aws.New(context.Background(), cloud.CloudConfig{
		Provider: "aws",
		Region:   "us-east-1",
	})
	if err != nil {
		tb.Fatalf("failed to create AWS adapter: %v", err)
	}

	return &PerformanceTestSuite{
		orchestrator: orchestrator,
		adapter:      adapter,
		tracker:      tracker,
	}
}

func BenchmarkAIOrchestrator(b *testing.B) {
	suite := NewPerformanceTestSuite(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Create a local random source to avoid global lock contention in parallel tests
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))

		for pb.Next() {
			ctx := context.Background()
			prompt := generateRandomPrompt(rng)
			risk := rng.Float64() * 10.0 // Generate risk between 0.0 and 10.0 to hit all tiers

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

func generateRandomPrompt(rng *rand.Rand) string {
	prompts := []string{
		"Optimize this EC2 instance for cost.",
		"Provide a cost analysis for our RDS fleet.",
		"Review S3 lifecycle policies for buckets tagged 'production'.",
		"What are the security implications of opening port 80 on this security group?",
		"Give me a summary of unused EBS volumes in the us-west-2 region.",
		"Can you recommend a right-sizing strategy for our Kubernetes cluster?",
		"Analyze the network traffic patterns for the last 24 hours.",
		"Check for compliance with CIS benchmarks for our Ubuntu servers.",
		"Generate a report of all IAM users with administrative privileges.",
		"Find all publicly accessible S3 buckets and suggest remediation.",
		"CRITICAL: Production database is unresponsive. Diagnose immediately.",
		"EMERGENCY: Detected unauthorized access in root account.",
		"FATAL: Kubernetes cluster lost quorum. Recovery steps needed.",
	}
	return prompts[rng.Intn(len(prompts))]
}
