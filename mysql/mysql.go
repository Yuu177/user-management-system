package mysql

import (
	"errors"
	"fmt"
	"geerpc/protocol"
	"geerpc/utils"

	// "github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mysql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	// _ "github.com/go-sql-driver/mysql"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	// "gorm.io/gorm"
)

// db连接
var db *gorm.DB

// 包初始化函数，可以用来初始化 gorm
func init() {
	// 配置 dsn
	// 账号
	username := "user"
	// 密码
	password := "user"
	// mysql 服务地址
	host := "127.0.0.1"
	// 端口
	port := 3306
	// 数据库名
	Dbname := "testdb01"

	// 拼接 mysql dsn，即拼接数据源，下方 {} 中的替换参数即可
	// {username}:{password}@tcp({host}:{port})/{Dbname}?charset=utf8&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=60s
	// timeout 是连接超时时间，readTimeout 是读超时时间，writeTimeout 是写超时时间，可以不填
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)

	// err
	var err error
	// 连接 mysql 获取 db 实例
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}

	// 设置数据库连接池参数
	sqlDB, _ := db.DB()
	// 设置数据库连接池最大连接数
	sqlDB.SetMaxOpenConns(100)
	// 连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭
	sqlDB.SetMaxIdleConns(20)

	db.AutoMigrate(&protocol.User{}) // 自动建表
}

// 获取 gorm db，其他包调用此方法即可拿到 db
// 无需担心不同协程并发时使用这个 db 对象会公用一个连接，因为 db 在调用其方法时候会从数据库连接池获取新的连接
func GetDB() *gorm.DB {
	return db
}

// 创建一个账号
func CreateAccount(user *protocol.User) error {
	if err := db.Create(&user).Error; err != nil {
		fmt.Println("插入失败", err)
		return err
	}
	return nil
}

// 登陆验证
func LoginAuth(userName string, password string) (bool, error) {
	var user protocol.User
	db.Where("UserName = ?", userName).First(&user)
	pwd := utils.Sha256(password)
	if user.Password == pwd {
		return true, nil
	}
	return false, nil
}

// GetProfile 获取用户信息.
func GetProfile(userName string) (protocol.User, error) {
	var user protocol.User
	err := db.Where("UserName = ?", userName).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return protocol.User{}, err
	}
	return user, nil
}
