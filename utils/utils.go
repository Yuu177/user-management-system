package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"geerpc/redis"
	"math/rand"
	"time"
)

const (
	magicNum = "token-tpy"
)

func createRandomNumber() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

func getCurTime() string {
	now := time.Now()
	dateStr := fmt.Sprintf("%d%d%d%d%d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	return dateStr
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha256(passwd string) string {
	rh := sha256.New()
	rh.Write([]byte(passwd))
	return hex.EncodeToString(rh.Sum(nil))
}

func GenerateToken(userName string) string {
	var token string
	token += magicNum + "-"
	token += userName + "-"
	token += getCurTime() + "-"
	token += createRandomNumber()
	return MD5(token)
}

func SaveToken(token, value string) {
	var exp time.Duration = 10 * time.Second
	redis.Set(token, value, exp) // 暂时设置秒
}

func GetValueByToken(token string) (string, error) {
	return redis.Get(token)
}
