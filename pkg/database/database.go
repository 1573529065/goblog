package database

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"goblog/pkg/logger"
	"time"
)

var DB *sql.DB

func Initialeze() {
	initDB()
	createTables()
}

func initDB() {
	var err error
	config := mysql.Config{
		User:                 "root",
		Passwd:               "123456",
		Addr:                 "192.168.1.133",
		Net:                  "tcp",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}
	fmt.Println(config.FormatDSN())

	//准备连接数据库数据池
	DB, err = sql.Open("mysql", config.FormatDSN())
	logger.LogError(err)

	// 设置最大连接数
	DB.SetMaxOpenConns(25)

	// 设置最大空闲连接数
	DB.SetMaxIdleConns(25)
	// 设置每个连接的过期时间
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 尝试连接, 失败会报错
	err = DB.Ping()
	logger.LogError(err)

}

func createTables() {
	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
	id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
	title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
	body longtext COLLATE utf8mb4_unicode_ci
	);
	`
	_, err := DB.Exec(createArticlesSQL)
	logger.LogError(err)
}
