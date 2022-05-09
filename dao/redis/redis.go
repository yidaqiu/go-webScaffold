package redis

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/go-redis/redis/v8"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

func Init() (err error) {
	if err = initClient(); err != nil {
		fmt.Printf("init redis client failed, err:%v\n", err)
		return
	}
	fmt.Println("connect redis success...")

	// 最后记得关闭连接，释放资源
	defer rdb.Close()

	return
}

// 初始化连接
func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			viper.GetString("redis.host"),
			viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"), // no password set
		DB:       viper.GetInt("redis.db"),          // use default DB
		PoolSize: viper.GetInt("redis.pool_size"),
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
