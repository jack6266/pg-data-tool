package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pg-data-tool/internal/logger"
)

// GetBackupFiles 获取目录下的所有备份文件
func GetBackupFiles(backupPath string) ([]string, error) {
	var backupFiles []string
	err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".sql" || ext == ".backup" || ext == ".tar" {
				backupFiles = append(backupFiles, path)
			}
		}
		return nil
	})
	return backupFiles, err
}

// ExtractDatabaseName 从备份文件名中提取数据库名
func ExtractDatabaseName(fileName string) string {
	logger.Info("正在从文件名提取数据库名: %s", fileName)

	// 移除文件扩展名
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	logger.Info("移除扩展名后的文件名: %s", baseName)

	// 找到最后一个下划线的位置
	lastUnderscoreIndex := strings.LastIndex(baseName, "_")
	if lastUnderscoreIndex == -1 {
		logger.Error("文件名 %s 中没有找到下划线分隔符", fileName)
		return ""
	}

	// 取最后一个下划线之前的部分作为数据库名
	dbName := baseName[:lastUnderscoreIndex]
	logger.Info("提取到的数据库名: %s", dbName)
	return dbName
}

// ValidateBackupFile 验证备份文件
func ValidateBackupFile(backupPath string, isRestoreAll bool) error {
	// 检查文件或目录是否存在
	fileInfo, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("错误：备份文件或目录 %s 不存在", backupPath)
	}

	if isRestoreAll {
		if !fileInfo.IsDir() && !strings.HasSuffix(backupPath, ".dir") {
			return fmt.Errorf("错误：全库还原需要指定备份目录或目录格式的备份文件")
		}
	}

	return nil
}
