package queue

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// RedisQueue implements the QueueProvider interface
type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(addr, password string) *RedisQueue {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisQueue{client: rdb}
}

func (q *RedisQueue) GetQueueDepth() int {
	val, err := q.client.LLen(context.Background(), "talos:task_queue").Result()
	if err != nil {
		return 0
	}
	return int(val)
}