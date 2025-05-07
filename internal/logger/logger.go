package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	logFile     *os.File
)

// Init 初始化日志系统
func Init() error {
	// 创建logs目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 生成日志文件名，格式：pg-data-tool_YYYYMMDD.log
	logFileName := filepath.Join("logs", fmt.Sprintf("pg-data-tool_%s.log", time.Now().Format("20060102")))

	// 打开日志文件
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 创建同时输出到控制台和文件的Writer
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// 初始化日志记录器
	infoLogger = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime)

	return nil
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	if infoLogger == nil {
		Init()
	}
	infoLogger.Printf(format, v...)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	if errorLogger == nil {
		Init()
	}
	errorLogger.Printf(format, v...)
}

// Fatal 记录致命错误并退出程序
func Fatal(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	}
	os.Exit(1)
}
