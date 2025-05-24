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
	"pg-data-tool/internal/utils"

	_ "github.com/lib/pq"
)

// PerformBackup 执行数据库备份
func PerformBackup(cfg *config.Config) error {
	logger.Info("开始执行数据库备份操作")
	logger.Info("连接参数: 主机=%s, 端口=%s, 用户=%s", cfg.Host, cfg.Port, cfg.User)
	logger.Info("备份格式: %s", cfg.Format)

	// 检查PostgreSQL系统数据库连接
	if err := utils.CheckPostgresConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password); err != nil {
		return err
	}

	// 创建备份目录（格式：backups-yymmddHHmiss）
	now := time.Now()
	backupDir := fmt.Sprintf("backups-%s-%s", now.Format("060102150405"), cfg.Host)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		err := fmt.Errorf("创建备份目录失败: %v", err)
		logger.Error(err.Error())
		return err
	}
	logger.Info("备份文件将保存在: %s", backupDir)

	if cfg.BackupAll {
		// 获取所有数据库列表
		databases, err := getAllDatabases(cfg)
		if err != nil {
			return err
		}

		logger.Info("开始全库备份，共发现 %d 个数据库", len(databases))
		for _, db := range databases {
			// 检查每个数据库的连接
			if err := utils.CheckDatabaseConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password, db); err != nil {
				logger.Error("数据库 %s 连接失败: %v", db, err)
				continue
			}

			if err := backupSingleDatabase(cfg, db, backupDir); err != nil {
				logger.Error("备份数据库 %s 失败: %v", db, err)
				continue
			}
		}
		logger.Info("全库备份完成")
	} else {
		if cfg.DBName == "" {
			err := fmt.Errorf("错误：必须指定数据库名称或使用 --backup-all 参数进行全库备份")
			logger.Error(err.Error())
			return err
		}

		// 检查目标数据库连接
		if err := utils.CheckDatabaseConnection(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName); err != nil {
			return err
		}

		if err := backupSingleDatabase(cfg, cfg.DBName, backupDir); err != nil {
			return err
		}
	}

	return nil
}

// getAllDatabases 获取所有数据库列表
func getAllDatabases(cfg *config.Config) ([]string, error) {
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
		AND datname NOT IN ('postgres', 'template0', 'template1')
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

// backupSingleDatabase 备份单个数据库
func backupSingleDatabase(cfg *config.Config, dbname, backupDir string) error {
	logger.Info("开始备份数据库: %s", dbname)

	// 获取数据库信息
	dbInfo, err := getDatabaseInfo(cfg, dbname)
	if err != nil {
		logger.Error("获取数据库信息失败: %v", err)
	} else {
		// 输出到控制台
		fmt.Println("\n数据库基本信息:")
		fmt.Printf("数据库名称: %s\n", dbname)
		fmt.Printf("数据库大小: %s\n", dbInfo.Size)
		fmt.Printf("表数量: %d\n", dbInfo.TableCount)
		fmt.Printf("索引数量: %d\n", dbInfo.IndexCount)
		fmt.Printf("视图数量: %d\n", dbInfo.ViewCount)
		fmt.Printf("函数数量: %d\n", dbInfo.FunctionCount)

		// 同时记录到日志
		logger.Info("数据库基本信息:")
		logger.Info("- 数据库大小: %s", dbInfo.Size)
		logger.Info("- 表数量: %d", dbInfo.TableCount)
		logger.Info("- 索引数量: %d", dbInfo.IndexCount)
		logger.Info("- 视图数量: %d", dbInfo.ViewCount)
		logger.Info("- 函数数量: %d", dbInfo.FunctionCount)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("150405")
	var backupFile string
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
		return fmt.Errorf("不支持的备份格式: %s", cfg.Format)
	}

	backupFile = filepath.Join(backupDir, fmt.Sprintf("%s_%s%s", dbname, timestamp, fileExt))

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
		// 创建目录
		if err := os.MkdirAll(backupFile, 0755); err != nil {
			return fmt.Errorf("创建备份目录失败: %v", err)
		}
	}

	// 添加输出文件参数
	args = append(args, "-f", backupFile, dbname)

	cmd := exec.Command("pg_dump", args...)
	logger.Info("执行命令: pg_dump %v", args)

	// 设置环境变量
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Errorf("备份数据库 %s 失败: %v\n错误输出: %s", dbname, err, output)
		logger.Error(errMsg.Error())
		return errMsg
	}

	logger.Info("数据库 %s 备份成功，文件保存在: %s", dbname, backupFile)
	return nil
}

// DatabaseInfo 存储数据库信息
type DatabaseInfo struct {
	Size          string
	TableCount    int
	IndexCount    int
	ViewCount     int
	FunctionCount int
	Tables        []TableInfo
}

// TableInfo 存储表信息
type TableInfo struct {
	Name       string
	Size       string
	RowCount   int64
	IndexCount int
}

// getDatabaseInfo 获取数据库信息
func getDatabaseInfo(cfg *config.Config, dbname string) (*DatabaseInfo, error) {
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

	// 获取表详细信息
	rows, err := db.Query(`
		SELECT 
			c.relname as table_name,
			pg_size_pretty(pg_total_relation_size(c.oid)) as size,
			(SELECT COUNT(*) FROM pg_indexes WHERE tablename = c.relname) as index_count,
			(SELECT n_live_tup FROM pg_stat_user_tables WHERE relname = c.relname) as row_count
		FROM pg_class c
		WHERE relkind = 'r' AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		ORDER BY c.relname
	`)
	if err != nil {
		return nil, fmt.Errorf("获取表详细信息失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table TableInfo
		err := rows.Scan(&table.Name, &table.Size, &table.IndexCount, &table.RowCount)
		if err != nil {
			return nil, fmt.Errorf("读取表信息失败: %v", err)
		}
		info.Tables = append(info.Tables, table)
	}

	return info, nil
}
