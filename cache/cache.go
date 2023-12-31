package cache

import (
	"context"

	"github.com/ccever1/gorm-cache/config"
	"github.com/ccever1/gorm-cache/data_layer"
	"github.com/ccever1/gorm-cache/util"
	"gorm.io/gorm"
)

type GormCache struct {
	Config     *config.CacheConfig
	Logger     config.LoggerInterface
	InstanceId string
	cache      data_layer.DataLayerInterface
}

func (c *GormCache) Name() string {
	return util.GormCachePrefix
}

func (c *GormCache) Initialize(db *gorm.DB) (err error) {
	err = db.Callback().Query().Before("gorm:query").Register("gorm:"+util.GormCachePrefix+":before_query", BeforeQuery(c))
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("*").Register("gorm:"+util.GormCachePrefix+":after_query", AfterQuery(c))
	if err != nil {
		return err
	}
	//fmt.Println("GormCache Initialize")
	return
}

func (c *GormCache) Init() error {
	if c.Config.RedisConfig == nil {
		panic("please init redis config!")
	}
	c.Config.RedisConfig.InitClient()
	c.InstanceId = util.GenInstanceId()
	prefix := util.GormCachePrefix + ":" + c.InstanceId
	c.cache = &data_layer.RedisLayer{}
	if c.Config.DebugLogger == nil {
		c.Config.DebugLogger = &config.DefaultLoggerImpl{}
	}
	c.Logger = c.Config.DebugLogger
	c.Logger.SetIsDebug(c.Config.DebugMode)

	err := c.cache.Init(c.Config, prefix)
	if err != nil {
		c.Logger.CtxError(context.Background(), "[Init] cache init error: %v", err)
		return err
	}
	return nil
}

func (c *GormCache) SetSearchCache(ctx context.Context, cacheValue string, ttl int64, tableName string,
	sql string, vars ...interface{}) error {
	key := util.GenSearchCacheKey(c.InstanceId, tableName, sql, vars...)
	return c.cache.SetKey(ctx, util.Kv{
		Key:   key,
		Value: cacheValue,
		TTL:   ttl,
	})
}

func (c *GormCache) GetSearchCache(ctx context.Context, tableName string, sql string, vars ...interface{}) (string, error) {
	key := util.GenSearchCacheKey(c.InstanceId, tableName, sql, vars...)
	return c.cache.GetValue(ctx, key)
}
