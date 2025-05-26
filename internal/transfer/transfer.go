package transfer

import (
	"fmt"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
	"strings"
)

// Transferer 处理数据库传输
type Transferer struct {
	cfg *config.Config
	ops *TransferOperations
	db  *DatabaseOperations
}

// NewTransferer 创建新的传输器实例
func NewTransferer(cfg *config.Config) *Transferer {
	return &Transferer{
		cfg: cfg,
		ops: NewTransferOperations(cfg),
		db:  NewDatabaseOperations(cfg),
	}
}

// PerformTransfer 执行数据库传输
func (t *Transferer) PerformTransfer() error {
	// 显示传输开始信息
	DisplayTransferStart(t.cfg)

	if t.cfg.TransferAll {
		return t.transferAllDatabases()
	}
	return t.transferSingleDatabase(t.cfg.DBName)
}

// transferAllDatabases 传输所有数据库
func (t *Transferer) transferAllDatabases() error {
	// 获取所有数据库列表
	databases, err := t.db.GetDatabaseList()
	if err != nil {
		return fmt.Errorf("获取数据库列表失败: %v", err)
	}

	// 传输统计
	total := len(databases)
	success := 0
	failed := 0

	// 传输每个数据库
	for _, db := range databases {
		// 获取并显示数据库信息
		info, err := t.db.GetDatabaseInfo(db)
		if err != nil {
			logger.Error("获取数据库 %s 信息失败: %v", db, err)
		} else {
			DisplayDatabaseInfo(info)
		}

		if err := t.transferSingleDatabase(db); err != nil {
			failed++
			DisplayTransferError(db, err)
			continue
		}
		success++
		DisplayTransferSuccess(db)
	}

	// 显示传输总结
	DisplayTransferSummary(total, success, failed)

	return nil
}

// transferSingleDatabase 传输单个数据库
func (t *Transferer) transferSingleDatabase(dbname string) error {
	logger.Info("开始传输数据库: %s", dbname)

	// 获取并显示源数据库信息
	info, err := t.db.GetDatabaseInfo(dbname)
	if err != nil {
		logger.Error("获取数据库 %s 信息失败: %v", dbname, err)
	} else {
		DisplayDatabaseInfo(info)
	}

	// 执行传输
	if err := t.ops.TransferDatabase(dbname); err != nil {
		return err
	}

	// 显示传输完成信息
	fmt.Println("\n==========================================")
	fmt.Println("🎉 数据库传输完成")
	fmt.Println("------------------------------------------")
	fmt.Printf("📌 数据库名称: %s\n", dbname)
	fmt.Printf("📌 目标主机: %s\n", t.cfg.TargetHost)

	// 获取并显示目标数据库基本信息
	stats, err := t.db.GetTargetDatabaseStats(dbname)
	if err != nil {
		logger.Error("获取目标数据库 %s 信息失败: %v", dbname, err)
	} else {
		// 解析并显示数据库统计信息
		parts := strings.Split(strings.TrimSpace(stats), "|")
		if len(parts) >= 5 {
			fmt.Println("\n📊 目标数据库基本信息：")
			fmt.Printf("   - 数据库大小: %s\n", strings.TrimSpace(parts[0]))
			fmt.Printf("   - 表数量: %s\n", strings.TrimSpace(parts[1]))
			fmt.Printf("   - 索引数量: %s\n", strings.TrimSpace(parts[2]))
			fmt.Printf("   - 视图数量: %s\n", strings.TrimSpace(parts[3]))
			fmt.Printf("   - 函数数量: %s\n", strings.TrimSpace(parts[4]))
		} else {
			logger.Error("解析目标数据库统计信息失败: 数据格式不正确")
		}
	}

	fmt.Println("==========================================")

	return nil
}
