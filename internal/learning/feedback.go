package learning

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// FeedbackType defines the user reaction
type FeedbackType string

const (
	FeedbackApprove FeedbackType = "APPROVE"
	FeedbackReject  FeedbackType = "REJECT"
	FeedbackModify  FeedbackType = "MODIFY"
)

// FeedbackEvent captures a user's interaction with an AI decision
type FeedbackEvent struct {
	ID        string       `json:"id"`
	ActionID  string       `json:"action_id"`
	Type      FeedbackType `json:"type"`
	UserID    string       `json:"user_id"`
	Comments  string       `json:"comments,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

// LearningEngine manages weight updates based on feedback
type LearningEngine struct {
	db      *sql.DB
	weights map[string]float64
	mu      sync.RWMutex
}

func NewLearningEngine(db *sql.DB) *LearningEngine {
	return &LearningEngine{
		db:      db,
		weights: make(map[string]float64),
	}
}

// RecordFeedback stores feedback and triggers learning
func (l *LearningEngine) RecordFeedback(ctx context.Context, feedback FeedbackEvent) error {
	// 1. Store raw feedback
	query := `INSERT INTO feedback_events (id, action_id, type, user_id, comments, timestamp) 
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := l.db.ExecContext(ctx, query, feedback.ID, feedback.ActionID, feedback.Type, feedback.UserID, feedback.Comments, feedback.Timestamp)
	if err != nil {
		return err
	}

	// 2. Trigger weight update (Reinforcement)
	go l.updateWeights(feedback)

	return nil
}

// updateWeights adjusts decision weights based on feedback
func (l *LearningEngine) updateWeights(feedback FeedbackEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Fetch action metadata (mocked)
	actionType := "scale_up_db" // In real implementation, fetch from DB using feedback.ActionID

	currentWeight := l.getWeight(actionType)
	learningRate := 0.05

	switch feedback.Type {
	case FeedbackApprove:
		// Reinforce: Increase trust in this action type
		l.weights[actionType] = currentWeight + learningRate
	case FeedbackReject:
		// Penalize: Decrease trust
		l.weights[actionType] = currentWeight - (learningRate * 2.0) // Penalize harder than reward
	case FeedbackModify:
		// Slight penalty + Context update (TODO)
		l.weights[actionType] = currentWeight - (learningRate * 0.5)
	}

	// Normalize
	if l.weights[actionType] > 1.0 {
		l.weights[actionType] = 1.0
	}
	if l.weights[actionType] < 0.1 {
		l.weights[actionType] = 0.1
	}

	// Persist new weight to DB
	// l.db.Exec("UPDATE decision_weights SET weight = ? WHERE action_type = ?", l.weights[actionType], actionType)
}

func (l *LearningEngine) getWeight(actionType string) float64 {
	if w, ok := l.weights[actionType]; ok {
		return w
	}
	return 0.5 // Default neutral weight
}

// GetActionConfidence returns the learned confidence score for an action type
func (l *LearningEngine) GetActionConfidence(actionType string) float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.getWeight(actionType)
}
