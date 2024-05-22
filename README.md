[![Apache-2.0 license](https://img.shields.io/badge/license-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

# gorm-cache

`gorm-cache` aims to provide a look-aside, almost-no-code-modification cache solution for gorm v2 users.
`gorm-cache` 是gorm的缓存中间件，在需要缓存的地方使用InstanceSet即可使用缓存。

Redis, where cached data stores in Redis (if you have multiple servers running the same procedure, they don't share the same space in Redis)

# Overview

即插即用

旁路缓存

数据源使用 Redis

会话级缓存

基于 Pacific73/gorm-cache 修改

# Install

go get github.com/ccever1/gorm-cache

## Usage

```go
package main

import (
	"github.com/ccever1/gorm-cache/cache"
	"github.com/ccever1/gorm-cache/config"
	"github.com/ccever1/gorm-cache/util"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "user:pass@tcp(127.0.0.1:3306)/database_name?charset=utf8mb4"
	lg := logger.Default.LogMode(logger.Info)
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: lg})

	redisClient := redis.NewClient(&redis.Options{
		DB:   2,
		Addr: "localhost:6379",
	})

	cache, _ := cache.NewGormCache(&config.CacheConfig{
		RedisConfig: cache.NewRedisConfigWithClient(redisClient),
		DebugMode:   true,
	})
	//More options in `config.config.go`
	db.Use(cache) // use gorm plugin

	var users []User

	dbx := db.Where("maxsuccessions > ?", 16).Session(&gorm.Session{})

	dbx.InstanceSet(util.GormCacheTTL, 5000).Find(&users) // search cache not hit, objects cached, 5000ms

	dbx.InstanceSet(util.GormCacheTTL, 5000).Find(&users) // search cache hit

	dbx.Find(&users) // do not use cache

}


type User struct {
	ID             int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`                                                // ID
	Maxsuccessions int    `gorm:"column:maxsuccessions;type:int(10);default:1;NOT NULL;comment:Maximum consecutive login days;" json:"maxsuccessions"` // 
	RealName       string `gorm:"column:real_name;comment:real name;size:50;" json:"real_name"`                                       // 
}

func (m *User) TableName() string {
	return "user"
}

```

