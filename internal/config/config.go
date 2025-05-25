package config

// Config 存储数据库连接和操作配置
type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	File         string
	BackupAll    bool   // 是否备份所有数据库
	Format       string // 备份格式：custom, plain, directory, tar
	RestoreAll   bool   // 是否还原所有数据库
	AutoCreateDB bool   // 是否自动创建数据库
	ParallelJobs int    // 并行作业数，用于 pg_dump 的 -j 参数
	UseParallel  bool   // 是否使用并行处理，默认为 false
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "postgres",
		Format:       "plain", // 默认使用plain格式（SQL文本格式）
		Password:     "Pw!123456",
		ParallelJobs: 4,     // 默认使用4个并行作业
		UseParallel:  false, // 默认不使用并行处理
	}
}
