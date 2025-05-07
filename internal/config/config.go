package config

// Config 存储数据库连接和操作配置
type Config struct {
	Host       string
	Port       string
	User       string
	Password   string
	DBName     string
	File       string
	BackupAll  bool   // 是否备份所有数据库
	Format     string // 备份格式：custom, plain, directory, tar
	RestoreAll bool   // 是否还原所有数据库
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{
		Host:   "localhost",
		Port:   "5432",
		User:   "postgres",
		Format: "plain", // 默认使用plain格式（SQL文本格式）
	}
}
