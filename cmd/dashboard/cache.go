package main

import (
	"context"
	"time"

	"github.com/Xover-Official/Xover/internal/cloud"
	"go.uber.org/zap"
)

// startResourceCacheRefresh runs a periodic background job to refresh the resource cache.
func (s *server) startResourceCacheRefresh(ctx context.Context) {
	s.logger.Info("starting resource cache refresh loop")
	ticker := time.NewTicker(5 * time.Minute) // Refresh every 5 minutes
	defer ticker.Stop()

	// Perform an initial refresh immediately on startup
	s.performCacheRefresh(ctx)

	for {
		select {
		case <-ticker.C:
			s.performCacheRefresh(ctx)
		case <-ctx.Done():
			s.logger.Info("stopping resource cache refresh loop")
			return
		}
	}
}

// performCacheRefresh contains the logic for a single refresh operation.
// It fetches resources from the cloud adapter and updates the shared cache.
func (s *server) performCacheRefresh(ctx context.Context) {
	s.resourceCache.refreshMu.Lock()
	defer s.resourceCache.refreshMu.Unlock()

	s.logger.Info("performing resource cache refresh")

	fetchCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	resources, err := s.adapter.FetchResources(fetchCtx)
	if err != nil {
		s.logger.Error("failed to fetch resources for cache", zap.Error(err))
		return // Keep stale data on failure
	}

	s.resourceCache.Lock()
	s.resourceCache.resources = resources
	s.resourceCache.fetchedAt = time.Now()
	s.resourceCache.Unlock()
	s.logger.Info("resource cache updated successfully", zap.Int("count", len(resources)))

	// Now, update derived caches
	s.updateResourceMetricsCache(resources)
	s.updateOptimizationSuggestionsCache(resources)
}

// updateResourceMetricsCache calculates and caches aggregate metrics.
func (s *server) updateResourceMetricsCache(resources []*cloud.ResourceV2) {
	s.logger.Info("performing resource metrics cache refresh")
	totalResources := len(resources)
	if totalResources == 0 {
		return
	}

	totalCost := 0.0
	underutilizedCount := 0
	cpuUsage := 0.0
	memoryUsage := 0.0

	for _, res := range resources {
		totalCost += res.CostPerMonth
		cpuUsage += res.CPUUsage
		memoryUsage += res.MemoryUsage

		if res.CPUUsage < 30 && res.MemoryUsage < 40 {
			underutilizedCount++
		}
	}

	avgCPU := cpuUsage / float64(totalResources)
	avgMemory := memoryUsage / float64(totalResources)
	underutilizedPercent := float64(underutilizedCount) / float64(totalResources) * 100

	metrics := &ResourceMetricsResponse{
		Status:                  "success",
		TotalResources:          totalResources,
		TotalMonthlyCost:        totalCost,
		AverageCPUUsage:         avgCPU,
		AverageMemoryUsage:      avgMemory,
		UnderutilizedCount:      underutilizedCount,
		UnderutilizedPercentage: underutilizedPercent,
		PotentialMonthlySavings: totalCost * (underutilizedPercent / 100),
		Timestamp:               time.Now(),
	}

	s.metricsCache.Lock()
	s.metricsCache.metrics = metrics
	s.metricsCache.fetchedAt = time.Now()
	s.metricsCache.Unlock()
	s.logger.Info("resource metrics cache updated successfully")
}

// updateOptimizationSuggestionsCache generates and caches all possible optimization suggestions.
func (s *server) updateOptimizationSuggestionsCache(resources []*cloud.ResourceV2) {
	s.logger.Info("performing optimization suggestions cache refresh")
	suggestions := make([]OptimizationSuggestion, 0)

	for _, res := range resources {
		// This logic is simplified; in a real scenario, it would call the AI orchestrator.
		if suggestion := generateSuggestionForResource(res); suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	response := &OptimizationSuggestionsResponse{
		Status:                "success",
		Suggestions:           suggestions,
		TotalSuggestions:      len(suggestions),
		TotalPotentialSavings: calculateTotalSavings(suggestions),
		Timestamp:             time.Now(),
	}

	s.suggestionsCache.Lock()
	s.suggestionsCache.suggestions = response
	s.suggestionsCache.fetchedAt = time.Now()
	s.suggestionsCache.Unlock()
	s.logger.Info("optimization suggestions cache updated successfully", zap.Int("suggestions_found", len(suggestions)))
}
