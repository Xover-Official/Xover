package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "success",
		"token_statistics": stats,
		"timestamp":        time.Now(),
	})
}

func (s *server) handleResourceMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.adapter == nil {
		respondWithError(w, http.StatusInternalServerError, "Cloud adapter not initialized")
		return
	}

	// Fetch resources and calculate metrics
	resources, err := s.adapter.FetchResources(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch resources: %v", err))
		return
	}

	// Calculate metrics
	totalResources := len(resources)
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":                    "success",
		"total_resources":           totalResources,
		"total_monthly_cost":        totalCost,
		"average_cpu_usage":         avgCPU,
		"average_memory_usage":      avgMemory,
		"underutilized_count":       underutilizedCount,
		"underutilized_percentage":  underutilizedPercent,
		"potential_monthly_savings": totalCost * (underutilizedPercent / 100),
		"timestamp":                 time.Now(),
	})
}

func (s *server) handleOptimizationSuggestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.adapter == nil {
		respondWithError(w, http.StatusInternalServerError, "Cloud adapter not initialized")
		return
	}

	// Get query parameters
	resourceType := r.URL.Query().Get("type")
	minSavings := r.URL.Query().Get("min_savings")

	// Fetch resources
	resources, err := s.adapter.FetchResources(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch resources: %v", err))
		return
	}

	suggestions := make([]map[string]interface{}, 0)

	for _, res := range resources {
		// Skip if resource type filter is set and doesn't match
		if resourceType != "" && res.Type != resourceType {
			continue
		}

		// Generate optimization suggestions based on utilization
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

		// Apply minimum savings filter
		if minSavings != "" {
			if minVal, err := parseFloat(minSavings); err == nil && estimatedSavings < minVal {
				continue
			}
		}

		if suggestion != "" {
			suggestions = append(suggestions, map[string]interface{}{
				"resource_id":       res.ID,
				"resource_type":     res.Type,
				"provider":          res.Provider,
				"region":            res.Region,
				"current_cost":      res.CostPerMonth,
				"cpu_usage":         res.CPUUsage,
				"memory_usage":      res.MemoryUsage,
				"suggestion":        suggestion,
				"estimated_savings": estimatedSavings,
				"priority":          priority,
				"reason":            generateOptimizationReason(res.CPUUsage, res.MemoryUsage, res.State),
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":                  "success",
		"suggestions":             suggestions,
		"total_suggestions":       len(suggestions),
		"total_potential_savings": calculateTotalSavings(suggestions),
		"timestamp":               time.Now(),
	})
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, fmt.Sprintf(`{"error": "%s"}`, message), code)
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

func calculateTotalSavings(suggestions []map[string]interface{}) float64 {
	total := 0.0
	for _, suggestion := range suggestions {
		if savings, ok := suggestion["estimated_savings"].(float64); ok {
			total += savings
		}
	}
	return total
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
