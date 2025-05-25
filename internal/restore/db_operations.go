package restore

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
)

// CreateDatabaseIfNotExists 如果数据库不存在则创建
func CreateDatabaseIfNotExists(cfg *config.Config, dbName string) error {
	// 连接到postgres数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("连接postgres数据库失败: %v", err)
	}
	defer db.Close()

	// 检查数据库是否存在
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pg_database WHERE datname = $1
		)
	`, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查数据库是否存在失败: %v", err)
	}

	if !exists {
		logger.Info("数据库 %s 不存在，正在创建...", dbName)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("创建数据库失败: %v", err)
		}
		logger.Info("数据库 %s 创建成功", dbName)
	} else {
		logger.Info("数据库 %s 已存在", dbName)
	}

	return nil
}

// RestoreSingleFile 还原单个备份文件到指定数据库
func RestoreSingleFile(cfg *config.Config, backupFile, dbName string) error {
	// 构建pg_restore命令
	args := []string{
		"-h", cfg.Host,
		"-p", cfg.Port,
		"-U", cfg.User,
		"-d", dbName,
		"-v",
		"--clean",     // 在还原前清除数据库对象
		"--if-exists", // 如果对象不存在则不报错
	}

	// 根据文件扩展名添加特定参数
	ext := strings.ToLower(filepath.Ext(backupFile))
	switch ext {
	case ".sql":
		// SQL文件使用psql命令
		args := []string{
			"-h", cfg.Host,
			"-p", cfg.Port,
			"-U", cfg.User,
			"-d", dbName,
			"-f", backupFile,
		}
		cmd := exec.Command("psql", args...)
		logger.Info("执行命令: psql -h %s -p %s -U %s -d %s -f %s", cfg.Host, cfg.Port, cfg.User, dbName, backupFile)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("还原失败: %v\n错误输出: %s", err, output)
		}
		return nil
	case ".backup", ".tar", ".dir":
		// 其他格式使用pg_restore命令
		args = append(args, backupFile)
		cmd := exec.Command("pg_restore", args...)
		logger.Info("执行命令: pg_restore %v", args)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("还原失败: %v\n错误输出: %s", err, output)
		}
		return nil
	default:
		return fmt.Errorf("不支持的备份文件格式: %s", ext)
	}
}
