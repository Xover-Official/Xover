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

// ClaudeClient implements Tier 3 (Arbiter) - Safety validation
type ClaudeClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	model      string
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:   apiKey,
		endpoint: "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "claude-3-5-sonnet-20240620",
	}
}

// Analyze implements AIClient interface
func (c *ClaudeClient) Analyze(ctx context.Context, request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

	reqBody := map[string]interface{}{
		"model":       c.model,
		"max_tokens":  request.MaxTokens,
		"temperature": request.Temperature,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": request.Prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := result.Content[0].Text
	tokensUsed := result.Usage.InputTokens + result.Usage.OutputTokens

	// Claude pricing: ~$0.003 per 1K input, $0.015 per 1K output
	inputCost := float64(result.Usage.InputTokens) * 0.000003
	outputCost := float64(result.Usage.OutputTokens) * 0.000015
	costUSD := inputCost + outputCost

	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		CostUSD:    costUSD,
		Model:      c.model,
		Latency:    time.Since(startTime),
		Confidence: 0.92, // High confidence - Arbiter role
	}, nil
}

func (c *ClaudeClient) GetEstimatedCost(request AIRequest) float64 {
	estimatedInputTokens := len(request.Prompt) / 4
	estimatedOutputTokens := request.MaxTokens

	inputCost := float64(estimatedInputTokens) * 0.000003
	outputCost := float64(estimatedOutputTokens) * 0.000015

	return inputCost + outputCost
}

func (c *ClaudeClient) GetModel() string {
	return c.model
}

func (c *ClaudeClient) GetTier() int {
	return 3
}

func (c *ClaudeClient) HealthCheck(ctx context.Context) error {
	testReq := AIRequest{
		Prompt:      "Hello",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := c.Analyze(ctx, testReq)
	return err
}
