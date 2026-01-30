package ai

import (
	"sync"
	"time"
)

// PerformanceMetrics tracks AI model performance
type PerformanceMetrics struct {
	Model            string        `json:"model"`
	TotalCalls       int64         `json:"total_calls"`
	SuccessfulCalls  int64         `json:"successful_calls"`
	FailedCalls      int64         `json:"failed_calls"`
	TotalLatency     time.Duration `json:"total_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	TotalTokens      int64         `json:"total_tokens"`
	TotalCost        float64       `json:"total_cost"`
	LastUsed         time.Time     `json:"last_used"`
	SuccessRate      float64       `json:"success_rate"`
	AvgLatency       time.Duration `json:"avg_latency"`
	AvgTokensPerCall float64       `json:"avg_tokens_per_call"`
}

// PerformanceTracker tracks performance metrics for all models
type PerformanceTracker struct {
	mu      sync.RWMutex
	metrics map[string]*PerformanceMetrics
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics: make(map[string]*PerformanceMetrics),
	}
}

// RecordSuccess records a successful AI call
func (p *PerformanceTracker) RecordSuccess(model string, latency time.Duration, tokens int, cost float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	m, exists := p.metrics[model]
	if !exists {
		m = &PerformanceMetrics{
			Model:      model,
			MinLatency: latency,
			MaxLatency: latency,
		}
		p.metrics[model] = m
	}

	m.TotalCalls++
	m.SuccessfulCalls++
	m.TotalLatency += latency
	m.TotalTokens += int64(tokens)
	m.TotalCost += cost
	m.LastUsed = time.Now()

	if latency < m.MinLatency {
		m.MinLatency = latency
	}
	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}

	p.recalculate(m)
}

// RecordFailure records a failed AI call
func (p *PerformanceTracker) RecordFailure(model string, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	m, exists := p.metrics[model]
	if !exists {
		m = &PerformanceMetrics{
			Model:      model,
			MinLatency: latency,
			MaxLatency: latency,
		}
		p.metrics[model] = m
	}

	m.TotalCalls++
	m.FailedCalls++
	m.TotalLatency += latency
	m.LastUsed = time.Now()

	p.recalculate(m)
}

// recalculate updates derived metrics
func (p *PerformanceTracker) recalculate(m *PerformanceMetrics) {
	if m.TotalCalls > 0 {
		m.SuccessRate = float64(m.SuccessfulCalls) / float64(m.TotalCalls) * 100
		m.AvgLatency = m.TotalLatency / time.Duration(m.TotalCalls)
	}

	if m.SuccessfulCalls > 0 {
		m.AvgTokensPerCall = float64(m.TotalTokens) / float64(m.SuccessfulCalls)
	}
}

// GetMetrics returns metrics for a specific model
func (p *PerformanceTracker) GetMetrics(model string) *PerformanceMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if m, exists := p.metrics[model]; exists {
		// Return a copy
		copy := *m
		return &copy
	}

	return nil
}

// GetAllMetrics returns all metrics
func (p *PerformanceTracker) GetAllMetrics() map[string]*PerformanceMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*PerformanceMetrics)
	for k, v := range p.metrics {
		copy := *v
		result[k] = &copy
	}

	return result
}

// GetBestModel returns the model with the best performance (lowest avg latency + high success rate)
func (p *PerformanceTracker) GetBestModel(minCalls int64) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var bestModel string
	var bestScore float64 = 999999999

	for model, m := range p.metrics {
		if m.TotalCalls < minCalls {
			continue
		}

		// Score: lower is better (latency in ms + penalty for failures)
		score := float64(m.AvgLatency.Milliseconds())
		if m.SuccessRate < 95 {
			score *= (100 - m.SuccessRate) / 100 // Penalty for low success rate
		}

		if score < bestScore {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}

// Reset clears all metrics
func (p *PerformanceTracker) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics = make(map[string]*PerformanceMetrics)
}
