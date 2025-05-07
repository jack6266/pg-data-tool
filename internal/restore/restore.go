package restore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
	"pg-data-tool/internal/utils"
)

// PerformRestore 执行数据库还原
func PerformRestore(cfg *config.Config) error {
	logger.Info("开始执行数据库还原操作")
	logger.Info("连接参数: 主机=%s, 端口=%s, 用户=%s", cfg.Host, cfg.Port, cfg.User)

	// 检查PostgreSQL系统数据库连接
	if err := utils.CheckPostgresConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password); err != nil {
		return err
	}

	if !cfg.RestoreAll && cfg.DBName == "" {
		err := fmt.Errorf("错误：必须指定数据库名称或使用 --restore-all 参数进行全库还原")
		logger.Error(err.Error())
		return err
	}

	if cfg.File == "" {
		err := fmt.Errorf("错误：必须指定备份文件路径")
		logger.Error(err.Error())
		return err
	}

	// 检查文件或目录是否存在
	fileInfo, err := os.Stat(cfg.File)
	if os.IsNotExist(err) {
		err := fmt.Errorf("错误：备份文件或目录 %s 不存在", cfg.File)
		logger.Error(err.Error())
		return err
	}

	if cfg.RestoreAll {
		if fileInfo.IsDir() {
			// 如果是目录，处理目录下的所有备份文件
			return restoreAllFromDirectory(cfg)
		} else {
			// 如果是单个文件，检查是否是目录格式的备份
			if strings.HasSuffix(cfg.File, ".dir") {
				return restoreAllFromDirectory(cfg)
			} else {
				return fmt.Errorf("错误：全库还原需要指定备份目录或目录格式的备份文件")
			}
		}
	} else {
		// 单库还原
		if err := utils.CheckDatabaseConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName); err != nil {
			return err
		}
		return restoreSingleDatabase(cfg)
	}
}

// restoreAllFromDirectory 从目录还原所有数据库
func restoreAllFromDirectory(cfg *config.Config) error {
	logger.Info("开始从目录还原所有数据库: %s", cfg.File)

	// 获取目录下的所有备份文件
	var backupFiles []string
	err := filepath.Walk(cfg.File, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 检查文件扩展名
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".sql" || ext == ".backup" || ext == ".tar" {
				backupFiles = append(backupFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历备份目录失败: %v", err)
	}

	if len(backupFiles) == 0 {
		return fmt.Errorf("错误：在目录 %s 中未找到有效的备份文件", cfg.File)
	}

	logger.Info("找到 %d 个备份文件", len(backupFiles))

	// 处理每个备份文件
	for _, backupFile := range backupFiles {
		// 从文件名中提取数据库名
		fileName := filepath.Base(backupFile)
		dbName := extractDatabaseName(fileName)
		if dbName == "" {
			logger.Error("无法从文件名 %s 提取数据库名，跳过此文件", fileName)
			continue
		}

		logger.Info("正在还原数据库 %s 从文件 %s", dbName, backupFile)

		// 检查数据库连接
		if err := utils.CheckDatabaseConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password, dbName); err != nil {
			logger.Error("数据库 %s 连接失败: %v", dbName, err)
			continue
		}

		// 执行还原
		if err := restoreSingleFile(cfg, backupFile, dbName); err != nil {
			logger.Error("还原数据库 %s 失败: %v", dbName, err)
			continue
		}

		logger.Info("数据库 %s 还原成功", dbName)
	}

	logger.Info("全库还原完成")
	return nil
}

// restoreSingleDatabase 还原单个数据库
func restoreSingleDatabase(cfg *config.Config) error {
	logger.Info("开始还原数据库: %s", cfg.DBName)
	return restoreSingleFile(cfg, cfg.File, cfg.DBName)
}

// restoreSingleFile 还原单个备份文件到指定数据库
func restoreSingleFile(cfg *config.Config, backupFile, dbName string) error {
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

// extractDatabaseName 从备份文件名中提取数据库名
func extractDatabaseName(fileName string) string {
	logger.Info("正在从文件名提取数据库名: %s", fileName)

	// 移除文件扩展名
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	logger.Info("移除扩展名后的文件名: %s", baseName)

	// 找到最后一个下划线的位置
	lastUnderscoreIndex := strings.LastIndex(baseName, "_")
	if lastUnderscoreIndex == -1 {
		logger.Error("文件名 %s 中没有找到下划线分隔符", fileName)
		return ""
	}

	// 取最后一个下划线之前的部分作为数据库名
	dbName := baseName[:lastUnderscoreIndex]
	logger.Info("提取到的数据库名: %s", dbName)
	return dbName
}
