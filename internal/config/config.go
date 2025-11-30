package config

import (
	"andorralee/internal/repositories"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/client"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DockerCli *client.Client
	MySQLDB   *gorm.DB
	SyncToken = getEnv("SYNC_TOKEN", "sync-dev-token")
)

// Config 应用配置
type Config struct {
	MySQL struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
	}
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	config := &Config{}

	// MySQL配置
	config.MySQL.Host = getEnv("MYSQL_HOST", "localhost")
	config.MySQL.Port = getEnv("MYSQL_PORT", "3306")
	config.MySQL.User = getEnv("MYSQL_USER", "root")
	config.MySQL.Password = getEnv("MYSQL_PASSWORD", "123456")
	config.MySQL.Database = getEnv("MYSQL_DATABASE", "andorralee")

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// InitDockerClient 初始化 Docker 客户端
func InitDockerClient() error {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		fmt.Println("Docker 客户端初始化失败: " + err.Error())
		return err
	}

	// 验证连接
	if _, err := cli.Ping(context.Background()); err != nil {
		fmt.Println("Docker 连接失败: " + err.Error())
		return err
	}

	DockerCli = cli
	fmt.Println("Docker 客户端初始化成功")
	return nil
}

// InitDatabase 初始化数据库
func InitDatabase() error {
	config := LoadConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		fmt.Println("数据库连接失败: " + err.Error())
		return err
	}

	MySQLDB = db
	fmt.Println("数据库连接成功")
	return nil
}

// InitTables 初始化数据库表
func InitTables() error {
	if MySQLDB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 自动迁移数据库表结构
	// 注意外键依赖顺序，例如 attack_session 依赖 attack_event(session_id) 的外键，需先迁移被引用表
	err := MySQLDB.AutoMigrate(
		&repositories.HoneypotTemplate{},
		&repositories.HoneypotInstance{},
		&repositories.HoneypotLog{},
		&repositories.SecurityRule{},
		&repositories.RuleLog{},
		&repositories.DockerImage{},
		&repositories.DockerImageLog{},
		&repositories.ContainerLogSegment{},
		&repositories.DockerContainer{},
		&repositories.HeraldingAuthLog{},
		&repositories.CowrieLog{},
		&repositories.MySQLHoneypotLog{},
		&repositories.MalwareSignature{},
		// 先迁移攻击事件，再迁移会话，避免外键顺序问题
		&repositories.AttackEvent{},
		&repositories.AttackSession{},
		&repositories.ScanResult{},
		&repositories.DetectionResult{},
		&repositories.ThreatIntelligence{},
		&repositories.SyncError{},
	)

	if err != nil {
		// 若为 GORM 在迁移期间尝试删除 attack_session.uk_session_id 导致的外键依赖错误，降级为告警并继续
		if strings.Contains(err.Error(), "Cannot drop index 'uk_session_id'") {
			fmt.Println("警告: 检测到 uk_session_id 被外键依赖，保持现有索引并继续")
			return nil
		}
		fmt.Println("数据库表初始化失败: " + err.Error())
		return err
	}

	fmt.Println("数据库表初始化成功")
	return nil
}
