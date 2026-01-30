package persistence

import (
	"time"
)

// Ledger defines the interface for persistence operations
type Ledger interface {
	RecordAction(action Action) error
	GetPendingActions() ([]Action, error)
	MarkComplete(actionID string) error
	MarkFailed(actionID string, errorMsg string) error
	GetActionByChecksum(checksum string) (*Action, error)
	GetStats() (map[string]int, error)
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
