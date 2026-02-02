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

// GeminiProClient implements Tier 2 (Strategist) - Deeper cost-benefit analysis
type GeminiProClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	model      string
}

// NewGeminiProClient creates a new Gemini Pro client
func NewGeminiProClient(apiKey string) *GeminiProClient {
	return &GeminiProClient{
		apiKey:   apiKey,
		endpoint: "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro:generateContent",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "gemini-1.5-pro",
	}
}

// Analyze implements AIClient interface
func (c *GeminiProClient) Analyze(ctx context.Context, request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

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

	// Gemini Pro pricing: ~$0.01 per 1K tokens (higher than Flash)
	costUSD := float64(tokensUsed) * 0.00001

	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		CostUSD:    costUSD,
		Model:      c.model,
		Latency:    time.Since(startTime),
		Confidence: 0.90, // Higher confidence than Tier 1
	}, nil
}

func (c *GeminiProClient) GetEstimatedCost(request AIRequest) float64 {
	estimatedTokens := len(request.Prompt)/4 + request.MaxTokens
	return float64(estimatedTokens) * 0.00001
}

func (c *GeminiProClient) GetModel() string {
	return c.model
}

func (c *GeminiProClient) GetTier() int {
	return 2
}

func (c *GeminiProClient) HealthCheck(ctx context.Context) error {
	testReq := AIRequest{
		Prompt:      "Hello",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := c.Analyze(ctx, testReq)
	return err
}
