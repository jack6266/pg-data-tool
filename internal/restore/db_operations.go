package restore

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"

	_ "github.com/lib/pq"
)

// CleanDatabaseStructure 清理数据库中的所有结构（表、视图、函数、序列等）
func CleanDatabaseStructure(cfg *config.Config, dbName string) error {
	logger.Info("开始清理数据库 %s 中的结构", dbName)

	// 连接到目标数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 删除视图
	if err := dropViews(db); err != nil {
		logger.Warn("删除视图失败: %v", err)
	}

	// 删除函数
	if err := dropFunctions(db); err != nil {
		logger.Warn("删除函数失败: %v", err)
	}

	// 删除表（会自动删除相关索引和约束）
	if err := dropTables(db); err != nil {
		logger.Warn("删除表失败: %v", err)
	}

	// 删除序列
	if err := dropSequences(db); err != nil {
		logger.Warn("删除序列失败: %v", err)
	}

	// 删除自定义类型
	if err := dropTypes(db); err != nil {
		logger.Warn("删除自定义类型失败: %v", err)
	}

	logger.Info("数据库 %s 结构清理完成", dbName)
	return nil
}

// dropViews 删除所有视图
func dropViews(db *sql.DB) error {
	logger.Info("开始删除视图")

	rows, err := db.Query(`
		SELECT viewname 
		FROM pg_views 
		WHERE schemaname = 'public'
		ORDER BY viewname
	`)
	if err != nil {
		return fmt.Errorf("获取视图列表失败: %v", err)
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		if err := rows.Scan(&viewName); err != nil {
			return fmt.Errorf("读取视图名失败: %v", err)
		}
		views = append(views, viewName)
	}

	for _, view := range views {
		logger.Info("删除视图: %s", view)
		if _, err := db.Exec(fmt.Sprintf("DROP VIEW IF EXISTS %s CASCADE", view)); err != nil {
			logger.Warn("删除视图 %s 失败: %v", view, err)
		}
	}

	logger.Info("已删除 %d 个视图", len(views))
	return nil
}

// dropFunctions 删除所有用户定义的函数
func dropFunctions(db *sql.DB) error {
	logger.Info("开始删除函数")

	rows, err := db.Query(`
		SELECT proname, pg_get_function_identity_arguments(oid) as args
		FROM pg_proc 
		WHERE pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		ORDER BY proname
	`)
	if err != nil {
		return fmt.Errorf("获取函数列表失败: %v", err)
	}
	defer rows.Close()

	var functions []struct {
		name string
		args string
	}

	for rows.Next() {
		var funcName, funcArgs string
		if err := rows.Scan(&funcName, &funcArgs); err != nil {
			return fmt.Errorf("读取函数信息失败: %v", err)
		}
		functions = append(functions, struct {
			name string
			args string
		}{funcName, funcArgs})
	}

	for _, function := range functions {
		logger.Info("删除函数: %s(%s)", function.name, function.args)
		if _, err := db.Exec(fmt.Sprintf("DROP FUNCTION IF EXISTS %s(%s) CASCADE", function.name, function.args)); err != nil {
			logger.Warn("删除函数 %s 失败: %v", function.name, err)
		}
	}

	logger.Info("已删除 %d 个函数", len(functions))
	return nil
}

// dropTables 删除所有表
func dropTables(db *sql.DB) error {
	logger.Info("开始删除表")

	rows, err := db.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		ORDER BY tablename
	`)
	if err != nil {
		return fmt.Errorf("获取表列表失败: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("读取表名失败: %v", err)
		}
		tables = append(tables, tableName)
	}

	for _, table := range tables {
		logger.Info("删除表: %s", table)
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			logger.Warn("删除表 %s 失败: %v", table, err)
		}
	}

	logger.Info("已删除 %d 个表", len(tables))
	return nil
}

// dropSequences 删除所有序列
func dropSequences(db *sql.DB) error {
	logger.Info("开始删除序列")

	rows, err := db.Query(`
		SELECT sequence_name 
		FROM information_schema.sequences 
		WHERE sequence_schema = 'public'
		ORDER BY sequence_name
	`)
	if err != nil {
		return fmt.Errorf("获取序列列表失败: %v", err)
	}
	defer rows.Close()

	var sequences []string
	for rows.Next() {
		var seqName string
		if err := rows.Scan(&seqName); err != nil {
			return fmt.Errorf("读取序列名失败: %v", err)
		}
		sequences = append(sequences, seqName)
	}

	for _, seq := range sequences {
		logger.Info("删除序列: %s", seq)
		if _, err := db.Exec(fmt.Sprintf("DROP SEQUENCE IF EXISTS %s CASCADE", seq)); err != nil {
			logger.Warn("删除序列 %s 失败: %v", seq, err)
		}
	}

	logger.Info("已删除 %d 个序列", len(sequences))
	return nil
}

// dropTypes 删除所有用户定义的类型
func dropTypes(db *sql.DB) error {
	logger.Info("开始删除自定义类型")

	rows, err := db.Query(`
		SELECT typname 
		FROM pg_type 
		WHERE typnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		AND typtype = 'e'  -- 只删除枚举类型
		ORDER BY typname
	`)
	if err != nil {
		return fmt.Errorf("获取类型列表失败: %v", err)
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var typeName string
		if err := rows.Scan(&typeName); err != nil {
			return fmt.Errorf("读取类型名失败: %v", err)
		}
		types = append(types, typeName)
	}

	for _, typ := range types {
		logger.Info("删除类型: %s", typ)
		if _, err := db.Exec(fmt.Sprintf("DROP TYPE IF EXISTS %s CASCADE", typ)); err != nil {
			logger.Warn("删除类型 %s 失败: %v", typ, err)
		}
	}

	logger.Info("已删除 %d 个类型", len(types))
	return nil
}

// CleanDatabaseData 清理数据库中的所有数据但保留结构
func CleanDatabaseData(cfg *config.Config, dbName string) error {
	logger.Info("开始清理数据库 %s 中的数据", dbName)

	// 连接到目标数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 获取所有用户表
	rows, err := db.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		ORDER BY tablename
	`)
	if err != nil {
		return fmt.Errorf("获取表列表失败: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("读取表名失败: %v", err)
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		logger.Info("数据库 %s 中没有找到用户表", dbName)
		return nil
	}

	logger.Info("找到 %d 个表，开始清理数据", len(tables))

	// 禁用外键约束检查
	if _, err := db.Exec("SET session_replication_role = replica;"); err != nil {
		logger.Warn("禁用外键约束失败，继续执行: %v", err)
	}

	// 清理每个表的数据
	for _, table := range tables {
		logger.Info("正在清理表: %s", table)
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			logger.Warn("清理表 %s 失败: %v", table, err)
		}
	}

	// 重新启用外键约束检查
	if _, err := db.Exec("SET session_replication_role = DEFAULT;"); err != nil {
		logger.Warn("重新启用外键约束失败: %v", err)
	}

	// 重置序列
	rows, err = db.Query(`
		SELECT sequence_name 
		FROM information_schema.sequences 
		WHERE sequence_schema = 'public'
	`)
	if err != nil {
		logger.Warn("获取序列列表失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var seqName string
			if err := rows.Scan(&seqName); err != nil {
				continue
			}
			if _, err := db.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seqName)); err != nil {
				logger.Warn("重置序列 %s 失败: %v", seqName, err)
			}
		}
	}

	logger.Info("数据库 %s 数据清理完成", dbName)
	return nil
}

// CreateDatabaseIfNotExists 如果数据库不存在则创建
func CreateDatabaseIfNotExists(cfg *config.Config, dbName string) error {
	// 连接到postgres数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("连接postgres数据库失败: %v", err)
	}
	defer db.Close()

	// 检查数据库是否存在
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pg_database WHERE datname = $1
		)
	`, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查数据库是否存在失败: %v", err)
	}

	if !exists {
		logger.Info("数据库 %s 不存在，正在创建...", dbName)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("创建数据库失败: %v", err)
		}
		logger.Info("数据库 %s 创建成功", dbName)
	} else {
		logger.Info("数据库 %s 已存在", dbName)
	}

	return nil
}

// RestoreSingleFile 还原单个备份文件到指定数据库
func RestoreSingleFile(cfg *config.Config, backupFile, dbName string) error {
	// 如果配置了清理结构选项，则先清理数据库结构
	if cfg.CleanStructure {
		logger.Info("配置了清理结构选项，开始清理数据库 %s 的结构", dbName)
		if err := CleanDatabaseStructure(cfg, dbName); err != nil {
			logger.Warn("清理数据库结构失败，继续执行还原: %v", err)
		}
	} else if cfg.CleanData {
		// 如果只配置了清理数据选项，则只清理数据库数据
		logger.Info("配置了清理数据选项，开始清理数据库 %s 的数据", dbName)
		if err := CleanDatabaseData(cfg, dbName); err != nil {
			logger.Warn("清理数据库数据失败，继续执行还原: %v", err)
		}
	}

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
		// 注意：对于SQL文件，数据清理已在函数开头执行
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
