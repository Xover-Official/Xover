package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type MemoryEntry struct {
	ResourceID string `json:"resource_id"`
	Action     string `json:"action"`
	Reasoning  string `json:"reasoning"`
	Timestamp  string `json:"timestamp"`
}

// ProjectMemory now uses Redis for distributed persistence
type ProjectMemory struct {
	client *redis.Client
	prefix string
}

func NewProjectMemory(client *redis.Client) *ProjectMemory {
	return &ProjectMemory{
		client: client,
		prefix: "talos:memory:",
	}
}

// AddEntry pushes a new memory entry to the Redis list
func (m *ProjectMemory) AddEntry(ctx context.Context, entry MemoryEntry) error {
	if entry.ResourceID == "" {
		return fmt.Errorf("cannot add memory entry: resource ID is empty")
	}

	if entry.ResourceID == "" {
		return fmt.Errorf("cannot add memory entry: resource ID is empty")
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal memory entry: %w", err)
	}

	// Use resource-specific key to prevent history eviction collisions
	key := m.prefix + entry.ResourceID

	// Push to head and keep last 50 entries per resource (sufficient for context window)
	pipe := m.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 49)
	_, err = pipe.Exec(ctx)
	return err
}

// GetDecisionsForResource retrieves entries for a specific resource
func (m *ProjectMemory) GetDecisionsForResource(ctx context.Context, id string) ([]MemoryEntry, error) {
	if id == "" {
		return nil, fmt.Errorf("cannot retrieve decisions: resource ID is empty")
	}

	// Fetch only items for this specific resource (O(1) lookup)
	key := m.prefix + id
	items, err := m.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var matched []MemoryEntry
	for _, item := range items {
		var entry MemoryEntry
		if err := json.Unmarshal([]byte(item), &entry); err == nil {
			matched = append(matched, entry)
		}
	}
	return matched, nil
}
