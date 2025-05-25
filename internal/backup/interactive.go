package backup

import (
	"fmt"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
)

// ConfirmBackup 交互式确认备份操作
func ConfirmBackup(cfg *config.Config) error {
	fmt.Println("\n=== 备份操作确认 ===")
	fmt.Printf("主机: %s\n", cfg.Host)
	fmt.Printf("端口: %s\n", cfg.Port)
	fmt.Printf("用户: %s\n", cfg.User)
	fmt.Printf("备份格式: %s\n", cfg.Format)
	if cfg.BackupAll {
		fmt.Println("备份范围: 所有数据库")
	} else {
		fmt.Printf("备份数据库: %s\n", cfg.DBName)
	}

	// 询问是否使用并行处理
	if cfg.Format == "custom" || cfg.Format == "directory" || cfg.Format == "tar" {
		fmt.Println("\n=== 并行处理选项 ===")
		fmt.Println("💡 并行处理说明：")
		fmt.Println("   - 当数据库大小超过2GB时，建议使用并行处理")
		fmt.Println("   - 并行处理会提高备份速度，但会增加CPU和内存使用")
		fmt.Println("   - 建议并行数设置为CPU核心数的1-2倍")
		fmt.Println("   - 当前默认并行数：", cfg.ParallelJobs)
		fmt.Println("------------------------------------------")
		fmt.Printf("是否使用并行处理？(y/N) ")
		var useParallel string
		fmt.Scanln(&useParallel)
		if useParallel == "y" || useParallel == "Y" {
			cfg.UseParallel = true
			fmt.Printf("并行作业数: %d\n", cfg.ParallelJobs)
		} else {
			cfg.UseParallel = false
			cfg.ParallelJobs = 1
			fmt.Println("已禁用并行处理，将使用单线程备份")
		}
	}

	fmt.Print("\n是否继续？(y/N) ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		return fmt.Errorf("操作已取消")
	}

	logger.Info("用户确认继续备份操作")
	return nil
}
