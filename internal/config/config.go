package config

import (
	"andorralee/internal/repositories"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/docker/docker/client"
	dameng "github.com/godoes/gorm-dameng"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DockerCli *client.Client
	MySQLDB   *gorm.DB
	DamengDB  *gorm.DB
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
	Dameng struct {
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

	// 达梦数据库配置
	config.Dameng.Host = getEnv("DAMENG_HOST", "localhost")
	config.Dameng.Port = getEnv("DAMENG_PORT", "5236")
	config.Dameng.User = getEnv("DAMENG_USER", "SYSDBA")
	config.Dameng.Password = getEnv("DAMENG_PASSWORD", "Dm123456")
	config.Dameng.Database = getEnv("DAMENG_DATABASE", "DOCKER_OPS")

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

// InitMySQL 初始化 MySQL 数据库
func InitMySQL() error {
	config := LoadConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("MySQL 连接失败: " + err.Error())
		return err
	}

	MySQLDB = db
	fmt.Println("MySQL 数据库连接成功")
	return nil
}

// InitDameng 初始化达梦数据库
func InitDameng() error {
	config := LoadConfig()

	options := map[string]string{
		"schema":         "SYSDBA",
		"appName":        "Andorralee Docker API",
		"connectTimeout": "30000",
	}

	// 将端口号转换为整数
	port, err := strconv.Atoi(config.Dameng.Port)
	if err != nil {
		return fmt.Errorf("无效的端口号: %v", err)
	}

	// 构建达梦数据库连接字符串
	dsn := dameng.BuildUrl(
		config.Dameng.User,
		config.Dameng.Password,
		config.Dameng.Host,
		port,
		options,
	)

	// 使用 GORM 打开达梦数据库连接
	db, err := gorm.Open(dameng.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("达梦数据库连接失败: " + err.Error())
		return err
	}

	DamengDB = db
	// 不在这里打印连接成功消息，只在main.go中打印
	return nil
}

// InitTables 初始化数据库表
func InitTables() error {
	if MySQLDB == nil {
		return fmt.Errorf("MySQL数据库未初始化")
	}

	// 自动迁移数据库表结构
	err := MySQLDB.AutoMigrate(
		&repositories.HoneypotTemplate{},
		&repositories.HoneypotInstance{},
		&repositories.HoneypotLog{},
		&repositories.Bait{},
		&repositories.SecurityRule{},
		&repositories.RuleLog{},
	)

	if err != nil {
		fmt.Println("MySQL数据库表初始化失败: " + err.Error())
		return err
	}

	fmt.Println("MySQL数据库表初始化成功")
	return nil
}
