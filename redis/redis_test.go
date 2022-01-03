package redis

import "testing"

// TestSetToken 测试SetToken函数.
func TestSetToken(t *testing.T) {
	var tests = []struct {
		userName string
		token    string
		exp      int64
	}{
		{"user0", "token", 5},
	}
	for _, test := range tests {
		if err := SetToken(test.userName, test.token, test.exp); err != nil {
			t.Errorf("SetToken not pass. userName:%s, token:%s, exp:%d, err:%q", test.userName, test.token, test.exp, err)
		}
	}
}

func TestCheckToken(t *testing.T) {
	var tests = []struct {
		userName string
		token    string
		ok       bool
	}{
		{"user0", "token", true},
		{"user1", "token1", false},
	}
	for _, test := range tests {
		if ok, err := CheckToken(test.userName, test.token); err != nil || ok != test.ok {
			t.Errorf("CheckToken no pass. userName:%s, token:%s, ok:%t, err:%q", test.userName, test.token, test.ok, err)
		}
	}
}
