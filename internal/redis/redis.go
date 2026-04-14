package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis interface {
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Set(ctx context.Context, key, value string, exp time.Duration) error
	SetNX(ctx context.Context, key, value string, exp time.Duration) (bool, error)
	HSet(ctx context.Context, key string, value map[string]any, exp time.Duration) error
	HGet(ctx context.Context, key, field string) (string, error)
	Publish(ctx context.Context, channel, message string) error
	Subscribe(ctx context.Context, channel string) *redis.PubSub
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, exp time.Duration) error
	MGet(ctx context.Context, keys []string) ([]any, error)
	MSet(ctx context.Context, data map[string]any) error
	Scan(ctx context.Context, match string, count int64) ([]string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Close() error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisInstance(client *redis.Client) Redis {
	return &redisClient{client: client}
}

func (r *redisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (r *redisClient) Set(ctx context.Context, key, value string, exp time.Duration) error {
	err := r.client.Set(ctx, key, value, exp).Err()
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func (r *redisClient) SetNX(ctx context.Context, key, value string, exp time.Duration) (bool, error) {
	res, err := r.client.SetNX(ctx, key, value, exp).Result()
	if err != nil {
		return res, fmt.Errorf("redis: %w", err)
	}

	return res, nil
}

func (r *redisClient) HSet(ctx context.Context, key string, value map[string]any, exp time.Duration) error {
	err := r.client.HSet(ctx, key, value).Err()
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	err = r.client.Expire(ctx, key, exp).Err()
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func (r *redisClient) HGet(ctx context.Context, key, field string) (string, error) {
	value, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		return "", fmt.Errorf("redis: %w", err)
	}

	return value, nil
}

func (r *redisClient) Publish(ctx context.Context, channel, message string) error {
	err := r.client.Publish(ctx, channel, message).Err()
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func (r *redisClient) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	pubsub := r.client.Subscribe(ctx, channel)
	return pubsub
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis: %w", err)
	}

	return value, nil
}

func (r *redisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	res, err := r.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: %w", err)
	}

	return res, nil
}

func (r *redisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	found, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: %w", err)
	}

	return found, nil
}

func (r *redisClient) Incr(ctx context.Context, key string) (int64, error) {
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: %w", err)
	}

	return count, nil
}

func (r *redisClient) Expire(ctx context.Context, key string, exp time.Duration) error {
	_, err := r.client.Expire(ctx, key, exp).Result()
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func (r *redisClient) MGet(ctx context.Context, keys []string) ([]any, error) {
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	if len(values) == 1 && values[0] == nil {
		return nil, nil
	}

	return values, nil
}

func (r *redisClient) MSet(ctx context.Context, data map[string]any) error {
	_, err := r.client.MSet(ctx, data).Result()
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func (r *redisClient) Scan(ctx context.Context, match string, count int64) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		k, c, err := r.client.Scan(ctx, cursor, match, count).Result()
		if err != nil {
			return nil, fmt.Errorf("redis: %w", err)
		}

		keys = append(keys, k...)
		if c == 0 {
			break
		}

		cursor = c
	}

	return keys, nil
}

func (r *redisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}

	return keys, nil
}

func (r *redisClient) Close() error {
	return r.client.Close()
}
