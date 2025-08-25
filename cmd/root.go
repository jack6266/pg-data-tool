package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"pg-data-tool/internal/backup"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/restore"
	"pg-data-tool/internal/transfer"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// 设置控制台编码为 UTF-8
func initConsoleEncoding() {
	if runtime.GOOS == "windows" {
		// 设置控制台代码页为 UTF-8 (65001)
		cmd := exec.Command("chcp", "65001")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("警告: 设置控制台编码失败: %v\n", err)
		}
	}
}

var cfg *config.Config

// promptForInput 用于交互式获取用户输入
func promptForInput(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// promptForBool 用于交互式获取布尔值输入
func promptForBool(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)
	defaultStr := "y"
	if !defaultValue {
		defaultStr = "n"
	}
	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}

// getInteractiveConfig 通过交互式方式获取配置
func getInteractiveConfig() *config.Config {
	cfg := config.NewConfig()

	// 获取操作类型
	fmt.Println("\n请选择操作类型：")
	fmt.Println("1. 备份")
	fmt.Println("2. 还原")
	fmt.Println("3. 数据传输")
	var choice string
	fmt.Print("请选择 [1/2/3]: ")
	fmt.Scanln(&choice)

	if choice == "1" {
		backupFlag = true
	} else if choice == "2" {
		restoreFlag = true
	} else if choice == "3" {
		transferFlag = true
	} else {
		fmt.Println("无效的选择")
		os.Exit(1)
	}

	// 获取数据库连接信息
	cfg.Host = promptForInput("数据库主机地址", cfg.Host)
	cfg.Port = promptForInput("数据库端口", cfg.Port)
	cfg.User = promptForInput("数据库用户名", cfg.User)
	cfg.Password = promptForInput("数据库密码", cfg.Password)

	if backupFlag {
		cfg.BackupAll = promptForBool("是否备份所有数据库", false)
		if !cfg.BackupAll {
			cfg.DBName = promptForInput("要备份的数据库名称", "")
		}
		cfg.Format = promptForInput("备份格式 (plain/custom/directory/tar)", cfg.Format)
		cfg.File = promptForInput("备份文件保存路径", "")

		// 只在directory模式下询问是否使用并行处理
		if cfg.Format == "directory" {
			cfg.UseParallel = promptForBool("是否使用并行处理", false)
			if cfg.UseParallel {
				parallelJobs := promptForInput("并行作业数", fmt.Sprintf("%d", cfg.ParallelJobs))
				if jobs, err := fmt.Sscanf(parallelJobs, "%d", &cfg.ParallelJobs); err != nil || jobs != 1 {
					fmt.Println("无效的并行作业数，使用默认值：", cfg.ParallelJobs)
				}
			} else {
				cfg.ParallelJobs = 1
				fmt.Println("已禁用并行处理，将使用单线程备份")
			}
		}
	} else if restoreFlag {
		cfg.RestoreAll = promptForBool("是否还原所有数据库", false)
		if !cfg.RestoreAll {
			cfg.DBName = promptForInput("要还原的数据库名称", "")
		}
		cfg.File = promptForInput("备份文件路径", "")
		cfg.AutoCreateDB = promptForBool("如果数据库不存在是否自动创建", false)

		// 清理选项
		fmt.Println("\n=== 清理选项 ===")
		fmt.Println("💡 清理说明：")
		fmt.Println("   - 清理数据：只删除表中的数据，保留表结构")
		fmt.Println("   - 清理结构：删除所有表、视图、函数、序列等数据库对象")
		fmt.Println("   - 清理结构包含清理数据的功能")
		fmt.Println("------------------------------------------")

		cfg.CleanStructure = promptForBool("是否在还原前清理数据库结构（表、视图、函数等）", false)
		if !cfg.CleanStructure {
			cfg.CleanData = promptForBool("是否在还原前清理数据库中的所有数据", false)
		}

		// 检查文件格式是否支持并行处理
		ext := strings.ToLower(filepath.Ext(cfg.File))
		if ext == ".dir" {
			cfg.UseParallel = promptForBool("是否使用并行处理", false)
			if cfg.UseParallel {
				parallelJobs := promptForInput("并行作业数", fmt.Sprintf("%d", cfg.ParallelJobs))
				if jobs, err := fmt.Sscanf(parallelJobs, "%d", &cfg.ParallelJobs); err != nil || jobs != 1 {
					fmt.Println("无效的并行作业数，使用默认值：", cfg.ParallelJobs)
				}
			} else {
				cfg.ParallelJobs = 1
				fmt.Println("已禁用并行处理，将使用单线程还原")
			}
		}
	} else if transferFlag {
		fmt.Println("\n=== 目标数据库信息 ===")
		cfg.TargetHost = promptForInput("目标数据库主机地址", "")
		cfg.TargetPort = promptForInput("目标数据库端口", "5432")
		cfg.TargetUser = promptForInput("目标数据库用户名", "erdcloud")
		cfg.TargetPassword = promptForInput("目标数据库密码", "Pw!123456")

		cfg.TransferAll = promptForBool("是否传输所有数据库(y/n)", false)
		if !cfg.TransferAll {
			cfg.DBName = promptForInput("要传输的数据库名称", "")
		}

		fmt.Println("\n=== 传输选项 ===")
		fmt.Println("💡 传输说明：")
		fmt.Println("   - 支持跨服务器数据库传输")
		fmt.Println("   - 可以选择是否包含大对象")
		fmt.Println("   - 可以选择是否包含索引")
		fmt.Println("   - 可以选择是否包含权限")
		fmt.Println("------------------------------------------")

		cfg.IncludeBlobs = promptForBool("是否包含大对象(y/n)", true)
		cfg.IncludeIndexes = promptForBool("是否包含索引(y/n)", true)
		cfg.IncludePrivileges = promptForBool("是否包含权限(y/n)", true)
	}

	return cfg
}

var (
	backupFlag   bool
	restoreFlag  bool
	transferFlag bool
)

func init() {
	cfg = config.NewConfig()

	rootCmd.Flags().StringVar(&cfg.Host, "host", cfg.Host, "数据库主机地址")
	rootCmd.Flags().StringVar(&cfg.Port, "port", cfg.Port, "数据库端口")
	rootCmd.Flags().StringVar(&cfg.User, "user", cfg.User, "数据库用户名")
	rootCmd.Flags().StringVar(&cfg.Password, "password", cfg.Password, "数据库密码")
	rootCmd.Flags().StringVar(&cfg.DBName, "dbname", "", "数据库名称（单库操作时必需）")
	rootCmd.Flags().BoolVar(&backupFlag, "backup", false, "执行备份操作")
	rootCmd.Flags().BoolVar(&restoreFlag, "restore", false, "执行还原操作")
	rootCmd.Flags().BoolVar(&transferFlag, "transfer", false, "执行数据传输操作")
	rootCmd.Flags().StringVar(&cfg.File, "file", "", "备份文件路径")
	rootCmd.Flags().BoolVar(&cfg.BackupAll, "backup-all", false, "备份所有数据库（仅备份时有效）")
	rootCmd.Flags().BoolVar(&cfg.RestoreAll, "restore-all", false, "还原所有数据库（仅还原时有效）")
	rootCmd.Flags().StringVarP(&cfg.Format, "format", "f", cfg.Format, "备份格式 (plain, custom, directory, tar)")

	// 备份命令
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "执行数据库备份",
		Long: `执行数据库备份操作，支持以下格式：
- plain: SQL文本格式（默认）
- custom: 二进制格式
- directory: 目录格式
- tar: tar归档格式`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backuper := backup.NewBackuper(cfg)
			return backuper.PerformBackup()
		},
	}
	backupCmd.Flags().StringVarP(&cfg.DBName, "dbname", "d", "", "要备份的数据库名称")
	backupCmd.Flags().BoolVarP(&cfg.BackupAll, "backup-all", "a", false, "备份所有数据库")
	backupCmd.Flags().StringVarP(&cfg.Format, "format", "f", cfg.Format, "备份格式 (plain, custom, directory, tar)")
	rootCmd.AddCommand(backupCmd)

	// 还原命令
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "执行数据库还原",
		RunE: func(cmd *cobra.Command, args []string) error {
			restorer := restore.NewRestorer(cfg)
			return restorer.PerformRestore()
		},
	}
	restoreCmd.Flags().StringVarP(&cfg.DBName, "dbname", "d", "", "要还原的数据库名称")
	restoreCmd.Flags().BoolVarP(&cfg.RestoreAll, "restore-all", "a", false, "还原所有数据库")
	restoreCmd.Flags().StringVarP(&cfg.File, "file", "f", "", "备份文件路径")
	restoreCmd.Flags().BoolVar(&cfg.AutoCreateDB, "auto-create-db", false, "如果数据库不存在则自动创建")
	restoreCmd.Flags().BoolVar(&cfg.CleanData, "clean-data", false, "在还原前清理数据库中的所有数据")
	restoreCmd.Flags().BoolVar(&cfg.CleanStructure, "clean-structure", false, "在还原前清理数据库中的所有结构（表、视图、函数等）")
	rootCmd.AddCommand(restoreCmd)

	// 数据传输命令
	transferCmd := &cobra.Command{
		Use:   "transfer",
		Short: "执行数据库传输",
		Long: `执行数据库传输操作，支持以下功能：
1. 跨服务器数据库传输
2. 可选择是否包含大对象
3. 可选择是否包含索引
4. 可选择是否包含权限`,
		RunE: func(cmd *cobra.Command, args []string) error {
			transferer := transfer.NewTransferer(cfg)
			return transferer.PerformTransfer()
		},
	}
	transferCmd.Flags().StringVarP(&cfg.DBName, "dbname", "d", "", "要传输的数据库名称")
	transferCmd.Flags().BoolVarP(&cfg.TransferAll, "transfer-all", "a", false, "传输所有数据库")
	transferCmd.Flags().StringVar(&cfg.TargetHost, "target-host", "", "目标数据库主机地址")
	transferCmd.Flags().StringVar(&cfg.TargetPort, "target-port", "5432", "目标数据库端口")
	transferCmd.Flags().StringVar(&cfg.TargetUser, "target-user", "", "目标数据库用户名")
	transferCmd.Flags().StringVar(&cfg.TargetPassword, "target-password", "", "目标数据库密码")
	transferCmd.Flags().BoolVar(&cfg.IncludeBlobs, "include-blobs", true, "是否包含大对象")
	transferCmd.Flags().BoolVar(&cfg.IncludeIndexes, "include-indexes", true, "是否包含索引")
	transferCmd.Flags().BoolVar(&cfg.IncludePrivileges, "include-privileges", true, "是否包含权限")
	rootCmd.AddCommand(transferCmd)
}

var rootCmd = &cobra.Command{
	Use:   "pg-data-tool",
	Short: "PostgreSQL数据库备份和还原工具",
	Long: `PostgreSQL数据备份还原工具，支持以下功能：
1. 数据库备份（支持单库和全库备份）
2. 数据库还原（支持单库和全库还原）
3. 数据库传输（支持跨服务器传输）
4. 支持多种备份格式：
   - plain: SQL文本格式（默认）
   - custom: 二进制格式
   - directory: 目录格式
   - tar: tar归档格式`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化控制台编码
		initConsoleEncoding()

		// 检查是否有命令行参数
		hasArgs := false
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			hasArgs = true
		})

		if !hasArgs {
			// 如果没有命令行参数，使用交互式模式
			cfg = getInteractiveConfig()
		}

		var err error
		if backupFlag {
			backuper := backup.NewBackuper(cfg)
			err = backuper.PerformBackup()
		} else if restoreFlag {
			restorer := restore.NewRestorer(cfg)
			err = restorer.PerformRestore()
		} else if transferFlag {
			transferer := transfer.NewTransferer(cfg)
			err = transferer.PerformTransfer()
		} else {
			fmt.Println("请指定 --backup、--restore 或 --transfer 参数")
			return
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}
