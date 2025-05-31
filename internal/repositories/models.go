package repositories

import (
	"time"
)

// HoneypotTemplate 蜜罐模板模型
type HoneypotTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url" gorm:"not null"`
	Port        int       `json:"port"`
	Type        string    `json:"type" gorm:"not null"` // web, ssh, ftp, etc.
	DeployCount int       `json:"deploy_count" gorm:"default:0"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// HoneypotInstance 蜜罐实例模型
type HoneypotInstance struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	Name          string           `json:"name" gorm:"not null"`
	Description   string           `json:"description"`
	TemplateID    uint             `json:"template_id" gorm:"not null"`
	Template      HoneypotTemplate `json:"template" gorm:"foreignKey:TemplateID"`
	Status        string           `json:"status" gorm:"not null"` // created, running, stopped, error
	Port          int              `json:"port"`
	IP            string           `json:"ip"`
	ContainerName string           `json:"container_name"`
	ContainerID   string           `json:"container_id"`
	CreateTime    time.Time        `json:"create_time"`
	UpdateTime    time.Time        `json:"update_time"`
}

// HoneypotLog 蜜罐日志模型
type HoneypotLog struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	InstanceID uint             `json:"instance_id" gorm:"not null"`
	Instance   HoneypotInstance `json:"instance" gorm:"foreignKey:InstanceID"`
	Type       string           `json:"type"`  // access, attack, info, etc.
	Level      string           `json:"level"` // info, warning, error, etc.
	Content    string           `json:"content" gorm:"not null"`
	SourceIP   string           `json:"source_ip"`
	CreateTime time.Time        `json:"create_time"`
}

// Bait 诱饵/蜜签模型
type Bait struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	Name        string           `json:"name" gorm:"not null"`
	Description string           `json:"description"`
	Type        string           `json:"type" gorm:"not null"` // file, link, credential, etc.
	Content     string           `json:"content" gorm:"not null"`
	InstanceID  uint             `json:"instance_id"`
	Instance    HoneypotInstance `json:"instance" gorm:"foreignKey:InstanceID"`
	IsDeployed  bool             `json:"is_deployed" gorm:"default:false"`
	DeployTime  time.Time        `json:"deploy_time"`
	CreateTime  time.Time        `json:"create_time"`
	UpdateTime  time.Time        `json:"update_time"`
}

// SecurityRule 安全规则模型
type SecurityRule struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Type        string    `json:"type" gorm:"not null"` // detection, prevention, response, etc.
	Condition   string    `json:"condition" gorm:"not null"`
	Action      string    `json:"action" gorm:"not null"`
	IsEnabled   bool      `json:"is_enabled" gorm:"default:true"`
	Priority    int       `json:"priority" gorm:"default:0"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// RuleLog 规则执行日志模型
type RuleLog struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	RuleID     uint             `json:"rule_id" gorm:"not null"`
	Rule       SecurityRule     `json:"rule" gorm:"foreignKey:RuleID"`
	InstanceID uint             `json:"instance_id"`
	Instance   HoneypotInstance `json:"instance" gorm:"foreignKey:InstanceID"`
	Result     string           `json:"result" gorm:"not null"` // success, failed, etc.
	Details    string           `json:"details"`
	CreateTime time.Time        `json:"create_time"`
}

// TableName 设置表名
func (HoneypotTemplate) TableName() string {
	return "honeypot_template"
}

func (HoneypotInstance) TableName() string {
	return "honeypot_instance"
}

func (HoneypotLog) TableName() string {
	return "honeypot_log"
}

func (Bait) TableName() string {
	return "bait"
}

func (SecurityRule) TableName() string {
	return "security_rule"
}

func (RuleLog) TableName() string {
	return "rule_log"
}
