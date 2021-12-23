package redisUtil

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

func init() {
	initClient()
}

// 初始化连接
func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		fmt.Println("redis can not conn")
		return err
	}
	return nil
}

// expire 失效时间。0 是永久
func Set(key string, value interface{}, expiration time.Duration) error { // 这里 value 还是换成 string 好一点感觉。因为这里反正存的是string
	err := rdb.Set(key, value, expiration).Err()
	if err != nil {
		log.Println("RedisSet Error! key:", key, "Details:", err.Error())
		return err
	}
	return nil
}

func Get(key string) (string, error) {
	val, err := rdb.Get(key).Result()
	if err != nil {
		fmt.Printf("redisUtil failed, err: %v\n", err)
		return "", err
	}
	return val, err
}

func Delete(key string) error {
	err := rdb.Del(key).Err()
	if err != nil {
		log.Println("redis delete ", key, " Error. Details: ", err.Error())
	}
	return err
}
