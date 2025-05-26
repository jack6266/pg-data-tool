package transfer

import (
	"fmt"
	"pg-data-tool/internal/config"
)

// DisplayTransferStart 显示传输开始信息
func DisplayTransferStart(cfg *config.Config) {
	fmt.Println("\n==========================================")
	fmt.Println("🚀 数据库传输开始")
	fmt.Println("------------------------------------------")

	fmt.Println("📌 源数据库信息：")
	fmt.Printf("   - 主机: %s\n", cfg.Host)
	fmt.Printf("   - 端口: %s\n", cfg.Port)
	fmt.Printf("   - 用户: %s\n", cfg.User)

	fmt.Println("\n📌 目标数据库信息：")
	fmt.Printf("   - 主机: %s\n", cfg.TargetHost)
	fmt.Printf("   - 端口: %s\n", cfg.TargetPort)
	fmt.Printf("   - 用户: %s\n", cfg.TargetUser)

	fmt.Println("\n⚙️ 传输选项：")
	if cfg.TransferAll {
		fmt.Println("   - 传输范围: 所有数据库")
		fmt.Println("   - 系统数据库将被自动排除")
	} else {
		fmt.Printf("   - 传输范围: 指定数据库 (%s)\n", cfg.DBName)
	}
	fmt.Printf("   - 包含大对象: %v\n", cfg.IncludeBlobs)
	fmt.Printf("   - 包含索引: %v\n", cfg.IncludeIndexes)
	fmt.Printf("   - 包含权限: %v\n", cfg.IncludePrivileges)

	fmt.Println("\n💡 传输说明：")
	fmt.Println("   - 如果目标数据库已存在，将会被覆盖")
	fmt.Println("   - 传输过程中请勿中断操作")
	fmt.Println("   - 大数据库传输可能需要较长时间")
	fmt.Println("   - 请确保目标服务器有足够空间")
	if cfg.TransferAll {
		fmt.Println("   - 全库传输将自动跳过系统数据库")
		fmt.Println("   - 建议先备份重要数据")
	}

	fmt.Println("------------------------------------------")
	fmt.Println("⏳ 正在开始传输...")
	fmt.Println("==========================================")
}

// DisplayDatabaseInfo 显示数据库信息
func DisplayDatabaseInfo(info *DatabaseInfo) {
	fmt.Println("\n=== 数据库信息 ===")
	fmt.Printf("名称: %s\n", info.Name)
	fmt.Printf("大小: %s\n", info.Size)
	fmt.Printf("所有者: %s\n", info.Owner)
	if info.Description != "" {
		fmt.Printf("描述: %s\n", info.Description)
	}
	fmt.Println("------------------------------------------")
}

// DisplayTransferSuccess 显示传输成功信息
func DisplayTransferSuccess(dbname string) {
	fmt.Printf("\n✅ 数据库 %s 传输成功\n", dbname)
}

// DisplayTransferError 显示传输错误信息
func DisplayTransferError(dbname string, err error) {
	fmt.Printf("\n❌ 数据库 %s 传输失败: %v\n", dbname, err)
	fmt.Println("\n可能的原因：")
	fmt.Println("1. 数据库连接失败")
	fmt.Println("2. 用户权限不足")
	fmt.Println("3. 目标数据库已存在")
	fmt.Println("4. 磁盘空间不足")
}

// DisplayTransferSummary 显示传输总结信息
func DisplayTransferSummary(total, success, failed int) {
	fmt.Println("\n==========================================")
	fmt.Println("📊 传输结果统计")
	fmt.Println("------------------------------------------")
	fmt.Printf("📝 总数据库数: %d\n", total)
	fmt.Printf("✅ 成功传输: %d\n", success)
	fmt.Printf("❌ 传输失败: %d\n", failed)

	// 计算成功率
	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}
	fmt.Printf("📈 成功率: %.1f%%\n", successRate)

	fmt.Println("------------------------------------------")
	if failed > 0 {
		fmt.Println("💡 提示：")
		fmt.Println("   - 请检查失败数据库的详细错误信息")
		fmt.Println("   - 可以尝试单独传输失败的数据库")
		fmt.Println("   - 查看日志获取更多详细信息")
	} else {
		fmt.Println("🎉 所有数据库传输完成！")
	}
	fmt.Println("==========================================")
}
