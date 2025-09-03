package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ANSI 颜色常量
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

var (
	infoLogger   *log.Logger
	warnLogger   *log.Logger
	errorLogger  *log.Logger
	logFile      *os.File
	colorEnabled bool
)

// colorize 为文本添加颜色（仅在控制台输出时）
func colorize(color, text string) string {
	if colorEnabled {
		return color + text + Reset
	}
	return text
}

// isColorTerminal 检查是否支持颜色输出
func isColorTerminal() bool {
	// 在 Windows 上检查是否支持 ANSI 颜色
	if runtime.GOOS == "windows" {
		// Windows 10 及以上版本支持 ANSI 颜色
		return true
	}
	// Unix 系统通常都支持颜色
	return true
}

// coloredWriter 包装 Writer 以在控制台输出时添加颜色
type coloredWriter struct {
	consoleWriter io.Writer
	fileWriter    io.Writer
	color         string
	prefix        string
}

func (cw *coloredWriter) Write(p []byte) (n int, err error) {
	// 向文件写入无颜色的内容
	if cw.fileWriter != nil {
		cw.fileWriter.Write(p)
	}

	// 向控制台写入带颜色的内容
	if cw.consoleWriter != nil && colorEnabled {
		coloredContent := string(p)
		// 替换前缀为带颜色的前缀
		if cw.prefix != "" {
			coloredPrefix := colorize(cw.color+Bold, cw.prefix)
			// 查找并替换前缀
			if len(coloredContent) > len(cw.prefix) {
				coloredContent = coloredPrefix + coloredContent[len(cw.prefix):]
			}
		}
		return cw.consoleWriter.Write([]byte(coloredContent))
	}

	return cw.consoleWriter.Write(p)
}

// Init 初始化日志系统
func Init() error {
	// 检查是否支持颜色输出
	colorEnabled = isColorTerminal()

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

	// 创建带颜色的Writer
	infoWriter := &coloredWriter{
		consoleWriter: os.Stdout,
		fileWriter:    logFile,
		color:         Green,
		prefix:        "[INFO] ",
	}

	warnWriter := &coloredWriter{
		consoleWriter: os.Stdout,
		fileWriter:    logFile,
		color:         Yellow,
		prefix:        "[WARN] ",
	}

	errorWriter := &coloredWriter{
		consoleWriter: os.Stdout,
		fileWriter:    logFile,
		color:         Red,
		prefix:        "[ERROR] ",
	}

	// 初始化日志记录器
	infoLogger = log.New(infoWriter, "[INFO] ", log.Ldate|log.Ltime)
	warnLogger = log.New(warnWriter, "[WARN] ", log.Ldate|log.Ltime)
	errorLogger = log.New(errorWriter, "[ERROR] ", log.Ldate|log.Ltime)

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

// Warn 记录警告日志
func Warn(format string, v ...interface{}) {
	if warnLogger == nil {
		Init()
	}
	warnLogger.Printf(format, v...)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	if errorLogger == nil {
		Init()
	}
	errorLogger.Printf(format, v...)
}

// Success 记录成功日志（绿色）
func Success(format string, v ...interface{}) {
	if infoLogger == nil {
		Init()
	}
	// 使用特殊的成功前缀
	if colorEnabled {
		successMsg := fmt.Sprintf(format, v...)
		fmt.Printf("%s[SUCCESS]%s %s %s\n",
			Bold+Green, Reset,
			time.Now().Format("2006/01/02 15:04:05"),
			successMsg)
		// 同时写入文件
		if logFile != nil {
			logFile.WriteString(fmt.Sprintf("[SUCCESS] %s %s\n",
				time.Now().Format("2006/01/02 15:04:05"),
				successMsg))
		}
	} else {
		infoLogger.Printf(format, v...)
	}
}

// Fatal 记录致命错误并退出程序
func Fatal(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	}
	os.Exit(1)
}

// SetColorEnabled 设置是否启用颜色输出
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// IsColorEnabled 返回是否启用了颜色输出
func IsColorEnabled() bool {
	return colorEnabled
}
