package memory

import (
	"context"
	"errors"
)

type Cache struct {
	cache map[string][]byte
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[string][]byte),
	}
}

func (c *Cache) Get(_ context.Context, key string) ([]byte, error) {
	if value, ok := c.cache[key]; ok {
		return value, nil
	}

	return nil, errors.New("key not found")
}

func (c *Cache) Set(_ context.Context, key string, value []byte) error {
	c.cache[key] = value
	return nil
}

func (c *Cache) Clear(_ context.Context) error {
	c.cache = make(map[string][]byte)
	return nil
}
