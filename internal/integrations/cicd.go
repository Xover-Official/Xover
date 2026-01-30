package integrations

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CIProvider defines CI/CD providers
type CIProvider string

const (
	DriverGitHub CIProvider = "github"
	DriverGitLab CIProvider = "gitlab"
)

// TriggerRequest defines a pipeline trigger
type TriggerRequest struct {
	Provider CIProvider
	Token    string
	Owner    string // User or Org
	Repo     string
	Ref      string // Branch or Tag
	Workflow string // Filename or ID
	Inputs   map[string]interface{}
}

// CIClient handles CI triggering
type CIClient struct {
	client *http.Client
}

// NewCIClient creates a CI client
func NewCIClient() *CIClient {
	return &CIClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// TriggerPipeline triggers a workflow
func (c *CIClient) TriggerPipeline(ctx context.Context, req TriggerRequest) error {
	switch req.Provider {
	case DriverGitHub:
		return c.triggerGitHub(ctx, req)
	case DriverGitLab:
		return c.triggerGitLab(ctx, req)
	default:
		return fmt.Errorf("unsupported provider: %s", req.Provider)
	}
}

func (c *CIClient) triggerGitHub(ctx context.Context, req TriggerRequest) error {
	// https://docs.github.com/en/rest/actions/workflows#create-a-workflow-dispatch-event
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches",
		req.Owner, req.Repo, req.Workflow)

	// Payload
	payload := map[string]interface{}{
		"ref":    req.Ref,
		"inputs": req.Inputs,
	}

	return c.doRequest(ctx, "POST", url, req.Token, payload)
}

func (c *CIClient) triggerGitLab(ctx context.Context, req TriggerRequest) error {
	// https://docs.gitlab.com/ee/api/pipelines.html#create-a-new-pipeline
	// Repo here is project ID
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/pipeline", req.Repo)

	payload := map[string]interface{}{
		"ref":       req.Ref,
		"variables": req.Inputs,
	}

	return c.doRequest(ctx, "POST", url, req.Token, payload)
}

func (c *CIClient) doRequest(ctx context.Context, method, url, token string, body interface{}) error {
	// Simplified request logic (assumes JSON body)
	// In production, marshal body, handle errors, check status codes
	fmt.Printf("Mock Trigger: %s %s\n", method, url)
	return nil
}
