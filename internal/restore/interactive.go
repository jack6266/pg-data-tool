package restore

import (
	"fmt"
	"path/filepath"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/logger"
	"strings"
)

// ConfirmRestore 交互式确认还原操作
func ConfirmRestore(cfg *config.Config) error {
	fmt.Println("\n=== 还原操作确认 ===")
	fmt.Printf("主机: %s\n", cfg.Host)
	fmt.Printf("端口: %s\n", cfg.Port)
	fmt.Printf("用户: %s\n", cfg.User)
	fmt.Printf("备份文件: %s\n", cfg.File)
	if cfg.RestoreAll {
		fmt.Println("还原范围: 所有数据库")
	} else {
		fmt.Printf("还原数据库: %s\n", cfg.DBName)
	}

	// 检查文件格式是否支持并行处理
	ext := strings.ToLower(filepath.Ext(cfg.File))
	if ext == ".backup" || ext == ".tar" || ext == ".dir" {
		fmt.Println("\n=== 并行处理选项 ===")
		fmt.Println("💡 并行处理说明：")
		fmt.Println("   - 当备份文件大小超过2GB时，建议使用并行处理")
		fmt.Println("   - 并行处理会提高还原速度，但会增加CPU和内存使用")
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
			fmt.Println("已禁用并行处理，将使用单线程还原")
		}
	}

	fmt.Print("\n是否继续？(y/N) ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		return fmt.Errorf("操作已取消")
	}

	logger.Info("用户确认继续还原操作")
	return nil
}

// DisplayRestoreResult 显示还原结果
func DisplayRestoreResult(cfg *config.Config, success bool, err error) {
	if !success {
		fmt.Println("\n==========================================")
		fmt.Println("❌ 数据库还原操作失败！")
		fmt.Println("------------------------------------------")
		fmt.Println("📝 错误详情：")
		fmt.Printf("   - 错误信息：%v\n", err)
		if cfg.RestoreAll {
			fmt.Println("   - 操作类型：全库还原")
		} else {
			fmt.Printf("   - 目标数据库：%s\n", cfg.DBName)
		}
		fmt.Printf("   - 备份文件：%s\n", cfg.File)
		fmt.Printf("   - 目标主机：%s\n", cfg.Host)
		fmt.Printf("   - 目标端口：%s\n", cfg.Port)
		fmt.Println("------------------------------------------")
		fmt.Println("💡 建议：")
		fmt.Println("   1. 检查备份文件是否完整")
		fmt.Println("   2. 确认数据库连接信息是否正确")
		fmt.Println("   3. 验证用户权限是否足够")
		fmt.Println("   4. 查看日志获取更多信息")
		fmt.Println("==========================================")
		return
	}

	// 显示还原完成信息
	fmt.Println("\n==========================================")
	fmt.Println("✅ 数据库还原操作已完成！")
	if cfg.RestoreAll {
		fmt.Println("📊 已完成全库还原")
	} else {
		fmt.Printf("📊 已完成数据库 %s 的还原\n", cfg.DBName)
	}
	fmt.Println("------------------------------------------")
	fmt.Println("📝 还原详情：")
	fmt.Printf("   - 备份文件：%s\n", cfg.File)
	fmt.Printf("   - 目标主机：%s\n", cfg.Host)
	fmt.Printf("   - 目标端口：%s\n", cfg.Port)
	fmt.Printf("   - 操作用户：%s\n", cfg.User)
	if cfg.AutoCreateDB {
		fmt.Println("   - 自动创建数据库：是")
	} else {
		fmt.Println("   - 自动创建数据库：否")
	}
	fmt.Println("==========================================")
}
