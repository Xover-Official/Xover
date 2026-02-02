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

// GeminiFlashClient implements Tier 1 (Sentinel) - Fast, cheap monitoring
type GeminiFlashClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	model      string
}

// NewGeminiFlashClient creates a new Gemini Flash client
func NewGeminiFlashClient(apiKey string) *GeminiFlashClient {
	return &GeminiFlashClient{
		apiKey:   apiKey,
		endpoint: "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "gemini-2.0-flash-exp",
	}
}

// Analyze implements AIClient interface
func (c *GeminiFlashClient) Analyze(ctx context.Context, request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

	// Build Gemini API request
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": request.Prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     request.Temperature,
			"maxOutputTokens": request.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API call
	url := fmt.Sprintf("%s?key=%s", c.endpoint, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := result.Candidates[0].Content.Parts[0].Text
	tokensUsed := result.UsageMetadata.TotalTokenCount

	// Calculate cost (Gemini Flash pricing: ~$0.002 per 1K tokens)
	costUSD := float64(tokensUsed) * 0.000002

	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		CostUSD:    costUSD,
		Model:      c.model,
		Latency:    time.Since(startTime),
		Confidence: 0.85, // Tier 1 has moderate confidence
	}, nil
}

// GetEstimatedCost estimates cost before making the call
func (c *GeminiFlashClient) GetEstimatedCost(request AIRequest) float64 {
	estimatedTokens := len(request.Prompt)/4 + request.MaxTokens
	return float64(estimatedTokens) * 0.000002
}

// GetModel returns the model identifier
func (c *GeminiFlashClient) GetModel() string {
	return c.model
}

// GetTier returns the tier level
func (c *GeminiFlashClient) GetTier() int {
	return 1
}

// HealthCheck verifies the API is accessible
func (c *GeminiFlashClient) HealthCheck(ctx context.Context) error {
	testReq := AIRequest{
		Prompt:      "Hello",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := c.Analyze(ctx, testReq)
	return err
}
