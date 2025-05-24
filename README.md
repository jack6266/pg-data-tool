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

#### 1. 单库备份

##### 基本用法
```bash
# 使用默认格式（plain）和默认密码
pg-data-tool-windows-amd64.exe --backup --dbname mydb

# 指定格式和默认密码
pg-data-tool-windows-amd64.exe --backup --dbname mydb --format custom    # 二进制格式
pg-data-tool-windows-amd64.exe --backup --dbname mydb --format directory # 目录格式
pg-data-tool-windows-amd64.exe --backup --dbname mydb --format tar       # tar格式
```

##### 控制台输出
执行备份时，会在控制台显示数据库的详细信息，格式如下：
```
数据库基本信息:
数据库名称: mydb
数据库大小: 1.2GB
表数量: 10
索引数量: 15
视图数量: 2
函数数量: 5
```

##### 备份信息
每次备份都会生成一个额外的 `.info` 文件，包含以下信息：
- 备份时间
- 数据库连接信息
- 数据库基本信息：
  - 数据库大小
  - 表数量
  - 索引数量
  - 视图数量
  - 函数数量
- 表详细信息：
  - 表名
  - 表大小
  - 行数
  - 索引数量

示例备份信息文件（.info）：
```json
{
  "backup_time": "2024-03-21 14:30:22",
  "host": "localhost",
  "port": "5432",
  "user": "postgres",
  "database": "mydb",
  "format": "custom",
  "file": "backups-240321143022-localhost/mydb_143022.backup",
  "database_info": {
    "size": "1.2GB",
    "table_count": 10,
    "index_count": 15,
    "view_count": 2,
    "function_count": 5,
    "tables": [
      {
        "name": "users",
        "size": "256MB",
        "row_count": 10000,
        "index_count": 3
      },
      {
        "name": "orders",
        "size": "512MB",
        "row_count": 50000,
        "index_count": 4
      }
    ]
  }
}
```

##### 日志信息
备份过程中的日志会记录以下信息：
```
[INFO] 开始执行数据库备份操作
[INFO] 连接参数: 主机=localhost, 端口=5432, 用户=postgres
[INFO] 备份格式: custom
[INFO] 备份文件将保存在: backups-240321143022-localhost
[INFO] 开始备份数据库: mydb
[INFO] 数据库基本信息:
[INFO] - 数据库大小: 1.2GB
[INFO] - 表数量: 10
[INFO] - 索引数量: 15
[INFO] - 视图数量: 2
[INFO] - 函数数量: 5
[INFO] 执行命令: pg_dump -h localhost -p 5432 -U postgres -F custom -v -f backups-240321143022-localhost/mydb_143022.backup mydb
[INFO] 数据库 mydb 备份成功，文件保存在: backups-240321143022-localhost/mydb_143022.backup
```

2. 全库备份：
```bash
# 使用默认格式和默认密码
pg-data-tool-windows-amd64.exe --backup --backup-all

# 指定格式和默认密码
pg-data-tool-windows-amd64.exe --backup --backup-all --format custom
```

### 还原数据库

1. 单库还原：
```bash
# SQL格式和默认密码
pg-data-tool-windows-amd64.exe --restore --dbname mydb --file backups/mydb_20240321_123456.sql

# 二进制格式和默认密码
pg-data-tool-windows-amd64.exe --restore --dbname mydb --file backups/mydb_20240321_123456.backup

# tar格式和默认密码
pg-data-tool-windows-amd64.exe --restore --dbname mydb --file backups/mydb_20240321_123456.tar
```

2. 全库还原：
```bash
# 从备份目录还原（使用默认密码）
pg-data-tool-windows-amd64.exe --restore --restore-all --file backups/

# 从目录格式备份还原（使用默认密码）
pg-data-tool-windows-amd64.exe --restore --restore-all --file backups/all_dbs_20240321_123456.dir
```

### 完整参数示例

```bash
# 使用默认密码
pg-data-tool-windows-amd64.exe --backup \
    --dbname mydb \
    --host localhost \
    --port 5432 \
    --user postgres \
    --format plain

# 指定自定义密码
pg-data-tool-windows-amd64.exe --backup \
    --dbname mydb \
    --host localhost \
    --port 5432 \
    --user postgres \
    --password your_password \
    --format plain
```

## 参数说明

### 全局参数
- `-H, --host`: 数据库主机地址（默认：localhost）
- `-p, --port`: 数据库端口（默认：5432）
- `-U, --user`: 数据库用户名（默认：postgres）
- `-W, --password`: 数据库密码（默认：Pw!123456）

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