package api

import (
	"net/http"
)

// @title Talos Guardian API
// @version 2.0
// @description Autonomous AI-powered cloud optimization platform
// @termsOfService https://talos.dev/terms

// @contact.name API Support
// @contact.url https://talos.dev/support
// @contact.email support@talos.dev

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @tag.name Health
// @tag.description Health check endpoints

// @tag.name AI
// @tag.description AI swarm operations

// @tag.name Resources
// @tag.description Cloud resource management

// @tag.name Optimization
// @tag.description Cost optimization operations

// @tag.name Analytics
// @tag.description ROI and analytics endpoints

// HealthCheck godoc
// @Summary Check system health
// @Description Get detailed health status of all components
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorResponse
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Implementation in health_handler.go
}

// GetSwarmStatus godoc
// @Summary Get AI swarm status
// @Description Get real-time status of all AI tiers
// @Tags AI
// @Accept json
// @Produce json
// @Success 200 {object} SwarmStatusResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /swarm/status [get]
func GetSwarmStatus(w http.ResponseWriter, r *http.Request) {}

// RunOptimization godoc
// @Summary Run optimization
// @Description Execute AI-driven optimization workflow
// @Tags Optimization
// @Accept json
// @Produce json
// @Param request body OptimizationRequest true "Optimization parameters"
// @Success 200 {object} OptimizationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /optimize [post]
func RunOptimization(w http.ResponseWriter, r *http.Request) {}

// GetResources godoc
// @Summary List cloud resources
// @Description Get all discovered cloud resources with costs
// @Tags Resources
// @Accept json
// @Produce json
// @Param provider query string false "Filter by provider (aws, gcp, azure)"
// @Param type query string false "Filter by resource type"
// @Success 200 {array} ResourceV2
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /resources [get]
func GetResources(w http.ResponseWriter, r *http.Request) {}

// GetROI godoc
// @Summary Get ROI metrics
// @Description Get AI cost vs cloud savings ROI data
// @Tags Analytics
// @Accept json
// @Produce json
// @Success 200 {object} ROIResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /roi [get]
func GetROI(w http.ResponseWriter, r *http.Request) {}

// ChatWithAI godoc
// @Summary Chat with AI
// @Description Interactive chat with AI swarm
// @Tags AI
// @Accept json
// @Produce json
// @Param message body ChatRequest true "Chat message"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /ai/chat [post]
func ChatWithAI(w http.ResponseWriter, r *http.Request) {}

// API Models for Swagger

type HealthResponse struct {
	Status    string                 `json:"status" example:"healthy"`
	Checks    map[string]interface{} `json:"checks"`
	Timestamp int64                  `json:"timestamp"`
}

type SwarmStatusResponse struct {
	ActiveTier    int          `json:"active_tier" example:"2"`
	TierStatus    []TierStatus `json:"tier_status"`
	CurrentAction string       `json:"current_action"`
	QueueDepth    int          `json:"queue_depth"`
}

type TierStatus struct {
	Tier          int     `json:"tier" example:"1"`
	Name          string  `json:"name" example:"Sentinel"`
	Model         string  `json:"model" example:"gemini-2.0-flash-exp"`
	Active        bool    `json:"active" example:"true"`
	RequestsToday int     `json:"requests_today" example:"450"`
	AvgLatencyMs  float64 `json:"avg_latency_ms" example:"850"`
	SuccessRate   float64 `json:"success_rate" example:"99.5"`
	Status        string  `json:"status" example:"healthy"`
}

type OptimizationRequest struct {
	Type      string  `json:"type" example:"full"`
	RiskLimit float64 `json:"risk_limit" example:"7.0"`
	DryRun    bool    `json:"dry_run" example:"false"`
}

type OptimizationResponse struct {
	OptimizationsFound int     `json:"optimizations_found" example:"12"`
	EstimatedSavings   float64 `json:"estimated_savings" example:"1250.50"`
	ActionsApplied     int     `json:"actions_applied" example:"8"`
	Status             string  `json:"status" example:"completed"`
}

type ResourceV2 struct {
	ID                string            `json:"id" example:"i-1234567890abcdef"`
	Type              string            `json:"type" example:"ec2"`
	Provider          string            `json:"provider" example:"aws"`
	Region            string            `json:"region" example:"us-east-1"`
	CostPerMonth      float64           `json:"cost_per_month" example:"73.50"`
	OptimizationScore float64           `json:"optimization_score" example:"45.5"`
	Tags              map[string]string `json:"tags"`
}

type ROIResponse struct {
	Ratio        float64 `json:"ratio" example:"8.5"`
	TotalSavings float64 `json:"total_savings" example:"12500.00"`
	TotalCost    float64 `json:"total_cost" example:"1470.00"`
	NetProfit    float64 `json:"net_profit" example:"11030.00"`
}

type ChatRequest struct {
	Message string `json:"message" example:"Should I optimize this EC2 instance?"`
}

type ChatResponse struct {
	Response string `json:"response" example:"Yes, based on 5% CPU usage..."`
	Model    string `json:"model" example:"gemini-pro"`
	Tier     int    `json:"tier" example:"2"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Missing required field"`
}
