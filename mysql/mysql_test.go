package mysql

import "testing"

func TestLoginAuth(t *testing.T) {
	var tests = []struct {
		userName, password string
		ok                 bool
	}{
		{"user0", "user", true},
		{"user1", "user", true},
		{"noExist", "noExist", false},
		{"", "", false},
		{"user0", "", false},
		{"", "user", false},
	}
	for _, test := range tests {
		if ok, err := LoginAuth(test.userName, test.password); err != nil || ok != test.ok {
			t.Errorf("LoginAuth no pass. userName:%s, password:%s, ok:%t", test.userName, test.password, test.ok)
		}
	}
}
