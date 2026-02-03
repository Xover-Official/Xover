package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/Xover-Official/Xover/internal/analytics"
	"github.com/Xover-Official/Xover/internal/cloud"
	"go.uber.org/zap"
)

// AICache defines the interface for caching AI responses.
// This decouples the orchestrator from the concrete Redis implementation.
type AICache interface {
	Get(ctx context.Context, prompt string) (*CachedResponse, error)
	Set(ctx context.Context, prompt string, response *AIResponse) error
	Close() error
}

// UnifiedOrchestrator manages AI calls through the factory with caching and retries
type UnifiedOrchestrator struct {
	factory      *AIClientFactory
	tokenTracker *analytics.TokenTracker
	cache        AICache
	logger       *zap.Logger
}

// NewUnifiedOrchestrator creates a new orchestrator with the given configuration and zap logger
func NewUnifiedOrchestrator(config *Config, tokenTracker *analytics.TokenTracker, logger *zap.Logger) (*UnifiedOrchestrator, error) {
	factory, err := NewAIClientFactory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client factory: %w", err)
	}

	var cache AICache
	if config.CacheEnabled && config.CacheAddr != "" {
		cache, err = NewRedisCache(config.CacheAddr, "", 0, time.Hour)
		if err != nil {
			logger.Info("Redis cache unavailable", zap.Error(err))
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
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if resource == nil {
		return nil, fmt.Errorf("resource is required")
	}

	// Check cache first
	if o.cache != nil {
		cached, err := o.cache.Get(ctx, prompt)
		if err == nil && cached != nil {
			o.logger.Info("Cache HIT", zap.String("resource_id", resource.ID))
			return cached.Response, nil
		}
	}

	// Get appropriate client for risk level
	client := o.factory.GetClientForRisk(riskScore)

	o.logger.Info("Routing to AI client", zap.Float64("risk_score", riskScore), zap.String("client_type", fmt.Sprintf("%T", client)))

	// Dynamic token allocation based on risk tier
	maxTokens := 1000
	if riskScore >= 7.0 {
		maxTokens = 4000 // High-risk tiers (Arbiter/Reasoning/Oracle) require more context
	}

	// Create request
	request := AIRequest{
		Prompt:       prompt,
		ResourceType: resource.Type,
		RiskScore:    riskScore,
		MaxTokens:    maxTokens,
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
	response, err := o.AnalyzeWithRetry(ctx, client, request, 3)
	if err != nil {
		o.logger.Error("AI analysis failed", zap.Error(err))
		return nil, err
	}

	// Track usage
	if o.tokenTracker != nil {
		o.tokenTracker.RecordUsage(response.Model, response.TokensUsed)
	}

	// Cache the response
	if o.cache != nil {
		if err := o.cache.Set(ctx, prompt, response); err != nil {
			o.logger.Warn("Failed to cache response", zap.Error(err))
		}
	}

	return response, nil
}

// analyzeWithRetry implements retry logic for AI calls
func (o *UnifiedOrchestrator) AnalyzeWithRetry(ctx context.Context, client AIClient, request AIRequest, maxRetries int) (*AIResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			o.logger.Info("Retrying AI analysis", zap.Int("attempt", attempt), zap.Duration("backoff", backoff))

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			case <-time.After(backoff):
				// Continue execution
			}
		}

		response, err := client.Analyze(ctx, request)
		if err == nil {
			return response, nil
		}

		lastErr = err
		o.logger.Warn("AI analysis attempt failed", zap.Int("attempt", attempt), zap.Error(err))

		// Fail fast if context is cancelled
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled during analysis: %w", ctx.Err())
		}
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
