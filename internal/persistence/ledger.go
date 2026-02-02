package persistence

import (
	"context"
	"time"
)

// Ledger defines the interface for persistence operations
type Ledger interface {
	RecordAction(ctx context.Context, action *Action) error
	GetPendingActions(ctx context.Context) ([]Action, error)
	MarkComplete(ctx context.Context, actionID string) error
	MarkFailed(ctx context.Context, actionID string, errorMsg string) error
	GetActionByChecksum(ctx context.Context, checksum string) (*Action, error)
	GetStats(ctx context.Context) (map[string]int, error)
	Close()
}

// Action represents an idempotent action in the system
type Action struct {
	ID               string
	ResourceID       string
	ActionType       string
	Status           string
	Checksum         string
	Payload          map[string]interface{}
	Reasoning        string
	RiskScore        float64
	EstimatedSavings float64
	CreatedAt        time.Time
	StartedAt        *time.Time
	CompletedAt      *time.Time
	ErrorMessage     string
}
