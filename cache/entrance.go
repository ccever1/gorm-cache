package cache

import "github.com/ccever1/ch-cache/config"

func NewChCache(cacheConfig *config.CacheConfig) (*ChCache, error) {
	cache := &ChCache{
		Config: cacheConfig,
	}
	err := cache.Init()
	return cache, err
}
