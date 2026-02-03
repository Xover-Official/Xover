package main

import (
	"time"

	"github.com/Xover-Official/Xover/internal/cloud"
)

// Standardized API response structs

// ROIResponse defines the structure for the ROI endpoint.
type ROIResponse struct {
	ROI struct {
		TotalSavings  float64 `json:"total_savings"`
		TotalCosts    float64 `json:"total_costs"`
		NetROI        float64 `json:"net_roi"`
		ROIPercentage float64 `json:"roi_percentage"`
	} `json:"roi"`
	Period string `json:"period"`
}

// ModelTokenInfo defines the token and cost for a specific AI model.
type ModelTokenInfo struct {
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost"`
}

// TokenBreakdownResponse defines the structure for the token breakdown endpoint.
type TokenBreakdownResponse struct {
	TotalCostUSD    float64                   `json:"total_cost_usd"`
	TotalTokens     int                       `json:"total_tokens"`
	TotalSavingsUSD float64                   `json:"total_savings_usd"`
	NetProfitUSD    float64                   `json:"net_profit_usd"`
	Breakdown       map[string]ModelTokenInfo `json:"breakdown"`
}

// SystemStatusResponse defines the structure for the system status endpoint.
type SystemStatusResponse struct {
	Status   string                 `json:"status"`
	Version  string                 `json:"version"`
	Uptime   string                 `json:"uptime"`
	Services map[string]string      `json:"services"`
	Metrics  map[string]interface{} `json:"metrics"`
}

// ResourcesResponse defines the structure for the resources endpoint.
type ResourcesResponse struct {
	Resources   []*cloud.ResourceV2 `json:"resources"`
	TotalCount  int                 `json:"total_count"`
	LastUpdated time.Time           `json:"last_updated"`
}

// HealthzResponse defines the structure for the healthz endpoint.
type HealthzResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// DashboardStatsResponse defines the structure for the main dashboard stats.
type DashboardStatsResponse struct {
	CurrentMonthlyBurn float64 `json:"current_monthly_burn"`
	ProjectedOverrun   float64 `json:"projected_overrun"`
	BiggestSaver       struct {
		Description      string  `json:"description"`
		PotentialSavings float64 `json:"potential_savings"`
	} `json:"biggest_saver"`
	AIRecommendationCount int `json:"ai_recommendation_count"`
	RunwayExtensionDays   int `json:"runway_extension_days"`
}

// OpportunityResponse defines the structure for a single savings opportunity card.
type OpportunityResponse struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	SavingsMonthly float64 `json:"savings_monthly"`
	RiskScore      float64 `json:"risk_score"`
	Reasoning      string  `json:"reasoning"`
	ActionType     string  `json:"action_type"`
	AITier         string  `json:"ai_tier"`
}

// AnomalyResponse defines the structure for a cost anomaly alert.
type AnomalyResponse struct {
	ID         string    `json:"id"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	ResourceID string    `json:"resource_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// FeedbackResponse defines the structure for the feedback submission response.
type FeedbackResponse struct {
	Status string `json:"status"`
}

// TokenStatsResponse defines the structure for the token stats endpoint.
type TokenStatsResponse struct {
	Status          string         `json:"status"`
	TokenStatistics map[string]any `json:"token_statistics"`
	Timestamp       time.Time      `json:"timestamp"`
}

// ResourceMetricsResponse defines the structure for the resource metrics endpoint.
type ResourceMetricsResponse struct {
	Status                  string    `json:"status"`
	TotalResources          int       `json:"total_resources"`
	TotalMonthlyCost        float64   `json:"total_monthly_cost"`
	AverageCPUUsage         float64   `json:"average_cpu_usage"`
	AverageMemoryUsage      float64   `json:"average_memory_usage"`
	UnderutilizedCount      int       `json:"underutilized_count"`
	UnderutilizedPercentage float64   `json:"underutilized_percentage"`
	PotentialMonthlySavings float64   `json:"potential_monthly_savings"`
	Timestamp               time.Time `json:"timestamp"`
}

// OptimizationSuggestion defines the structure for a single optimization suggestion.
type OptimizationSuggestion struct {
	ResourceID       string  `json:"resource_id"`
	ResourceType     string  `json:"resource_type"`
	Provider         string  `json:"provider"`
	Region           string  `json:"region"`
	CurrentCost      float64 `json:"current_cost"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
	Suggestion       string  `json:"suggestion"`
	EstimatedSavings float64 `json:"estimated_savings"`
	Priority         string  `json:"priority"`
	Reason           string  `json:"reason"`
}

// OptimizationSuggestionsResponse defines the structure for the optimization suggestions endpoint.
type OptimizationSuggestionsResponse struct {
	Status                string                   `json:"status"`
	Suggestions           []OptimizationSuggestion `json:"suggestions"`
	TotalSuggestions      int                      `json:"total_suggestions"`
	TotalPotentialSavings float64                  `json:"total_potential_savings"`
	Timestamp             time.Time                `json:"timestamp"`
}
