package util

import "errors"

var SearchCacheHit = errors.New("search cache hit1")

type Kv struct {
	Key   string
	Value string
	TTL   int64
}

const (
	GormCachePrefix = "chcache"
	GormCacheTTL    = "gorm:" + GormCachePrefix + ":iscache"
)
