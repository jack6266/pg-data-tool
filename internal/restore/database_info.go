package restore

import (
	"database/sql"
	"fmt"
)

// GetDatabaseInfo 获取数据库信息
func GetDatabaseInfo(db *sql.DB, dbName string) (*DatabaseInfo, error) {
	info := &DatabaseInfo{}

	// 获取数据库大小
	var size string
	err := db.QueryRow(`
		SELECT pg_size_pretty(pg_database_size($1))
	`, dbName).Scan(&size)
	if err != nil {
		return nil, fmt.Errorf("获取数据库大小失败: %v", err)
	}
	info.Size = size

	// 获取表数量
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public'
	`).Scan(&info.TableCount)
	if err != nil {
		return nil, fmt.Errorf("获取表数量失败: %v", err)
	}

	// 获取索引数量
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'
	`).Scan(&info.IndexCount)
	if err != nil {
		return nil, fmt.Errorf("获取索引数量失败: %v", err)
	}

	// 获取视图数量
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_views WHERE schemaname = 'public'
	`).Scan(&info.ViewCount)
	if err != nil {
		return nil, fmt.Errorf("获取视图数量失败: %v", err)
	}

	// 获取函数数量
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_proc WHERE pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
	`).Scan(&info.FunctionCount)
	if err != nil {
		return nil, fmt.Errorf("获取函数数量失败: %v", err)
	}

	return info, nil
}
