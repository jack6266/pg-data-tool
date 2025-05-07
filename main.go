package main

import (
	"fmt"
	"os"

	"pg-data-tool/cmd"
	"pg-data-tool/internal/logger"
)

var Version string

func main() {
	// 初始化日志系统
	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// 执行命令
	if err := cmd.Execute(); err != nil {
		logger.Error("执行命令失败: %v", err)
		os.Exit(1)
	}
}
