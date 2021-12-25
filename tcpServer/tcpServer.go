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
	var services UserServices
	//注册服务.
	panicIfErr(geerpc.Register(&services)) //注册 user 的所有方法. like: user.loginAuth()
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

type UserServices struct{}

// SignUpService 注册接口的实际服务，同时用于在注册时向rpc传递参数类型.
func (s *UserServices) SignUp(arg protocol.User, reply *protocol.RespSignUp) error {
	if arg.UserName == "" || arg.Password == "" {
		reply.Ret = 1
		return nil
	}
	if arg.NickName == "" {
		arg.NickName = arg.UserName
	}

	if err := mysql.CreateAccount(&arg); err != nil {
		reply.Ret = 2
		log.Printf("tcp.signUp: mysql.CreateAccount failed. usernam:%s, err:%q\n", arg.UserName, err)
		return nil
	}

	reply.Ret = 0
	return nil
}

// LoginService 登录接口的实际服务，同时用于在注册时向rpc传递参数类型.
func (s *UserServices) Login(arg protocol.ReqLogin, reply *protocol.RespLogin) error {
	ok, err := mysql.LoginAuth(arg.UserName, arg.Password)
	if err != nil {
		reply.Ret = 2
		log.Printf("tcp.login: mysql.LoginAuth failed. usernam:%s, err:%q\n", arg.UserName, err)
		return nil
	}
	//账号或密码不正确.
	if !ok {
		reply.Ret = 1
		return nil
	}
	t := token.GenerateToken(arg.UserName)
	err = token.SaveToken(t, arg.UserName, int64(config.TokenMaxExTime))
	if err != nil {
		reply.Ret = 2
		log.Printf("tcp.login: redis.SetToken failed. usernam:%s, token:%s, err:%q\n", arg.UserName, t, err)
		return nil
	}
	reply.Ret = 0
	reply.Token = t
	log.Printf("tcp.login: login done. username:%s\n", arg.UserName)
	return nil
}

// GetProfileService 获取信息接口的实际服务，同时用于在注册时向rpc传递参数类型.
func GetProfileService(arg protocol.ReqGetProfile) (reply protocol.RespGetProfile) {
	// 校验token
	ok, err := checkToken(arg.UserName, arg.Token)
	if err != nil {
		reply.Ret = 3
		log.Printf("tcp.getProfile: checkToken failed. usernam:%s, token:%s, err:%q\n", arg.UserName, arg.Token, err)
		return
	}
	if !ok {
		reply.Ret = 1
		return
	}

	//redis没有数据，从mysql里取.
	user, err := mysql.GetProfile(arg.UserName)
	if err != nil {
		reply.Ret = 3
		log.Printf("mysql tcp.getProfile: mysql.GetProfile failed. username:%s, err:%q\n", arg.UserName, err)
		return
	}
	log.Printf("tcp.getProfile done. username:%s\n", arg.UserName)
	return protocol.RespGetProfile{Ret: 0, UserName: user.UserName, NickName: user.NickName, PicName: user.PicName}

}

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

// checkToken  检查Token
func checkToken(userName string, tk string) (bool, error) {
	// 压测token
	if tk == "test" {
		return true, nil
	}
	return token.CheckToken(userName, tk)
}
