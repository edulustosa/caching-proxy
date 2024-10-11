package cache

import (
	"context"
	"errors"
)

type OriginResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

type Cache interface {
	Get(ctx context.Context, key string) (OriginResponse, error)
	Set(ctx context.Context, key string, value OriginResponse) error
	Clear(ctx context.Context) error
}

type Memory struct {
	cache map[string]OriginResponse
}

func NewMemoryCache() *Memory {
	return &Memory{
		cache: make(map[string]OriginResponse),
	}
}

func (c *Memory) Get(_ context.Context, key string) (OriginResponse, error) {
	if value, ok := c.cache[key]; ok {
		return value, nil
	}

	return OriginResponse{}, errors.New("key not found")
}

func (c *Memory) Set(_ context.Context, key string, value OriginResponse) error {
	c.cache[key] = value
	return nil
}

func (c *Memory) Clear(_ context.Context) error {
	c.cache = make(map[string]OriginResponse)
	return nil
}
