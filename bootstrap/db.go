package bootstrap

import (
	"goblog/pkg/model"
	"time"
)

func SetupDB() {
	// 建立数据库连接池
	db := model.ConnectDB()

	// 命令行打印数据库请求的信息
	sqlDB, _ := db.DB()

	// 设置最大连接数
	sqlDB.SetMaxOpenConns(100)

	// 最大空闲
	sqlDB.SetMaxIdleConns(25)
	// 过期时间
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
}
