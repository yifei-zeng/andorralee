package config

import (
	"andorralee/internal/repositories"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	DBMode string // mysql | dameng
	MySQL  struct {
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

	// 数据库模式 (默认 mysql)
	config.DBMode = strings.ToLower(getEnv("DB_MODE", "mysql"))

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

	port, err := strconv.Atoi(config.Dameng.Port)
	if err != nil {
		return fmt.Errorf("无效的端口号: %v", err)
	}

	// 构建达梦数据库连接字符串 (dm://user:pwd@host:port)
	dsn := dameng.BuildUrl(
		config.Dameng.User,
		config.Dameng.Password,
		config.Dameng.Host,
		port,
		options,
	)

	db, err := gorm.Open(dameng.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("达梦数据库连接失败: " + err.Error())
		return err
	}
	DamengDB = db
	return nil
}

// InitTables 初始化数据库表
func InitTables() error {
	// 优先使用当前模式选择的数据库
	active := GetActiveDB()
	if active == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 自动迁移数据库表结构（GORM 跨驱动相同调用）
	err := active.AutoMigrate(
		&repositories.HoneypotTemplate{},
		&repositories.HoneypotInstance{},
		&repositories.HoneypotLog{},
		&repositories.Bait{},
		&repositories.SecurityRule{},
		&repositories.RuleLog{},
		&repositories.DockerImage{},
		&repositories.DockerImageLog{},
		&repositories.ContainerLogSegment{},
		&repositories.DockerContainer{},
		&repositories.HeadlingAuthLog{},
		&repositories.CowrieLog{},
		&repositories.MalwareSignature{},
		&repositories.ScanResult{},
		&repositories.DetectionResult{},
		&repositories.AttackSession{},
		&repositories.AttackEvent{},
		&repositories.ThreatIntelligence{},
		&repositories.HoneytokenEvent{},
	)

	if err != nil {
		fmt.Println("数据库表初始化失败: " + err.Error())
		return err
	}
	fmt.Println("数据库表初始化成功")
	return nil
}

// InitDamengTables 初始化达梦数据库表
func InitDamengTables() error {
	if DamengDB == nil {
		return fmt.Errorf("达梦数据库未初始化")
	}

	// 自动迁移数据库表结构
	err := DamengDB.AutoMigrate(
		&repositories.HoneypotTemplate{},
		&repositories.HoneypotInstance{},
		&repositories.HoneypotLog{},
		&repositories.Bait{},
		&repositories.SecurityRule{},
		&repositories.RuleLog{},
		&repositories.DockerImage{},
		&repositories.DockerImageLog{},
		&repositories.ContainerLogSegment{},
		&repositories.DockerContainer{},
		&repositories.HeadlingAuthLog{},
		&repositories.CowrieLog{},
		&repositories.MalwareSignature{},
		&repositories.ScanResult{},
		&repositories.DetectionResult{},
		&repositories.AttackSession{},
		&repositories.AttackEvent{},
		&repositories.ThreatIntelligence{},
		&repositories.HoneytokenEvent{},
	)

	if err != nil {
		fmt.Println("达梦数据库表初始化失败: " + err.Error())
		return err
	}

	fmt.Println("达梦数据库表初始化成功")
	return nil
}

// GetActiveDB 根据 DB_MODE 返回当前激活的 *gorm.DB
func GetActiveDB() *gorm.DB {
	mode := strings.ToLower(getEnv("DB_MODE", "mysql"))
	if mode == "dameng" {
		if DamengDB != nil {
			return DamengDB
		}
		// 回退到 MySQL
		return MySQLDB
	}
	// 默认 mysql
	if MySQLDB != nil {
		return MySQLDB
	}
	return DamengDB
}

// GetDBMode 返回当前数据库模式
func GetDBMode() string {
	return strings.ToLower(getEnv("DB_MODE", "mysql"))
}

// GetDBByMode 根据传入模式返回对应数据库连接(可能为nil)
func GetDBByMode(mode string) *gorm.DB {
	switch strings.ToLower(mode) {
	case "mysql":
		return MySQLDB
	case "dameng":
		return DamengDB
	default:
		return nil
	}
}

// SchemaCheckResult 结构自检结果
type SchemaCheckResult struct {
	MissingTables []string
	MissingColumns map[string][]string
}

// SelfCheckSchema 简单自检：检查关键表与列是否存在（达梦/ MySQL 通用）
func SelfCheckSchema() SchemaCheckResult {
	db := GetActiveDB()
	result := SchemaCheckResult{MissingTables: []string{}, MissingColumns: map[string][]string{}}
	if db == nil { return result }

	// INFORMATION_SCHEMA 在不同数据库差异，这里采用 GORM Migrator 接口尝试
	migrator := db.Migrator()
	required := map[string][]string{
		"malware_signature":      {"id","name","pattern","type","severity","is_active"},
		"threat_intelligence":    {"id","indicator_type","indicator_value","severity","is_active"},
	}
	for tbl, cols := range required {
		if !migrator.HasTable(tbl) {
			result.MissingTables = append(result.MissingTables, tbl)
			continue
		}
		missCols := []string{}
		for _, c := range cols {
			if !migrator.HasColumn(tbl, c) { missCols = append(missCols, c) }
		}
		if len(missCols) > 0 { result.MissingColumns[tbl] = missCols }
	}
	return result
}

// LogSchemaCheck 执行并输出结果
func LogSchemaCheck() {
	res := SelfCheckSchema()
	if len(res.MissingTables)==0 && len(res.MissingColumns)==0 {
		fmt.Println("[SchemaCheck] 所有关键表与列存在")
		return
	}
	fmt.Println("[SchemaCheck] 检测到结构缺失:")
	if len(res.MissingTables)>0 { fmt.Println("  缺失表:", strings.Join(res.MissingTables, ",")) }
	for tbl, cols := range res.MissingColumns { fmt.Printf("  表 %s 缺失列: %s\n", tbl, strings.Join(cols, ",")) }
}

// 延迟执行自检：等待迁移完成
func InitSchemaSelfCheckAsync() {
	go func(){
		time.Sleep(2 * time.Second)
		LogSchemaCheck()
	}()
}
