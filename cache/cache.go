package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

var SearchCacheHit = errors.New("search cache hit1")

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
func BeforeQuery(cache *ChCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {

		isC, ok := db.InstanceGet("gorm:chcache:iscache")
		if !ok || !isC.(bool) {
			db.Error = nil
			return
		}
		callbacks.BuildQuerySQL(db)
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context

		sql := db.Statement.SQL.String()
		db.InstanceSet("gorm:chcache:sql", sql)
		db.InstanceSet("gorm:chcache:vars", db.Statement.Vars)
		fmt.Println("ChCache BeforeQuery")

		cacheValue, err := cache.GetSearchCache(ctx, tableName, sql, db.Statement.Vars...)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				s := fmt.Sprintf("chcache[BeforeQuery] get cache value for sql %s error: %v", sql, err)
				fmt.Println(s)
			}
			db.Error = nil
			return
		}

		s := fmt.Sprintf("chcache[BeforeQuery] get value: %s", cacheValue)
		fmt.Println(s)
		rowsAffectedPos := strings.Index(cacheValue, "|")
		db.RowsAffected, err = strconv.ParseInt(cacheValue[:rowsAffectedPos], 10, 64)
		if err != nil {

			s := fmt.Sprintf("chcache[BeforeQuery] unmarshal rows affected cache error: %v", err)
			fmt.Println(s)
			db.Error = nil
			return
		}
		err = json.Unmarshal([]byte(cacheValue[rowsAffectedPos+1:]), db.Statement.Dest)
		if err != nil {
			s := fmt.Sprintf("chcache[BeforeQuery] unmarshal search cache error: %v", err)
			fmt.Println(s)
			db.Error = nil
			return
		}

		db.Error = SearchCacheHit

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

func AfterQuery(cache *ChCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		isC, ok := db.InstanceGet("gorm:chcache:iscache")
		if !ok || !isC.(bool) {
			db.Error = nil
			return
		}
		s, _ := db.InstanceGet("gorm:chcache:sql")
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context
		sqlObj, _ := db.InstanceGet("gorm:chcache:sql")
		sql := sqlObj.(string)
		varObj, _ := db.InstanceGet("gorm:chcache:vars")
		vars := varObj.([]interface{})
		if db.Error == nil {

			cacheBytes, err := json.Marshal(db.Statement.Dest)
			if err != nil {
				s := fmt.Sprintf("chcache[AfterQuery] cannot marshal cache for sql: %s, not cached", sql)
				fmt.Println(s)
				return
			}
			err = cache.SetSearchCache(ctx, fmt.Sprintf("%d|", db.RowsAffected)+string(cacheBytes), tableName, sql, vars...)
			if err != nil {
				s := fmt.Sprintf("chcache[AfterQuery] set search cache for sql: %s error: %v", sql, err)
				fmt.Println(s)
				return
			}

			s = fmt.Sprintf("chcache[AfterQuery] sql %s cached", sql)
			fmt.Println(s)

			return
		}
		if errors.Is(db.Error, SearchCacheHit) {
			// search cache hit
			db.Error = nil
			return
		}
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
func (c *ChCache) GetSearchCache(ctx context.Context, tableName string, sql string, vars ...interface{}) (string, error) {
	key := GenSearchCacheKey(c.InstanceId, tableName, sql, vars...)
	return c.cache.GetValue(ctx, key)
}
