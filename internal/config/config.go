package config

// Config 存储配置信息
type Config struct {
	// 源数据库连接信息
	Host     string
	Port     string
	User     string
	Password string
	DBName   string

	// 目标数据库连接信息（用于数据传输）
	TargetHost     string
	TargetPort     string
	TargetUser     string
	TargetPassword string

	// 备份相关配置
	BackupAll bool
	Format    string
	File      string

	// 还原相关配置
	RestoreAll   bool
	AutoCreateDB bool // 是否自动创建数据库

	// 传输相关配置
	TransferAll       bool
	IncludeBlobs      bool
	IncludeIndexes    bool
	IncludePrivileges bool

	// 并行处理相关配置
	UseParallel  bool
	ParallelJobs int
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{
		Host:           "localhost",
		Port:           "5432",
		User:           "postgres",
		Format:         "plain",
		ParallelJobs:   4,
		TransferAll:    false,
		UseParallel:    false,
		Password:       "Pw!123456", // 设置源数据库默认密码
		TargetPassword: "Pw!123456", // 设置目标数据库默认密码
	}
}
