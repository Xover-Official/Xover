package ai

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements caching for AI responses
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr string, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
		prefix: "talos:ai:",
	}, nil
}

// Get retrieves a cached AI response
func (c *RedisCache) Get(ctx context.Context, prompt string) (*CachedResponse, error) {
	key := c.makeKey(prompt)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var cached CachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	return &cached, nil
}

// Set stores an AI response in cache
func (c *RedisCache) Set(ctx context.Context, prompt string, response *AIResponse) error {
	key := c.makeKey(prompt)

	cached := CachedResponse{
		Response:  response,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// makeKey creates a cache key from the prompt
func (c *RedisCache) makeKey(prompt string) string {
	hash := md5.Sum([]byte(prompt))
	return c.prefix + hex.EncodeToString(hash[:])
}

// GetStats returns cache statistics
func (c *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info := c.client.Info(ctx, "stats")
	keyspace := c.client.Info(ctx, "keyspace")

	return map[string]interface{}{
		"stats":    info.Val(),
		"keyspace": keyspace.Val(),
	}, nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// CachedResponse wraps an AI response with cache metadata
type CachedResponse struct {
	Response  *AIResponse
	CachedAt  time.Time
	ExpiresAt time.Time
}
