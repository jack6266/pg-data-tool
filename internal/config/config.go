package config

import (
	"os"
	"path/filepath"
)

// Config 存储配置信息
type Config struct {
	// 源数据库连接信息
	Host     string
	Port     string
	User     string
	Password string
	DBName   string

	// 目标数据库连接信息（用于数据传输）
	TargetHost     string
	TargetPort     string
	TargetUser     string
	TargetPassword string

	// 备份类型和相关配置
	BackupType     string // logical, physical, incremental
	BackupAll      bool
	Format         string
	File           string
	BaseBackupPath string // 基础备份路径（用于增量备份）
	WALArchiveDir  string // WAL归档目录（用于物理备份）
	MaxWalSize     string // 最大WAL大小
	CheckpointLSN  string // 检查点LSN（用于增量备份）

	// 还原类型和相关配置
	RestoreType    string // logical, physical, incremental
	RestoreAll     bool
	AutoCreateDB   bool   // 是否自动创建数据库
	CleanData      bool   // 是否在还原前清理数据库数据
	CleanStructure bool   // 是否在还原前清理数据库结构（表、视图、函数等）
	DropDatabase   bool   // 是否在还原前删除并重新创建数据库
	TargetTime     string // 恢复到指定时间点（PITR）

	// 传输相关配置
	TransferAll       bool
	IncludeBlobs      bool
	IncludeIndexes    bool
	IncludePrivileges bool

	// 并行处理相关配置
	UseParallel  bool
	ParallelJobs int
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	// 获取当前工作目录的绝对路径
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "." // 如果获取失败，使用相对路径
	} else {
		// 确保是绝对路径
		currentDir, _ = filepath.Abs(currentDir)
	}

	// WAL归档目录使用绝对路径
	walArchiveDir := filepath.Join(currentDir, "wal_archive")

	return &Config{
		Host:           "localhost",
		Port:           "5432",
		User:           "erdcloud",
		Format:         "plain",
		File:           currentDir, // 默认备份路径为当前目录的绝对路径
		BackupType:     "logical",  // 默认逻辑备份
		RestoreType:    "logical",  // 默认逻辑还原
		ParallelJobs:   4,
		TransferAll:    false,
		UseParallel:    false,
		MaxWalSize:     "1GB",         // 默认WAL大小
		Password:       "Pw!123456",   // 设置源数据库默认密码
		TargetPassword: "Pw!123456",   // 设置目标数据库默认密码
		BaseBackupPath: currentDir,    // 基础备份路径也设为当前目录的绝对路径
		WALArchiveDir:  walArchiveDir, // WAL归档目录设为当前目录下的子目录的绝对路径
	}
}

// NormalizePath 规范化路径，将相对路径转换为绝对路径
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}

	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(path) {
		return path
	}

	// 将相对路径转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		// 如果转换失败，返回原路径
		return path
	}

	return absPath
}

// SetFile 设置文件路径并规范化
func (c *Config) SetFile(path string) {
	c.File = NormalizePath(path)
}

// SetBaseBackupPath 设置基础备份路径并规范化
func (c *Config) SetBaseBackupPath(path string) {
	c.BaseBackupPath = NormalizePath(path)
}

// SetWALArchiveDir 设置WAL归档目录并规范化
func (c *Config) SetWALArchiveDir(path string) {
	c.WALArchiveDir = NormalizePath(path)
}
