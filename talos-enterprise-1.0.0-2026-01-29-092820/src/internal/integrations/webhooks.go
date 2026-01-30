package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Webhook defines a registered webhook
type Webhook struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"` // e.g., "alert", "deploy"
	Secret    string    `json:"secret"` // For signing
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

// WebhookRegistry manages webhooks
type WebhookRegistry struct {
	webhooks map[string]*Webhook
	mu       sync.RWMutex
	client   *http.Client
}

// NewWebhookRegistry creates a registry
func NewWebhookRegistry() *WebhookRegistry {
	return &WebhookRegistry{
		webhooks: make(map[string]*Webhook),
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

// Register registers a new webhook
func (r *WebhookRegistry) Register(url string, events []string, secret string) *Webhook {
	r.mu.Lock()
	defer r.mu.Unlock()

	hook := &Webhook{
		ID:        uuid.New().String(),
		URL:       url,
		Events:    events,
		Secret:    secret,
		CreatedAt: time.Now(),
		Active:    true,
	}
	r.webhooks[hook.ID] = hook
	return hook
}

// Unregister removes a webhook
func (r *WebhookRegistry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.webhooks, id)
}

// DispatchEvent sends an event to all subscribed webhooks
func (r *WebhookRegistry) DispatchEvent(ctx context.Context, eventType string, payload interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, hook := range r.webhooks {
		if !hook.Active {
			continue
		}

		// Check if subscribed
		subscribed := false
		for _, e := range hook.Events {
			if e == "*" || e == eventType {
				subscribed = true
				break
			}
		}

		if subscribed {
			go r.sendWebhook(context.Background(), hook, eventType, payload)
		}
	}
}

func (r *WebhookRegistry) sendWebhook(ctx context.Context, hook *Webhook, eventType string, payload interface{}) {
	body := map[string]interface{}{
		"id":        uuid.New().String(),
		"event":     eventType,
		"timestamp": time.Now().UTC(),
		"payload":   payload,
	}

	data, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	// Retry logic (3 times)
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", hook.URL, bytes.NewBuffer(data))
		if err != nil {
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Talos-Event", eventType)
		// Add signature header here (HMAC-SHA256) using hook.Secret

		resp, err := r.client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return // Success
		}

		if resp != nil {
			resp.Body.Close()
		}

		// Exponential backoff
		time.Sleep(time.Duration(1<<i) * time.Second)
	}

	fmt.Printf("Failed to send webhook to %s after retries\n", hook.URL)
}
