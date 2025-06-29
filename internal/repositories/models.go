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

// HoneypotSession 蜜罐会话模型
type HoneypotSession struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	SessionID       string     `json:"session_id" gorm:"size:36;not null;uniqueIndex;comment:会话唯一ID"`
	SourceIP        string     `json:"source_ip" gorm:"size:45;not null;index;comment:攻击者IP"`
	SourcePort      uint16     `json:"source_port" gorm:"not null;comment:攻击者端口"`
	DestinationIP   string     `json:"destination_ip" gorm:"size:45;not null;comment:目标IP"`
	DestinationPort uint16     `json:"destination_port" gorm:"not null;comment:目标端口"`
	Protocol        string     `json:"protocol" gorm:"size:20;not null;index;comment:协议类型"`
	ContainerID     string     `json:"container_id" gorm:"size:64;index;comment:关联的容器ID"`
	ContainerName   string     `json:"container_name" gorm:"size:100;comment:容器名称"`
	StartTime       time.Time  `json:"start_time" gorm:"not null;comment:会话开始时间"`
	EndTime         *time.Time `json:"end_time" gorm:"comment:会话结束时间"`
	DurationSeconds *int       `json:"duration_seconds" gorm:"comment:会话持续时长(秒)"`
	Status          string     `json:"status" gorm:"type:enum('active','closed','timeout');default:'active';comment:会话状态"`
	EventCount      int        `json:"event_count" gorm:"default:0;comment:事件数量"`
	AuthAttempts    int        `json:"auth_attempts" gorm:"default:0;comment:认证尝试次数"`
	CommandCount    int        `json:"command_count" gorm:"default:0;comment:命令执行次数"`
	LastActivity    time.Time  `json:"last_activity" gorm:"not null;comment:最后活动时间"`
	UserAgent       string     `json:"user_agent" gorm:"size:500;comment:用户代理"`
	ClientInfo      string     `json:"client_info" gorm:"size:255;comment:客户端信息"`
	Fingerprint     string     `json:"fingerprint" gorm:"size:64;comment:客户端指纹"`
	CreatedAt       time.Time  `json:"created_at" gorm:"not null;comment:记录创建时间"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"not null;comment:记录更新时间"`
}

func (HoneypotSession) TableName() string {
	return "honeypot_session"
}

// SessionEvent 会话事件模型
type SessionEvent struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"size:36;not null;index;comment:会话ID"`
	EventType string    `json:"event_type" gorm:"type:enum('connect','auth','command','disconnect','error');not null;comment:事件类型"`
	EventTime time.Time `json:"event_time" gorm:"type:datetime(6);not null;comment:事件时间"`
	Username  string    `json:"username" gorm:"size:255;comment:用户名"`
	Password  string    `json:"password" gorm:"size:255;comment:密码"`
	Command   string    `json:"command" gorm:"type:text;comment:命令内容"`
	Response  string    `json:"response" gorm:"type:text;comment:响应内容"`
	Success   *bool     `json:"success" gorm:"comment:操作是否成功"`
	ErrorMsg  string    `json:"error_msg" gorm:"size:500;comment:错误信息"`
	Details   string    `json:"details" gorm:"type:text;comment:详细信息"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;comment:记录创建时间"`
}

func (SessionEvent) TableName() string {
	return "session_event"
}

// ContainerRuntimeLog 容器运行时日志模型
type ContainerRuntimeLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	LogID           string    `json:"log_id" gorm:"size:36;not null;uniqueIndex;comment:日志唯一ID"`
	ContainerID     string    `json:"container_id" gorm:"size:64;not null;index;comment:Docker容器ID"`
	ContainerName   string    `json:"container_name" gorm:"size:100;comment:容器名称"`
	ImageName       string    `json:"image_name" gorm:"size:200;comment:镜像名称"`
	LogTimestamp    time.Time `json:"log_timestamp" gorm:"type:datetime(6);not null;index;comment:日志产生时间戳"`
	LogLevel        string    `json:"log_level" gorm:"type:enum('DEBUG','INFO','WARN','ERROR','FATAL');not null;comment:日志级别"`
	EventType       string    `json:"event_type" gorm:"type:enum('connection','authentication','command','response','disconnection','error','system');not null;index;comment:事件类型"`
	SourceIP        string    `json:"source_ip" gorm:"size:45;index;comment:源IP地址"`
	SourcePort      *uint16   `json:"source_port" gorm:"comment:源端口"`
	DestinationIP   string    `json:"destination_ip" gorm:"size:45;comment:目标IP地址"`
	DestinationPort *uint16   `json:"destination_port" gorm:"comment:目标端口"`
	Protocol        string    `json:"protocol" gorm:"type:enum('tcp','udp','http','https','ssh','telnet','ftp','smtp','mysql','redis','other');not null;index;comment:协议类型"`
	SessionID       string    `json:"session_id" gorm:"size:36;index;comment:会话ID"`
	Username        string    `json:"username" gorm:"size:255;index;comment:用户名"`
	Password        string    `json:"password" gorm:"size:255;comment:密码"`
	PasswordHash    string    `json:"password_hash" gorm:"size:255;comment:密码哈希值"`
	AuthSuccess     *bool     `json:"auth_success" gorm:"comment:认证是否成功"`
	Command         string    `json:"command" gorm:"type:text;comment:执行的命令"`
	CommandArgs     string    `json:"command_args" gorm:"type:text;comment:命令参数"`
	Response        string    `json:"response" gorm:"type:text;comment:命令响应内容"`
	ResponseCode    *int      `json:"response_code" gorm:"comment:响应状态码"`
	ExecutionTimeMs *int      `json:"execution_time_ms" gorm:"comment:命令执行时间(毫秒)"`
	UserAgent       string    `json:"user_agent" gorm:"size:500;comment:用户代理字符串"`
	ClientInfo      string    `json:"client_info" gorm:"size:255;comment:客户端信息"`
	Fingerprint     string    `json:"fingerprint" gorm:"size:64;comment:客户端指纹"`
	RequestHeaders  string    `json:"request_headers" gorm:"type:json;comment:HTTP请求头"`
	ResponseHeaders string    `json:"response_headers" gorm:"type:json;comment:HTTP响应头"`
	RequestBody     string    `json:"request_body" gorm:"type:text;comment:请求体内容"`
	ResponseBody    string    `json:"response_body" gorm:"type:text;comment:响应体内容"`
	FilePath        string    `json:"file_path" gorm:"size:500;comment:涉及的文件路径"`
	FileOperation   string    `json:"file_operation" gorm:"type:enum('read','write','create','delete','modify','execute');comment:文件操作类型"`
	ProcessID       *int      `json:"process_id" gorm:"comment:进程ID"`
	ProcessName     string    `json:"process_name" gorm:"size:100;comment:进程名称"`
	ParentProcessID *int      `json:"parent_process_id" gorm:"comment:父进程ID"`
	CPUUsage        *float64  `json:"cpu_usage" gorm:"type:decimal(5,2);comment:CPU使用率(%)"`
	MemoryUsage     *int64    `json:"memory_usage" gorm:"comment:内存使用量(字节)"`
	NetworkBytesIn  *int64    `json:"network_bytes_in" gorm:"comment:网络入流量(字节)"`
	NetworkBytesOut *int64    `json:"network_bytes_out" gorm:"comment:网络出流量(字节)"`
	ErrorMessage    string    `json:"error_message" gorm:"type:text;comment:错误信息"`
	StackTrace      string    `json:"stack_trace" gorm:"type:text;comment:错误堆栈"`
	RawLogLine      string    `json:"raw_log_line" gorm:"type:text;not null;comment:原始日志行"`
	ParsedFields    string    `json:"parsed_fields" gorm:"type:json;comment:解析出的额外字段"`
	Tags            string    `json:"tags" gorm:"type:json;comment:标签信息"`
	SeverityScore   *int      `json:"severity_score" gorm:"comment:威胁严重程度评分(1-10)"`
	IsMalicious     bool      `json:"is_malicious" gorm:"default:false;comment:是否为恶意行为"`
	DetectionRules  string    `json:"detection_rules" gorm:"type:json;comment:触发的检测规则"`
	Geolocation     string    `json:"geolocation" gorm:"type:json;comment:IP地理位置信息"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;comment:记录创建时间"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null;comment:记录更新时间"`
}

func (ContainerRuntimeLog) TableName() string {
	return "container_runtime_log"
}

// ContainerSessionSummary 容器会话汇总模型
type ContainerSessionSummary struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	SessionID          string     `json:"session_id" gorm:"size:36;not null;uniqueIndex;comment:会话唯一ID"`
	ContainerID        string     `json:"container_id" gorm:"size:64;not null;index;comment:Docker容器ID"`
	ContainerName      string     `json:"container_name" gorm:"size:100;comment:容器名称"`
	SourceIP           string     `json:"source_ip" gorm:"size:45;not null;index;comment:攻击者IP"`
	SourcePort         *uint16    `json:"source_port" gorm:"comment:攻击者端口"`
	DestinationIP      string     `json:"destination_ip" gorm:"size:45;comment:目标IP"`
	DestinationPort    *uint16    `json:"destination_port" gorm:"comment:目标端口"`
	Protocol           string     `json:"protocol" gorm:"size:20;not null;index;comment:主要协议"`
	StartTime          time.Time  `json:"start_time" gorm:"type:datetime(6);not null;index;comment:会话开始时间"`
	EndTime            *time.Time `json:"end_time" gorm:"type:datetime(6);comment:会话结束时间"`
	DurationSeconds    *int       `json:"duration_seconds" gorm:"comment:会话持续时长(秒)"`
	TotalEvents        int        `json:"total_events" gorm:"default:0;comment:总事件数"`
	ConnectionEvents   int        `json:"connection_events" gorm:"default:0;comment:连接事件数"`
	AuthAttempts       int        `json:"auth_attempts" gorm:"default:0;comment:认证尝试次数"`
	SuccessfulAuths    int        `json:"successful_auths" gorm:"default:0;comment:成功认证次数"`
	FailedAuths        int        `json:"failed_auths" gorm:"default:0;comment:失败认证次数"`
	CommandExecutions  int        `json:"command_executions" gorm:"default:0;comment:命令执行次数"`
	SuccessfulCommands int        `json:"successful_commands" gorm:"default:0;comment:成功命令数"`
	FailedCommands     int        `json:"failed_commands" gorm:"default:0;comment:失败命令数"`
	FileOperations     int        `json:"file_operations" gorm:"default:0;comment:文件操作次数"`
	ErrorEvents        int        `json:"error_events" gorm:"default:0;comment:错误事件数"`
	UniqueUsernames    int        `json:"unique_usernames" gorm:"default:0;comment:尝试的用户名数量"`
	UniquePasswords    int        `json:"unique_passwords" gorm:"default:0;comment:尝试的密码数量"`
	UniqueCommands     int        `json:"unique_commands" gorm:"default:0;comment:执行的唯一命令数"`
	TotalBytesIn       int64      `json:"total_bytes_in" gorm:"default:0;comment:总入流量(字节)"`
	TotalBytesOut      int64      `json:"total_bytes_out" gorm:"default:0;comment:总出流量(字节)"`
	MaxCPUUsage        *float64   `json:"max_cpu_usage" gorm:"type:decimal(5,2);comment:最大CPU使用率"`
	MaxMemoryUsage     *int64     `json:"max_memory_usage" gorm:"comment:最大内存使用量"`
	ClientInfo         string     `json:"client_info" gorm:"size:255;comment:客户端信息"`
	UserAgent          string     `json:"user_agent" gorm:"size:500;comment:用户代理"`
	Fingerprint        string     `json:"fingerprint" gorm:"size:64;comment:客户端指纹"`
	Geolocation        string     `json:"geolocation" gorm:"type:json;comment:IP地理位置"`
	ThreatLevel        string     `json:"threat_level" gorm:"type:enum('low','medium','high','critical');comment:威胁等级"`
	IsSuccessfulBreach bool       `json:"is_successful_breach" gorm:"default:false;comment:是否成功入侵"`
	AttackPatterns     string     `json:"attack_patterns" gorm:"type:json;comment:识别的攻击模式"`
	SessionStatus      string     `json:"session_status" gorm:"type:enum('active','completed','timeout','error');default:'active';comment:会话状态"`
	CreatedAt          time.Time  `json:"created_at" gorm:"not null;comment:记录创建时间"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"not null;comment:记录更新时间"`
}

func (ContainerSessionSummary) TableName() string {
	return "container_session_summary"
}
