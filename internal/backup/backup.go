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

// NewBackuper 创建新的备份器
func NewBackuper(cfg *config.Config) *Backuper {
	return &Backuper{cfg: cfg}
}

// displayBackupSuccess 显示备份成功信息
func (b *Backuper) displayBackupSuccess(dbname, backupFile string) {
	fmt.Println("\n==========================================")
	fmt.Println("✅ 数据库备份操作已完成！")
	fmt.Println("------------------------------------------")
	fmt.Println("📝 备份详情：")
	fmt.Printf("   - 数据库名称：%s\n", dbname)
	fmt.Printf("   - 备份文件：%s\n", backupFile)
	fmt.Printf("   - 备份格式：%s\n", b.cfg.Format)
	fmt.Printf("   - 目标主机：%s\n", b.cfg.Host)
	fmt.Printf("   - 目标端口：%s\n", b.cfg.Port)
	fmt.Printf("   - 操作用户：%s\n", b.cfg.User)
	fmt.Println("==========================================")
}

// displayBackupError 显示备份错误信息
func (b *Backuper) displayBackupError(err error, operation string, dbname string) {
	fmt.Println("\n==========================================")
	fmt.Println("❌ 备份操作失败！")
	fmt.Println("------------------------------------------")
	fmt.Println("📝 错误详情：")
	fmt.Printf("   - 操作类型：%s\n", operation)
	fmt.Printf("   - 错误信息：%v\n", err)
	if b.cfg.BackupAll {
		fmt.Println("   - 操作模式：全库备份")
	} else {
		fmt.Printf("   - 目标数据库：%s\n", dbname)
	}
	fmt.Printf("   - 备份格式：%s\n", b.cfg.Format)
	fmt.Printf("   - 目标主机：%s\n", b.cfg.Host)
	fmt.Printf("   - 目标端口：%s\n", b.cfg.Port)
	fmt.Println("------------------------------------------")
	fmt.Println("💡 建议：")
	fmt.Println("   1. 检查数据库连接是否正常")
	fmt.Println("   2. 确认用户权限是否足够")
	fmt.Println("   3. 验证磁盘空间是否充足")
	fmt.Println("   4. 查看日志获取更多信息")
	fmt.Println("==========================================")
	logger.Error("%s失败: %v", operation, err)
}

// PerformBackup 执行数据库备份
func (b *Backuper) PerformBackup() error {
	logger.Info("开始执行数据库备份操作")
	logger.Info("连接参数: 主机=%s, 端口=%s, 用户=%s", b.cfg.Host, b.cfg.Port, b.cfg.User)
	logger.Info("备份格式: %s", b.cfg.Format)

	// 检查PostgreSQL系统数据库连接
	if err := utils.CheckPostgresConnection(b.cfg.Host, b.cfg.Port, b.cfg.User, b.cfg.Password); err != nil {
		b.displayBackupError(err, "检查数据库连接", "")
		return err
	}

	// 创建备份目录
	backupDir, err := b.createBackupDir()
	if err != nil {
		b.displayBackupError(err, "创建备份目录", "")
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
		b.displayBackupError(backupErr, "执行备份", b.cfg.DBName)
		return backupErr
	}

	return nil
}

// createBackupDir 创建备份目录
func (b *Backuper) createBackupDir() (string, error) {
	now := time.Now()
	backupDir := fmt.Sprintf("backups-%s-%s", now.Format("060102150405"), b.cfg.Host)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		err := fmt.Errorf("创建备份目录失败: %v", err)
		logger.Error(err.Error())
		return "", err
	}
	logger.Info("备份文件将保存在: %s", backupDir)
	return backupDir, nil
}

// backupAll 执行全库备份
func (b *Backuper) backupAll(backupDir string) error {
	// 获取所有数据库列表
	databases, err := b.getAllDatabases()
	if err != nil {
		b.displayBackupError(err, "获取数据库列表", "")
		return err
	}

	logger.Info("开始全库备份，共发现 %d 个数据库", len(databases))
	successCount := 0
	for _, db := range databases {
		if err := b.processDatabase(db, backupDir); err != nil {
			b.displayBackupError(err, fmt.Sprintf("备份数据库 %s", db), db)
			continue
		}
		successCount++
	}

	fmt.Println("\n==========================================")
	fmt.Println("📊 全库备份完成统计：")
	fmt.Printf("   - 总数据库数：%d\n", len(databases))
	fmt.Printf("   - 成功备份数：%d\n", successCount)
	fmt.Printf("   - 失败数量：%d\n", len(databases)-successCount)
	fmt.Println("==========================================")

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
	if err := b.displayDatabaseInfo(dbname); err != nil {
		logger.Error("获取数据库信息失败: %v", err)
	}

	// 执行备份
	return b.backupSingleDatabase(dbname, backupDir)
}

// getAllDatabases 获取所有数据库列表
func (b *Backuper) getAllDatabases() ([]string, error) {
	// 连接到postgres数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		b.cfg.Host, b.cfg.Port, b.cfg.User, b.cfg.Password)
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

// displayDatabaseInfo 显示数据库信息
func (b *Backuper) displayDatabaseInfo(dbname string) error {
	dbInfo, err := b.getDatabaseInfo(dbname)
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

// getDatabaseInfo 获取数据库信息
func (b *Backuper) getDatabaseInfo(dbname string) (*DatabaseInfo, error) {
	// 连接到数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		b.cfg.Host, b.cfg.Port, b.cfg.User, b.cfg.Password, dbname)
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

// backupSingleDatabase 备份单个数据库
func (b *Backuper) backupSingleDatabase(dbname, backupDir string) error {
	logger.Info("开始备份数据库: %s", dbname)

	// 生成备份文件名
	backupFile, err := b.generateBackupFileName(dbname, backupDir)
	if err != nil {
		b.displayBackupError(err, "生成备份文件名", dbname)
		return err
	}

	// 构建并执行备份命令
	cmd, err := b.buildBackupCommand(dbname, backupFile)
	if err != nil {
		b.displayBackupError(err, "构建备份命令", dbname)
		return err
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Errorf("备份数据库 %s 失败: %v\n错误输出: %s", dbname, err, output)
		b.displayBackupError(errMsg, "执行备份命令", dbname)
		return errMsg
	}

	b.displayBackupSuccess(dbname, backupFile)
	logger.Info("数据库 %s 备份成功，文件保存在: %s", dbname, backupFile)
	return nil
}

// generateBackupFileName 生成备份文件名
func (b *Backuper) generateBackupFileName(dbname, backupDir string) (string, error) {
	timestamp := time.Now().Format("150405")
	var fileExt string

	// 根据备份格式设置文件扩展名
	switch b.cfg.Format {
	case "custom":
		fileExt = ".backup"
	case "plain":
		fileExt = ".sql"
	case "directory":
		fileExt = ".dir"
	case "tar":
		fileExt = ".tar"
	default:
		return "", fmt.Errorf("不支持的备份格式: %s", b.cfg.Format)
	}

	backupFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s%s", dbname, timestamp, fileExt))

	// 如果是目录格式，需要创建目录
	if b.cfg.Format == "directory" {
		if err := os.MkdirAll(backupFile, 0755); err != nil {
			return "", fmt.Errorf("创建备份目录失败: %v", err)
		}
	}

	return backupFile, nil
}

// buildBackupCommand 构建备份命令
func (b *Backuper) buildBackupCommand(dbname, backupFile string) (*exec.Cmd, error) {
	// 构建pg_dump命令
	args := []string{
		"-h", b.cfg.Host,
		"-p", b.cfg.Port,
		"-U", b.cfg.User,
		"-F", b.cfg.Format,
		"-v",
	}

	// 根据格式添加特定参数
	switch b.cfg.Format {
	case "custom":
		args = append(args, "-b") // 包含大对象
	}

	// 添加输出文件参数
	args = append(args, "-f", backupFile, dbname)

	cmd := exec.Command("pg_dump", args...)
	logger.Info("执行命令: pg_dump %v", args)

	// 设置环境变量
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", b.cfg.Password))

	return cmd, nil
}
