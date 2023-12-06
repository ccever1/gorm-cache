package cache

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ChCache struct {
	InstanceId string
	cache      DataLayerInterface
}

func GenInstanceId() string {
	charList := []byte("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().Unix())
	length := 5
	str := make([]byte, 0)
	for i := 0; i < length; i++ {
		str = append(str, charList[rand.Intn(len(charList))])
	}
	return string(str)
}
func (c *ChCache) Init() error {
	c.InstanceId = GenInstanceId()
	c.cache = &RedisLayer{}
	c.cache.Init("chcache")
	return nil
}
func NewChCache() (*ChCache, error) {
	cache := &ChCache{}
	err := cache.Init()
	return cache, err
}

func (c *ChCache) Name() string {
	return "chcache"
}
func (c *ChCache) Initialize(db *gorm.DB) (err error) {
	err = db.Callback().Query().Before("gorm:query").Register("gorm:chcache:before_query", BeforeQuery(c))
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("*").Register("gorm:chcache:after_query", AfterQuery(c))
	if err != nil {
		return err
	}
	fmt.Println("ChCache Initialize")
	return
}
func BeforeQuery(c *ChCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		sql := db.Statement.SQL.String()
		db.InstanceSet("gorm:chcache:sql", sql)
		db.InstanceSet("gorm:chcache:vars", db.Statement.Vars)
		fmt.Println("ChCache BeforeQuery")
		db.Error = nil
		return
	}
}

type RedisLayer struct {
	client    *redis.Client
	ttl       int64
	keyPrefix string

	batchExistSha string
	cleanCacheSha string
}

func AfterQuery(c *ChCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		s, _ := db.InstanceGet("gorm:chcache:sql")
		fmt.Println(s)
	}
}

func (c *ChCache) SetSearchCache(ctx context.Context, cacheValue string, tableName string,
	sql string, vars ...interface{}) error {
	key := GenSearchCacheKey(c.InstanceId, tableName, sql, vars...)
	return c.cache.SetKey(ctx, Kv{
		Key:   key,
		Value: cacheValue,
	})
}
func GenSearchCacheKey(instanceId string, tableName string, sql string, vars ...interface{}) string {
	buf := strings.Builder{}
	buf.WriteString(sql)
	for _, v := range vars {
		pv := reflect.ValueOf(v)
		if pv.Kind() == reflect.Ptr {
			buf.WriteString(fmt.Sprintf(":%v", pv.Elem()))
		} else {
			buf.WriteString(fmt.Sprintf(":%v", v))
		}
	}
	return fmt.Sprintf("%s:%s:s:%s:%s", "chcache", instanceId, tableName, buf.String())
}

type Kv struct {
	Key   string
	Value string
}

type DataLayerInterface interface {
	Init(prefix string) error
	// read

	GetValue(ctx context.Context, key string) (string, error)

	// write

	SetKey(ctx context.Context, kv Kv) error
}

func (r *RedisLayer) SetKey(ctx context.Context, kv Kv) error {
	return r.client.Set(ctx, kv.Key, kv.Value, time.Duration(RandFloatingInt64(r.ttl))*time.Millisecond).Err()
}
func RandFloatingInt64(v int64) int64 {
	randNum := rand.Float64()*0.2 + 0.9
	return int64(float64(v) * randNum)
}
func (r *RedisLayer) GetValue(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}
func (r *RedisLayer) Init(prefix string) error {
	opt := &redis.Options{Addr: "localhost:6379"}
	r.client = redis.NewClient(opt)

	r.ttl = 50000

	r.keyPrefix = prefix
	return nil
}
