package backup

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
	"pg-data-tool/internal/utils"

	_ "github.com/lib/pq"
)

// PerformBackup 执行数据库备份
func PerformBackup(cfg *config.Config) error {
	logger.Info("开始执行数据库备份操作")
	logger.Info("连接参数: 主机=%s, 端口=%s, 用户=%s", cfg.Host, cfg.Port, cfg.User)
	logger.Info("备份格式: %s", cfg.Format)

	// 检查PostgreSQL系统数据库连接
	if err := utils.CheckPostgresConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password); err != nil {
		return err
	}

	// 创建备份目录（格式：backups-yymmddHHmiss）
	now := time.Now()
	backupDir := fmt.Sprintf("backups-%s-%s", now.Format("060102150405"), cfg.Host)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		err := fmt.Errorf("创建备份目录失败: %v", err)
		logger.Error(err.Error())
		return err
	}
	logger.Info("备份文件将保存在: %s", backupDir)

	if cfg.BackupAll {
		// 获取所有数据库列表
		databases, err := getAllDatabases(cfg)
		if err != nil {
			return err
		}

		logger.Info("开始全库备份，共发现 %d 个数据库", len(databases))
		for _, db := range databases {
			// 检查每个数据库的连接
			if err := utils.CheckDatabaseConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password, db); err != nil {
				logger.Error("数据库 %s 连接失败: %v", db, err)
				continue
			}

			if err := backupSingleDatabase(cfg, db, backupDir); err != nil {
				logger.Error("备份数据库 %s 失败: %v", db, err)
				continue
			}
		}
		logger.Info("全库备份完成")
	} else {
		if cfg.DBName == "" {
			err := fmt.Errorf("错误：必须指定数据库名称或使用 --backup-all 参数进行全库备份")
			logger.Error(err.Error())
			return err
		}

		// 检查目标数据库连接
		if err := utils.CheckDatabaseConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName); err != nil {
			return err
		}

		if err := backupSingleDatabase(cfg, cfg.DBName, backupDir); err != nil {
			return err
		}
	}

	return nil
}

// getAllDatabases 获取所有数据库列表
func getAllDatabases(cfg *config.Config) ([]string, error) {
	// 连接到postgres数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 查询所有非系统数据库
	rows, err := db.Query(`
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false 
		AND datname NOT IN ('postgres', 'template0', 'template1')
		ORDER BY datname
	`)
	if err != nil {
		return nil, fmt.Errorf("查询数据库列表失败: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			return nil, fmt.Errorf("读取数据库名称失败: %v", err)
		}
		databases = append(databases, dbname)
	}

	return databases, nil
}

// backupSingleDatabase 备份单个数据库
func backupSingleDatabase(cfg *config.Config, dbname, backupDir string) error {
	logger.Info("开始备份数据库: %s", dbname)

	// 生成备份文件名
	timestamp := time.Now().Format("150405")
	var backupFile string
	var fileExt string

	// 根据备份格式设置文件扩展名
	switch cfg.Format {
	case "custom":
		fileExt = ".backup"
	case "plain":
		fileExt = ".sql"
	case "directory":
		fileExt = ".dir"
	case "tar":
		fileExt = ".tar"
	default:
		return fmt.Errorf("不支持的备份格式: %s", cfg.Format)
	}

	backupFile = filepath.Join(backupDir, fmt.Sprintf("%s_%s%s", dbname, timestamp, fileExt))

	// 构建pg_dump命令
	args := []string{
		"-h", cfg.Host,
		"-p", cfg.Port,
		"-U", cfg.User,
		"-F", cfg.Format,
		"-v",
		"--encoding=UTF8",
	}

	// 根据格式添加特定参数
	switch cfg.Format {
	case "custom":
		args = append(args, "-b") // 包含大对象
	case "directory":
		// 创建目录
		if err := os.MkdirAll(backupFile, 0755); err != nil {
			return fmt.Errorf("创建备份目录失败: %v", err)
		}
	}

	// 添加输出文件参数
	args = append(args, "-f", backupFile, dbname)

	cmd := exec.Command("pg_dump", args...)
	logger.Info("执行命令: pg_dump %v", args)

	// 设置环境变量
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Errorf("备份数据库 %s 失败: %v\n错误输出: %s", dbname, err, output)
		logger.Error(errMsg.Error())
		return errMsg
	}

	logger.Info("数据库 %s 备份成功，文件保存在: %s", dbname, backupFile)
	return nil
}
