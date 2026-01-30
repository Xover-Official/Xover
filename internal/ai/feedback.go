package ai

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackPositive FeedbackType = "positive"
	FeedbackNegative FeedbackType = "negative"
	FeedbackNeutral  FeedbackType = "neutral"
)

// Feedback represents user feedback on an AI decision
type Feedback struct {
	ID            string       `json:"id"`
	Model         string       `json:"model"`
	Prompt        string       `json:"prompt"`
	Response      string       `json:"response"`
	FeedbackType  FeedbackType `json:"feedback_type"`
	Comment       string       `json:"comment,omitempty"`
	Timestamp     time.Time    `json:"timestamp"`
	RiskScore     float64      `json:"risk_score"`
	WasApplied    bool         `json:"was_applied"`
	ActualOutcome string       `json:"actual_outcome,omitempty"`
}

// FeedbackStore stores and manages feedback for reinforcement learning
type FeedbackStore struct {
	mu        sync.RWMutex
	feedback  []Feedback
	storePath string
}

// NewFeedbackStore creates a new feedback store
func NewFeedbackStore(storePath string) *FeedbackStore {
	store := &FeedbackStore{
		feedback:  make([]Feedback, 0),
		storePath: storePath,
	}
	store.load()
	return store
}

// Add adds new feedback
func (f *FeedbackStore) Add(feedback Feedback) {
	f.mu.Lock()
	defer f.mu.Unlock()

	feedback.Timestamp = time.Now()
	f.feedback = append(f.feedback, feedback)

	f.persist()
}

// GetAll returns all feedback
func (f *FeedbackStore) GetAll() []Feedback {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]Feedback, len(f.feedback))
	copy(result, f.feedback)
	return result
}

// GetByModel returns feedback for a specific model
func (f *FeedbackStore) GetByModel(model string) []Feedback {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]Feedback, 0)
	for _, fb := range f.feedback {
		if fb.Model == model {
			result = append(result, fb)
		}
	}
	return result
}

// GetStats returns feedback statistics
func (f *FeedbackStore) GetStats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := map[string]interface{}{
		"total":    len(f.feedback),
		"positive": 0,
		"negative": 0,
		"neutral":  0,
		"by_model": make(map[string]map[string]int),
	}

	for _, fb := range f.feedback {
		switch fb.FeedbackType {
		case FeedbackPositive:
			stats["positive"] = stats["positive"].(int) + 1
		case FeedbackNegative:
			stats["negative"] = stats["negative"].(int) + 1
		case FeedbackNeutral:
			stats["neutral"] = stats["neutral"].(int) + 1
		}

		modelStats, ok := stats["by_model"].(map[string]map[string]int)[fb.Model]
		if !ok {
			modelStats = map[string]int{"positive": 0, "negative": 0, "neutral": 0}
			stats["by_model"].(map[string]map[string]int)[fb.Model] = modelStats
		}

		switch fb.FeedbackType {
		case FeedbackPositive:
			modelStats["positive"]++
		case FeedbackNegative:
			modelStats["negative"]++
		case FeedbackNeutral:
			modelStats["neutral"]++
		}
	}

	// Calculate approval rate
	total := len(f.feedback)
	if total > 0 {
		positive := stats["positive"].(int)
		stats["approval_rate"] = float64(positive) / float64(total) * 100
	}

	return stats
}

// persist saves feedback to disk
func (f *FeedbackStore) persist() {
	if f.storePath == "" {
		return
	}

	data, err := json.MarshalIndent(f.feedback, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(f.storePath, data, 0644)
}

// load loads feedback from disk
func (f *FeedbackStore) load() {
	if f.storePath == "" {
		return
	}

	data, err := os.ReadFile(f.storePath)
	if err != nil {
		return
	}

	json.Unmarshal(data, &f.feedback)
}

// GetRecentNegative returns recent negative feedback for analysis
func (f *FeedbackStore) GetRecentNegative(limit int) []Feedback {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]Feedback, 0)
	count := 0

	// Iterate in reverse (most recent first)
	for i := len(f.feedback) - 1; i >= 0 && count < limit; i-- {
		if f.feedback[i].FeedbackType == FeedbackNegative {
			result = append(result, f.feedback[i])
			count++
		}
	}

	return result
}
