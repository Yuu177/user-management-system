package main

import (
	"geerpc/config"
	"geerpc/mysql"
	"geerpc/protocol"
	"geerpc/rpc"
	"geerpc/token"
	"log"
)

func main() {
	var services UserServices
	// //注册服务.
	// panicIfErr(geerpc.Register(&services)) //注册 user 的所有方法. like: user.loginAuth()
	// // panicIfErr(server.Register("GetProfile", GetProfile, GetProfileService))
	// // panicIfErr(server.Register("UpdateProfilePic", UpdateProfilePic, UpdateProfilePicService))
	// // panicIfErr(server.Register("UpdateNickName", UpdateNickName, UpdateNickNameService))

	// //监听并且处理连接.
	// l, err := net.Listen("tcp", config.TCPServerAddr)
	// if err != nil {
	// 	log.Println("network error:", err)
	// }

	// geerpc.Accept(l)

	s := rpc.NewServer()
	s.Register(&services)
	s.ListenAndServe(config.TCPServerAddr)
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
	log.Println("tpy...login")
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
func (s *UserServices) GetProfile(arg protocol.ReqGetProfile, reply *protocol.RespGetProfile) error {
	// 校验token
	ok, err := checkToken(arg.UserName, arg.Token)
	if err != nil {
		reply.Ret = 3
		log.Printf("tcp.getProfile: checkToken failed. usernam:%s, token:%s, err:%q\n", arg.UserName, arg.Token, err)
		return err
	}
	if !ok {
		reply.Ret = 1
		return err
	}

	user, err := mysql.GetProfile(arg.UserName)
	if err != nil {
		reply.Ret = 3
		log.Printf("mysql tcp.getProfile: mysql.GetProfile failed. username:%s, err:%q\n", user.UserName, err)
		return err
	}

	log.Printf("tcp.getProfile done. username:%s\n", user.UserName)
	reply.Ret = 0
	reply.UserName = user.UserName
	reply.NickName = user.NickName
	reply.PicName = user.PicName
	// 下面这种是错误的
	// reply = &protocol.RespGetProfile{Ret: 0, UserName: user.UserName, NickName: user.NickName, PicName: user.PicName}
	log.Println(reply)
	return nil

}

// UpdateProfilePicService 更新头像接口的实际服务(picName/FileName)，同时用于在注册时向rpc传递参数类型.
func (s *UserServices) UpdateProfilePic(arg protocol.ReqUpdateProfilePic, reply *protocol.RespUpdateProfilePic) error {
	// 校验token.
	ok, err := checkToken(arg.UserName, arg.Token)
	if err != nil {
		reply.Ret = 3
		log.Printf("tcp.updateProfilePic: checkToken failed. username:%s, token:%s, err:%q", arg.UserName, arg.Token, err)
		return err
	}
	if !ok {
		reply.Ret = 1
		return err
	}

	// 写入数据库.
	ok, err = mysql.UpdateProfilePic(arg.UserName, arg.FileName)
	if err != nil {
		reply.Ret = 3
		log.Printf("tcp.updateProfilePic: mysql.UpdateProfilePic failed. username:%s, filename:%s, err:%q", arg.UserName, arg.FileName, err)
		return err
	}
	if !ok {
		reply.Ret = 2
		return err
	}
	reply.Ret = 0
	log.Printf("tcp.updateProfilePic done. username:%s, filename:%s", arg.UserName, arg.FileName)
	return nil
}

// UpdateNickNameService 更新昵称接口的实际服务(NickName)，同时用于在注册时向rpc传递参数类型.
func (s *UserServices) UpdateNickName(arg protocol.ReqUpdateNickName, reply *protocol.RespUpdateNickName) error {
	// 校验token.
	ok, err := checkToken(arg.UserName, arg.Token)
	if err != nil {
		reply.Ret = 3
		log.Printf("tcp.updateNickName: checkToken failed. username:%s, token:%s, err:%q\n", arg.UserName, arg.Token, err)
		return err
	}
	if !ok {
		reply.Ret = 1
		return err
	}

	// 写入数据库.
	ok, err = mysql.UpdateNickName(arg.UserName, arg.NickName)
	if err != nil {
		reply.Ret = 3
		log.Printf("tcp.updateNickName: mysql.UpdateNickName failed. username:%s, nickname:%s, err:%q\n", arg.UserName, arg.NickName, err)
		return err
	}
	if !ok {
		reply.Ret = 2
		return err
	}
	reply.Ret = 0
	log.Printf("tcp.updateNickName done. username:%s, nickname:%s\n", arg.UserName, arg.NickName)
	return nil
}

// checkToken  检查Token
func checkToken(userName string, tk string) (bool, error) {
	// 压测token
	if tk == "test" {
		return true, nil
	}
	return token.CheckToken(userName, tk)
}
