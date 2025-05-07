package utils

import (
	"database/sql"
	"fmt"
	"time"

	"pg-data-tool/internal/logger"

	_ "github.com/lib/pq"
)

// CheckDatabaseConnection 检查数据库连接
func CheckDatabaseConnection(host, port, user, password, dbname string) error {
	logger.Info("正在检查数据库连接...")

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// 设置连接超时
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	// 设置连接超时
	db.SetConnMaxLifetime(time.Second * 5)
	db.SetMaxOpenConns(1)

	// 尝试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 检查数据库版本
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
		return fmt.Errorf("获取数据库版本失败: %v", err)
	}

	logger.Info("数据库连接成功")
	logger.Info("数据库版本: %s", version)
	return nil
}

// CheckPostgresConnection 检查PostgreSQL系统数据库连接
func CheckPostgresConnection(host, port, user, password string) error {
	return CheckDatabaseConnection(host, port, user, password, "postgres")
}
