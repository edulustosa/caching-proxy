package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type OriginResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

// Interface representing caching operations
type Cache interface {
	Get(ctx context.Context, key string) (OriginResponse, error)
	Set(ctx context.Context, key string, value OriginResponse) error
	Clear(ctx context.Context) error
}

type MemoryCache struct {
	cache map[string]OriginResponse
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: make(map[string]OriginResponse),
	}
}

func (c *MemoryCache) Get(_ context.Context, key string) (OriginResponse, error) {
	if value, ok := c.cache[key]; ok {
		return value, nil
	}

	return OriginResponse{}, errors.New("key not found")
}

func (c *MemoryCache) Set(_ context.Context, key string, value OriginResponse) error {
	c.cache[key] = value
	return nil
}

func (c *MemoryCache) Clear(_ context.Context) error {
	c.cache = make(map[string]OriginResponse)
	return nil
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(url string) (*RedisCache, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("error parsing redis url: %w", err)
	}

	client := redis.NewClient(opt)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("error connecting to redis: %w", err)
	}

	return &RedisCache{client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (OriginResponse, error) {
	val, err := c.client.JSONGet(ctx, key, ".").Result()
	if err != nil {
		return OriginResponse{}, err
	}

	var resp OriginResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return OriginResponse{}, err
	}

	return resp, nil
}

func (c *RedisCache) Set(
	ctx context.Context,
	key string,
	value OriginResponse,
) error {
	return c.client.JSONSet(ctx, key, ".", value).Err()
}

func (c *RedisCache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}
