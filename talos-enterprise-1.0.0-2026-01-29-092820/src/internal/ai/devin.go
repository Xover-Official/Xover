package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DevinClient implements Tier 5 (Oracle) - Critical infrastructure operations
type DevinClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	model      string
}

// NewDevinClient creates a new Devin client
func NewDevinClient(apiKey string) *DevinClient {
	return &DevinClient{
		apiKey:   apiKey,
		endpoint: "https://api.devin.ai/v1/completions",
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for complex operations
		},
		model: "devin-1",
	}
}

// Analyze implements AIClient interface for critical decisions
func (c *DevinClient) Analyze(request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

	// Enhanced prompt for critical infrastructure decisions
	criticalPrompt := fmt.Sprintf(`CRITICAL INFRASTRUCTURE DECISION REQUIRED:

%s

IMPORTANT:
- This is a high-risk operation affecting production infrastructure
- Provide extremely detailed analysis and safeguards
- List all potential failure modes
- Suggest rollback procedures
- Confidence threshold: >95%%

Provide comprehensive analysis:`, request.Prompt)

	reqBody := map[string]interface{}{
		"prompt":      criticalPrompt,
		"max_tokens":  request.MaxTokens,
		"temperature": request.Temperature,
		"context": map[string]interface{}{
			"risk_score":    request.RiskScore,
			"resource_type": request.ResourceType,
			"critical":      true,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(request.Context, "POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("X-Devin-Mode", "critical")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Completion string `json:"completion"`
		Metadata   struct {
			TokensUsed int      `json:"tokens_used"`
			Confidence float64  `json:"confidence"`
			RiskLevel  string   `json:"risk_level"`
			Safeguards []string `json:"safeguards"`
		} `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	content := result.Completion
	tokensUsed := result.Metadata.TokensUsed

	// Devin pricing: Premium tier - ~$0.50 per 1K tokens (most expensive)
	costUSD := float64(tokensUsed) * 0.0005

	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		CostUSD:    costUSD,
		Model:      c.model,
		Latency:    time.Since(startTime),
		Confidence: result.Metadata.Confidence,
		Reasoning:  fmt.Sprintf("Critical analysis with %d safeguards", len(result.Metadata.Safeguards)),
	}, nil
}

func (c *DevinClient) GetEstimatedCost(request AIRequest) float64 {
	estimatedTokens := len(request.Prompt)/4 + request.MaxTokens
	return float64(estimatedTokens) * 0.0005
}

func (c *DevinClient) GetModel() string {
	return c.model
}

func (c *DevinClient) GetTier() int {
	return 5
}

func (c *DevinClient) HealthCheck(ctx context.Context) error {
	testReq := AIRequest{
		Context:     ctx,
		Prompt:      "System health check",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := c.Analyze(testReq)
	return err
}
