package backup

import (
	"fmt"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
)

// DisplayBackupSuccess 显示备份成功信息
func DisplayBackupSuccess(cfg *config.Config, dbname, backupFile string) {
	fmt.Println("\n==========================================")
	fmt.Println("✅ 数据库备份操作已完成！")
	fmt.Println("------------------------------------------")
	fmt.Println("📝 备份详情：")
	fmt.Printf("   - 数据库名称：%s\n", dbname)
	fmt.Printf("   - 备份文件：%s\n", backupFile)
	fmt.Printf("   - 备份格式：%s\n", cfg.Format)
	fmt.Printf("   - 目标主机：%s\n", cfg.Host)
	fmt.Printf("   - 目标端口：%s\n", cfg.Port)
	fmt.Printf("   - 操作用户：%s\n", cfg.User)
	fmt.Println("==========================================")
}

// DisplayBackupError 显示备份错误信息
func DisplayBackupError(cfg *config.Config, err error, operation string, dbname string) {
	fmt.Println("\n==========================================")
	fmt.Println("❌ 备份操作失败！")
	fmt.Println("------------------------------------------")
	fmt.Println("📝 错误详情：")
	fmt.Printf("   - 操作类型：%s\n", operation)
	fmt.Printf("   - 错误信息：%v\n", err)
	if cfg.BackupAll {
		fmt.Println("   - 操作模式：全库备份")
	} else {
		fmt.Printf("   - 目标数据库：%s\n", dbname)
	}
	fmt.Printf("   - 备份格式：%s\n", cfg.Format)
	fmt.Printf("   - 目标主机：%s\n", cfg.Host)
	fmt.Printf("   - 目标端口：%s\n", cfg.Port)
	fmt.Println("------------------------------------------")
	fmt.Println("💡 建议：")
	fmt.Println("   1. 检查数据库连接是否正常")
	fmt.Println("   2. 确认用户权限是否足够")
	fmt.Println("   3. 验证磁盘空间是否充足")
	fmt.Println("   4. 查看日志获取更多信息")
	fmt.Println("==========================================")
	logger.Error("%s失败: %v", operation, err)
}

// DisplayBackupSummary 显示备份统计信息
func DisplayBackupSummary(totalCount, successCount int) {
	fmt.Println("\n==========================================")
	fmt.Println("📊 全库备份完成统计：")
	fmt.Printf("   - 总数据库数：%d\n", totalCount)
	fmt.Printf("   - 成功备份数：%d\n", successCount)
	fmt.Printf("   - 失败数量：%d\n", totalCount-successCount)
	fmt.Println("==========================================")
}
