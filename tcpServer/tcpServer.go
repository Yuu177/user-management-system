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
		resp.Ret = protocol.UserNameOrPasswordNull
		return nil
	}
	if req.NickName == "" {
		req.NickName = req.UserName
	}

	if err := mysql.CreateUser(req.UserName, req.NickName, req.Password); err != nil {
		resp.Ret = protocol.UserCreateFailed
		log.Printf("tcp.signUp: mysql.CreateUser failed. usernam:%s, err:%q\n", req.UserName, err)
		return nil
	}

	resp.Ret = protocol.Success
	return nil
}

// 登陆
func (s *UserServices) Login(req protocol.ReqLogin, resp *protocol.RespLogin) error {
	ok, err := loginAuth(req)
	if err != nil {
		resp.Ret = protocol.LoginFailed // 内部出错
		return nil
	}
	if !ok {
		resp.Ret = protocol.UserNameOrPasswordError // 用户名或密码错误
		return nil
	}
	// 登陆成功，更新 redis 中的数据
	token := utils.GetToken(req.UserName)
	err = updateCache(req, token)
	if err != nil {
		resp.Ret = protocol.LoginFailed
		return nil
	}

	// 一切正常
	resp.Ret = protocol.Success
	resp.Token = token
	log.Printf("tcp.login: login done. username:%s\n", req.UserName)
	return nil
}

func updateCache(req protocol.ReqLogin, token string) error {
	redis.SetPassword(req.UserName, req.Password)
	err := redis.SetToken(req.UserName, token, int64(config.MaxExTime))
	if err != nil {
		log.Printf("tcp.login: redis.SetToken failed. usernam:%s, token:%s, err:%q\n", req.UserName, token, err)
		return nil
	}
	return nil
}

func loginAuth(req protocol.ReqLogin) (bool, error) {
	// 先从 redis 中验证，如果 redis 通过就不用验证了。
	ok := redis.LoginAuth(req.UserName, req.Password)
	if ok {
		log.Printf("redis ok, username:%s\n", req.UserName)
		return true, nil
	}

	// 如果 redis 不通过（redis 中没有数据或者数据有问题），再从 mysql 中查看
	ok, err := mysql.LoginAuth(req.UserName, req.Password)
	if err != nil {
		log.Printf("tcp.login: mysql.LoginAuth failed. username:%s, err:%q\n", req.UserName, err)
		return false, nil
	}
	return ok, nil
}

// 获取用户基本信息
func (s *UserServices) GetProfile(req protocol.ReqGetProfile, resp *protocol.RespGetProfile) error {
	// 校验token
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = protocol.GetProfileFailed
		log.Printf("tcp.getProfile: checkToken failed. usernam:%s, token:%s, err:%q\n", req.UserName, req.Token, err)
		return nil
	}
	if !ok {
		resp.Ret = protocol.TokenCheckFailed
		return nil
	}

	userProfile, isRead := getUserProfile(req)
	if !isRead {
		resp.Ret = protocol.DataIsNil
		return nil
	}

	log.Printf("tcp.getProfile done. username:%s\n", req.UserName)
	*resp = protocol.RespGetProfile{Ret: protocol.Success, UserName: req.UserName, NickName: userProfile.NickName, PicName: userProfile.PicName}
	return nil
}

func getUserProfile(req protocol.ReqGetProfile) (protocol.UserProfile, bool) {
	// 先尝试从redis取数据
	userProfile, isRead := redis.GetProfile(req.UserName)
	if isRead {
		// redis 中读取到数据
		log.Printf("redis tcp.getProfile done. username:%s\n", req.UserName)
		return userProfile, true
	}

	// redis 没有读取到数据，从 mysql 里取
	userProfile, isRead = mysql.GetProfile(req.UserName)
	if !isRead {
		log.Printf("mysql tcp.getProfile: mysql.GetProfile failed. username:%s\n", req.UserName)
		return protocol.UserProfile{}, false
	}

	// 向 redis 插入数据
	redis.SetNickNameAndPicName(req.UserName, userProfile.NickName, userProfile.PicName)

	return userProfile, true
}

// 更新头像
func (s *UserServices) UpdateProfilePic(req protocol.ReqUpdateProfilePic, resp *protocol.RespUpdateProfilePic) error {
	// 校验token.
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = protocol.UpdateFailed
		log.Printf("tcp.updateProfilePic: checkToken failed. username:%s, token:%s, err:%q\n", req.UserName, req.Token, err)
		return nil
	}
	if !ok {
		resp.Ret = protocol.TokenCheckFailed
		return nil
	}

	// 使 redis 对应的数据失效（由于数据将会被修改）
	if err := redis.InvaildCache(req.UserName); err != nil {
		resp.Ret = protocol.UpdateFailed
		log.Printf("tcp.updateProfilePic: redis.InvaildCache failed. username:%s, err:%q\n", req.UserName, err)
		return nil
	}

	// 写入数据库
	ok, err = mysql.UpdateProfilePic(req.UserName, req.FileName)
	if err != nil {
		resp.Ret = protocol.UpdateFailed
		log.Printf("tcp.updateProfilePic: mysql.UpdateProfilePic failed. username:%s, filename:%s, err:%q\n", req.UserName, req.FileName, err)
		return nil
	}
	if !ok {
		resp.Ret = protocol.UserNotExist
		return nil
	}
	resp.Ret = protocol.Success
	log.Printf("tcp.updateProfilePic done. username:%s, filename:%s\n", req.UserName, req.FileName)
	return nil
}

// 更新昵称
func (s *UserServices) UpdateNickName(req protocol.ReqUpdateNickName, resp *protocol.RespUpdateNickName) error {
	// 校验token
	ok, err := checkToken(req.UserName, req.Token)
	if err != nil {
		resp.Ret = protocol.UpdateFailed
		log.Printf("tcp.updateNickName: checkToken failed. username:%s, token:%s, err:%q\n", req.UserName, req.Token, err)
		return nil
	}
	if !ok {
		resp.Ret = protocol.TokenCheckFailed
		return nil
	}
	// 使 redis 对应的数据失效（由于数据将会被修改）
	if err := redis.InvaildCache(req.UserName); err != nil {
		resp.Ret = protocol.UpdateFailed
		log.Printf("tcp.updateNickName: redis.InvaildCache failed. username:%s, err:%q\n", req.UserName, err)
		return nil
	}
	// 写入数据库
	ok, err = mysql.UpdateNickName(req.UserName, req.NickName)
	if err != nil {
		resp.Ret = protocol.UpdateFailed
		log.Printf("tcp.updateNickName: mysql.UpdateNickName failed. username:%s, nickname:%s, err:%q\n", req.UserName, req.NickName, err)
		return nil
	}
	if !ok {
		resp.Ret = protocol.UserNotExist
		return nil
	}
	resp.Ret = protocol.Success
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
