package devin

import (
	"context"
	"fmt"
	"time"
)

// TaskRequest represents a complex coding task for Devin
type TaskRequest struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Files       []string `json:"files"` // Context files
	BranchName  string   `json:"branch_name"`
	AutoMerge   bool     `json:"auto_merge"`
}

// TaskStatus represents the status of a Devin task
type TaskStatus struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"` // RUNNING, COMPLETED, FAILED, USR_ACTION_REQUIRED
	PRLink    string `json:"pr_link,omitempty"`
	Logs      string `json:"logs,omitempty"`
	UpdatedAt time.Time
}

// Client handles interaction with the Devin API
type Client struct {
	APIKey string
}

func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// DelegateTask sends a complex refactoring task to Devin
func (c *Client) DelegateTask(ctx context.Context, req TaskRequest) (*TaskStatus, error) {
	fmt.Printf("🤖 Delegating task to Devin: %s\n", req.Description)

	// Mock API call to Devin
	// resp, err := http.Post("https://api.devin.ai/v1/sessions", ...)

	return &TaskStatus{
		TaskID:    "devin-sess-12345",
		Status:    "RUNNING",
		UpdatedAt: time.Now(),
	}, nil
}

// GetStatus checks the progress of a task
func (c *Client) GetStatus(ctx context.Context, taskID string) (*TaskStatus, error) {
	// Mock status check
	return &TaskStatus{
		TaskID: "devin-sess-12345",
		Status: "COMPLETED",
		PRLink: "https://github.com/org/repo/pull/101",
		Logs:   "Analyzed 15 files. Refactored logic in auth.go. Ran tests: PASS.",
	}, nil
}

// ReviewChanges acts as the 'Talos Governance' layer
// Talos reviews Devin's work before allowing merge
func (c *Client) ReviewChanges(ctx context.Context, prLink string) (bool, string) {
	// Here Talos would use its own reasoning engine to check the PR
	// e.g., run policy checks, security scans, heuristic analysis

	fmt.Printf("🛡️ Talos Governance: Reviewing Devin's PR %s\n", prLink)

	// Mock approval
	return true, "Changes comply with security policy SEC-001 and passing all tests."
}
