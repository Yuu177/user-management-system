package geerpc

import (
	"fmt"
	"html/template"
	"log"
	"mySql/dao"
	"net/http"
	"token/tokenUtil"
)

type debugLogin struct{}

func (server *debugLogin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("/Users/tpy/codeTest/userSystem/rpc/tmpl/login.tmpl")
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}
	// 利用给定数据渲染模板, 并将结果写入w
	// tmpl.Execute(w, "小明")
	tmpl.Execute(w, nil)
}

type debugUserProfile struct{}

type userProfile struct {
	Name string
	Pwd  string
}

func (server *debugUserProfile) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("/Users/tpy/codeTest/userSystem/rpc/tmpl/user_profile.tmpl")
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}

	// read cookie token
	c, err := req.Cookie("tpy-token")
	if err == nil {
		userName, err := tokenUtil.GetValueByToken(c.Value)
		if err == nil {
			tmpl.Execute(w, &userProfile{Name: userName, Pwd: "is have token"})
			return
		}
		log.Println("redis token not exist: ", c.Value)
	} else {
		log.Println("not exist tpy-token")
	}

	name := req.FormValue("name")
	pwd := req.FormValue("pwd")
	user := &userProfile{Name: name, Pwd: pwd}
	log.Println("get user: ", user)
	token := tokenUtil.GenerateToken(name)
	tokenUtil.SaveToken(token, name) // redis, key: token, value: userName
	// set token to cookie
	cookie := http.Cookie{Name: "tpy-token", Value: token}
	http.SetCookie(w, &cookie)
	// 利用给定数据渲染模板, 并将结果写入 w
	tmpl.Execute(w, user)
}

type UserInfo struct {
	Name     string
	Nickname string
	Password string
	Picture  string
}

var db = dao.GetDB()

func insert(info *UserInfo) {
	db.Create(info)
}

func update(info *UserInfo) {
	db.Save(info)
}

func selectByUserName(name string) UserInfo {
	var info UserInfo
	db.Find(&info, "name=?", name)
	return info
}

type registUser struct{}

func (server *registUser) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	
}
