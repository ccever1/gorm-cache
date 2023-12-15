package data_layer

import (
	"context"

	"github.com/ccever1/gorm-cache/config"
	"github.com/ccever1/gorm-cache/util"
)

type DataLayerInterface interface {
	Init(config *config.CacheConfig, prefix string) error
	// read

	GetValue(ctx context.Context, key string) (string, error)

	// write

	SetKey(ctx context.Context, kv util.Kv) error
}
