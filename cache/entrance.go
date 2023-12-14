package cache

import (
	"github.com/ccever1/ch-cache/config"
	"github.com/go-redis/redis/v8"
)

func NewChCache(cacheConfig *config.CacheConfig) (*ChCache, error) {
	cache := &ChCache{
		Config: cacheConfig,
	}
	err := cache.Init()
	return cache, err
}
func NewRedisConfigWithClient(client *redis.Client) *config.RedisConfig {
	return &config.RedisConfig{
		Mode:   config.RedisConfigModeRaw,
		Client: client,
	}
}
