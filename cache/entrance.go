package cache

import (
	"github.com/ccever1/gorm-cache/config"
	"github.com/go-redis/redis/v8"
)

func NewGormCache(cacheConfig *config.CacheConfig) (*GormCache, error) {
	cache := &GormCache{
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
