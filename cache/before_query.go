package cache

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/ccever1/ch-cache/util"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

func BeforeQuery(cache *ChCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {

		isC, ok := db.InstanceGet("gorm:" + util.GormCachePrefix + ":iscache")
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
		db.InstanceSet("gorm:"+util.GormCachePrefix+":sql", sql)
		db.InstanceSet("gorm:"+util.GormCachePrefix+":vars", db.Statement.Vars)
		//fmt.Println("ChCache BeforeQuery")

		cacheValue, err := cache.GetSearchCache(ctx, tableName, sql, db.Statement.Vars...)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				cache.Logger.CtxError(ctx, "[BeforeQuery] get cache value for sql %s error: %v", sql, err)
			}
			db.Error = nil
			return
		}
		cache.Logger.CtxInfo(ctx, "[BeforeQuery] get value: %s", cacheValue)
		rowsAffectedPos := strings.Index(cacheValue, "|")
		db.RowsAffected, err = strconv.ParseInt(cacheValue[:rowsAffectedPos], 10, 64)
		if err != nil {
			cache.Logger.CtxError(ctx, "[BeforeQuery] unmarshal rows affected cache error: %v", err)
			db.Error = nil
			return
		}
		err = json.Unmarshal([]byte(cacheValue[rowsAffectedPos+1:]), db.Statement.Dest)
		if err != nil {
			cache.Logger.CtxError(ctx, "[BeforeQuery] unmarshal search cache error: %v", err)
			db.Error = nil
			return
		}

		db.Error = util.SearchCacheHit

		return
	}
}
