package dao

import (
	"fmt"
	"log"
	"report/src/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// db连接
var db *gorm.DB

// setup 初始化连接
func setup() {
	conn, err := gorm.Open("mysql", "user:user@(127.0.0.1:3306)/testdb01?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Print(err.Error())
	}
	sqlDB, err := conn.DB()
	if err != nil {
		fmt.Error("connect db server failed.")
	}
	sqlDB.SetMaxIdleConns(10)                   // sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxOpenConns(100)                  // sets the maximum number of open connections to the database.
	sqlDB.SetConnMaxLifetime(time.Second * 600) // sets the maximum amount of time a connection may be reused.
	db = conn
}

// GetDB 开放给外部获得db连接
func GetDB() *gorm.DB {
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
