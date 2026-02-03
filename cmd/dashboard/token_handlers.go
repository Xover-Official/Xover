package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Xover-Official/Xover/internal/cloud"
)

// Additional handlers that aren't in main.go

func (s *server) handleTokenStats(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.tracker == nil {
		respondWithError(w, http.StatusInternalServerError, "Token tracker not initialized")
		return
	}

	stats := s.tracker.GetStats()
	resp := TokenStatsResponse{
		Status:          "success",
		TokenStatistics: stats,
		Timestamp:       time.Now(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *server) handleResourceMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	s.metricsCache.RLock()
	metrics := s.metricsCache.metrics
	s.metricsCache.RUnlock()

	if metrics == nil {
		respondWithError(w, http.StatusServiceUnavailable, "Metrics cache is not populated yet. Please try again in a moment.")
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

func (s *server) handleOptimizationSuggestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.adapter == nil {
		respondWithError(w, http.StatusInternalServerError, "System not initialized")
		return
	}

	s.suggestionsCache.RLock()
	allSuggestions := s.suggestionsCache.suggestions
	s.suggestionsCache.RUnlock()

	if allSuggestions == nil {
		respondWithError(w, http.StatusServiceUnavailable, "Suggestions cache is not populated yet. Please try again in a moment.")
		return
	}

	// Filter cached suggestions based on query parameters
	resourceType := r.URL.Query().Get("type")
	minSavings := r.URL.Query().Get("min_savings")

	filteredSuggestions := make([]OptimizationSuggestion, 0)
	for _, suggestion := range allSuggestions.Suggestions {
		if resourceType != "" && suggestion.ResourceType != resourceType {
			continue
		}

		if minSavings != "" {
			if minVal, err := parseFloat(minSavings); err == nil && suggestion.EstimatedSavings < minVal {
				continue
			}
		}
		filteredSuggestions = append(filteredSuggestions, suggestion)
	}

	// Re-calculate total savings for the filtered list
	finalResponse := OptimizationSuggestionsResponse{
		Status:                "success",
		Suggestions:           filteredSuggestions,
		TotalSuggestions:      len(filteredSuggestions),
		TotalPotentialSavings: calculateTotalSavings(filteredSuggestions),
		Timestamp:             time.Now(),
	}

	json.NewEncoder(w).Encode(finalResponse)
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// generateSuggestionForResource is a helper for the caching worker.
func generateSuggestionForResource(res *cloud.ResourceV2) *OptimizationSuggestion {
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
		}
	}
	return nil
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

func calculateTotalSavings(suggestions []OptimizationSuggestion) float64 {
	total := 0.0
	for _, suggestion := range suggestions {
		total += suggestion.EstimatedSavings
	}
	return total
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
