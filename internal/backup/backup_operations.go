package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
)

// CreateBackupDir 创建备份目录
func CreateBackupDir(cfg *config.Config) (string, error) {
	now := time.Now()
	backupDir := fmt.Sprintf("backups-%s-%s", now.Format("060102150405"), cfg.Host)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		err := fmt.Errorf("创建备份目录失败: %v", err)
		logger.Error(err.Error())
		return "", err
	}
	logger.Info("备份文件将保存在: %s", backupDir)
	return backupDir, nil
}

// GenerateBackupFileName 生成备份文件名
func GenerateBackupFileName(cfg *config.Config, dbname, backupDir string) (string, error) {
	timestamp := time.Now().Format("150405")
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
		return "", fmt.Errorf("不支持的备份格式: %s", cfg.Format)
	}

	backupFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s%s", dbname, timestamp, fileExt))

	// 如果是目录格式，需要创建目录
	if cfg.Format == "directory" {
		if err := os.MkdirAll(backupFile, 0755); err != nil {
			return "", fmt.Errorf("创建备份目录失败: %v", err)
		}
	}

	return backupFile, nil
}

// BuildBackupCommand 构建备份命令
func BuildBackupCommand(cfg *config.Config, dbname, backupFile string) (*exec.Cmd, error) {
	// 构建pg_dump命令
	args := []string{
		"-h", cfg.Host,
		"-p", cfg.Port,
		"-U", cfg.User,
		"-F", cfg.Format,
		"-v",
	}

	// 根据格式添加特定参数
	switch cfg.Format {
	case "custom":
		args = append(args, "-b") // 包含大对象
	case "directory":
		// 只在directory模式下启用并行处理
		if cfg.UseParallel {
			args = append(args, "-j", fmt.Sprintf("%d", cfg.ParallelJobs))
		}
	case "tar":
		// tar格式不支持并行处理
	}

	// 添加输出文件参数
	args = append(args, "-f", backupFile, dbname)

	cmd := exec.Command("pg_dump", args...)
	logger.Info("执行命令: pg_dump %v", args)

	// 设置环境变量
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	return cmd, nil
}

// BackupSingleDatabase 备份单个数据库
func BackupSingleDatabase(cfg *config.Config, dbname, backupDir string) error {
	logger.Info("开始备份数据库: %s", dbname)

	// 生成备份文件名
	backupFile, err := GenerateBackupFileName(cfg, dbname, backupDir)
	if err != nil {
		return err
	}

	// 构建并执行备份命令
	cmd, err := BuildBackupCommand(cfg, dbname, backupFile)
	if err != nil {
		return err
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("备份数据库 %s 失败: %v\n错误输出: %s", dbname, err, output)
	}

	logger.Info("数据库 %s 备份成功，文件保存在: %s", dbname, backupFile)
	return nil
}
