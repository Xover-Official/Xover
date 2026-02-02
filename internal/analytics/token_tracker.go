package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/project-atlas/atlas/internal/logger"
	"go.uber.org/zap"
)

// TokenUsage tracks usage for a specific AI model
type TokenUsage struct {
	Tokens   int     `json:"tokens"`
	CostUSD  float64 `json:"cost_usd"`
	Requests int     `json:"requests"`
}

// TokenTracker tracks AI token usage and calculates ROI
type TokenTracker struct {
	mu              sync.RWMutex
	TotalTokens     int                   `json:"total_tokens"`
	TotalCostUSD    float64               `json:"total_cost_usd"`
	TotalSavingsUSD float64               `json:"total_savings_usd"`
	NetROI          float64               `json:"net_roi"`
	ModelBreakdown  map[string]TokenUsage `json:"model_breakdown"`
	StartTime       time.Time             `json:"start_time"`
	persistPath     string
	stopChan        chan struct{}
	dirty           bool
}

// Model pricing (per 1M tokens)
var modelPricing = map[string]float64{
	"gemini-2.0-flash-exp":        0.075, // $0.075 per 1M tokens
	"gemini-1.5-pro":              2.50,  // $2.50 per 1M tokens
	"anthropic/claude-3.5-sonnet": 3.00,  // $3.00 per 1M tokens
	"openai/gpt-5-mini":           0.10,  // $0.10 per 1M tokens (estimated)
	"devin":                       10.00, // $10.00 per request (flat fee)
}

// NewTokenTracker creates a new token tracker
func NewTokenTracker(persistPath string) *TokenTracker {
	tracker := &TokenTracker{
		ModelBreakdown: make(map[string]TokenUsage),
		StartTime:      time.Now(),
		persistPath:    persistPath,
		stopChan:       make(chan struct{}),
	}

	// Try to load existing data
	tracker.Load()

	// Start persistence loop
	go tracker.persistLoop()

	return tracker
}

func (t *TokenTracker) persistLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			if t.dirty {
				t.saveInternal()
				t.dirty = false
			}
			t.mu.Unlock()
		case <-t.stopChan:
			t.Close()
			return
		}
	}
}

// Close stops the tracker and performs a final save
func (t *TokenTracker) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.dirty {
		t.saveInternal()
		t.dirty = false
	}
}

// RecordUsage records token usage for a specific model
func (t *TokenTracker) RecordUsage(model string, tokens int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Calculate cost
	pricePerMillion, ok := modelPricing[model]
	if !ok {
		pricePerMillion = 1.0 // Default fallback
	}

	var costUSD float64
	if model == "devin" {
		costUSD = pricePerMillion // Flat fee
	} else {
		costUSD = (float64(tokens) / 1_000_000.0) * pricePerMillion
	}

	// Update totals
	t.TotalTokens += tokens
	t.TotalCostUSD += costUSD

	// Update model breakdown
	usage := t.ModelBreakdown[model]
	usage.Tokens += tokens
	usage.CostUSD += costUSD
	usage.Requests++
	t.ModelBreakdown[model] = usage

	// Recalculate ROI
	t.calculateROI()

	// Mark as dirty
	t.dirty = true
}

func (t *TokenTracker) RecordSavings(savingsUSD float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.TotalSavingsUSD += savingsUSD
	t.calculateROI()
	t.dirty = true
}

// calculateROI calculates the net ROI (must be called with lock held)
func (t *TokenTracker) calculateROI() {
	if t.TotalCostUSD > 0 {
		t.NetROI = ((t.TotalSavingsUSD - t.TotalCostUSD) / t.TotalCostUSD) * 100
	} else {
		t.NetROI = 0
	}
}

// GetROI returns the current ROI percentage
func (t *TokenTracker) GetROI() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.NetROI
}

// GetStats returns current statistics
func (t *TokenTracker) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"total_tokens":      t.TotalTokens,
		"total_cost_usd":    t.TotalCostUSD,
		"total_savings_usd": t.TotalSavingsUSD,
		"net_roi":           t.NetROI,
		"net_profit_usd":    t.TotalSavingsUSD - t.TotalCostUSD,
		"model_breakdown":   t.ModelBreakdown,
		"uptime_hours":      time.Since(t.StartTime).Hours(),
	}
}

// GenerateReport generates a human-readable report
func (t *TokenTracker) GenerateReport() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	report := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════╗
║           TALOS AI TOKEN & ROI REPORT                      ║
╚════════════════════════════════════════════════════════════╝

📊 SUMMARY
  • Total Tokens Used:     %d
  • Total AI Cost:          $%.2f
  • Total Savings:          $%.2f
  • Net Profit:             $%.2f
  • ROI:                    %.1f%%
  • Uptime:                 %.1f hours

🤖 MODEL BREAKDOWN
`, t.TotalTokens, t.TotalCostUSD, t.TotalSavingsUSD,
		t.TotalSavingsUSD-t.TotalCostUSD, t.NetROI, time.Since(t.StartTime).Hours())

	for model, usage := range t.ModelBreakdown {
		report += fmt.Sprintf("  • %-30s %8d tokens | %4d requests | $%.2f\n",
			model, usage.Tokens, usage.Requests, usage.CostUSD)
	}

	return report
}

// saveInternal persists the tracker state to disk (lock assumed held)
func (t *TokenTracker) saveInternal() {
	if t.persistPath == "" {
		return
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		logger.GetLogger().Error("Failed to marshal token tracker data", zap.Error(err))
		return
	}

	if err := os.WriteFile(t.persistPath, data, 0644); err != nil {
		logger.GetLogger().Error("Failed to save token tracker data", zap.String("path", t.persistPath), zap.Error(err))
	}
}

// GetBreakdown returns model breakdown statistics
func (t *TokenTracker) GetBreakdown() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	breakdown := make(map[string]interface{})
	for model, usage := range t.ModelBreakdown {
		breakdown[model] = map[string]interface{}{
			"tokens":   usage.Tokens,
			"cost":     usage.CostUSD,
			"requests": usage.Requests,
		}
	}

	// Add totals
	breakdown["total_cost"] = t.TotalCostUSD
	breakdown["projected_savings"] = t.TotalSavingsUSD
	breakdown["roi"] = t.NetROI

	return breakdown
}

// TrackAI is a convenience method that combines recording usage and savings
func (t *TokenTracker) TrackAI(model string, tokens int, costUSD, savingsUSD float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Update totals
	t.TotalTokens += tokens
	t.TotalCostUSD += costUSD
	t.TotalSavingsUSD += savingsUSD

	// Update model breakdown
	usage := t.ModelBreakdown[model]
	usage.Tokens += tokens
	usage.CostUSD += costUSD
	usage.Requests++
	t.ModelBreakdown[model] = usage

	// Recalculate ROI
	t.calculateROI()

	// Mark as dirty
	t.dirty = true
}

// Load loads the tracker state from disk
func (t *TokenTracker) Load() error {
	if t.persistPath == "" {
		return nil
	}

	data, err := os.ReadFile(t.persistPath)
	if err != nil {
		return err // File doesn't exist yet, that's okay
	}

	return json.Unmarshal(data, t)
}
