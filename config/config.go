package config

type CacheConfig struct {
	RedisConfig *RedisConfig
	CacheTTL    int64
	DebugMode   bool
	DebugLogger LoggerInterface
}
