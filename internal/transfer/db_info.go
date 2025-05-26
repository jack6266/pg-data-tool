package transfer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DatabaseInfo 存储数据库信息
type DatabaseInfo struct {
	Name        string
	Size        string
	Owner       string
	Description string
}

// GetDatabaseInfo 获取数据库信息
func (d *DatabaseOperations) GetDatabaseInfo(dbname string) (*DatabaseInfo, error) {
	// 获取数据库大小
	sizeCmd := exec.Command("psql",
		"-h", d.cfg.Host,
		"-p", d.cfg.Port,
		"-U", d.cfg.User,
		"-t",
		"-c", fmt.Sprintf("SELECT pg_size_pretty(pg_database_size('%s'));", dbname))

	sizeCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.Password))
	sizeOutput, err := sizeCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取数据库大小失败: %v", err)
	}

	// 获取数据库所有者
	ownerCmd := exec.Command("psql",
		"-h", d.cfg.Host,
		"-p", d.cfg.Port,
		"-U", d.cfg.User,
		"-t",
		"-c", fmt.Sprintf("SELECT d.datname, u.usename FROM pg_database d JOIN pg_user u ON d.datdba = u.usesysid WHERE d.datname = '%s';", dbname))

	ownerCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.Password))
	ownerOutput, err := ownerCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取数据库所有者失败: %v", err)
	}

	// 获取数据库描述
	descCmd := exec.Command("psql",
		"-h", d.cfg.Host,
		"-p", d.cfg.Port,
		"-U", d.cfg.User,
		"-t",
		"-c", fmt.Sprintf("SELECT description FROM pg_description WHERE objoid = (SELECT oid FROM pg_database WHERE datname = '%s');", dbname))

	descCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.Password))
	descOutput, err := descCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取数据库描述失败: %v", err)
	}

	// 解析所有者信息
	ownerParts := strings.Fields(string(ownerOutput))
	owner := "unknown"
	if len(ownerParts) >= 2 {
		owner = ownerParts[1]
	}

	return &DatabaseInfo{
		Name:        dbname,
		Size:        strings.TrimSpace(string(sizeOutput)),
		Owner:       owner,
		Description: strings.TrimSpace(string(descOutput)),
	}, nil
}

// GetDatabaseList 获取数据库列表
func (d *DatabaseOperations) GetDatabaseList() ([]string, error) {
	cmd := exec.Command("psql",
		"-h", d.cfg.Host,
		"-p", d.cfg.Port,
		"-U", d.cfg.User,
		"-t",
		"-c", "SELECT datname FROM pg_database WHERE datistemplate = false;")

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", d.cfg.Password))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取数据库列表失败: %v", err)
	}

	// 处理输出
	databases := strings.Split(string(output), "\n")
	var result []string
	for _, db := range databases {
		db = strings.TrimSpace(db)
		if db != "" && !isSystemDatabase(db) {
			result = append(result, db)
		}
	}

	return result, nil
}

// isSystemDatabase 判断是否为系统数据库
func isSystemDatabase(dbname string) bool {
	systemDBs := map[string]bool{
		"postgres":           true,
		"template0":          true,
		"template1":          true,
		"information_schema": true,
	}
	return systemDBs[dbname]
}
