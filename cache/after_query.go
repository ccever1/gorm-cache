package cache

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ccever1/gorm-cache/util"
	"gorm.io/gorm"
)

func AfterQuery(cache *ChCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		isC, ok := db.InstanceGet("gorm:" + util.GormCachePrefix + ":iscache")
		if !ok || !isC.(bool) {
			db.Error = nil
			return
		}
		// s, _ := db.InstanceGet("gorm:"+util.GormCachePrefix+":sql")
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context
		sqlObj, _ := db.InstanceGet("gorm:" + util.GormCachePrefix + ":sql")
		sql := sqlObj.(string)
		varObj, _ := db.InstanceGet("gorm:" + util.GormCachePrefix + ":vars")
		vars := varObj.([]interface{})
		if db.Error == nil {
			cache.Logger.CtxInfo(ctx, "[AfterQuery] start to set search cache for sql: %s", sql)
			cacheBytes, err := json.Marshal(db.Statement.Dest)
			if err != nil {
				cache.Logger.CtxError(ctx, "[AfterQuery] cannot marshal cache for sql: %s, not cached", sql)
				return
			}
			cache.Logger.CtxInfo(ctx, "[AfterQuery] set cache: %v", string(cacheBytes))
			err = cache.SetSearchCache(ctx, fmt.Sprintf("%d|", db.RowsAffected)+string(cacheBytes), tableName, sql, vars...)
			if err != nil {
				cache.Logger.CtxError(ctx, "[AfterQuery] set search cache for sql: %s error: %v", sql, err)
				return
			}
			cache.Logger.CtxInfo(ctx, "[AfterQuery] sql %s cached", sql)
			return
		}
		if errors.Is(db.Error, util.SearchCacheHit) {
			// search cache hit
			db.Error = nil
			return
		}
	}
}
