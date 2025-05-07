# PostgreSQL 数据备份还原工具

这是一个用于PostgreSQL数据库备份和还原的命令行工具，支持Windows和Linux系统。

## 功能特点

- 支持数据库备份和还原
- 支持单库和全库操作
- 支持多种备份格式（plain、custom、directory、tar）
- 自动检测数据库连接
- 详细的日志记录
- 跨平台支持（Windows/Linux）

## 系统要求

- Go 1.21 或更高版本
- PostgreSQL 客户端工具（pg_dump、pg_restore、psql）
- Windows 或 Linux 操作系统

## 安装

1. 克隆仓库：
```bash
git clone https://github.com/yourusername/pg-data-tool.git
cd pg-data-tool
```

2. 安装依赖：
```bash
go mod download
```

3. 编译程序：

Windows:
```bash
.\build.bat
```

Linux:
```bash
chmod +x build.sh
./build.sh
```

## 使用方法

### 备份数据库

1. 单库备份：
```bash
# 使用默认格式（plain）
pg-data-tool backup -d mydb

# 指定格式
pg-data-tool backup -d mydb -f custom    # 二进制格式
pg-data-tool backup -d mydb -f directory # 目录格式
pg-data-tool backup -d mydb -f tar       # tar格式
```

2. 全库备份：
```bash
# 使用默认格式
pg-data-tool backup -a

# 指定格式
pg-data-tool backup -a -f custom
```

### 还原数据库

1. 单库还原：
```bash
# SQL格式
pg-data-tool restore -d mydb -f backups/mydb_20240321_123456.sql

# 二进制格式
pg-data-tool restore -d mydb -f backups/mydb_20240321_123456.backup

# tar格式
pg-data-tool restore -d mydb -f backups/mydb_20240321_123456.tar
```

2. 全库还原：
```bash
# 从备份目录还原
pg-data-tool restore -a -f backups/

# 从目录格式备份还原
pg-data-tool restore -a -f backups/all_dbs_20240321_123456.dir
```

### 完整参数示例

```bash
pg-data-tool backup -d mydb \
    -H localhost \
    -p 5432 \
    -U postgres \
    -W your_password \
    -f plain
```

## 参数说明

### 全局参数
- `-H, --host`: 数据库主机地址（默认：localhost）
- `-p, --port`: 数据库端口（默认：5432）
- `-U, --user`: 数据库用户名（默认：postgres）
- `-W, --password`: 数据库密码

### 备份参数
- `-d, --dbname`: 要备份的数据库名称
- `-a, --backup-all`: 备份所有数据库
- `-f, --format`: 备份格式（可选：plain、custom、directory、tar）

### 还原参数
- `-d, --dbname`: 要还原的数据库名称
- `-a, --restore-all`: 还原所有数据库
- `-f, --file`: 备份文件或目录路径

## 备份格式说明

- `plain`: SQL文本格式（默认）
  - 文件扩展名：.sql
  - 优点：可读性好，可直接编辑
  - 适用场景：小型数据库，需要查看或编辑备份内容

- `custom`: 二进制格式
  - 文件扩展名：.backup
  - 优点：压缩率高，支持并行备份
  - 适用场景：大型数据库，需要快速备份还原

- `directory`: 目录格式
  - 文件扩展名：.dir
  - 优点：支持并行备份，每个表一个文件
  - 适用场景：需要选择性还原表

- `tar`: tar归档格式
  - 文件扩展名：.tar
  - 优点：标准归档格式，便于传输
  - 适用场景：需要跨平台传输备份

## 注意事项

1. 确保PostgreSQL客户端工具已正确安装并添加到系统PATH中
2. 备份文件保存在`backups-yymmddHHmiss`目录下，例如：`backups-240321143022/`
3. 备份文件名格式：`数据库名_时分秒.扩展名`
4. 还原操作会覆盖目标数据库中的现有数据，请谨慎操作
5. 全库还原时，如果某个数据库还原失败，会继续处理其他数据库
6. 所有操作都会记录详细日志，存放在`logs`目录下

## 日志说明

日志文件保存在`logs`目录下，格式为：`pg-data-tool_YYYYMMDD.log`

日志包含以下信息：
- 操作类型（备份/还原）
- 连接参数
- 执行命令
- 操作结果
- 错误信息（如果有）

## 常见问题

1. 连接数据库失败
   - 检查数据库服务是否运行
   - 验证连接参数是否正确
   - 确认用户权限是否足够

2. 备份/还原失败
   - 检查磁盘空间是否充足
   - 确认文件权限是否正确
   - 查看详细日志了解具体原因

3. 全库操作失败
   - 检查是否所有数据库都有访问权限
   - 确认备份文件命名是否符合规范
   - 查看日志了解具体失败的数据库 


## 使用示例：

pg-data-tool-windows-amd64.exe --backup --backup-all --host 192.168.12.175 --port 5432 --user postgres --password xxxx

pg-data-tool-windows-amd64.exe --restore --restore-all --host 192.168.12.175 --port 5432 --user postgres --password xxxx --file ./backups-250507
