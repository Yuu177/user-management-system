package geerpc

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"mySql/dao"
	"net/http"
	"token/tokenUtil"

	"github.com/jinzhu/gorm"
)

type User struct {
	Name     string
	Nickname string
	Password string
	Picture  string
}

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

func (server *debugUserProfile) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("/Users/tpy/codeTest/userSystem/rpc/tmpl/user.tmpl")
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}

	// read cookie token
	c, err := req.Cookie("tpy-token")
	if err == nil {
		userName, err := tokenUtil.GetValueByToken(c.Value)
		if err == nil {
			tmpl.Execute(w, &User{Name: userName, Password: "is have token"})
			return
		}
		log.Println("redis token not exist: ", c.Value)
	} else {
		log.Println("not exist tpy-token")
	}

	name := req.FormValue("name")
	pwd := req.FormValue("pwd")
	user := &User{Name: name, Password: pwd}
	log.Println("get user: ", user)
	token := tokenUtil.GenerateToken(name)
	tokenUtil.SaveToken(token, name) // redis, key: token, value: userName
	// set token to cookie
	cookie := http.Cookie{Name: "tpy-token", Value: token}
	http.SetCookie(w, &cookie)
	// 利用给定数据渲染模板, 并将结果写入 w
	tmpl.Execute(w, user)
}

var db = dao.GetDB()

func insert(info *User) {
	db.Create(info)
}

func update(info *User) {
	db.Save(info)
}

func selectUserByName(name string) (User, error) {
	var user User
	err := db.Where("name  = ?", name).First(&user)

	// if errors.Is(err, gorm.ErrRecordNotFound) {
	// 	log.Println("not found", name)
	// } else if err != nil {
	// 	log.Println("select error", err)
	// }

	return user, err
}

type registUser struct{}

func (server *registUser) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := req.FormValue("name")
	nickname := req.FormValue("nickname")
	password := req.FormValue("password")
	picture := req.FormValue("picture")

	var user = &User{
		Name:     name,
		Nickname: nickname,
		Password: password,
		Picture:  picture,
	}

	// 判断用户是否存在
	_, err := selectUserByName(user.Name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("not found", name)
		insert(user)
	} else if err != nil {
		log.Println("select error, please try later", err)
	} else {
		log.Println("username is exist, please change name")
	}

}
