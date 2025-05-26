package transfer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"pg-data-tool/internal/config"
)

// DatabaseOperations 处理数据库操作
type DatabaseOperations struct {
	cfg *config.Config
}

// NewDatabaseOperations 创建新的数据库操作实例
func NewDatabaseOperations(cfg *config.Config) *DatabaseOperations {
	return &DatabaseOperations{
		cfg: cfg,
	}
}

// CreateDatabase 创建数据库
func (d *DatabaseOperations) CreateDatabase(dbname string) error {
	cmd := exec.Command("psql",
		"-h", d.cfg.TargetHost,
		"-p", d.cfg.TargetPort,
		"-U", d.cfg.TargetUser,
		"-c", fmt.Sprintf("CREATE DATABASE %s;", dbname))

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.TargetPassword))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	return nil
}

// DropDatabase 删除数据库
func (d *DatabaseOperations) DropDatabase(dbname string) error {
	cmd := exec.Command("psql",
		"-h", d.cfg.TargetHost,
		"-p", d.cfg.TargetPort,
		"-U", d.cfg.TargetUser,
		"-c", fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbname))

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.TargetPassword))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("删除数据库失败: %v", err)
	}

	return nil
}

// DatabaseExists 检查数据库是否存在
func (d *DatabaseOperations) DatabaseExists(dbname string) (bool, error) {
	cmd := exec.Command("psql",
		"-h", d.cfg.TargetHost,
		"-p", d.cfg.TargetPort,
		"-U", d.cfg.TargetUser,
		"-t",
		"-c", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s';", dbname))

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.TargetPassword))

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("检查数据库是否存在失败: %v", err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// GetTargetDatabaseStats 获取目标数据库基本信息
func (d *DatabaseOperations) GetTargetDatabaseStats(dbname string) (string, error) {
	cmd := exec.Command("psql",
		"-h", d.cfg.TargetHost,
		"-p", d.cfg.TargetPort,
		"-U", d.cfg.TargetUser,
		"-d", dbname, // 指定数据库名
		"-t",
		"-c", `
			SELECT 
				pg_size_pretty(pg_database_size(current_database())) as size,
				(SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public') as table_count,
				(SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public') as index_count,
				(SELECT COUNT(*) FROM pg_views WHERE schemaname = 'public') as view_count,
				(SELECT COUNT(*) FROM pg_proc WHERE pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')) as function_count;
		`)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.TargetPassword))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取目标数据库信息失败: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}
