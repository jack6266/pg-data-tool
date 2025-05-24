package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"pg-data-tool/internal/backup"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/restore"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
	var choice string
	fmt.Print("请选择 [1/2]: ")
	fmt.Scanln(&choice)

	if choice == "1" {
		backupFlag = true
	} else if choice == "2" {
		restoreFlag = true
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
	} else if restoreFlag {
		cfg.RestoreAll = promptForBool("是否还原所有数据库", false)
		if !cfg.RestoreAll {
			cfg.DBName = promptForInput("要还原的数据库名称", "")
		}
		cfg.File = promptForInput("备份文件路径", "")
	}

	return cfg
}

var (
	backupFlag  bool
	restoreFlag bool
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
	rootCmd.AddCommand(restoreCmd)
}

var rootCmd = &cobra.Command{
	Use:   "pg-data-tool",
	Short: "PostgreSQL数据库备份和还原工具",
	Long: `PostgreSQL数据备份还原工具，支持以下功能：
1. 数据库备份（支持单库和全库备份）
2. 数据库还原（支持单库和全库还原）
3. 支持多种备份格式：
   - plain: SQL文本格式（默认）
   - custom: 二进制格式
   - directory: 目录格式
   - tar: tar归档格式`,
	Run: func(cmd *cobra.Command, args []string) {
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
		} else {
			fmt.Println("请指定 --backup 或 --restore 参数")
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
