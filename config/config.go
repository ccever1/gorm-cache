package config

type CacheConfig struct {
	RedisConfig *RedisConfig
	DebugMode   bool
	DebugLogger LoggerInterface
}
