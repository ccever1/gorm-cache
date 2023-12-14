package config

import (
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Options *redis.Options
	Client  *redis.Client

	once sync.Once
}

func (c *RedisConfig) InitClient() *redis.Client {
	c.once.Do(func() {
		c.Client = redis.NewClient(c.Options)
	})
	return c.Client
}
