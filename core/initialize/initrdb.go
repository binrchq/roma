package initialize

import (
	"binrc.com/roma/core/global"
	"github.com/redis/go-redis/v9"
)

func InitRDB() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     global.CONFIG.Database.RdbUrl,
		Password: global.CONFIG.Database.RdbPasswd, // no password set
		DB:       0,                                // use default DB
	})
	return rdb
}
