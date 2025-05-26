package transfer

import (
	"fmt"
	"os"
	"os/exec"

	"pg-data-tool/internal/config"
)

// TransferOperations 处理传输操作
type TransferOperations struct {
	cfg *config.Config
	db  *DatabaseOperations
}

// NewTransferOperations 创建新的传输操作实例
func NewTransferOperations(cfg *config.Config) *TransferOperations {
	return &TransferOperations{
		cfg: cfg,
		db:  NewDatabaseOperations(cfg),
	}
}

// TransferDatabase 传输数据库
func (t *TransferOperations) TransferDatabase(dbname string) error {
	// 检查目标数据库是否存在
	exists, err := t.db.DatabaseExists(dbname)
	if err != nil {
		return err
	}

	// 如果数据库存在，先删除
	if exists {
		if err := t.db.DropDatabase(dbname); err != nil {
			return err
		}
	}

	// 创建新数据库
	if err := t.db.CreateDatabase(dbname); err != nil {
		return err
	}

	// 构建pg_dump命令
	dumpArgs := []string{
		"-h", t.cfg.Host,
		"-p", t.cfg.Port,
		"-U", t.cfg.User,
		"-F", "custom",
		"-v",
	}

	// 添加选项
	if t.cfg.IncludeBlobs {
		dumpArgs = append(dumpArgs, "-b")
	}
	if !t.cfg.IncludeIndexes {
		dumpArgs = append(dumpArgs, "--no-indexes")
	}
	if !t.cfg.IncludePrivileges {
		dumpArgs = append(dumpArgs, "--no-acl")
	}

	// 添加数据库名
	dumpArgs = append(dumpArgs, dbname)

	// 构建pg_restore命令
	restoreArgs := []string{
		"-h", t.cfg.TargetHost,
		"-p", t.cfg.TargetPort,
		"-U", t.cfg.TargetUser,
		"-d", dbname,
		"-v",
	}

	// 创建管道
	dumpCmd := exec.Command("pg_dump", dumpArgs...)
	restoreCmd := exec.Command("pg_restore", restoreArgs...)

	// 设置环境变量
	dumpCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", t.cfg.Password))
	restoreCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", t.cfg.TargetPassword))

	// 连接命令
	restoreCmd.Stdin, _ = dumpCmd.StdoutPipe()
	restoreCmd.Stdout = os.Stdout
	restoreCmd.Stderr = os.Stderr

	// 执行命令
	if err := restoreCmd.Start(); err != nil {
		return fmt.Errorf("启动还原命令失败: %v", err)
	}
	if err := dumpCmd.Run(); err != nil {
		return fmt.Errorf("执行备份命令失败: %v", err)
	}
	if err := restoreCmd.Wait(); err != nil {
		return fmt.Errorf("等待还原命令完成失败: %v", err)
	}

	return nil
}
