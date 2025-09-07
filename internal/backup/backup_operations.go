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

	_ "github.com/lib/pq"
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
	switch cfg.BackupType {
	case "logical":
		return LogicalBackupSingleDatabase(cfg, dbname, backupDir)
	case "physical":
		return PhysicalBackup(cfg, backupDir)
	case "incremental":
		return IncrementalBackup(cfg, backupDir)
	default:
		return fmt.Errorf("不支持的备份类型: %s", cfg.BackupType)
	}
}

// LogicalBackupSingleDatabase 逻辑备份单个数据库
func LogicalBackupSingleDatabase(cfg *config.Config, dbname, backupDir string) error {
	logger.Info("开始逻辑备份数据库: %s", dbname)

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
		return fmt.Errorf("逻辑备份数据库 %s 失败: %v\n错误输出: %s", dbname, err, output)
	}

	logger.Success("数据库 %s 逻辑备份成功，文件保存在: %s", dbname, backupFile)
	return nil
}

// PhysicalBackup 执行物理热备
func PhysicalBackup(cfg *config.Config, backupDir string) error {
	logger.Info("开始执行物理热备")

	// 创建物理备份目录
	physicalBackupDir := filepath.Join(backupDir, "physical_backup")
	if err := os.MkdirAll(physicalBackupDir, 0755); err != nil {
		return fmt.Errorf("创建物理备份目录失败: %v", err)
	}

	// 启用WAL归档（如果需要）
	if err := enableWALArchiving(cfg); err != nil {
		logger.Warn("启用WAL归档失败: %v", err)
	}

	// 执行pg_basebackup
	args := []string{
		"-h", cfg.Host,
		"-p", cfg.Port,
		"-U", cfg.User,
		"-D", physicalBackupDir,
		"-Ft", // tar格式
		"-z",  // 压缩
		"-P",  // 显示进度
		"-v",  // 详细输出
		"-W",  // 等待WAL归档
	}

	cmd := exec.Command("pg_basebackup", args...)
	logger.Info("执行命令: pg_basebackup %v", args)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("物理热备失败: %v\n错误输出: %s", err, output)
	}

	logger.Success("物理热备成功，备份保存在: %s", physicalBackupDir)
	return nil
}

// IncrementalBackup 执行增量备份
func IncrementalBackup(cfg *config.Config, backupDir string) error {
	logger.Info("开始执行增量备份")

	// 检查是否有基础备份
	if cfg.BaseBackupPath == "" {
		return fmt.Errorf("增量备份需要指定基础备份路径 (--base-backup-path)")
	}

	// 获取当前LSN
	currentLSN, err := getCurrentLSN(cfg)
	if err != nil {
		return fmt.Errorf("获取当前LSN失败: %v", err)
	}

	logger.Info("当前LSN: %s", currentLSN)

	// 创建增量备份目录
	incrementalBackupDir := filepath.Join(backupDir, fmt.Sprintf("incremental_%s", time.Now().Format("150405")))
	if err := os.MkdirAll(incrementalBackupDir, 0755); err != nil {
		return fmt.Errorf("创建增量备份目录失败: %v", err)
	}

	// 使用pg_receivewal接收WAL文件
	args := []string{
		"-h", cfg.Host,
		"-p", cfg.Port,
		"-U", cfg.User,
		"-D", incrementalBackupDir,
		"-v",
		"--synchronous",
	}

	// 如果有检查点LSN，添加起始位置
	if cfg.CheckpointLSN != "" {
		args = append(args, "-S", cfg.CheckpointLSN)
	}

	cmd := exec.Command("pg_receivewal", args...)
	logger.Info("执行命令: pg_receivewal %v", args)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	// 创建增量备份信息文件
	infoFile := filepath.Join(incrementalBackupDir, "backup_info.txt")
	infoContent := fmt.Sprintf("备份类型: 增量备份\n时间: %s\n基础备份: %s\n起始LSN: %s\n当前LSN: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		cfg.BaseBackupPath,
		cfg.CheckpointLSN,
		currentLSN)

	if err := os.WriteFile(infoFile, []byte(infoContent), 0644); err != nil {
		logger.Warn("写入备份信息文件失败: %v", err)
	}

	logger.Success("增量备份设置完成，WAL文件将保存在: %s", incrementalBackupDir)
	logger.Info("请在需要时停止pg_receivewal进程")

	return nil
}

// enableWALArchiving 启用WAL归档
func enableWALArchiving(cfg *config.Config) error {
	logger.Info("检查WAL归档配置")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 检查archive_mode
	var archiveMode string
	err = db.QueryRow("SHOW archive_mode").Scan(&archiveMode)
	if err != nil {
		return fmt.Errorf("查询archive_mode失败: %v", err)
	}

	if archiveMode != "on" {
		logger.Warn("WAL归档未启用 (archive_mode = %s)，建议在postgresql.conf中设置 archive_mode = on", archiveMode)
	} else {
		logger.Info("WAL归档已启用")
	}

	return nil
}

// getCurrentLSN 获取当前LSN
func getCurrentLSN(cfg *config.Config) (string, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	var lsn string
	err = db.QueryRow("SELECT pg_current_wal_lsn()").Scan(&lsn)
	if err != nil {
		return "", fmt.Errorf("获取当前LSN失败: %v", err)
	}

	return lsn, nil
}
