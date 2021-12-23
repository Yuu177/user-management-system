package mySqlUtil

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// UserInfo 用户信息
type UserInfo struct {
	Name     string
	Nickname string
	Password string
	Picture  string // 保存用户头像路径
}

var db *gorm.DB

func NewConnection() *gorm.DB {
	conn, err := gorm.Open("mysql", "user:user@(127.0.0.1:3306)/testdb01?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Print(err.Error())
	}
	return conn
}

func Setup() {
	db = NewConnection()
	db.DB().SetMaxIdleConns(10)                   //最大空闲连接数
	db.DB().SetMaxOpenConns(30)                   //最大连接数
	db.DB().SetConnMaxLifetime(time.Second * 300) //设置连接空闲超时
	//db.LogMode(true)
}

func GetDB() *gorm.DB {
	if err := db.DB().Ping(); err != nil {
	   db.Close()
	   db = NewConnection()
	}
	return db
 }

func main() {
	db, err := gorm.Open("mysql", "user:user@(127.0.0.1:3306)/testdb01?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 自动迁移。这个时候 UserInfo 的数据库表会自动建立
	db.AutoMigrate(&UserInfo{})

	u1 := UserInfo{1, "枯藤", "男", "篮球"}
	u2 := UserInfo{2, "topgoer.com", "女", "足球"}
	// 创建记录（插入）
	db.Create(&u1)
	db.Create(&u2)

	// 查询
	var u = new(UserInfo)
	db.First(u) // 查询出来的数据保存到 u 中。
	fmt.Printf("%#v\n", u)

	var uu UserInfo
	db.Find(&uu, "hobby=?", "足球")
	fmt.Printf("%#v\n", uu)

	// 更新
	db.Model(&u).Update("hobby", "双色球") // 只更新 hobby 字段
	// 更新整条记录
	// u.Gender = "女"
	// u.Hobby = "football"
	// db.Save(&u)

	// 删除
	db.Delete(&u)
}
