package models

import "time"

type ActionStatus string

const (
	StatusPending   ActionStatus = "PENDING"
	StatusCompleted ActionStatus = "COMPLETED"
	StatusFailed    ActionStatus = "FAILED"
)

type ActionRecord struct {
	RequestID  string       `json:"request_id"`
	Checksum   string       `json:"checksum"`
	Status     ActionStatus `json:"status"`
	ResourceID string       `json:"resource_id,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}
