package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient abstract client for Redis (Cluster or Standalone)
type RedisClient struct {
	Client redis.UniversalClient
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addrs    []string // Host:Port addresses
	Password string
	DB       int
	Master   string // For Sentinel
	Cluster  bool
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	var rdb redis.UniversalClient

	if config.Cluster {
		// Cluster mode
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    config.Addrs,
			Password: config.Password,
		})
	} else if config.Master != "" {
		// Sentinel mode
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    config.Master,
			SentinelAddrs: config.Addrs,
			Password:      config.Password,
			DB:            config.DB,
		})
	} else {
		// Standalone mode
		rdb = redis.NewClient(&redis.Options{
			Addr:     config.Addrs[0],
			Password: config.Password,
			DB:       config.DB,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{Client: rdb}, nil
}

// Set stores a value
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonVal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, key, jsonVal, expiration).Err()
}

// Get retrieves a value
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return errors.New("key not found")
	} else if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// Close closes the connection
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
