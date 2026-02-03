package main

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// --- API Handlers ---

func (s *server) handleROI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := ROIResponse{
		Period: "30 days",
	}
	resp.ROI.TotalSavings = 1250.50
	resp.ROI.TotalCosts = 342.75
	resp.ROI.NetROI = 907.75
	resp.ROI.ROIPercentage = 264.8

	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleTokenBreakdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	stats := s.tracker.GetStats()

	// Safely extract values from the map to prevent panics from type assertions.
	safeFloat64 := func(m map[string]interface{}, key string) float64 {
		if val, ok := m[key].(float64); ok {
			return val
		}
		return 0.0
	}
	safeInt := func(m map[string]interface{}, key string) int {
		if val, ok := m[key].(float64); ok { // JSON numbers are often float64
			return int(val)
		}
		if val, ok := m[key].(int); ok {
			return val
		}
		return 0
	}

	resp := TokenBreakdownResponse{
		TotalCostUSD:    safeFloat64(stats, "total_cost_usd"),
		TotalTokens:     safeInt(stats, "total_tokens"),
		TotalSavingsUSD: safeFloat64(stats, "total_savings_usd"),
		NetProfitUSD:    safeFloat64(stats, "net_profit_usd"),
		Breakdown: map[string]ModelTokenInfo{
			"sentinel":   {Tokens: 1500, Cost: 0.75},
			"strategist": {Tokens: 800, Cost: 1.20},
			"arbiter":    {Tokens: 400, Cost: 0.80},
			"reasoning":  {Tokens: 200, Cost: 0.60},
		},
	}
	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := SystemStatusResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Uptime:  "2h 34m",
		Services: map[string]string{
			"ai_orchestrator": "online",
			"cloud_adapter":   "online",
			"redis":           "online",
			"database":        "online",
		},
		Metrics: map[string]interface{}{
			"active_optimizations": 3,
			"resources_monitored":  127,
			"cost_savings_today":   45.75,
		},
	}
	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleResources(w http.ResponseWriter, r *http.Request) {
	// The cache is now updated by a background worker.
	// This handler just serves the latest cached data.
	s.resourceCache.RLock()
	defer s.resourceCache.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := ResourcesResponse{
		Resources:   s.resourceCache.resources,
		TotalCount:  len(s.resourceCache.resources),
		LastUpdated: s.resourceCache.fetchedAt,
	}
	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := HealthzResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	}
	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

// --- New Dashboard Handlers (Prioritize & Simplify) ---

func (s *server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Returns key numbers for the hero section:
	// 1. Current monthly burn
	// 2. Projected overrun
	// 3. Biggest saver this week
	// 4. AI recommendation count
	resp := DashboardStatsResponse{
		CurrentMonthlyBurn:    12450.00,
		ProjectedOverrun:      1200.00,
		AIRecommendationCount: 14,
		RunwayExtensionDays:   24,
	}
	resp.BiggestSaver.Description = "Idle EC2 Instances (3)"
	resp.BiggestSaver.PotentialSavings = 1200.00

	json.NewEncoder(w).Encode(resp)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleOpportunities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Returns top savings opportunities as cards
	// Includes AI reasoning and risk score for "Smart Data Presentation"
	opportunities := []OpportunityResponse{
		{
			ID:             "opt-101",
			Title:          "Resize Production DB",
			Description:    "Downgrade db.r5.2xlarge to db.r5.xlarge",
			SavingsMonthly: 450.00,
			RiskScore:      2.5,
			Reasoning:      "CPU utilization < 15% for 30 days. Memory usage stable at 40%.",
			ActionType:     "rightsizing",
			AITier:         "Strategist",
		},
		{
			ID:             "opt-102",
			Title:          "Spot Instance Migration",
			Description:    "Move batch processing cluster to Spot",
			SavingsMonthly: 800.00,
			RiskScore:      4.0,
			Reasoning:      "Workload is fault-tolerant and stateless. Spot history shows 99.9% availability in this AZ.",
			ActionType:     "arbitrage",
			AITier:         "Reasoning",
		},
	}

	json.NewEncoder(w).Encode(opportunities)
	if err := json.NewEncoder(w).Encode(opportunities); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleAnomalies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Returns active anomaly alerts for prominent banners
	anomalies := []AnomalyResponse{
		{
			ID:         "anom-55",
			Severity:   "warning",
			Message:    "Cost spike detected in us-east-1: +40% vs average",
			ResourceID: "nat-gateway-0x82...",
			Timestamp:  time.Now().Add(-2 * time.Hour),
		},
	}

	json.NewEncoder(w).Encode(anomalies)
	if err := json.NewEncoder(w).Encode(anomalies); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}

func (s *server) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Log feedback for AI improvement
	// In a real implementation, this would save to the FeedbackStore
	s.logger.Info("User feedback received",
		zap.String("endpoint", "submit_feedback"),
	)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FeedbackResponse{Status: "received"})
	if err := json.NewEncoder(w).Encode(FeedbackResponse{Status: "received"}); err != nil {
		s.logger.Error("failed to write response", zap.Error(err))
	}
}
