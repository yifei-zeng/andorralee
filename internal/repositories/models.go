package repositories

import (
	"time"
)

// HoneypotTemplate 蜜罐模板模型
type HoneypotTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:50;not null;comment:蜜罐名称"`
	Protocol    string    `json:"protocol" gorm:"size:20;not null;comment:协议类型"`
	ImportTime  time.Time `json:"import_time" gorm:"not null;comment:导入时间"`
	DeployCount int       `json:"deploy_count" gorm:"default:0;comment:已部署数量"`
}

// HoneypotInstance 蜜罐实例模型
type HoneypotInstance struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"size:50;not null;comment:实例名称"`
	HoneypotName  string    `json:"honeypot_name" gorm:"size:100;not null;comment:蜜罐名称"`
	ContainerName string    `json:"container_name" gorm:"size:50;not null;comment:容器名称"`
	ContainerID   string    `json:"container_id" gorm:"size:64;comment:Docker容器ID"`
	IP            string    `json:"ip" gorm:"size:45;not null;comment:IP地址"`
	HoneypotIP    string    `json:"honeypot_ip" gorm:"size:45;comment:蜜罐IP地址"`
	Port          int       `json:"port" gorm:"not null;comment:端口号"`
	Protocol      string    `json:"protocol" gorm:"size:20;not null;comment:协议类型"`
	InterfaceType string    `json:"interface_type" gorm:"size:50;comment:蜜罐接口类型"`
	Status        string    `json:"status" gorm:"size:20;not null;default:created;comment:部署状态"`
	ImageName     string    `json:"image_name" gorm:"size:200;comment:Docker镜像名称"`
	ImageID       string    `json:"image_id" gorm:"size:100;comment:Docker镜像ID"`
	PortMappings  string    `json:"port_mappings" gorm:"type:json;comment:端口映射配置"`
	Environment   string    `json:"environment" gorm:"type:json;comment:环境变量配置"`
	CreateTime    time.Time `json:"create_time" gorm:"not null;comment:创建时间"`
	UpdateTime    time.Time `json:"update_time" gorm:"comment:更新时间"`
	Description   string    `json:"description" gorm:"type:text;comment:描述"`
}

// SecurityRule 安全规则模型
type SecurityRule struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	RuleName          string `json:"rule_name" gorm:"size:50;not null;comment:规则名称"`
	TriggerConditions string `json:"trigger_conditions" gorm:"type:text;not null;comment:触发条件"`
	Actions           string `json:"actions" gorm:"type:text;not null;comment:执行动作"`
	IsEnabled         bool   `json:"is_enabled" gorm:"default:1;comment:启用状态(1启用,0禁用)"`
}

// HoneypotLog 蜜罐日志模型
type HoneypotLog struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	InstanceID uint             `json:"instance_id" gorm:"not null;comment:蜜罐实例ID"`
	Instance   HoneypotInstance `json:"instance" gorm:"foreignKey:InstanceID"`
	LogType    string           `json:"log_type" gorm:"size:20;not null;comment:日志类型"`
	Content    string           `json:"content" gorm:"type:text;not null;comment:日志内容"`
	LogTime    time.Time        `json:"log_time" gorm:"not null;comment:记录时间"`
}

// RuleLog 规则日志模型
type RuleLog struct {
	ID       uint         `json:"id" gorm:"primaryKey"`
	RuleID   uint         `json:"rule_id" gorm:"not null;comment:规则ID"`
	RuleName string       `json:"rule_name" gorm:"size:50;not null;comment:规则名称"`
	Content  string       `json:"content" gorm:"type:text;not null;comment:日志内容"`
	LogTime  time.Time    `json:"log_time" gorm:"not null;comment:记录时间"`
	Rule     SecurityRule `json:"rule" gorm:"foreignKey:RuleID"`
}

// Bait 诱饵模型
type Bait struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	Name       string           `json:"name" gorm:"size:50;not null;comment:诱饵名称"`
	FileType   string           `json:"file_type" gorm:"size:10;not null;comment:文件类型"`
	IsDeployed bool             `json:"is_deployed" gorm:"default:0;comment:投放状态(1已投放,0未投放)"`
	CreateTime time.Time        `json:"create_time" gorm:"not null;comment:创建时间"`
	InstanceID uint             `json:"instance_id" gorm:"comment:关联蜜罐实例"`
	Instance   HoneypotInstance `json:"instance" gorm:"foreignKey:InstanceID"`
}

// TableName 设置表名
func (HoneypotTemplate) TableName() string {
	return "honeypot_template"
}

func (HoneypotInstance) TableName() string {
	return "honeypot_instance"
}

func (SecurityRule) TableName() string {
	return "security_rule"
}

func (HoneypotLog) TableName() string {
	return "honeypot_log"
}

func (RuleLog) TableName() string {
	return "rule_log"
}

func (Bait) TableName() string {
	return "bait"
}

// DockerImage Docker镜像模型
type DockerImage struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ImageID    string    `json:"image_id" gorm:"size:100;not null;comment:镜像ID"`
	Repository string    `json:"repository" gorm:"size:100;comment:仓库名称"`
	Tag        string    `json:"tag" gorm:"size:50;comment:标签"`
	Digest     string    `json:"digest" gorm:"size:100;comment:摘要"`
	Size       int64     `json:"size" gorm:"comment:镜像大小(字节)"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null;comment:创建时间"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"not null;comment:更新时间"`
}

// DockerImageLog Docker镜像操作日志模型
type DockerImageLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ImageID   string    `json:"image_id" gorm:"size:100;comment:镜像ID"`
	ImageName string    `json:"image_name" gorm:"size:200;comment:镜像名称(包含仓库和标签)"`
	Operation string    `json:"operation" gorm:"size:20;not null;comment:操作类型(pull/delete/tag/inspect)"`
	Details   string    `json:"details" gorm:"type:text;comment:操作详情"`
	Status    string    `json:"status" gorm:"size:10;not null;comment:操作状态(success/failed)"`
	Message   string    `json:"message" gorm:"type:text;comment:状态消息"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;comment:创建时间"`
}

// ContainerLogSegment 容器日志分析结果模型
type ContainerLogSegment struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	ContainerID   string     `json:"container_id" gorm:"size:64;not null;comment:容器ID"`
	ContainerName string     `json:"container_name" gorm:"size:100;comment:容器名称"`
	SegmentType   string     `json:"segment_type" gorm:"size:20;not null;comment:日志段类型(error/warning/info/debug)"`
	Content       string     `json:"content" gorm:"type:text;not null;comment:日志内容"`
	Timestamp     *time.Time `json:"timestamp" gorm:"comment:日志时间戳"`
	LineNumber    int        `json:"line_number" gorm:"comment:行号"`
	Component     string     `json:"component" gorm:"size:50;comment:组件名称"`
	SeverityLevel string     `json:"severity_level" gorm:"size:10;comment:严重程度"`
	CreatedAt     time.Time  `json:"created_at" gorm:"not null;comment:分析时间"`
}

// DockerContainer Docker容器管理模型
type DockerContainer struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ContainerID   string    `json:"container_id" gorm:"size:64;not null;comment:Docker容器ID"`
	ContainerName string    `json:"container_name" gorm:"size:100;not null;comment:容器名称"`
	ImageID       string    `json:"image_id" gorm:"size:100;comment:关联的镜像ID"`
	ImageName     string    `json:"image_name" gorm:"size:200;comment:镜像名称"`
	Status        string    `json:"status" gorm:"size:20;comment:容器状态(running/stopped/exited等)"`
	Ports         string    `json:"ports" gorm:"type:json;comment:端口映射信息"`
	Environment   string    `json:"environment" gorm:"type:json;comment:环境变量"`
	CreatedAt     time.Time `json:"created_at" gorm:"not null;comment:创建时间"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"not null;comment:更新时间"`
}

// TableName 设置表名
func (DockerImage) TableName() string {
	return "docker_image"
}

func (DockerImageLog) TableName() string {
	return "docker_image_log"
}

func (ContainerLogSegment) TableName() string {
	return "container_log_segment"
}

func (DockerContainer) TableName() string {
	return "docker_container"
}

// HeadlingAuthLog Headling认证日志模型
type HeadlingAuthLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Timestamp       time.Time `json:"timestamp" gorm:"type:datetime(6);not null;comment:捕获到认证行为的时间戳"`
	AuthID          string    `json:"auth_id" gorm:"size:36;not null;uniqueIndex;comment:此次认证行为的唯一ID"`
	SessionID       string    `json:"session_id" gorm:"size:36;not null;index;comment:所属会话ID"`
	SourceIP        string    `json:"source_ip" gorm:"size:45;not null;index;comment:攻击者IP"`
	SourcePort      uint      `json:"source_port" gorm:"not null;comment:攻击者使用的端口"`
	DestinationIP   string    `json:"destination_ip" gorm:"size:45;not null;index;comment:被攻击的蜜罐容器IP"`
	DestinationPort uint      `json:"destination_port" gorm:"not null;comment:目标端口"`
	Protocol        string    `json:"protocol" gorm:"size:20;not null;index;comment:使用的协议"`
	Username        string    `json:"username" gorm:"size:255;not null;index;comment:攻击者输入的用户名"`
	Password        string    `json:"password" gorm:"size:255;not null;comment:攻击者输入的密码"`
	PasswordHash    string    `json:"password_hash" gorm:"size:255;comment:密码hash值"`
	ContainerID     string    `json:"container_id" gorm:"size:64;index;comment:关联的容器ID"`
	ContainerName   string    `json:"container_name" gorm:"size:100;comment:容器名称"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;comment:记录创建时间"`
}

// HeadlingAuthStatistics Headling认证统计模型
type HeadlingAuthStatistics struct {
	LogDate         string    `json:"log_date"`
	Protocol        string    `json:"protocol"`
	TotalAttempts   int       `json:"total_attempts"`
	UniqueIPs       int       `json:"unique_ips"`
	UniqueUsernames int       `json:"unique_usernames"`
	UniqueSessions  int       `json:"unique_sessions"`
	FirstAttempt    time.Time `json:"first_attempt"`
	LastAttempt     time.Time `json:"last_attempt"`
}

// AttackerIPStatistics 攻击者IP统计模型
type AttackerIPStatistics struct {
	SourceIP              string    `json:"source_ip"`
	TotalAttempts         int       `json:"total_attempts"`
	ProtocolsUsed         int       `json:"protocols_used"`
	UsernamesTried        int       `json:"usernames_tried"`
	PortsTargeted         int       `json:"ports_targeted"`
	FirstSeen             time.Time `json:"first_seen"`
	LastSeen              time.Time `json:"last_seen"`
	AttackDurationMinutes int       `json:"attack_duration_minutes"`
}

func (HeadlingAuthLog) TableName() string {
	return "headling_auth_log"
}

// CowrieLog Cowrie蜜罐日志模型
type CowrieLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	EventTime       time.Time `json:"event_time" gorm:"type:datetime(6);not null;comment:事件发生的精确时间戳"`
	AuthID          string    `json:"auth_id" gorm:"size:36;not null;uniqueIndex;comment:认证行为的唯一ID"`
	SessionID       string    `json:"session_id" gorm:"size:36;not null;index;comment:会话ID"`
	SourceIP        string    `json:"source_ip" gorm:"size:15;not null;index;comment:攻击者IP"`
	SourcePort      uint16    `json:"source_port" gorm:"not null;comment:攻击者使用的端口"`
	DestinationIP   string    `json:"destination_ip" gorm:"size:15;not null;index;comment:蜜罐容器IP"`
	DestinationPort uint16    `json:"destination_port" gorm:"not null;comment:目标端口"`
	Protocol        string    `json:"protocol" gorm:"type:enum('http','ssh','telnet','ftp','smb','other');not null;index;comment:使用的协议类型"`
	ClientInfo      string    `json:"client_info" gorm:"size:255;comment:客户端信息"`
	Fingerprint     string    `json:"fingerprint" gorm:"size:64;comment:客户端指纹"`
	Username        string    `json:"username" gorm:"size:255;index;comment:攻击者输入的用户名"`
	Password        string    `json:"password" gorm:"size:255;comment:攻击者输入的密码"`
	PasswordHash    string    `json:"password_hash" gorm:"size:255;comment:密码哈希值"`
	Command         string    `json:"command" gorm:"type:text;comment:攻击者执行的命令内容"`
	CommandFound    *bool     `json:"command_found" gorm:"index;comment:命令是否被系统识别"`
	RawLog          string    `json:"raw_log" gorm:"type:text;not null;comment:原始日志内容"`
	ContainerID     string    `json:"container_id" gorm:"size:64;index;comment:关联的容器ID"`
	ContainerName   string    `json:"container_name" gorm:"size:100;comment:容器名称"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;comment:记录创建时间"`
}

// CowrieStatistics Cowrie日志统计模型
type CowrieStatistics struct {
	LogDate         string    `json:"log_date"`
	Protocol        string    `json:"protocol"`
	TotalEvents     int       `json:"total_events"`
	UniqueIPs       int       `json:"unique_ips"`
	UniqueSessions  int       `json:"unique_sessions"`
	AuthAttempts    int       `json:"auth_attempts"`
	CommandAttempts int       `json:"command_attempts"`
	ValidCommands   int       `json:"valid_commands"`
	FirstEvent      time.Time `json:"first_event"`
	LastEvent       time.Time `json:"last_event"`
}

// CowrieAttackerBehavior Cowrie攻击者行为统计模型
type CowrieAttackerBehavior struct {
	SourceIP                string    `json:"source_ip"`
	TotalEvents             int       `json:"total_events"`
	ProtocolsUsed           int       `json:"protocols_used"`
	SessionsCreated         int       `json:"sessions_created"`
	AuthAttempts            int       `json:"auth_attempts"`
	CommandsExecuted        int       `json:"commands_executed"`
	ValidCommands           int       `json:"valid_commands"`
	UsernamesTried          int       `json:"usernames_tried"`
	UniqueFingerprints      int       `json:"unique_fingerprints"`
	FirstSeen               time.Time `json:"first_seen"`
	LastSeen                time.Time `json:"last_seen"`
	ActivityDurationMinutes int       `json:"activity_duration_minutes"`
}

// CowrieCommandStatistics Cowrie命令统计模型
type CowrieCommandStatistics struct {
	Command        string    `json:"command"`
	UsageCount     int       `json:"usage_count"`
	UniqueIPs      int       `json:"unique_ips"`
	UniqueSessions int       `json:"unique_sessions"`
	CommandFound   bool      `json:"command_found"`
	FirstUsed      time.Time `json:"first_used"`
	LastUsed       time.Time `json:"last_used"`
}

func (CowrieLog) TableName() string {
	return "cowrie_log"
}

// MalwareSignature 病毒特征模型
type MalwareSignature struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null;comment:特征名称"`
	Pattern     string    `json:"pattern" gorm:"type:text;not null;comment:特征模式"`
	Type        string    `json:"type" gorm:"size:20;not null;comment:类型(hash/string/regex)"`
	Severity    string    `json:"severity" gorm:"size:20;not null;comment:严重程度(low/medium/high/critical)"`
	Description string    `json:"description" gorm:"type:text;comment:描述"`
	CreateTime  time.Time `json:"create_time" gorm:"not null;comment:创建时间"`
	UpdateTime  time.Time `json:"update_time" gorm:"comment:更新时间"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:1;comment:是否激活"`
}

// ScanResult 扫描结果模型
type ScanResult struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FileHash       string    `json:"file_hash" gorm:"size:64;not null;index;comment:文件SHA256哈希"`
	MD5Hash        string    `json:"md5_hash" gorm:"size:32;comment:文件MD5哈希"`
	FileName       string    `json:"file_name" gorm:"size:255;not null;comment:文件名"`
	FileSize       int64     `json:"file_size" gorm:"not null;comment:文件大小(字节)"`
	ScanTime       time.Time `json:"scan_time" gorm:"not null;comment:扫描时间"`
	IsInfected     bool      `json:"is_infected" gorm:"not null;comment:是否感染"`
	ThreatLevel    string    `json:"threat_level" gorm:"size:20;comment:威胁等级"`
	DetectionCount int       `json:"detection_count" gorm:"default:0;comment:检测到的威胁数量"`
	ScanDuration   int64     `json:"scan_duration_ms" gorm:"comment:扫描耗时(毫秒)"`
	SourceIP       string    `json:"source_ip" gorm:"size:45;comment:上传者IP"`
	UserAgent      string    `json:"user_agent" gorm:"type:text;comment:用户代理"`
}

// DetectionResult 检测结果模型
type DetectionResult struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	ScanResultID  uint             `json:"scan_result_id" gorm:"not null;comment:扫描结果ID"`
	ScanResult    ScanResult       `json:"scan_result" gorm:"foreignKey:ScanResultID"`
	SignatureID   uint             `json:"signature_id" gorm:"not null;comment:特征ID"`
	Signature     MalwareSignature `json:"signature" gorm:"foreignKey:SignatureID"`
	SignatureName string           `json:"signature_name" gorm:"size:100;comment:特征名称"`
	MatchType     string           `json:"match_type" gorm:"size:20;comment:匹配类型"`
	MatchContent  string           `json:"match_content" gorm:"type:text;comment:匹配内容"`
	Severity      string           `json:"severity" gorm:"size:20;comment:严重程度"`
	CreatedAt     time.Time        `json:"created_at" gorm:"not null;comment:创建时间"`
}

// AttackSession 攻击会话模型
type AttackSession struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	SessionID       string     `json:"session_id" gorm:"size:36;not null;uniqueIndex;comment:会话唯一ID"`
	SourceIP        string     `json:"source_ip" gorm:"size:45;not null;index;comment:攻击者IP"`
	SourcePort      uint       `json:"source_port" gorm:"comment:攻击者端口"`
	DestinationIP   string     `json:"destination_ip" gorm:"size:45;not null;comment:目标IP"`
	DestinationPort uint       `json:"destination_port" gorm:"not null;comment:目标端口"`
	Protocol        string     `json:"protocol" gorm:"size:20;not null;comment:协议类型"`
	StartTime       time.Time  `json:"start_time" gorm:"not null;comment:会话开始时间"`
	EndTime         *time.Time `json:"end_time" gorm:"comment:会话结束时间"`
	Duration        int64      `json:"duration" gorm:"comment:会话持续时间(秒)"`
	Status          string     `json:"status" gorm:"size:20;not null;default:active;comment:会话状态"`
	AuthAttempts    int        `json:"auth_attempts" gorm:"default:0;comment:认证尝试次数"`
	CommandCount    int        `json:"command_count" gorm:"default:0;comment:执行命令次数"`
	ThreatLevel     string     `json:"threat_level" gorm:"size:20;comment:威胁等级"`
	ContainerID     string     `json:"container_id" gorm:"size:64;comment:关联容器ID"`
	ContainerName   string     `json:"container_name" gorm:"size:100;comment:容器名称"`
	UserAgent       string     `json:"user_agent" gorm:"type:text;comment:用户代理"`
	Fingerprint     string     `json:"fingerprint" gorm:"size:64;comment:客户端指纹"`
}

// AttackEvent 攻击事件模型
type AttackEvent struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	SessionID   string        `json:"session_id" gorm:"size:36;not null;index;comment:关联会话ID"`
	Session     AttackSession `json:"session" gorm:"foreignKey:SessionID;references:SessionID"`
	EventType   string        `json:"event_type" gorm:"size:30;not null;comment:事件类型"`
	EventTime   time.Time     `json:"event_time" gorm:"not null;comment:事件时间"`
	SourceIP    string        `json:"source_ip" gorm:"size:45;not null;index;comment:攻击者IP"`
	Protocol    string        `json:"protocol" gorm:"size:20;not null;comment:协议"`
	Username    string        `json:"username" gorm:"size:255;comment:用户名"`
	Password    string        `json:"password" gorm:"size:255;comment:密码"`
	Command     string        `json:"command" gorm:"type:text;comment:执行的命令"`
	Payload     string        `json:"payload" gorm:"type:text;comment:攻击载荷"`
	Result      string        `json:"result" gorm:"type:text;comment:执行结果"`
	ThreatLevel string        `json:"threat_level" gorm:"size:20;comment:威胁等级"`
	IsBlocked   bool          `json:"is_blocked" gorm:"default:0;comment:是否被阻断"`
	RawData     string        `json:"raw_data" gorm:"type:text;comment:原始数据"`
}

// ThreatIntelligence 威胁情报模型
type ThreatIntelligence struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	IndicatorType  string    `json:"indicator_type" gorm:"size:20;not null;comment:指标类型(ip/domain/hash/url)"`
	IndicatorValue string    `json:"indicator_value" gorm:"size:255;not null;index;comment:指标值"`
	ThreatType     string    `json:"threat_type" gorm:"size:50;comment:威胁类型"`
	Confidence     int       `json:"confidence" gorm:"comment:置信度(0-100)"`
	Severity       string    `json:"severity" gorm:"size:20;comment:严重程度"`
	Source         string    `json:"source" gorm:"size:100;comment:情报来源"`
	Description    string    `json:"description" gorm:"type:text;comment:描述"`
	FirstSeen      time.Time `json:"first_seen" gorm:"comment:首次发现时间"`
	LastSeen       time.Time `json:"last_seen" gorm:"comment:最后发现时间"`
	IsActive       bool      `json:"is_active" gorm:"column:is_active;default:1;comment:是否激活"`

	Tags           string    `json:"tags" gorm:"type:json;comment:标签"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null;comment:创建时间"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// HoneytokenEvent 蜜签事件模型
type HoneytokenEvent struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	TokenID      string    `json:"token_id" gorm:"size:36;not null;index;comment:蜜签ID"`
	TokenType    string    `json:"token_type" gorm:"size:50;not null;comment:蜜签类型"`
	TokenName    string    `json:"token_name" gorm:"size:100;comment:蜜签名称"`
	TriggerTime  time.Time `json:"trigger_time" gorm:"not null;comment:触发时间"`
	SourceIP     string    `json:"source_ip" gorm:"size:45;not null;index;comment:触发者IP"`
	UserAgent    string    `json:"user_agent" gorm:"type:text;comment:用户代理"`
	RequestPath  string    `json:"request_path" gorm:"size:500;comment:请求路径"`
	RequestData  string    `json:"request_data" gorm:"type:text;comment:请求数据"`
	ResponseCode int       `json:"response_code" gorm:"comment:响应码"`
	Location     string    `json:"location" gorm:"size:255;comment:蜜签位置"`
	Description  string    `json:"description" gorm:"type:text;comment:描述"`
	ThreatLevel  string    `json:"threat_level" gorm:"size:20;comment:威胁等级"`
	IsProcessed  bool      `json:"is_processed" gorm:"default:0;comment:是否已处理"`
}

// TableName 设置表名
func (MalwareSignature) TableName() string {
	return "malware_signature"
}

func (ScanResult) TableName() string {
	return "scan_result"
}

func (DetectionResult) TableName() string {
	return "detection_result"
}

func (AttackSession) TableName() string {
	return "attack_session"
}

func (AttackEvent) TableName() string {
	return "attack_event"
}

func (ThreatIntelligence) TableName() string {
	return "threat_intelligence"
}

func (HoneytokenEvent) TableName() string {
	return "honeytoken_event"
}
