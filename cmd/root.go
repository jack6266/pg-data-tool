package cmd

import (
	"fmt"
	"os"

	"pg-data-tool/internal/backup"
	"pg-data-tool/internal/config"
	"pg-data-tool/internal/restore"

	"github.com/spf13/cobra"
)

var cfg *config.Config

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
		var err error
		if backupFlag {
			err = backup.PerformBackup(cfg)
		} else if restoreFlag {
			err = restore.PerformRestore(cfg)
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

var (
	backupFlag  bool
	restoreFlag bool
)

func init() {
	cfg = config.NewConfig()

	rootCmd.Flags().StringVar(&cfg.Host, "host", cfg.Host, "数据库主机地址")
	rootCmd.Flags().StringVar(&cfg.Port, "port", cfg.Port, "数据库端口")
	rootCmd.Flags().StringVar(&cfg.User, "user", cfg.User, "数据库用户名")
	rootCmd.Flags().StringVar(&cfg.Password, "password", "", "数据库密码")
	rootCmd.Flags().StringVar(&cfg.DBName, "dbname", "", "数据库名称（单库操作时必需）")
	rootCmd.Flags().BoolVar(&backupFlag, "backup", false, "执行备份操作")
	rootCmd.Flags().BoolVar(&restoreFlag, "restore", false, "执行还原操作")
	rootCmd.Flags().StringVar(&cfg.File, "file", "", "备份文件路径")
	rootCmd.Flags().BoolVar(&cfg.BackupAll, "backup-all", false, "备份所有数据库（仅备份时有效）")
	rootCmd.Flags().BoolVar(&cfg.RestoreAll, "restore-all", false, "还原所有数据库（仅还原时有效）")

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
			return backup.PerformBackup(cfg)
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
			return restore.PerformRestore(cfg)
		},
	}
	restoreCmd.Flags().StringVarP(&cfg.DBName, "dbname", "d", "", "要还原的数据库名称")
	restoreCmd.Flags().BoolVarP(&cfg.RestoreAll, "restore-all", "a", false, "还原所有数据库")
	restoreCmd.Flags().StringVarP(&cfg.File, "file", "f", "", "备份文件路径")
	rootCmd.AddCommand(restoreCmd)
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}
