package ai

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
)

// UnifiedOrchestrator manages AI calls through a factory, adding caching, retries, and fallbacks.
type UnifiedOrchestrator struct {
	factory      *AIClientFactory
	tokenTracker *analytics.TokenTracker
	cache        *RedisCache
	logger       *slog.Logger
}

// NewUnifiedOrchestrator creates a production-ready orchestrator.
func NewUnifiedOrchestrator(cfg *Config, tokenTracker *analytics.TokenTracker, logger *slog.Logger) (*UnifiedOrchestrator, error) {
	factory, err := NewAIClientFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client factory: %w", err)
	}

	var cache *RedisCache
	if cfg.CacheEnabled && cfg.CacheAddr != "" {
		cache, err = NewRedisCache(cfg.CacheAddr, "", 0, 1*time.Hour)
		if err != nil {
			logger.Warn("Redis cache unavailable, continuing without cache", "error", err)
		} else {
			logger.Info("Redis cache for AI responses enabled")
		}
	}

	return &UnifiedOrchestrator{
		factory:      factory,
		tokenTracker: tokenTracker,
		cache:        cache,
		logger:       logger,
	}, nil
}

// Analyze routes a request to the appropriate AI tier with caching, retries, and fallbacks.
func (o *UnifiedOrchestrator) Analyze(ctx context.Context, prompt string, riskScore float64, resource *cloud.ResourceV2) (*AIResponse, error) {
	// 1. Check cache
	if o.cache != nil {
		cached, err := o.cache.Get(ctx, prompt)
		if err == nil && cached != nil {
			o.logger.Info("AI response cache HIT", "model", cached.Response.Model)
			return cached.Response, nil
		}
	}

	// 2. Get appropriate client for the risk level
	client := o.factory.GetClientForRisk(riskScore)
	o.logger.Info("Routing to AI tier", "tier", client.GetTier(), "model", client.GetModel(), "risk", riskScore)

	request := AIRequest{
		Context:      ctx,
		Prompt:       prompt,
		ResourceType: resource.Type,
		RiskScore:    riskScore,
		MaxTokens:    1000,
		Temperature:  0.3,
	}

	// 3. Attempt analysis with retry logic
	response, err := o.analyzeWithRetry(client, request, 3)
	if err != nil {
		o.logger.Error("AI analysis failed after retries, attempting fallback", "tier", client.GetTier(), "error", err)
		return o.fallbackAnalyze(ctx, request, client.GetTier())
	}

	// 4. Track usage and savings
	if o.tokenTracker != nil {
		o.tokenTracker.RecordUsage(response.Model, response.TokensUsed)
		// In a real scenario, savings would be calculated based on the response content.
		// For now, we'll use the resource's estimated savings if available.
		o.tokenTracker.RecordSavings(resource.EstimatedSavings)
	}

	// 5. Cache the successful response
	if o.cache != nil {
		if err := o.cache.Set(ctx, prompt, response); err != nil {
			o.logger.Warn("Failed to cache AI response", "error", err)
		}
	}

	o.logger.Info("AI analysis successful",
		"model", response.Model,
		"tokens", response.TokensUsed,
		"cost_usd", response.CostUSD,
		"latency_ms", response.Latency.Milliseconds())

	return response, nil
}

// analyzeWithRetry implements exponential backoff retry logic.
func (o *UnifiedOrchestrator) analyzeWithRetry(client AIClient, request AIRequest, maxRetries int) (*AIResponse, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			o.logger.Info("Retrying AI call", "attempt", attempt+1, "max_attempts", maxRetries, "backoff", backoff)
			time.Sleep(backoff)
		}

		response, err := client.Analyze(request)
		if err == nil {
			return response, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

// fallbackAnalyze attempts to use a cheaper, more reliable model if the primary one fails.
func (o *UnifiedOrchestrator) fallbackAnalyze(ctx context.Context, request AIRequest, failedTier int) (*AIResponse, error) {
	if failedTier <= 1 {
		return nil, fmt.Errorf("tier 1 analysis failed, no lower tier to fall back to")
	}

	// Fallback to the most reliable, cheapest client (Tier 1)
	fallbackClient := o.factory.GetClientForRisk(1.0)
	o.logger.Info("Attempting fallback analysis", "fallback_tier", fallbackClient.GetTier(), "model", fallbackClient.GetModel())

	response, err := fallbackClient.Analyze(request)
	if err != nil {
		return nil, fmt.Errorf("fallback analysis also failed: %w", err)
	}

	// Track usage for the fallback model
	if o.tokenTracker != nil {
		o.tokenTracker.RecordUsage(response.Model, response.TokensUsed)
	}

	return response, nil
}

// Close gracefully shuts down the orchestrator's dependencies.
func (o *UnifiedOrchestrator) Close() error {
	if o.cache != nil {
		return o.cache.Close()
	}

	return nil
}
