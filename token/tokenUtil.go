package tokenUtil

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	redisUtil "geerpc/redis"
	"math/rand"
	"time"
)

const (
	magicNum = "token-tpy"
)

func createRandomNumber() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func getCurTime() string {
	now := time.Now()
	dateStr := fmt.Sprintf("%d%d%d%d%d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	return dateStr
}

func GenerateToken(userName string) string {
	var token string
	token += magicNum + "-"
	token += userName + "-"
	token += getCurTime() + "-"
	token += createRandomNumber()
	return md5V(token)
}

func SaveToken(token, value string) {
	var exp time.Duration = 10 * time.Second
	redisUtil.Set(token, value, exp) // 暂时设置秒
}

func GetValueByToken(token string) (string, error) {
	return redisUtil.Get(token)
}
