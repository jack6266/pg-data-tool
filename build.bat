@echo off
chcp 65001 > nul
echo 开始构建 pg-data-tool...

:: 设置版本号
set VERSION=1.0.0

:: 创建构建目录
if not exist build mkdir build

:: 构建Windows版本
echo 构建Windows版本...
set GOOS=windows
set GOARCH=amd64
go build -o build/pg-data-tool-windows-amd64.exe -ldflags "-X main.Version=%VERSION%"

:: 构建Linux版本
echo 构建Linux版本...
set GOOS=linux
set GOARCH=amd64
go build -o build/pg-data-tool-linux-amd64 -ldflags "-X main.Version=%VERSION%"

echo 构建完成！
echo 输出文件：
echo - build/pg-data-tool-windows-amd64.exe
echo - build/pg-data-tool-linux-amd64 