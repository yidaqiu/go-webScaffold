package redis

import (
	"context"
	"fmt"
	"ginframe/webScaffold/settings"

	"github.com/go-redis/redis/v8"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

func Init(redisConfig *settings.RedisConfig) (err error) {
	if err = initClient(redisConfig); err != nil {
		fmt.Printf("init redis client failed, err:%v\n", err)
		return
	}
	fmt.Println("connect redis success...")

	// 最后记得关闭连接，释放资源
	defer rdb.Close()

	return
}

// 初始化连接
func initClient(redisConfig *settings.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			redisConfig.Host,
			redisConfig.Port),
		Password: redisConfig.Password, // no password set
		DB:       redisConfig.Db,       // use default DB
		PoolSize: redisConfig.PoolSize,
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	rdb.Close()
}
