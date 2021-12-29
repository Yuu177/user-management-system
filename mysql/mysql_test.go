package mysql

import (
	"strconv"
	"testing"
)

const (
	MaxData = 10000000
)

// 初始化测试数据库,创建 10 000 000 个数据.
func TestCreateAccount(t *testing.T) {
	for i := 0; i < MaxData; i++ {
		userName := "bot" + strconv.Itoa(i)
		if err := CreateAccount(userName, "1234"); err != nil {
			t.Errorf("CreateAccount didn't pass. username:%s, err:%q", userName, err)
		}
		if err := CreateProfile(userName, "bot"); err != nil {
			t.Errorf("CreateProfile didn't pass. username:%s, err:%q", userName, err)
		}
		if i%100 == 0 {
			t.Logf("now is %d", i)
		}
	}
}
