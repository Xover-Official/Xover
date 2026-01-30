package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides caching for AI responses and rate limiting
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr string, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// CacheAIResponse caches an AI model response
func (r *RedisCache) CacheAIResponse(key string, response string, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, response, ttl).Err()
}

// GetCachedResponse retrieves a cached AI response
func (r *RedisCache) GetCachedResponse(key string) (string, bool) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", false // Cache miss
	}
	if err != nil {
		return "", false // Error treated as miss
	}
	return val, true
}

// IncrementRateLimit increments the rate limit counter for an API key
func (r *RedisCache) IncrementRateLimit(apiKey string) (int, error) {
	key := fmt.Sprintf("ratelimit:%s", apiKey)

	// Increment counter
	count, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiry on first increment
	if count == 1 {
		r.client.Expire(r.ctx, key, time.Minute)
	}

	return int(count), nil
}

// CheckRateLimit checks if rate limit is exceeded
func (r *RedisCache) CheckRateLimit(apiKey string, maxRequests int) (bool, error) {
	count, err := r.IncrementRateLimit(apiKey)
	if err != nil {
		return false, err
	}
	return count <= maxRequests, nil
}

// GetCacheStats returns cache statistics
func (r *RedisCache) GetCacheStats() (map[string]interface{}, error) {
	info, err := r.client.Info(r.ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	dbSize, err := r.client.DBSize(r.ctx).Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"db_size": dbSize,
		"info":    info,
	}, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GenerateCacheKey creates a consistent cache key for AI requests
func GenerateCacheKey(model string, prompt string) string {
	// Use first 100 chars of prompt to avoid key size issues
	if len(prompt) > 100 {
		prompt = prompt[:100]
	}
	return fmt.Sprintf("ai:%s:%s", model, prompt)
}
