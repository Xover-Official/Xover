package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackClient sends notifications to Slack
type SlackClient struct {
	webhookURL string
}

// NewSlackClient creates a new Slack client
func NewSlackClient(webhookURL string) *SlackClient {
	return &SlackClient{webhookURL: webhookURL}
}

// SendOptimizationNotification sends a notification about an optimization
func (s *SlackClient) SendOptimizationNotification(resource, action string, savings float64, risk float64) error {
	message := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "🤖 Talos Optimization Alert",
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Resource:*\n`%s`", resource)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Action:*\n%s", action)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Savings:*\n$%.2f/mo", savings)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Risk:*\n%.1f/10", risk)},
				},
			},
		},
	}

	return s.sendMessage(message)
}

func (s *SlackClient) sendMessage(message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

// JiraClient creates tasks in Jira
type JiraClient struct {
	baseURL  string
	username string
	apiToken string
	project  string
}

// NewJiraClient creates a new Jira client
func NewJiraClient(baseURL, username, apiToken, project string) *JiraClient {
	return &JiraClient{
		baseURL:  baseURL,
		username: username,
		apiToken: apiToken,
		project:  project,
	}
}

// CreateOptimizationTask creates a Jira task for an optimization
func (j *JiraClient) CreateOptimizationTask(resource, action string, savings float64) (string, error) {
	issue := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.project,
			},
			"summary":     fmt.Sprintf("Cloud Optimization: %s", resource),
			"description": fmt.Sprintf("Talos recommends: %s\nEstimated savings: $%.2f/mo", action, savings),
			"issuetype": map[string]string{
				"name": "Task",
			},
			"labels": []string{"talos", "cloud-optimization", "cost-savings"},
		},
	}

	payload, err := json.Marshal(issue)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", j.baseURL+"/rest/api/3/issue", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(j.username, j.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if key, ok := result["key"].(string); ok {
		return key, nil
	}

	return "", fmt.Errorf("failed to create issue")
}

// ClickUpClient creates tasks in ClickUp
type ClickUpClient struct {
	apiToken string
	listID   string
}

// NewClickUpClient creates a new ClickUp client
func NewClickUpClient(apiToken, listID string) *ClickUpClient {
	return &ClickUpClient{
		apiToken: apiToken,
		listID:   listID,
	}
}

// CreateOptimizationTask creates a ClickUp task
func (c *ClickUpClient) CreateOptimizationTask(resource, action string, savings float64, risk float64) (string, error) {
	task := map[string]interface{}{
		"name":        fmt.Sprintf("Optimize: %s", resource),
		"description": fmt.Sprintf("**Action:** %s\n**Savings:** $%.2f/mo\n**Risk:** %.1f/10", action, savings, risk),
		"tags":        []string{"talos", "optimization"},
		"priority":    getPriority(risk),
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.clickup.com/api/v2/list/%s/task", c.listID), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("failed to create task")
}

func getPriority(risk float64) int {
	if risk < 3 {
		return 4 // Low
	} else if risk < 7 {
		return 3 // Normal
	}
	return 2 // High
}
