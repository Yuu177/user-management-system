package token

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	redis "geerpc/redis"
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

func SaveToken(token string, value string, exp int64) error {
	return redis.Set(token, value, exp)
}

func CheckToken(input string, token string) (bool, error) {
	val, err := redis.Get(token)
	if err != nil {
		return false, err
	}

	if input == val {
		return true, nil
	}

	return false, nil
}
