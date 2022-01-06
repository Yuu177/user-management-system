package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"text/template"
	"userSystem/config"
	"userSystem/protocol"
	"userSystem/rpc"
	"userSystem/utils"
)

// 模版参数
var loginTemplate *template.Template
var profileTemplate *template.Template
var jumpTemplate *template.Template

// 用于向 login.html 模版传递参数.
type LoginResponse struct {
	Msg string
}

// 用于向 profile.html 模版传递参数.
type ProfileResponse struct {
	UserName string
	NickName string
	PicName  string
}

// 用于向 jump.html 模版传递参数.
type JumpResponse struct {
	Msg string
}

var rpcClient rpc.RPCClient

// init 提前解析 html 文件。程序用到即可直接使用，避免多次解析
func init() {
	loginTemplate = template.Must(template.ParseFiles("../templates/login.html"))
	profileTemplate = template.Must(template.ParseFiles("../templates/profile.html"))
	jumpTemplate = template.Must(template.ParseFiles("../templates/jump.html"))
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 初始化 rpc 客户端并且连接 rpc 服务器
	var err error
	rpcClient, err = rpc.Client(config.TCPClientPoolSize, config.TCPServerAddr)
	if err != nil {
		panic(err)
	}
	// 静态文件服务
	// 让文件服务器使用 config.StaticFilePath 目录下的文件，响应 url以 /static/ 开头的 http请求
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticFilePath))))

	// 注册 http 请求对应的处理函数
	// http.HandleFunc("/", GetProfile) 如果路由为 / 那么其他的 /abcd 请求也会到这里
	http.HandleFunc("/index", GetProfile)
	http.HandleFunc("/signUp", SignUp)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/profile", GetProfile)
	http.HandleFunc("/updateNickName", UpdateNickName)
	http.HandleFunc("/uploadFile", UploadProfilePicture)

	http.ListenAndServe(config.HTTPServerAddr, nil)
}

// ============================ /
// 注册账号
func SignUp(rw http.ResponseWriter, req *http.Request) {
	// 处理 http post方法
	if req.Method == "POST" {
		arg, err := signUpReqConvRpcArg(rw, req)
		if err != nil {
			return
		}
		reply := protocol.RespSignUp{}
		if err = callSignUp(rw, arg, &reply); err != nil {
			return
		}
		handleSignUpRet(rw, arg, reply)
	}
}

// 登录接口
func Login(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		arg, err := loginReqConvRpcArg(rw, req)
		if err != nil {
			return
		}
		reply := protocol.RespLogin{}
		if err = callLogin(rw, arg, &reply); err != nil {
			return
		}
		handleLoginRet(rw, arg, reply)
	}
}

// 获得用户信息
func GetProfile(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		arg, err := getProfileReqConvRpcArg(rw, req)
		if err != nil {
			return
		}
		reply := protocol.RespGetProfile{}
		if err = callGetProfile(rw, arg, &reply); err != nil {
			return
		}
		handleGetProfileRet(rw, arg, reply)
	}
}

// 更新昵称
func UpdateNickName(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		arg, err := updateNickNameReqConvRpcArg(rw, req)
		if err != nil {
			return
		}
		reply := protocol.RespUpdateNickName{}
		if err = callUpdateNickName(rw, arg, &reply); err != nil {
			return
		}
		handleUpdateNickNameRet(rw, arg, reply)
	}
}

// 上传并更新头像
func UploadProfilePicture(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		arg, err := uploadProfilePictureReqConvRpcArg(rw, req)
		if err != nil {
			return
		}
		reply := protocol.RespUpdateProfilePic{}
		if err = callUploadProfilePicture(rw, arg, &reply); err != nil {
			return
		}
		handleUploadProfilePictureRet(rw, arg, reply)
	}
}

// =============================== //

// 调用远程 rpc 服务, 将数据存入到数据库
func callSignUp(rw http.ResponseWriter, arg protocol.ReqSignUp, reply *protocol.RespSignUp) error {
	if err := rpcClient.Call("UserServices.SignUp", arg, reply); err != nil {
		log.Printf("http.SignUp: Call SignUp failed. username:%s, err:%q\n", arg.UserName, err)
		rw.Write([]byte("创建账号失败！"))
		return err
	}
	return nil
}

func handleSignUpRet(rw http.ResponseWriter, arg protocol.ReqSignUp, reply protocol.RespSignUp) {
	switch reply.Ret {
	case 0:
		rw.Write([]byte("创建账号成功！"))
	case 1:
		rw.Write([]byte("用户名或密码错误！"))
	default:
		rw.Write([]byte("创建账号失败！"))
	}
	log.Printf("http.SignUp: SignUp done. username:%s, ret:%d\n", arg.UserName, reply.Ret)
}

func signUpReqConvRpcArg(rw http.ResponseWriter, req *http.Request) (protocol.ReqSignUp, error) {
	// 获取请求各个字段值
	userName := req.FormValue("username")
	password := req.FormValue("password")
	nickName := req.FormValue("nickname")

	if userName == "" || password == "" {
		rw.Write([]byte("username and password couldn't be NULL!"))
		return protocol.ReqSignUp{}, errors.New("username and password couldn't be NULL")
	}
	log.Printf("userName = %s, password = %s, nickName = %s\n", userName, password, nickName)
	arg := protocol.ReqSignUp{
		UserName: userName,
		Password: password,
		NickName: nickName,
	}
	return arg, nil
}

// 调用远程 rpc 服务, 主要对登陆账号密码进行验证
func callLogin(rw http.ResponseWriter, arg protocol.ReqLogin, reply *protocol.RespLogin) error {
	if err := rpcClient.Call("UserServices.Login", arg, &reply); err != nil {
		log.Printf("http.Login: Call Login failed. username:%s, err:%q\n", arg.UserName, err)
		// 重新登录
		templateLogin(rw, LoginResponse{Msg: "登录失败！"})
		return err
	}
	return nil
}

func handleLoginRet(rw http.ResponseWriter, arg protocol.ReqLogin, reply protocol.RespLogin) {
	switch reply.Ret {
	case 0:
		// 登陆成功将 username,token 作为 Cookies 发送给客户端
		cookie := http.Cookie{Name: "username", Value: arg.UserName, MaxAge: config.MaxExTime}
		http.SetCookie(rw, &cookie)
		cookie = http.Cookie{Name: "token", Value: reply.Token, MaxAge: config.MaxExTime}
		http.SetCookie(rw, &cookie)

		templateJump(rw, JumpResponse{Msg: "登录成功！"})
	case 1:
		templateLogin(rw, LoginResponse{Msg: "用户名或密码错误！"})
	default:
		templateLogin(rw, LoginResponse{Msg: "登录失败！"})
	}
	log.Printf("http.Login: Login done. username:%s, ret:%d\n", arg.UserName, reply.Ret)
}

func loginReqConvRpcArg(rw http.ResponseWriter, req *http.Request) (protocol.ReqLogin, error) {
	userName := req.FormValue("username")
	password := req.FormValue("password")
	if userName == "" || password == "" {
		// 重新登录
		templateLogin(rw, LoginResponse{Msg: "用户名和密码不能为空！"})
		return protocol.ReqLogin{}, errors.New("用户名和密码不能为空")
	}

	arg := protocol.ReqLogin{
		UserName: userName,
		Password: password,
	}

	return arg, nil
}

// 调用远程 rpc 服务, 获取用户对应的信息
func callGetProfile(rw http.ResponseWriter, arg protocol.ReqGetProfile, reply *protocol.RespGetProfile) error {
	if err := rpcClient.Call("UserServices.GetProfile", arg, &reply); err != nil {
		log.Printf("http.GetProfile: Call GetProfile failed. username:%s, err:%q\n", arg.UserName, err)
		// templateJump(rw, JumpResponse{Msg: "获取用户信息失败！"})
		templateLogin(rw, LoginResponse{Msg: "用户登陆过期，请重新登陆"})
		return err
	}
	return nil
}

func handleGetProfileRet(rw http.ResponseWriter, arg protocol.ReqGetProfile, reply protocol.RespGetProfile) {
	switch reply.Ret {
	case 0:
		if reply.PicName == "" {
			reply.PicName = config.DefaultImagePath
		}
		log.Println(reply)
		// 将用户的信息返回给对应的用户
		templateProfile(rw, ProfileResponse{
			UserName: reply.UserName,
			NickName: reply.NickName,
			PicName:  reply.PicName})
	case 1:
		templateLogin(rw, LoginResponse{Msg: "请重新登录！"})
	case 2:
		templateJump(rw, JumpResponse{Msg: "用户不存在！"})
	default:
		templateJump(rw, JumpResponse{Msg: "获取用户信息失败！"})
	}
	log.Printf("http.GetProfile: GetProfile done. username:%s, ret:%d\n", arg.UserName, reply.Ret)
}

func getProfileReqConvRpcArg(rw http.ResponseWriter, req *http.Request) (protocol.ReqGetProfile, error) {
	// 获取 token, 没有 token 则重新登陆
	token, err := getToken(rw, req)
	if err != nil {
		return protocol.ReqGetProfile{}, err
	}
	// 获取用户名，如果为空从 cookie 获取
	userName := req.FormValue("username")
	if userName == "" {
		nameCookie, err := req.Cookie("username")
		if err != nil {
			templateLogin(rw, LoginResponse{Msg: ""}) // 如果 cookie 中也没有用户名，需要重新登陆
			return protocol.ReqGetProfile{}, errors.New("not find username")
		}
		userName = nameCookie.Value
	}

	arg := protocol.ReqGetProfile{
		UserName: userName,
		Token:    token,
	}

	return arg, nil
}

//调用远程 rpc 服务, 修改用户的 nickName 信息
func callUpdateNickName(rw http.ResponseWriter, arg protocol.ReqUpdateNickName, reply *protocol.RespUpdateNickName) error {
	if err := rpcClient.Call("UserServices.UpdateNickName", arg, &reply); err != nil {
		log.Printf("http.UpdateNickName: Call UpdateNickName failed. username:%s, err:%q", arg.UserName, err)
		templateJump(rw, JumpResponse{Msg: "修改昵称失败！"})
		return err
	}
	return nil
}

func handleUpdateNickNameRet(rw http.ResponseWriter, arg protocol.ReqUpdateNickName, reply protocol.RespUpdateNickName) {
	switch reply.Ret {
	case 0:
		templateJump(rw, JumpResponse{Msg: "修改昵称成功！"})
	case 1:
		templateLogin(rw, LoginResponse{Msg: "请重新登录！"})
	case 2:
		templateJump(rw, JumpResponse{Msg: "用户不存在！"})
	default:
		templateJump(rw, JumpResponse{Msg: "修改昵称失败！"})

	}
	log.Printf("http.UpdateNickName: UpdateNickName done. username:%s, nickname:%s, ret:%d", arg.UserName, arg.NickName, reply.Ret)
}

func updateNickNameReqConvRpcArg(rw http.ResponseWriter, req *http.Request) (protocol.ReqUpdateNickName, error) {
	// 获取 token, 没有 token 则重新登陆
	token, err := getToken(rw, req)
	if err != nil {
		return protocol.ReqUpdateNickName{}, err
	}
	userName := req.FormValue("username")
	nickName := req.FormValue("nickname")

	arg := protocol.ReqUpdateNickName{
		UserName: userName,
		NickName: nickName,
		Token:    token,
	}

	return arg, nil
}

// 调用远程rpc服务, 修改用户的头像 pickName 的路径
func callUploadProfilePicture(rw http.ResponseWriter, arg protocol.ReqUpdateProfilePic, reply *protocol.RespUpdateProfilePic) error {
	if err := rpcClient.Call("UserServices.UpdateProfilePic", arg, &reply); err != nil {
		log.Printf("http.UploadProfilePicture: Call UploadProfilePic failed. username:%s, err:%q", arg.UserName, err)
		templateJump(rw, JumpResponse{Msg: "修改头像失败！"})
		return err
	}
	return nil
}

func handleUploadProfilePictureRet(rw http.ResponseWriter, arg protocol.ReqUpdateProfilePic, reply protocol.RespUpdateProfilePic) {
	switch reply.Ret {
	case 0:
		templateJump(rw, JumpResponse{Msg: "修改头像成功！"})
	case 1:
		templateLogin(rw, LoginResponse{Msg: "请重新登录！"})
	case 2:
		templateJump(rw, JumpResponse{Msg: "用户不存在！"})
	default:
		templateJump(rw, JumpResponse{Msg: "修改头像失败！"})
	}
	log.Printf("http.UploadProfilePicture: UploadProfilePicture done. username:%s, filepath:%s, ret:%d", arg.UserName, arg.FileName, reply.Ret)
}

func uploadProfilePictureReqConvRpcArg(rw http.ResponseWriter, req *http.Request) (protocol.ReqUpdateProfilePic, error) {
	// 获取token, 没有token则重新登陆
	token, err := getToken(rw, req)
	if err != nil {
		return protocol.ReqUpdateProfilePic{}, err
	}
	userName := req.FormValue("username")
	fileName, err := getFileName(userName, rw, req)
	if err != nil {
		return protocol.ReqUpdateProfilePic{}, err
	}

	arg := protocol.ReqUpdateProfilePic{
		UserName: userName,
		FileName: fileName,
		Token:    token,
	}

	return arg, nil
}

func getToken(rw http.ResponseWriter, req *http.Request) (string, error) {
	token, err := req.Cookie("token")
	if err != nil {
		log.Printf("get token failed. err:%q", err)
		templateLogin(rw, LoginResponse{})
		return "", errors.New("get token failed")
	}

	return token.Value, nil
}

func getFileName(userName string, rw http.ResponseWriter, req *http.Request) (string, error) {
	// 获取图片文件
	file, head, err := req.FormFile("image")
	if err != nil {
		templateJump(rw, JumpResponse{Msg: "获取图片失败！"})
		log.Printf("http.UploadProfilePicture: get file name failed. username:%s, err:%q", userName, err)
		return "", errors.New("get file name failed")
	}
	defer file.Close()

	// 检测文件合法性，并且随机生成一个文件名，拷贝 newName
	newName, isLegal := utils.CheckAndCreateFileName(head.Filename)
	if !isLegal {
		templateJump(rw, JumpResponse{Msg: "文件格式不合法！"})
		return "", errors.New("文件格式不合法！")
	}
	filePath := config.StaticFilePath + newName
	fileName := newName

	dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		templateJump(rw, JumpResponse{Msg: "文件打开出错！"})
		return "", errors.New("文件打开出错！")
	}
	defer dstFile.Close()

	// 拷贝文件。拷贝上传的文件到 static 文件夹中
	_, err = io.Copy(dstFile, file)
	if err != nil {
		templateJump(rw, JumpResponse{Msg: "文件拷贝出错！"})
		return "", errors.New("文件拷贝出错！")
	}

	return fileName, nil
}

// http 登陆页面
func templateLogin(rw http.ResponseWriter, reply LoginResponse) {
	if err := loginTemplate.Execute(rw, reply); err != nil {
		log.Printf("http.templateLogin: %q\n", err)
	}
}

// http 编辑页面
func templateProfile(rw http.ResponseWriter, reply ProfileResponse) {
	if err := profileTemplate.Execute(rw, reply); err != nil {
		log.Printf("http.templateProfile: %q\n", err)
	}
}

// http 应答信息页面
func templateJump(rw http.ResponseWriter, reply JumpResponse) {
	if err := jumpTemplate.Execute(rw, reply); err != nil {
		log.Printf("http.templateJump: %q\n", err)
	}
}
