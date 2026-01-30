package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/project-atlas/atlas/internal/analytics"
)

// UnifiedClientFactory creates AI clients using OpenRouter for most models
type UnifiedClientFactory struct {
	openRouter   *OpenRouterClient
	devinClient  *DevinClient
	modelMapping map[int]string // tier -> OpenRouter model name
}

// NewUnifiedClientFactory creates a factory using OpenRouter
func NewUnifiedClientFactory(openrouterKey, devinKey string) *UnifiedClientFactory {
	return &UnifiedClientFactory{
		openRouter:  NewOpenRouterClient(openrouterKey),
		devinClient: NewDevinClient(devinKey),
		modelMapping: map[int]string{
			1: "google/gemini-2.0-flash-exp", // Tier 1: Sentinel (Free!)
			2: "google/gemini-pro",           // Tier 2: Strategist
			3: "anthropic/claude-3.5-sonnet", // Tier 3: Arbiter
			4: "openai/gpt-4o-mini",          // Tier 4: Reasoning
			5: "devin",                       // Tier 5: Oracle (direct API)
		},
	}
}

// GetClientForRisk returns appropriate model based on risk score
func (f *UnifiedClientFactory) GetClientForRisk(risk float64) (string, bool) {
	tier := f.riskToTier(risk)
	model := f.modelMapping[tier]
	useDevin := tier == 5

	return model, useDevin
}

func (f *UnifiedClientFactory) riskToTier(risk float64) int {
	switch {
	case risk < 3.0:
		return 1
	case risk < 5.0:
		return 2
	case risk < 7.0:
		return 3
	case risk < 9.0:
		return 4
	default:
		return 5
	}
}

// UnifiedOrchestrator manages AI calls through OpenRouter + Devin
type UnifiedOrchestrator struct {
	factory      *UnifiedClientFactory
	tokenTracker *analytics.TokenTracker
	cache        *RedisCache
	logger       *log.Logger
}

// NewUnifiedOrchestrator creates orchestrator with OpenRouter
func NewUnifiedOrchestrator(openrouterKey, devinKey, cacheAddr string, tokenTracker *analytics.TokenTracker, logger *log.Logger) (*UnifiedOrchestrator, error) {
	factory := NewUnifiedClientFactory(openrouterKey, devinKey)

	var cache *RedisCache
	var err error
	if cacheAddr != "" {
		cache, err = NewRedisCache(cacheAddr, "", 0, time.Hour)
		if err != nil {
			logger.Printf("⚠️  Redis cache unavailable: %v (continuing without cache)", err)
		} else {
			logger.Printf("✅ Redis cache enabled")
		}
	}

	return &UnifiedOrchestrator{
		factory:      factory,
		tokenTracker: tokenTracker,
		cache:        cache,
		logger:       logger,
	}, nil
}

// Analyze routes request to appropriate AI tier
func (o *UnifiedOrchestrator) Analyze(ctx context.Context, prompt string, riskScore float64, resourceType string, estimatedSavings float64) (*AIResponse, error) {
	// Check cache
	if o.cache != nil {
		cached, err := o.cache.Get(ctx, prompt)
		if err == nil && cached != nil {
			o.logger.Printf("💾 Cache HIT - saved API call!")
			return cached.Response, nil
		}
	}

	// Get model for risk level
	model, useDevin := o.factory.GetClientForRisk(riskScore)
	tier := o.factory.riskToTier(riskScore)

	o.logger.Printf("🤖 Routing to Tier %d (%s), risk: %.1f", tier, model, riskScore)

	request := AIRequest{
		Context:      ctx,
		Prompt:       prompt,
		ResourceType: resourceType,
		RiskScore:    riskScore,
		MaxTokens:    1000,
		Temperature:  0.3,
	}

	var response *AIResponse
	var err error

	// Use Devin directly for Tier 5, OpenRouter for everything else
	if useDevin {
		response, err = o.factory.devinClient.Analyze(request)
	} else {
		response, err = o.analyzeWithRetry(request, model, 3)
	}

	if err != nil {
		o.logger.Printf("❌ Tier %d failed: %v, trying fallback", tier, err)
		return o.fallbackAnalyze(ctx, request, tier, estimatedSavings)
	}

	// Track usage
	if o.tokenTracker != nil {
		o.tokenTracker.RecordUsage(response.Model, response.TokensUsed)
		o.tokenTracker.RecordSavings(estimatedSavings)
	}

	// Cache response
	if o.cache != nil {
		o.cache.Set(ctx, prompt, response)
	}

	o.logger.Printf("✅ Success - Tokens: %d, Cost: $%.4f, Latency: %v",
		response.TokensUsed, response.CostUSD, response.Latency)

	return response, nil
}

// analyzeWithRetry implements retry logic
func (o *UnifiedOrchestrator) analyzeWithRetry(request AIRequest, model string, maxRetries int) (*AIResponse, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			o.logger.Printf("🔄 Retry %d/%d after %v", attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
		}

		response, err := o.factory.openRouter.Analyze(request, model)
		if err == nil {
			return response, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

// fallbackAnalyze falls back to lower tiers
func (o *UnifiedOrchestrator) fallbackAnalyze(_ context.Context, request AIRequest, failedTier int, estimatedSavings float64) (*AIResponse, error) {
	// Try progressively cheaper models
	fallbackModels := []struct {
		tier  int
		model string
	}{
		{1, "google/gemini-2.0-flash-exp"},
		{2, "google/gemini-pro"},
		{3, "anthropic/claude-3.5-sonnet"},
	}

	for _, fb := range fallbackModels {
		if fb.tier >= failedTier {
			continue
		}

		o.logger.Printf("🔄 Fallback to Tier %d (%s)", fb.tier, fb.model)

		response, err := o.factory.openRouter.Analyze(request, fb.model)
		if err == nil {
			if o.tokenTracker != nil {
				o.tokenTracker.RecordUsage(response.Model, response.TokensUsed)
				o.tokenTracker.RecordSavings(estimatedSavings)
			}
			return response, nil
		}

		o.logger.Printf("❌ Fallback Tier %d failed: %v", fb.tier, err)
	}

	return nil, fmt.Errorf("all fallback tiers exhausted")
}

// HealthCheckAll checks all tiers
func (o *UnifiedOrchestrator) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for tier, model := range o.factory.modelMapping {
		tierName := fmt.Sprintf("Tier %d (%s)", tier, model)

		if tier == 5 {
			// Devin direct API
			err := o.factory.devinClient.HealthCheck(ctx)
			results[tierName] = err
		} else {
			// OpenRouter
			err := o.factory.openRouter.HealthCheck(ctx, model)
			results[tierName] = err
		}

		if results[tierName] == nil {
			o.logger.Printf("🟢 %s: HEALTHY", tierName)
		} else {
			o.logger.Printf("🔴 %s: %v", tierName, results[tierName])
		}
	}

	return results
}

// Close cleans up resources
func (o *UnifiedOrchestrator) Close() error {
	if o.cache != nil {
		return o.cache.Close()
	}
	return nil
}
