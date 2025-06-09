package backup

import (
	"database/sql"
	"fmt"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
)

// GetDatabaseInfo 获取数据库信息
func GetDatabaseInfo(cfg *config.Config, dbname string) (*DatabaseInfo, error) {
	// 连接到数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	info := &DatabaseInfo{}

	// 获取数据库大小
	var size string
	err = db.QueryRow(`
		SELECT pg_size_pretty(pg_database_size($1))
	`, dbname).Scan(&size)
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

// DisplayDatabaseInfo 显示数据库信息
func DisplayDatabaseInfo(cfg *config.Config, dbname string) error {
	dbInfo, err := GetDatabaseInfo(cfg, dbname)
	if err != nil {
		return err
	}

	// 输出到控制台
	fmt.Println("\n数据库基本信息:")
	fmt.Printf("数据库名称: %s\n", dbname)
	fmt.Printf("数据库大小: %s\n", dbInfo.Size)
	fmt.Printf("表数量: %d\n", dbInfo.TableCount)
	fmt.Printf("索引数量: %d\n", dbInfo.IndexCount)
	fmt.Printf("视图数量: %d\n", dbInfo.ViewCount)
	fmt.Printf("函数数量: %d\n", dbInfo.FunctionCount)

	// 记录到日志
	logger.Info("数据库基本信息:")
	logger.Info("- 数据库大小: %s", dbInfo.Size)
	logger.Info("- 表数量: %d", dbInfo.TableCount)
	logger.Info("- 索引数量: %d", dbInfo.IndexCount)
	logger.Info("- 视图数量: %d", dbInfo.ViewCount)
	logger.Info("- 函数数量: %d", dbInfo.FunctionCount)

	return nil
}

// GetAllDatabases 获取所有数据库列表
func GetAllDatabases(cfg *config.Config) ([]string, error) {
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
		AND datname NOT IN ('postgres', 'template0', 'template1', 'erdcloud')
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
