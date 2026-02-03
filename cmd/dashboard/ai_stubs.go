package main

import (
	"context"

	"github.com/Xover-Official/Xover/internal/cloud"
)

// AIOrchestrator defines the interface required by the dashboard for generating suggestions.
// This allows for mocking the AI backend during testing.
type AIOrchestrator interface {
	GenerateOptimizationSuggestion(ctx context.Context, res *cloud.ResourceV2) (*OptimizationSuggestion, error)
}

// MockOrchestrator is a mock implementation of AIOrchestrator for use in environments
// where the full AI backend is not available. It uses the old hardcoded logic.
type MockOrchestrator struct{}

// NewMockOrchestrator creates a new mock orchestrator.
func NewMockOrchestrator() *MockOrchestrator {
	return &MockOrchestrator{}
}

func (m *MockOrchestrator) GenerateOptimizationSuggestion(ctx context.Context, res *cloud.ResourceV2) (*OptimizationSuggestion, error) {
	// This mock re-implements the simplified logic previously in token_handlers.go
	// A real implementation would call different AI models based on complexity.
	var suggestion string
	var estimatedSavings float64
	var priority string

	if res.CPUUsage < 20 && res.MemoryUsage < 30 {
		suggestion = "resize_down"
		estimatedSavings = res.CostPerMonth * 0.5 // 50% savings
		priority = "high"
	} else if res.CPUUsage < 40 && res.MemoryUsage < 50 {
		suggestion = "rightsize"
		estimatedSavings = res.CostPerMonth * 0.25 // 25% savings
		priority = "medium"
	} else if res.State == "stopped" {
		suggestion = "terminate_if_unused"
		estimatedSavings = res.CostPerMonth
		priority = "high"
	}

	if suggestion != "" {
		return &OptimizationSuggestion{
			ResourceID: res.ID, ResourceType: res.Type, Provider: res.Provider,
			Region: res.Region, CurrentCost: res.CostPerMonth, CPUUsage: res.CPUUsage,
			MemoryUsage: res.MemoryUsage, Suggestion: suggestion, EstimatedSavings: estimatedSavings,
			Priority: priority, Reason: generateOptimizationReason(res.CPUUsage, res.MemoryUsage, res.State),
		}, nil
	}
	return nil, nil
}

func generateOptimizationReason(cpuUsage, memoryUsage float64, state string) string {
	if state == "stopped" {
		return "Resource is stopped and incurring costs"
	}
	if cpuUsage < 20 && memoryUsage < 30 {
		return "Very low utilization - consider significant downsizing"
	}
	if cpuUsage < 40 && memoryUsage < 50 {
		return "Low utilization - consider rightsizing"
	}
	return "Resource appears to be appropriately sized"
}
