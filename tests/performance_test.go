package tests

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"testing"

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

func NewPerformanceTestSuite() *PerformanceTestSuite {
    tracker := analytics.NewTokenTracker("")

    config := &ai.Config{
        GeminiAPIKey: "test-key",
        ClaudeAPIKey: "test-key",
        GPT5APIKey:   "test-key",
    }

    orchestrator, err := ai.NewUnifiedOrchestrator(config, tracker, slog.Default())
    if err != nil {
        panic(fmt.Errorf("failed to create orchestrator: %w", err))
    }

    adapter, err := aws.New(context.Background(), cloud.CloudConfig{
        Provider: "aws",
        Region:   "us-east-1",
    })
    if err != nil {
        panic(fmt.Errorf("failed to create AWS adapter: %w", err))
    }

    return &PerformanceTestSuite{
        orchestrator: orchestrator,
        adapter:      adapter,
        tracker:      tracker,
    }
}

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

func generateRandomPrompt() string {
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
    }
    return prompts[rand.Intn(len(prompts))]
}
