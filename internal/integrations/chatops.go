package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ChatPlatform defines supported platforms
type ChatPlatform string

const (
	PlatformSlack   ChatPlatform = "slack"
	PlatformTeams   ChatPlatform = "teams"
	PlatformDiscord ChatPlatform = "discord"
)

// ChatMessage represents a generic chat message
type ChatMessage struct {
	Platform    ChatPlatform
	WebhookURL  string
	Title       string
	Text        string
	Color       string // hex code
	Fields      []Field
	ActionURL   string
	ActionLabel string
}

type Field struct {
	Name   string
	Value  string
	Inline bool
}

// ChatOpsClient handles sending messages to chat platforms
type ChatOpsClient struct {
	client *http.Client
}

// NewChatOpsClient creates a new client
func NewChatOpsClient() *ChatOpsClient {
	return &ChatOpsClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends a message to the configured platform
func (c *ChatOpsClient) Send(ctx context.Context, msg ChatMessage) error {
	var payload interface{}
	var err error

	switch msg.Platform {
	case PlatformSlack:
		payload = c.buildSlackPayload(msg)
	case PlatformTeams:
		payload = c.buildTeamsPayload(msg)
	case PlatformDiscord:
		payload = c.buildDiscordPayload(msg)
	default:
		return fmt.Errorf("unsupported platform: %s", msg.Platform)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", msg.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed: %d", resp.StatusCode)
	}

	return nil
}

// Slack Payload Builder
func (c *ChatOpsClient) buildSlackPayload(msg ChatMessage) map[string]interface{} {
	fields := make([]map[string]interface{}, 0)
	for _, f := range msg.Fields {
		fields = append(fields, map[string]interface{}{
			"title": f.Name,
			"value": f.Value,
			"short": f.Inline,
		})
	}

	attachment := map[string]interface{}{
		"title":  msg.Title,
		"text":   msg.Text,
		"color":  msg.Color,
		"fields": fields,
		"footer": "Talos Guardian",
		"ts":     time.Now().Unix(),
	}

	if msg.ActionURL != "" {
		attachment["actions"] = []map[string]interface{}{
			{
				"type":  "button",
				"text":  msg.ActionLabel,
				"url":   msg.ActionURL,
				"style": "primary",
			},
		}
	}

	return map[string]interface{}{
		"attachments": []interface{}{attachment},
	}
}

// Teams Payload Builder (Adaptive Card)
func (c *ChatOpsClient) buildTeamsPayload(msg ChatMessage) map[string]interface{} {
	facts := make([]map[string]string, 0)
	for _, f := range msg.Fields {
		facts = append(facts, map[string]string{
			"title": f.Name,
			"value": f.Value,
		})
	}

	card := map[string]interface{}{
		"type":       "MessageCard",
		"context":    "http://schema.org/extensions",
		"themeColor": msg.Color,
		"summary":    msg.Title,
		"sections": []map[string]interface{}{
			{
				"activityTitle":    msg.Title,
				"activitySubtitle": "Talos Notification",
				"facts":            facts,
				"text":             msg.Text,
			},
		},
	}

	if msg.ActionURL != "" {
		card["potentialAction"] = []map[string]interface{}{
			{
				"@type": "OpenUri",
				"name":  msg.ActionLabel,
				"targets": []map[string]string{
					{"os": "default", "uri": msg.ActionURL},
				},
			},
		}
	}

	return card
}

// Discord Payload Builder
func (c *ChatOpsClient) buildDiscordPayload(msg ChatMessage) map[string]interface{} {
	fields := make([]map[string]interface{}, 0)
	for _, f := range msg.Fields {
		fields = append(fields, map[string]interface{}{
			"name":   f.Name,
			"value":  f.Value,
			"inline": f.Inline,
		})
	}

	// Discord uses integer colors
	color := 0x00FF00 // Default Green
	// Simplify color parsing (omitted)

	embed := map[string]interface{}{
		"title":       msg.Title,
		"description": msg.Text,
		"color":       color,
		"fields":      fields,
		"footer": map[string]string{
			"text": "Talos Guardian",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if msg.ActionURL != "" {
		embed["url"] = msg.ActionURL
	}

	return map[string]interface{}{
		"embeds": []interface{}{embed},
	}
}
