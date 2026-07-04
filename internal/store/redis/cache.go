package redisstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type JSONCache struct {
	client *redis.Client
}

func NewJSONCache(client *redis.Client) *JSONCache {
	return &JSONCache{client: client}
}

func (c *JSONCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal(value, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (c *JSONCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *JSONCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

func (c *JSONCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
