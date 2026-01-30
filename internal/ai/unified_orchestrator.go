package ai

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
)

// UnifiedOrchestrator manages AI calls through the factory with caching and retries
type UnifiedOrchestrator struct {
	factory      *AIClientFactory
	tokenTracker *analytics.TokenTracker
	cache        *RedisCache
	logger       *slog.Logger
}

// NewUnifiedOrchestrator creates a new orchestrator with the given configuration
func NewUnifiedOrchestrator(config *Config, tokenTracker *analytics.TokenTracker, logger *slog.Logger) (*UnifiedOrchestrator, error) {
	factory, err := NewAIClientFactory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client factory: %w", err)
	}

	var cache *RedisCache
	if config.CacheEnabled && config.CacheAddr != "" {
		cache, err = NewRedisCache(config.CacheAddr, "", 0, time.Hour)
		if err != nil {
			logger.Info("Redis cache unavailable", "error", err)
		} else {
			logger.Info("Redis cache enabled")
		}
	}

	return &UnifiedOrchestrator{
		factory:      factory,
		tokenTracker: tokenTracker,
		cache:        cache,
		logger:       logger,
	}, nil
}

// Analyze routes request to appropriate AI tier based on risk score
func (o *UnifiedOrchestrator) Analyze(ctx context.Context, prompt string, riskScore float64, resource *cloud.ResourceV2) (*AIResponse, error) {
	// Check cache first
	if o.cache != nil {
		cached, err := o.cache.Get(ctx, prompt)
		if err == nil && cached != nil {
			o.logger.Info("Cache HIT")
			return cached.Response, nil
		}
	}

	// Get appropriate client for risk level
	client := o.factory.GetClientForRisk(riskScore)

	o.logger.Info("Routing to AI client", "risk_score", riskScore, "client_type", fmt.Sprintf("%T", client))

	// Create request
	request := AIRequest{
		Prompt:       prompt,
		ResourceType: resource.Type,
		RiskScore:    riskScore,
		MaxTokens:    1000,
		Temperature:  0.3,
		Metadata: map[string]interface{}{
			"resource_id":    resource.ID,
			"provider":       resource.Provider,
			"region":         resource.Region,
			"cpu_usage":      resource.CPUUsage,
			"memory_usage":   resource.MemoryUsage,
			"cost_per_month": resource.CostPerMonth,
		},
	}

	// Analyze with retry logic
	response, err := o.analyzeWithRetry(client, request, 3)
	if err != nil {
		o.logger.Error("AI analysis failed", "error", err)
		return nil, err
	}

	// Track usage
	if o.tokenTracker != nil {
		o.tokenTracker.RecordUsage(response.Model, response.TokensUsed)
	}

	// Cache the response
	if o.cache != nil {
		if err := o.cache.Set(ctx, prompt, response); err != nil {
			o.logger.Warn("Failed to cache response", "error", err)
		}
	}

	return response, nil
}

// analyzeWithRetry implements retry logic for AI calls
func (o *UnifiedOrchestrator) analyzeWithRetry(client AIClient, request AIRequest, maxRetries int) (*AIResponse, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			o.logger.Info("Retrying AI analysis", "attempt", attempt+1, "max_retries", maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		response, err := client.Analyze(request)
		if err == nil {
			return response, nil
		}

		lastErr = err
		o.logger.Warn("AI analysis attempt failed", "attempt", attempt+1, "error", err)
	}

	return nil, fmt.Errorf("AI analysis failed after %d attempts: %w", maxRetries, lastErr)
}

// GetFactory returns the underlying AI client factory for advanced usage
func (o *UnifiedOrchestrator) GetFactory() *AIClientFactory {
	return o.factory
}

// Close cleans up resources
func (o *UnifiedOrchestrator) Close() error {
	if o.cache != nil {
		return o.cache.Close()
	}
	return nil
}
