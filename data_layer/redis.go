package data_layer

import (
	"context"
	"time"

	"github.com/ccever1/gorm-cache/config"
	"github.com/ccever1/gorm-cache/util"
	"github.com/go-redis/redis/v8"
)

type RedisLayer struct {
	client    *redis.Client
	logger    config.LoggerInterface
	keyPrefix string
}

func (r *RedisLayer) Init(conf *config.CacheConfig, prefix string) error {
	if conf.RedisConfig.Mode == config.RedisConfigModeOptions {
		r.client = redis.NewClient(conf.RedisConfig.Options)
	} else {
		r.client = conf.RedisConfig.Client
	}
	r.logger = conf.DebugLogger
	r.logger.SetIsDebug(conf.DebugMode)
	r.keyPrefix = prefix
	return nil
}
func (r *RedisLayer) GetValue(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisLayer) SetKey(ctx context.Context, kv util.Kv) error {
	ttl := kv.TTL
	return r.client.Set(ctx, kv.Key, kv.Value, time.Duration(util.RandFloatingInt64(ttl))*time.Millisecond).Err()
}
