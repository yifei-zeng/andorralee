package config

// HoneypotType 蜜罐类型
type HoneypotType string

const (
	HoneypotTypeSSH    HoneypotType = "ssh"    // SSH蜜罐
	HoneypotTypeHTTP   HoneypotType = "http"   // HTTP蜜罐
	HoneypotTypeMySQL  HoneypotType = "mysql"  // MySQL蜜罐
	HoneypotTypeRedis  HoneypotType = "redis"  // Redis蜜罐
	HoneypotTypeCustom HoneypotType = "custom" // 自定义蜜罐
)

// HoneypotResources 蜜罐资源限制
type HoneypotResources struct {
	CPULimit    string `json:"cpu_limit"`
	MemoryLimit string `json:"memory_limit"`
}

// HoneypotConfig 蜜罐配置
type HoneypotConfig struct {
	Type        HoneypotType      `json:"type"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Ports       map[string]string `json:"ports"`
	Environment map[string]string `json:"environment"`
	Resources   HoneypotResources `json:"resources"`
}

// DefaultHoneypotConfigs 默认蜜罐配置
var DefaultHoneypotConfigs = map[HoneypotType]HoneypotConfig{
	HoneypotTypeSSH: {
		Type:  HoneypotTypeSSH,
		Name:  "ssh-honeypot",
		Image: "dtagdevsec/sshpot:latest",
		Ports: map[string]string{
			"22": "2222",
		},
		Environment: map[string]string{
			"SSH_PORT": "22",
		},
		Resources: HoneypotResources{
			CPULimit:    "0.5",
			MemoryLimit: "512m",
		},
	},
	HoneypotTypeHTTP: {
		Type:  HoneypotTypeHTTP,
		Name:  "http-honeypot",
		Image: "dtagdevsec/httppot:latest",
		Ports: map[string]string{
			"80": "8080",
		},
		Environment: map[string]string{
			"HTTP_PORT": "80",
		},
		Resources: HoneypotResources{
			CPULimit:    "0.5",
			MemoryLimit: "512m",
		},
	},
	HoneypotTypeMySQL: {
		Type:  HoneypotTypeMySQL,
		Name:  "mysql-honeypot",
		Image: "dtagdevsec/mysqlpot:latest",
		Ports: map[string]string{
			"3306": "3306",
		},
		Environment: map[string]string{
			"MYSQL_PORT": "3306",
		},
		Resources: HoneypotResources{
			CPULimit:    "0.5",
			MemoryLimit: "512m",
		},
	},
	HoneypotTypeRedis: {
		Type:  HoneypotTypeRedis,
		Name:  "redis-honeypot",
		Image: "dtagdevsec/redispot:latest",
		Ports: map[string]string{
			"6379": "6379",
		},
		Environment: map[string]string{
			"REDIS_PORT": "6379",
		},
		Resources: HoneypotResources{
			CPULimit:    "0.5",
			MemoryLimit: "512m",
		},
	},
}
