package mysql

import (
	"fmt"
	"geerpc/protocol"
	"geerpc/utils"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/gorm"
)

// db连接
var db *gorm.DB

func setup() {
	conn, err := gorm.Open("mysql", "user:user@(127.0.0.1:3306)/testdb01?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Print(err.Error())
	}
	sqlDB, err := conn.DB()
	if err != nil {
		fmt.Println("connect db server failed.")
	}
	sqlDB.SetMaxIdleConns(10)                   // sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxOpenConns(100)                  // sets the maximum number of open connections to the database.
	sqlDB.SetConnMaxLifetime(time.Second * 600) // sets the maximum amount of time a connection may be reused.
	db = conn
}

func getDB() *gorm.DB {
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Errorf("connect db server failed.")
		setup()
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		setup()
	}
	return db
}

func init() {
	db = getDB()
	db.AutoMigrate(&protocol.User{})
}

// 创建一个账号
func CreateAccount(user *protocol.User) error {
	err := db.Create(&user)
	if err != nil {
		return err
	}
	return err
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
