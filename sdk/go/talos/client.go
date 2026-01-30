// Talos Guardian Go SDK
// Official Go client for the Talos platform

package talos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is the Talos API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Talos client
func NewClient(baseURL string, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Health checks system health
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var health HealthResponse
	err := c.get(ctx, "/health", &health)
	return &health, err
}

// GetSwarmStatus gets AI swarm status
func (c *Client) GetSwarmStatus(ctx context.Context) (*SwarmStatus, error) {
	var status SwarmStatus
	err := c.get(ctx, "/api/swarm/live", &status)
	return &status, err
}

// RunOptimization runs an optimization workflow
func (c *Client) RunOptimization(ctx context.Context, req OptimizationRequest) (*OptimizationResponse, error) {
	var result OptimizationResponse
	err := c.post(ctx, "/api/optimize", req, &result)
	return &result, err
}

// GetResources gets cloud resources
func (c *Client) GetResources(ctx context.Context, provider, resourceType string) ([]Resource, error) {
	url := "/api/resources"
	if provider != "" || resourceType != "" {
		url += "?"
		if provider != "" {
			url += "provider=" + provider
		}
		if resourceType != "" {
			if provider != "" {
				url += "&"
			}
			url += "type=" + resourceType
		}
	}

	var resources []Resource
	err := c.get(ctx, url, &resources)
	return resources, err
}

// GetROI gets ROI metrics
func (c *Client) GetROI(ctx context.Context) (*ROI, error) {
	var roi ROI
	err := c.get(ctx, "/api/roi", &roi)
	return &roi, err
}

// Chat sends a message to the AI
func (c *Client) Chat(ctx context.Context, message string) (string, error) {
	req := ChatRequest{Message: message}
	var resp ChatResponse

	if err := c.post(ctx, "/api/ai/chat", req, &resp); err != nil {
		return "", err
	}

	return resp.Response, nil
}

// HTTP helpers

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("API error: %s", errResp.Error)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// Types

type HealthResponse struct {
	Status    string                 `json:"status"`
	Checks    map[string]interface{} `json:"checks"`
	Timestamp int64                  `json:"timestamp"`
}

type SwarmStatus struct {
	ActiveTier    int          `json:"active_tier"`
	TierStatus    []TierStatus `json:"tier_status"`
	CurrentAction string       `json:"current_action"`
	QueueDepth    int          `json:"queue_depth"`
}

type TierStatus struct {
	Tier          int     `json:"tier"`
	Name          string  `json:"name"`
	Model         string  `json:"model"`
	Active        bool    `json:"active"`
	RequestsToday int     `json:"requests_today"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	SuccessRate   float64 `json:"success_rate"`
	Status        string  `json:"status"`
}

type OptimizationRequest struct {
	Type      string  `json:"type"`
	RiskLimit float64 `json:"risk_limit"`
	DryRun    bool    `json:"dry_run"`
}

type OptimizationResponse struct {
	OptimizationsFound int     `json:"optimizations_found"`
	EstimatedSavings   float64 `json:"estimated_savings"`
	ActionsApplied     int     `json:"actions_applied"`
	Status             string  `json:"status"`
}

type Resource struct {
	ID                string            `json:"id"`
	Type              string            `json:"type"`
	Provider          string            `json:"provider"`
	Region            string            `json:"region"`
	CostPerMonth      float64           `json:"cost_per_month"`
	OptimizationScore float64           `json:"optimization_score"`
	Tags              map[string]string `json:"tags"`
}

type ROI struct {
	Ratio        float64 `json:"ratio"`
	TotalSavings float64 `json:"total_savings"`
	TotalCost    float64 `json:"total_cost"`
	NetProfit    float64 `json:"net_profit"`
}

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
	Tier     int    `json:"tier"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
