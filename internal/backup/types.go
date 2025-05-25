package backup

import "pg-data-tool/internal/config"

// DatabaseInfo 存储数据库信息
type DatabaseInfo struct {
	Size          string
	TableCount    int
	IndexCount    int
	ViewCount     int
	FunctionCount int
}

// Backuper 数据库备份器
type Backuper struct {
	cfg *config.Config
}
