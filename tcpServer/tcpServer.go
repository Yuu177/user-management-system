package main

import (
	"geerpc"
	"geerpc/config"
	"geerpc/mysql"
	"geerpc/protocol"
	"geerpc/token"
	"log"
	"net"
)

func main() {
	var u user
	//注册服务.
	panicIfErr(geerpc.Register(&u)) //注册 user 的所有方法. like: user.loginAuth()
	// panicIfErr(server.Register("GetProfile", GetProfile, GetProfileService))
	// panicIfErr(server.Register("UpdateProfilePic", UpdateProfilePic, UpdateProfilePicService))
	// panicIfErr(server.Register("UpdateNickName", UpdateNickName, UpdateNickNameService))

	//监听并且处理连接.
	l, err := net.Listen("tcp", config.TCPServerAddr)
	if err != nil {
		log.Fatal("network error:", err)
	}

	geerpc.Accept(l)
}

// panicIfErr 错误包裹函数.
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type user struct{}

// SignUp 注册接口.
func (u *user) SignUp(v interface{}) interface{} {
	return SignUpService(*v.(*protocol.User))
}

// Login 登录接口.
func (u *user) Login(v interface{}) interface{} {
	return LoginService(*v.(*protocol.ReqLogin))
}

// // GetProfile 获取信息接口.
// func GetProfile(v interface{}) interface{} {
// 	return GetProfileService(*v.(*protocol.ReqGetProfile))
// }

// // UpdateProfilePic 更新头像接口.
// func UpdateProfilePic(v interface{}) interface{} {
// 	return UpdateProfilePicService(*v.(*protocol.ReqUpdateProfilePic))
// }

// // UpdateNickName 更新昵称接口
// func UpdateNickName(v interface{}) interface{} {
// 	return UpdateNickNameService(*v.(*protocol.ReqUpdateNickName))
// }

// SignUpService 注册接口的实际服务，同时用于在注册时向rpc传递参数类型.
func SignUpService(user protocol.User) (reply protocol.RespSignUp) {
	if user.UserName == "" || user.Password == "" {
		reply.Ret = 1
		return
	}
	if user.NickName == "" {
		user.NickName = user.UserName
	}

	if err := mysql.CreateAccount(&user); err != nil {
		reply.Ret = 2
		log.Fatal("tcp.signUp: mysql.CreateAccount failed. usernam:%s, err:%q", arg.UserName, err)
		return
	}

	reply.Ret = 0
	return
}

// LoginService 登录接口的实际服务，同时用于在注册时向rpc传递参数类型.
func LoginService(arg protocol.ReqLogin) (reply protocol.RespLogin) {
	ok, err := mysql.LoginAuth(arg.UserName, arg.Password)
	if err != nil {
		reply.Ret = 2
		log.Fatal("tcp.login: mysql.LoginAuth failed. usernam:%s, err:%q", arg.UserName, err)
		return
	}
	//账号或密码不正确.
	if !ok {
		reply.Ret = 1
		return
	}
	token := token.GenerateToken(arg.UserName)
	err = redis.SetToken(arg.UserName, token, int64(config.TokenMaxExTime))
	if err != nil {
		reply.Ret = 2
		log.Fatal("tcp.login: redis.SetToken failed. usernam:%s, token:%s, err:%q", arg.UserName, token, err)
		return
	}
	reply.Ret = 0
	reply.Token = token
	log.Infof("tcp.login: login done. username:%s", arg.UserName)
	return
}

// // GetProfileService 获取信息接口的实际服务，同时用于在注册时向rpc传递参数类型.
// func GetProfileService(arg protocol.ReqGetProfile) (reply protocol.RespGetProfile) {
// 	// 校验token
// 	ok, err := checkToken(arg.UserName, arg.Token)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.getProfile: checkToken failed. usernam:%s, token:%s, err:%q", arg.UserName, arg.Token, err)
// 		return
// 	}
// 	if !ok {
// 		reply.Ret = 1
// 		return
// 	}

// 	// 先尝试从redis取数据.
// 	nickName, picName, hasData, err := redis.GetProfile(arg.UserName)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.getProfile: redis.GetProfile failed. username:%s, err:%q", arg.UserName, err)
// 		return
// 	}
// 	if hasData {
// 		log.Infof("redis tcp.getProfile done. username:%s", arg.UserName)
// 		return protocol.RespGetProfile{Ret: 0, UserName: arg.UserName, NickName: nickName, PicName: picName}
// 	}

// 	//redis没有数据，从mysql里取.
// 	nickName, picName, hasData, err = mysql.GetProfile(arg.UserName)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("mysql tcp.getProfile: mysql.GetProfile failed. username:%s, err:%q", arg.UserName, err)
// 		return
// 	}
// 	if hasData {
// 		// 向redis插入数据.
// 		redis.SetNickNameAndPicName(arg.UserName, nickName, picName)
// 	} else {
// 		reply.Ret = 2
// 		log.Fatal("tcp.getProfile: mysql.GetProfile can't find username. username:%s", arg.UserName)
// 		return
// 	}
// 	log.Infof("tcp.getProfile done. username:%s", arg.UserName)
// 	return protocol.RespGetProfile{Ret: 0, UserName: arg.UserName, NickName: nickName, PicName: picName}

// }

// // UpdateProfilePicService 更新头像接口的实际服务(picName/FileName)，同时用于在注册时向rpc传递参数类型.
// func UpdateProfilePicService(arg protocol.ReqUpdateProfilePic) (reply protocol.RespUpdateProfilePic) {
// 	// 校验token.
// 	ok, err := checkToken(arg.UserName, arg.Token)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.updateProfilePic: checkToken failed. username:%s, token:%s, err:%q", arg.UserName, arg.Token, err)
// 		return
// 	}
// 	if !ok {
// 		reply.Ret = 1
// 		return
// 	}

// 	// 使redis对应的数据失效（由于数据将会被修改）.
// 	if err := redis.InvaildCache(arg.UserName); err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.updateProfilePic: redis.InvaildCache failed. username:%s, err:%q", arg.UserName, err)
// 		return
// 	}
// 	// 写入数据库.
// 	ok, err = mysql.UpdateProfilePic(arg.UserName, arg.FileName)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.updateProfilePic: mysql.UpdateProfilePic failed. username:%s, filename:%s, err:%q", arg.UserName, arg.FileName, err)
// 		return
// 	}
// 	if !ok {
// 		reply.Ret = 2
// 		return
// 	}
// 	reply.Ret = 0
// 	log.Infof("tcp.updateProfilePic done. username:%s, filename:%s", arg.UserName, arg.FileName)
// 	return
// }

// // UpdateNickNameService 更新昵称接口的实际服务(NickName)，同时用于在注册时向rpc传递参数类型.
// func UpdateNickNameService(arg protocol.ReqUpdateNickName) (reply protocol.RespUpdateNickName) {
// 	// 校验token.
// 	ok, err := checkToken(arg.UserName, arg.Token)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.updateNickName: checkToken failed. username:%s, token:%s, err:%q", arg.UserName, arg.Token, err)
// 		return
// 	}
// 	if !ok {
// 		reply.Ret = 1
// 		return
// 	}
// 	// 使redis对应的数据失效（由于数据将会被修改）.
// 	if err := redis.InvaildCache(arg.UserName); err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.updateNickName: redis.InvaildCache failed. username:%s, err:%q", arg.UserName, err)
// 		return
// 	}
// 	// 写入数据库.
// 	ok, err = mysql.UpdateNikcName(arg.UserName, arg.NickName)
// 	if err != nil {
// 		reply.Ret = 3
// 		log.Fatal("tcp.updateNickName: mysql.UpdateNikcName failed. username:%s, nickname:%s, err:%q", arg.UserName, arg.NickName, err)
// 		return
// 	}
// 	if !ok {
// 		reply.Ret = 2
// 		return
// 	}
// 	reply.Ret = 0
// 	log.Infof("tcp.updateNickName done. username:%s, nickname:%s", arg.UserName, arg.NickName)
// 	return
// }

//checkToken  检查Token
func checkToken(userName string, token string) (bool, error) {
	// 压测token
	if token == "test" {
		return true, nil
	}
	return redis.CheckToken(userName, token)
}
