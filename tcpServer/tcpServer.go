package main

import (
	"log"
	"userSystem/config"
	"userSystem/mysql"
	"userSystem/protocol"
	"userSystem/redis"
	"userSystem/rpc"
	"userSystem/utils"
)

type UserServices struct{}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var services UserServices
	s := rpc.NewServer()
	s.Register(&services)
	s.ListenAndServe(config.TCPServerAddr)
}

// 注册
func (s *UserServices) SignUp(req protocol.ReqSignUp, resp *protocol.RespSignUp) error {
	if req.UserName == "" || req.Password == "" {
		resp.Ret = 1
		return nil
	}
	if req.NickName == "" {
		req.NickName = req.UserName
	}

	if err := mysql.CreateUser(req.UserName, req.NickName, req.Password); err != nil {
		resp.Ret = 2
		log.Printf("tcp.signUp: mysql.CreateUser failed. usernam:%s, err:%q\n", req.UserName, err)
		return err
	}

	// if err := mysql.CreateAccount(req.UserName, req.Password); err != nil {
	// 	resp.Ret = 2
	// 	log.Printf("tcp.signUp: mysql.CreateAccount failed. usernam:%s, err:%q\n", req.UserName, err)
	// 	return err
	// }
	// if err := mysql.CreateProfile(req.UserName, req.NickName); err != nil {
	// 	resp.Ret = 2
	// 	log.Printf("tcp.signUp: mysql.CreateProfile failed. usernam:%s, err:%q\n", req.UserName, err)
	// 	return err
	// }

	resp.Ret = 0
	return nil
}

// 登陆
func (s *UserServices) Login(req protocol.ReqLogin, resp *protocol.RespLogin) error {
	var err error
	var ok bool
	if jump := redis.LoginAuth(req.UserName, req.Password); jump {
		goto OK_LOGIN
	}

	ok, err = mysql.LoginAuth(req.UserName, req.Password)
	if err != nil {
		resp.Ret = 2
		log.Printf("tcp.login: mysql.LoginAuth failed. usernam:%s, err:%q\n", req.UserName, err)
		return err
	}
	//账号或密码不正确.
	if !ok {
		resp.Ret = 1
		return nil
	}
OK_LOGIN:
	redis.SetPassword(req.UserName, req.Password)
	token := utils.GetToken(req.UserName)
	err = redis.SetToken(req.UserName, token, int64(config.MaxExTime))
	if err != nil {
		resp.Ret = 2
		log.Printf("tcp.login: redis.SetToken failed. usernam:%s, token:%s, err:%q\n", req.UserName, token, err)
		return err
	}
	resp.Ret = 0
	resp.Token = token
	log.Printf("tcp.login: login done. username:%s\n", req.UserName)
	return nil
}

// 获取用户基本信息
func (s *UserServices) GetProfile(req protocol.ReqGetProfile, resp *protocol.RespGetProfile) error {
	// 校验token
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = 3
		log.Printf("tcp.getProfile: checkToken failed. usernam:%s, token:%s, err:%q\n", req.UserName, req.Token, err)
		return err
	}
	if !ok {
		resp.Ret = 1
		return nil
	}

	// 先尝试从redis取数据
	userProfile, isRead := redis.GetProfile(req.UserName)
	if isRead {
		// redis 中读取到数据
		log.Printf("redis tcp.getProfile done. username:%s\n", req.UserName)
		*resp = protocol.RespGetProfile{Ret: 0, UserName: req.UserName, NickName: userProfile.NickName, PicName: userProfile.PicName}
		return nil
	}

	// redis 没有读取到数据，从 mysql 里取
	userProfile, isRead = mysql.GetProfile(req.UserName)
	if !isRead {
		resp.Ret = 2
		log.Printf("mysql tcp.getProfile: mysql.GetProfile failed. username:%s, err:%q\n", req.UserName, err)
		return err
	}
	// 向 redis 插入数据
	redis.SetNickNameAndPicName(req.UserName, userProfile.NickName, userProfile.PicName)

	log.Printf("tcp.getProfile done. username:%s\n", req.UserName)
	*resp = protocol.RespGetProfile{Ret: 0, UserName: req.UserName, NickName: userProfile.NickName, PicName: userProfile.PicName}
	return nil
}

// 更新头像
func (s *UserServices) UpdateProfilePic(req protocol.ReqUpdateProfilePic, resp *protocol.RespUpdateProfilePic) error {
	// 校验token.
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = 3
		log.Printf("tcp.updateProfilePic: checkToken failed. username:%s, token:%s, err:%q\n", req.UserName, req.Token, err)
		return err
	}
	if !ok {
		resp.Ret = 1
		return nil
	}

	// 使 redis 对应的数据失效（由于数据将会被修改）
	if err := redis.InvaildCache(req.UserName); err != nil {
		resp.Ret = 3
		log.Printf("tcp.updateProfilePic: redis.InvaildCache failed. username:%s, err:%q\n", req.UserName, err)
		return err
	}

	// 写入数据库
	ok, err = mysql.UpdateProfilePic(req.UserName, req.FileName)
	if err != nil {
		resp.Ret = 3
		log.Printf("tcp.updateProfilePic: mysql.UpdateProfilePic failed. username:%s, filename:%s, err:%q\n", req.UserName, req.FileName, err)
		return err
	}
	if !ok {
		resp.Ret = 2
		return nil
	}
	resp.Ret = 0
	log.Printf("tcp.updateProfilePic done. username:%s, filename:%s\n", req.UserName, req.FileName)
	return nil
}

// 更新昵称
func (s *UserServices) UpdateNickName(req protocol.ReqUpdateNickName, resp *protocol.RespUpdateNickName) error {
	// 校验token
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = 3
		log.Printf("tcp.updateNickName: checkToken failed. username:%s, token:%s, err:%q\n", req.UserName, req.Token, err)
		return err
	}
	if !ok {
		resp.Ret = 1
		return nil
	}
	// 使 redis 对应的数据失效（由于数据将会被修改）
	if err := redis.InvaildCache(req.UserName); err != nil {
		resp.Ret = 3
		log.Printf("tcp.updateNickName: redis.InvaildCache failed. username:%s, err:%q\n", req.UserName, err)
		return err
	}
	// 写入数据库
	ok, err = mysql.UpdateNickName(req.UserName, req.NickName)
	if err != nil {
		resp.Ret = 3
		log.Printf("tcp.updateNickName: mysql.UpdateNickName failed. username:%s, nickname:%s, err:%q\n", req.UserName, req.NickName, err)
		return err
	}
	if !ok {
		resp.Ret = 2
		return nil
	}
	resp.Ret = 0
	log.Printf("tcp.updateNickName done. username:%s, nickname:%s\n", req.UserName, req.NickName)
	return nil
}

// 检查Token
func checkToken(userName string, token string) (bool, error) {
	// 压测token
	if token == "test" {
		return true, nil
	}
	return redis.CheckToken(userName, token)
}
