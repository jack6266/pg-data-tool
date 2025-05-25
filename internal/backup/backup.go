package backup

import (
	"fmt"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
	"pg-data-tool/internal/utils"
)

// NewBackuper 创建新的备份器
func NewBackuper(cfg *config.Config) *Backuper {
	return &Backuper{cfg: cfg}
}

// PerformBackup 执行数据库备份
func (b *Backuper) PerformBackup() error {
	logger.Info("开始执行数据库备份操作")
	logger.Info("连接参数: 主机=%s, 端口=%s, 用户=%s", b.cfg.Host, b.cfg.Port, b.cfg.User)
	logger.Info("备份格式: %s", b.cfg.Format)

	// 检查PostgreSQL系统数据库连接
	if err := utils.CheckPostgresConnection(b.cfg.Host, b.cfg.Port, b.cfg.User, b.cfg.Password); err != nil {
		DisplayBackupError(b.cfg, err, "检查数据库连接", "")
		return err
	}

	// 创建备份目录
	backupDir, err := CreateBackupDir(b.cfg)
	if err != nil {
		DisplayBackupError(b.cfg, err, "创建备份目录", "")
		return err
	}

	// 执行备份
	var backupErr error
	if b.cfg.BackupAll {
		backupErr = b.backupAll(backupDir)
	} else {
		backupErr = b.backupSingle(backupDir)
	}

	if backupErr != nil {
		DisplayBackupError(b.cfg, backupErr, "执行备份", b.cfg.DBName)
		return backupErr
	}

	return nil
}

// backupAll 执行全库备份
func (b *Backuper) backupAll(backupDir string) error {
	// 获取所有数据库列表
	databases, err := GetAllDatabases(b.cfg)
	if err != nil {
		DisplayBackupError(b.cfg, err, "获取数据库列表", "")
		return err
	}

	logger.Info("开始全库备份，共发现 %d 个数据库", len(databases))
	successCount := 0
	for _, db := range databases {
		if err := b.processDatabase(db, backupDir); err != nil {
			DisplayBackupError(b.cfg, err, fmt.Sprintf("备份数据库 %s", db), db)
			continue
		}
		successCount++
	}

	DisplayBackupSummary(len(databases), successCount)
	logger.Info("全库备份完成")
	return nil
}

// backupSingle 执行单库备份
func (b *Backuper) backupSingle(backupDir string) error {
	if b.cfg.DBName == "" {
		return fmt.Errorf("错误：必须指定数据库名称或使用 --backup-all 参数进行全库备份")
	}

	// 检查目标数据库连接
	if err := utils.CheckDatabaseConnection(b.cfg.Host, b.cfg.Port, b.cfg.User, b.cfg.Password, b.cfg.DBName); err != nil {
		return err
	}

	return b.processDatabase(b.cfg.DBName, backupDir)
}

// processDatabase 处理单个数据库的备份
func (b *Backuper) processDatabase(dbname, backupDir string) error {
	// 检查数据库连接
	if err := utils.CheckDatabaseConnection(b.cfg.Host, b.cfg.Port, b.cfg.User, b.cfg.Password, dbname); err != nil {
		return fmt.Errorf("数据库 %s 连接失败: %v", dbname, err)
	}

	// 获取并显示数据库信息
	if err := DisplayDatabaseInfo(b.cfg, dbname); err != nil {
		logger.Error("获取数据库信息失败: %v", err)
	}

	// 执行备份
	if err := BackupSingleDatabase(b.cfg, dbname, backupDir); err != nil {
		return err
	}

	// 显示备份成功信息
	backupFile, err := GenerateBackupFileName(b.cfg, dbname, backupDir)
	if err != nil {
		return err
	}
	DisplayBackupSuccess(b.cfg, dbname, backupFile)

	return nil
}
