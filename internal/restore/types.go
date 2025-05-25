package restore

import "pg-data-tool/internal/config"

// DatabaseInfo 存储数据库信息
type DatabaseInfo struct {
	Size          string
	TableCount    int
	IndexCount    int
	ViewCount     int
	FunctionCount int
}

// Restorer 数据库还原器
type Restorer struct {
	cfg *config.Config
}
