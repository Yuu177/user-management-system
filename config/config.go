package config

import "time"

const (
	// redis 地址.
	RedisAddr string = "localhost:6379"
	// 连接 redis 最多的连接(Maximum number of socket connections.).
	RedisPoolSize int = 30
	// 生存时间（token 和 cookie）
	MaxExTime int = 30

	// 连接数据库地址.
	MysqlDB string = "user:user@tcp(127.0.0.1:3306)/testdb01?charset=utf8&parseTime=True&loc=Local"
	// 数据库一个连接的最大生命周期.
	// ConnMaxLifetime time.Duration = 2 * time.Second
	ConnMaxLifetime time.Duration = 14400 * time.Second
	// 连接池中最大空闲连接数.
	MaxIdleConns int = 2000
	// 同时连接数据库中最多连接数.
	MaxOpenConns int = 2000

	// TCP 服务日志.
	TCPServerLogPath string = "./log/tcp_server.log"
	// TCPServerAddr tcp server ip:port.
	TCPServerAddr string = "[127.0.0.1]:8888"
	// 客户端tcp连接池大小.
	TCPClientPoolSize int = 2000

	// HTTP服务日志.
	HTTPServerLogPath string = "./log/http_server.log"
	// HTTP服务地址.
	HTTPServerAddr string = "127.0.0.1:1088"

	// 静态文件服务地址.
	StaticFilePath string = "../static/"

	// 默认头像.
	DefaultImagePath string = "tpy.jpeg"
)
