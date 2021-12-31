package redis

import (
	"log"
	"time"
	"userSystem/config"
	"userSystem/protocol"
	"userSystem/utils"

	"github.com/go-redis/redis"
)

var rdb *redis.Client

func init() {
	initClient()
}

// 初始化连接
func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",                   // no password set
		DB:       0,                    // use default DB
		PoolSize: config.RedisPoolSize, // 设置 redis 连接池大小
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		log.Println("redis can not conn")
		return err
	}
	return nil
}

func set(key string, value interface{}, expiration int64) error {
	err := rdb.Set(key, value, time.Duration(expiration*1e9)).Err()
	if err != nil {
		log.Println("RedisSet Error! key:", key, "Details:", err.Error())
		return err
	}
	return nil
}

func get(key string) (string, error) {
	val, err := rdb.Get(key).Result()
	if err != nil {
		log.Printf("redisUtil failed, err: %v\n", err)
		return "", err
	}
	return val, err
}

// 登陆验证
func LoginAuth(userName string, password string) bool {
	var pwdKey = userName + "_pwd"
	if isExits := rdb.Exists(pwdKey).Val(); isExits == 0 { // 0 为不存在
		return false
	}
	pwd, err := get(pwdKey)
	if err != nil {
		return false
	}
	if pwd == utils.MD5(password) {
		return true
	}
	return false
}

func SetPassword(userName string, password string) {
	set(userName+"_pwd", utils.MD5(password), int64(config.MaxExTime))
}

// 获取用户信息
func GetProfile(userName string) (protocol.UserProfile, bool) {
	if isExits := rdb.Exists(userName).Val(); isExits == 0 { // 0 为不存在
		return protocol.UserProfile{}, false
	}

	vals, err := rdb.HGetAll(userName).Result()
	if err != nil {
		return protocol.UserProfile{}, false
	}
	return protocol.UserProfile{NickName: vals["nick_name"], PicName: vals["pic_name"]}, true
}

// 设置昵称和头像
func SetNickNameAndPicName(userName string, nickName string, picName string) error {
	fields := map[string]interface{}{
		"nick_name": nickName,
		"pic_name":  picName,
	}
	err := rdb.HMSet(userName, fields).Err()
	if err != nil {
		return err
	}
	return nil
}

// 删除 redis 中的数据，主要用于写入数据库之前，保持数据一致。
func InvaildCache(userName string) error {
	err := rdb.Del(userName).Err()
	if err != nil {
		return err
	}
	return nil
}

// 设置存活时间并保存 token 到 redis
func SetToken(userName string, token string, expiration int64) error {
	return set("auth_"+userName, token, expiration)
}

func CheckToken(userName string, token string) (bool, error) {
	val, err := get("auth_" + userName)
	if err != nil {
		return false, err
	}
	return token == val, nil
}
