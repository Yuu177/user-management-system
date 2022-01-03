package main

import (
	"testing"
	"userSystem/protocol"
)

var servicesTest UserServices

func TestLogin(t *testing.T) {
	var tests = []struct {
		req  protocol.ReqLogin
		resp protocol.RespLogin
		ret  int
	}{
		{protocol.ReqLogin{UserName: "user0", Password: "user"}, protocol.RespLogin{}, 0},
	}

	for _, test := range tests {
		err := servicesTest.Login(test.req, &test.resp)
		if test.resp.Ret != test.ret && err != nil {
			t.Errorf("Login no pass. username:%s, password:%s, ret:%d", test.req.UserName, test.req.Password, test.ret)
		}
	}
}
