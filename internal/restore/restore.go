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

// Restorer 数据库还原器
type Restorer struct {
	cfg *config.Config
}

// NewRestorer 创建新的还原器
func NewRestorer(cfg *config.Config) *Restorer {
	return &Restorer{cfg: cfg}
}

// PerformRestore 执行数据库还原
func (r *Restorer) PerformRestore() error {
	logger.Info("开始执行数据库还原操作")
	logger.Info("连接参数: 主机=%s, 端口=%s, 用户=%s", r.cfg.Host, r.cfg.Port, r.cfg.User)

	// 交互式确认
	if err := r.confirmRestore(); err != nil {
		return err
	}

	// 检查PostgreSQL系统数据库连接
	if err := utils.CheckPostgresConnection(r.cfg.Host, r.cfg.Port, r.cfg.User, r.cfg.Password); err != nil {
		return err
	}

	// 验证参数
	if err := r.validateParams(); err != nil {
		return err
	}

	// 执行还原
	var err error
	if r.cfg.RestoreAll {
		err = r.restoreAll()
	} else {
		err = r.restoreSingle()
	}

	if err != nil {
		fmt.Println("\n==========================================")
		fmt.Println("❌ 数据库还原操作失败！")
		fmt.Println("------------------------------------------")
		fmt.Println("📝 错误详情：")
		fmt.Printf("   - 错误信息：%v\n", err)
		if r.cfg.RestoreAll {
			fmt.Println("   - 操作类型：全库还原")
		} else {
			fmt.Printf("   - 目标数据库：%s\n", r.cfg.DBName)
		}
		fmt.Printf("   - 备份文件：%s\n", r.cfg.File)
		fmt.Printf("   - 目标主机：%s\n", r.cfg.Host)
		fmt.Printf("   - 目标端口：%s\n", r.cfg.Port)
		fmt.Println("------------------------------------------")
		fmt.Println("💡 建议：")
		fmt.Println("   1. 检查备份文件是否完整")
		fmt.Println("   2. 确认数据库连接信息是否正确")
		fmt.Println("   3. 验证用户权限是否足够")
		fmt.Println("   4. 查看日志获取更多信息")
		fmt.Println("==========================================")
		logger.Error("数据库还原操作失败: %v", err)
		return err
	}

	// 显示还原完成信息
	fmt.Println("\n==========================================")
	fmt.Println("✅ 数据库还原操作已完成！")
	if r.cfg.RestoreAll {
		fmt.Println("📊 已完成全库还原")
	} else {
		fmt.Printf("📊 已完成数据库 %s 的还原\n", r.cfg.DBName)
	}
	fmt.Println("------------------------------------------")
	fmt.Println("📝 还原详情：")
	fmt.Printf("   - 备份文件：%s\n", r.cfg.File)
	fmt.Printf("   - 目标主机：%s\n", r.cfg.Host)
	fmt.Printf("   - 目标端口：%s\n", r.cfg.Port)
	fmt.Printf("   - 操作用户：%s\n", r.cfg.User)
	if r.cfg.AutoCreateDB {
		fmt.Println("   - 自动创建数据库：是")
	} else {
		fmt.Println("   - 自动创建数据库：否")
	}
	fmt.Println("==========================================")
	logger.Info("数据库还原操作已完成")

	return nil
}

// confirmRestore 交互式确认还原操作
func (r *Restorer) confirmRestore() error {
	fmt.Printf("警告：即将还原数据库到以下服务器：\n")
	if r.cfg.RestoreAll {
		fmt.Printf("操作：全库还原\n")
	} else {
		fmt.Printf("操作：还原数据库 %s\n", r.cfg.DBName)
	}
	fmt.Printf("备份文件：%s\n", r.cfg.File)

	// 交互式选择是否自动创建数据库
	if !r.cfg.RestoreAll {
		fmt.Printf("是否自动创建数据库（如果不存在）？(y/N): ")
		var createDB string
		fmt.Scanln(&createDB)
		r.cfg.AutoCreateDB = createDB == "y" || createDB == "Y"
		fmt.Printf("自动创建数据库：%v\n", r.cfg.AutoCreateDB)
	} else {
		fmt.Printf("是否自动创建数据库（如果不存在）？(y/N): ")
		var createDB string
		fmt.Scanln(&createDB)
		r.cfg.AutoCreateDB = createDB == "y" || createDB == "Y"
		fmt.Printf("自动创建数据库：%v\n", r.cfg.AutoCreateDB)
	}

	fmt.Printf("警告：还原操作将覆盖目标数据库中的现有数据。\n")
	fmt.Printf("是否继续？(y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		return fmt.Errorf("用户取消还原操作")
	}

	return nil
}

// validateParams 验证参数
func (r *Restorer) validateParams() error {
	if !r.cfg.RestoreAll && r.cfg.DBName == "" {
		return fmt.Errorf("错误：必须指定数据库名称或使用 --restore-all 参数进行全库还原")
	}

	if r.cfg.File == "" {
		return fmt.Errorf("错误：必须指定备份文件路径")
	}

	// 检查文件或目录是否存在
	fileInfo, err := os.Stat(r.cfg.File)
	if os.IsNotExist(err) {
		return fmt.Errorf("错误：备份文件或目录 %s 不存在", r.cfg.File)
	}

	if r.cfg.RestoreAll {
		if !fileInfo.IsDir() && !strings.HasSuffix(r.cfg.File, ".dir") {
			return fmt.Errorf("错误：全库还原需要指定备份目录或目录格式的备份文件")
		}
	}

	return nil
}

// restoreAll 执行全库还原
func (r *Restorer) restoreAll() error {
	if fileInfo, _ := os.Stat(r.cfg.File); fileInfo.IsDir() || strings.HasSuffix(r.cfg.File, ".dir") {
		return r.restoreAllFromDirectory()
	}
	return fmt.Errorf("错误：全库还原需要指定备份目录或目录格式的备份文件")
}

// restoreSingle 执行单库还原
func (r *Restorer) restoreSingle() error {
	if r.cfg.AutoCreateDB {
		if err := r.createDatabaseIfNotExists(r.cfg.DBName); err != nil {
			return err
		}
	} else {
		if err := utils.CheckDatabaseConnection(r.cfg.Host, r.cfg.Port, r.cfg.User, r.cfg.Password, r.cfg.DBName); err != nil {
			return err
		}
	}
	return r.restoreSingleDatabase()
}

// restoreAllFromDirectory 从目录还原所有数据库
func (r *Restorer) restoreAllFromDirectory() error {
	logger.Info("开始从目录还原所有数据库: %s", r.cfg.File)

	// 获取目录下的所有备份文件
	backupFiles, err := r.getBackupFiles()
	if err != nil {
		return err
	}

	if len(backupFiles) == 0 {
		return fmt.Errorf("错误：在目录 %s 中未找到有效的备份文件", r.cfg.File)
	}

	logger.Info("找到 %d 个备份文件", len(backupFiles))

	// 处理每个备份文件
	successCount := 0
	for _, backupFile := range backupFiles {
		if err := r.processBackupFile(backupFile); err != nil {
			logger.Error("处理备份文件 %s 失败: %v", backupFile, err)
			continue
		}
		successCount++
	}

	fmt.Println("\n==========================================")
	fmt.Println("📊 还原完成统计")
	fmt.Println("------------------------------------------")
	fmt.Printf("   - 总文件数：%d\n", len(backupFiles))
	fmt.Printf("   - 成功还原：%d\n", successCount)
	fmt.Printf("   - 失败数量：%d\n", len(backupFiles)-successCount)
	fmt.Printf("   - 成功率：%.2f%%\n", float64(successCount)/float64(len(backupFiles))*100)
	fmt.Println("------------------------------------------")
	if successCount == len(backupFiles) {
		fmt.Println("✅ 所有数据库还原成功！")
	} else if successCount > 0 {
		fmt.Println("⚠️ 部分数据库还原成功，请检查失败项。")
	} else {
		fmt.Println("❌ 所有数据库还原失败！")
	}
	fmt.Println("==========================================")

	logger.Info("全库还原完成")
	return nil
}

// getBackupFiles 获取目录下的所有备份文件
func (r *Restorer) getBackupFiles() ([]string, error) {
	var backupFiles []string
	err := filepath.Walk(r.cfg.File, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".sql" || ext == ".backup" || ext == ".tar" {
				backupFiles = append(backupFiles, path)
			}
		}
		return nil
	})
	return backupFiles, err
}

// processBackupFile 处理单个备份文件
func (r *Restorer) processBackupFile(backupFile string) error {
	fileName := filepath.Base(backupFile)
	dbName := r.extractDatabaseName(fileName)
	if dbName == "" {
		return fmt.Errorf("无法从文件名 %s 提取数据库名", fileName)
	}

	logger.Info("正在还原数据库 %s 从文件 %s", dbName, backupFile)

	// 处理数据库创建
	if r.cfg.AutoCreateDB {
		if err := r.createDatabaseIfNotExists(dbName); err != nil {
			return fmt.Errorf("创建数据库 %s 失败: %v", dbName, err)
		}
	} else {
		if err := utils.CheckDatabaseConnection(r.cfg.Host, r.cfg.Port, r.cfg.User, r.cfg.Password, dbName); err != nil {
			return fmt.Errorf("数据库 %s 连接失败: %v", dbName, err)
		}
	}

	// 执行还原
	if err := r.restoreSingleFile(backupFile, dbName); err != nil {
		return fmt.Errorf("还原数据库 %s 失败: %v", dbName, err)
	}

	// 显示还原后的数据库信息
	if err := r.displayRestoredDatabaseInfo(dbName); err != nil {
		logger.Error("获取数据库 %s 信息失败: %v", dbName, err)
	}

	logger.Info("数据库 %s 还原成功", dbName)
	return nil
}

// restoreSingleDatabase 还原单个数据库
func (r *Restorer) restoreSingleDatabase() error {
	logger.Info("开始还原数据库: %s", r.cfg.DBName)

	if err := r.restoreSingleFile(r.cfg.File, r.cfg.DBName); err != nil {
		return err
	}

	if err := r.displayRestoredDatabaseInfo(r.cfg.DBName); err != nil {
		logger.Error("获取数据库信息失败: %v", err)
	}

	fmt.Printf("\n✅ 数据库 %s 还原成功！\n", r.cfg.DBName)
	return nil
}

// displayRestoredDatabaseInfo 显示还原后的数据库信息
func (r *Restorer) displayRestoredDatabaseInfo(dbName string) error {
	dbInfo, err := r.getDatabaseInfo(dbName)
	if err != nil {
		return err
	}

	// 输出到控制台
	fmt.Println("\n数据库基本信息:")
	fmt.Printf("数据库名称: %s\n", dbName)
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

// createDatabaseIfNotExists 如果数据库不存在则创建
func (r *Restorer) createDatabaseIfNotExists(dbName string) error {
	// 连接到postgres数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		r.cfg.Host, r.cfg.Port, r.cfg.User, r.cfg.Password)
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

// getDatabaseInfo 获取数据库信息
func (r *Restorer) getDatabaseInfo(dbName string) (*DatabaseInfo, error) {
	// 连接到数据库
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		r.cfg.Host, r.cfg.Port, r.cfg.User, r.cfg.Password, dbName)
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

// restoreSingleFile 还原单个备份文件到指定数据库
func (r *Restorer) restoreSingleFile(backupFile, dbName string) error {
	// 构建pg_restore命令
	args := []string{
		"-h", r.cfg.Host,
		"-p", r.cfg.Port,
		"-U", r.cfg.User,
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
		args := []string{
			"-h", r.cfg.Host,
			"-p", r.cfg.Port,
			"-U", r.cfg.User,
			"-d", dbName,
			"-f", backupFile,
		}
		cmd := exec.Command("psql", args...)
		logger.Info("执行命令: psql -h %s -p %s -U %s -d %s -f %s", r.cfg.Host, r.cfg.Port, r.cfg.User, dbName, backupFile)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", r.cfg.Password))
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
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", r.cfg.Password))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("还原失败: %v\n错误输出: %s", err, output)
		}
		return nil
	default:
		return fmt.Errorf("不支持的备份文件格式: %s", ext)
	}
}

// extractDatabaseName 从备份文件名中提取数据库名
func (r *Restorer) extractDatabaseName(fileName string) string {
	logger.Info("正在从文件名提取数据库名: %s", fileName)

	// 移除文件扩展名
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	logger.Info("移除扩展名后的文件名: %s", baseName)

	// 找到最后一个下划线的位置
	lastUnderscoreIndex := strings.LastIndex(baseName, "_")
	if lastUnderscoreIndex == -1 {
		logger.Error("文件名 %s 中没有找到下划线分隔符", fileName)
		return ""
	}

	// 取最后一个下划线之前的部分作为数据库名
	dbName := baseName[:lastUnderscoreIndex]
	logger.Info("提取到的数据库名: %s", dbName)
	return dbName
}
