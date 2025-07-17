# 蜜罐容器日志结构设计

## 📋 需求分析

根据项目需求，容器日志需要记录：
1. **连接时间与IP** - 攻击者连接的时间和IP地址
2. **协议与端口** - 使用的协议类型和端口信息
3. **用户名与密码** - 认证尝试的凭据信息
4. **命令与响应** - 执行的命令和系统响应
5. **会话持续时长** - 从连接到断开的完整时间

## 🗄️ 容器日志数据库结构

### 1. 主日志表 `container_runtime_log`

```sql
CREATE TABLE `container_runtime_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `log_id` VARCHAR(36) NOT NULL COMMENT '日志唯一ID',
    `container_id` VARCHAR(64) NOT NULL COMMENT 'Docker容器ID',
    `container_name` VARCHAR(100) COMMENT '容器名称',
    `image_name` VARCHAR(200) COMMENT '镜像名称',
    `log_timestamp` DATETIME(6) NOT NULL COMMENT '日志产生时间戳(微秒精度)',
    `log_level` ENUM('DEBUG','INFO','WARN','ERROR','FATAL') NOT NULL COMMENT '日志级别',
    `event_type` ENUM('connection','authentication','command','response','disconnection','error','system') NOT NULL COMMENT '事件类型',
    `source_ip` VARCHAR(45) COMMENT '源IP地址(支持IPv6)',
    `source_port` SMALLINT UNSIGNED COMMENT '源端口',
    `destination_ip` VARCHAR(45) COMMENT '目标IP地址',
    `destination_port` SMALLINT UNSIGNED COMMENT '目标端口',
    `protocol` ENUM('tcp','udp','http','https','ssh','telnet','ftp','smtp','mysql','redis','other') NOT NULL COMMENT '协议类型',
    `session_id` VARCHAR(36) COMMENT '会话ID',
    `username` VARCHAR(255) COMMENT '用户名',
    `password` VARCHAR(255) COMMENT '密码(明文)',
    `password_hash` VARCHAR(255) COMMENT '密码哈希值',
    `auth_success` BOOLEAN COMMENT '认证是否成功',
    `command` TEXT COMMENT '执行的命令',
    `command_args` TEXT COMMENT '命令参数',
    `response` TEXT COMMENT '命令响应内容',
    `response_code` INT COMMENT '响应状态码',
    `execution_time_ms` INT COMMENT '命令执行时间(毫秒)',
    `user_agent` VARCHAR(500) COMMENT '用户代理字符串',
    `client_info` VARCHAR(255) COMMENT '客户端信息',
    `fingerprint` VARCHAR(64) COMMENT '客户端指纹',
    `request_headers` JSON COMMENT 'HTTP请求头',
    `response_headers` JSON COMMENT 'HTTP响应头',
    `request_body` TEXT COMMENT '请求体内容',
    `response_body` TEXT COMMENT '响应体内容',
    `file_path` VARCHAR(500) COMMENT '涉及的文件路径',
    `file_operation` ENUM('read','write','create','delete','modify','execute') COMMENT '文件操作类型',
    `process_id` INT COMMENT '进程ID',
    `process_name` VARCHAR(100) COMMENT '进程名称',
    `parent_process_id` INT COMMENT '父进程ID',
    `cpu_usage` DECIMAL(5,2) COMMENT 'CPU使用率(%)',
    `memory_usage` BIGINT COMMENT '内存使用量(字节)',
    `network_bytes_in` BIGINT COMMENT '网络入流量(字节)',
    `network_bytes_out` BIGINT COMMENT '网络出流量(字节)',
    `error_message` TEXT COMMENT '错误信息',
    `stack_trace` TEXT COMMENT '错误堆栈',
    `raw_log_line` TEXT NOT NULL COMMENT '原始日志行',
    `parsed_fields` JSON COMMENT '解析出的额外字段',
    `tags` JSON COMMENT '标签信息',
    `severity_score` TINYINT COMMENT '威胁严重程度评分(1-10)',
    `is_malicious` BOOLEAN DEFAULT FALSE COMMENT '是否为恶意行为',
    `detection_rules` JSON COMMENT '触发的检测规则',
    `geolocation` JSON COMMENT 'IP地理位置信息',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_log_id` (`log_id`),
    KEY `idx_container_id` (`container_id`),
    KEY `idx_timestamp` (`log_timestamp`),
    KEY `idx_source_ip` (`source_ip`),
    KEY `idx_session_id` (`session_id`),
    KEY `idx_event_type` (`event_type`),
    KEY `idx_protocol` (`protocol`),
    KEY `idx_username` (`username`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='容器运行时日志表';
```

### 2. 会话汇总表 `container_session_summary`

```sql
CREATE TABLE `container_session_summary` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `session_id` VARCHAR(36) NOT NULL COMMENT '会话唯一ID',
    `container_id` VARCHAR(64) NOT NULL COMMENT 'Docker容器ID',
    `container_name` VARCHAR(100) COMMENT '容器名称',
    `source_ip` VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    `source_port` SMALLINT UNSIGNED COMMENT '攻击者端口',
    `destination_ip` VARCHAR(45) COMMENT '目标IP',
    `destination_port` SMALLINT UNSIGNED COMMENT '目标端口',
    `protocol` VARCHAR(20) NOT NULL COMMENT '主要协议',
    `start_time` DATETIME(6) NOT NULL COMMENT '会话开始时间',
    `end_time` DATETIME(6) COMMENT '会话结束时间',
    `duration_seconds` INT COMMENT '会话持续时长(秒)',
    `total_events` INT DEFAULT 0 COMMENT '总事件数',
    `connection_events` INT DEFAULT 0 COMMENT '连接事件数',
    `auth_attempts` INT DEFAULT 0 COMMENT '认证尝试次数',
    `successful_auths` INT DEFAULT 0 COMMENT '成功认证次数',
    `failed_auths` INT DEFAULT 0 COMMENT '失败认证次数',
    `command_executions` INT DEFAULT 0 COMMENT '命令执行次数',
    `successful_commands` INT DEFAULT 0 COMMENT '成功命令数',
    `failed_commands` INT DEFAULT 0 COMMENT '失败命令数',
    `file_operations` INT DEFAULT 0 COMMENT '文件操作次数',
    `error_events` INT DEFAULT 0 COMMENT '错误事件数',
    `unique_usernames` INT DEFAULT 0 COMMENT '尝试的用户名数量',
    `unique_passwords` INT DEFAULT 0 COMMENT '尝试的密码数量',
    `unique_commands` INT DEFAULT 0 COMMENT '执行的唯一命令数',
    `total_bytes_in` BIGINT DEFAULT 0 COMMENT '总入流量(字节)',
    `total_bytes_out` BIGINT DEFAULT 0 COMMENT '总出流量(字节)',
    `max_cpu_usage` DECIMAL(5,2) COMMENT '最大CPU使用率',
    `max_memory_usage` BIGINT COMMENT '最大内存使用量',
    `client_info` VARCHAR(255) COMMENT '客户端信息',
    `user_agent` VARCHAR(500) COMMENT '用户代理',
    `fingerprint` VARCHAR(64) COMMENT '客户端指纹',
    `geolocation` JSON COMMENT 'IP地理位置',
    `threat_level` ENUM('low','medium','high','critical') COMMENT '威胁等级',
    `is_successful_breach` BOOLEAN DEFAULT FALSE COMMENT '是否成功入侵',
    `attack_patterns` JSON COMMENT '识别的攻击模式',
    `session_status` ENUM('active','completed','timeout','error') DEFAULT 'active' COMMENT '会话状态',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_session_id` (`session_id`),
    KEY `idx_container_id` (`container_id`),
    KEY `idx_source_ip` (`source_ip`),
    KEY `idx_start_time` (`start_time`),
    KEY `idx_protocol` (`protocol`),
    KEY `idx_threat_level` (`threat_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='容器会话汇总表';
```

### 3. 攻击模式表 `attack_pattern_detection`

```sql
CREATE TABLE `attack_pattern_detection` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `detection_id` VARCHAR(36) NOT NULL COMMENT '检测唯一ID',
    `container_id` VARCHAR(64) NOT NULL COMMENT '容器ID',
    `session_id` VARCHAR(36) COMMENT '关联会话ID',
    `source_ip` VARCHAR(45) NOT NULL COMMENT '攻击源IP',
    `pattern_type` ENUM('brute_force','sql_injection','xss','command_injection','directory_traversal','port_scan','dos','malware','backdoor','privilege_escalation','data_exfiltration','other') NOT NULL COMMENT '攻击模式类型',
    `pattern_name` VARCHAR(100) NOT NULL COMMENT '模式名称',
    `confidence_score` DECIMAL(3,2) NOT NULL COMMENT '置信度(0.00-1.00)',
    `severity_level` ENUM('info','low','medium','high','critical') NOT NULL COMMENT '严重程度',
    `detection_time` DATETIME(6) NOT NULL COMMENT '检测时间',
    `evidence_logs` JSON COMMENT '证据日志ID列表',
    `attack_vector` VARCHAR(200) COMMENT '攻击向量',
    `target_service` VARCHAR(50) COMMENT '目标服务',
    `payload` TEXT COMMENT '攻击载荷',
    `mitigation_action` VARCHAR(100) COMMENT '缓解措施',
    `is_blocked` BOOLEAN DEFAULT FALSE COMMENT '是否已阻止',
    `false_positive` BOOLEAN DEFAULT FALSE COMMENT '是否为误报',
    `analyst_notes` TEXT COMMENT '分析师备注',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_detection_id` (`detection_id`),
    KEY `idx_container_id` (`container_id`),
    KEY `idx_session_id` (`session_id`),
    KEY `idx_source_ip` (`source_ip`),
    KEY `idx_pattern_type` (`pattern_type`),
    KEY `idx_detection_time` (`detection_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='攻击模式检测表';
```

## 📊 日志字段说明

### 核心字段（满足考核要求）

| 字段名 | 类型 | 说明 | 考核要求对应 |
|--------|------|------|-------------|
| `log_timestamp` | DATETIME(6) | 日志产生的精确时间戳 | ✅ 连接时间 |
| `source_ip` | VARCHAR(45) | 攻击者IP地址 | ✅ IP信息 |
| `source_port` | SMALLINT | 攻击者端口 | ✅ 端口信息 |
| `destination_port` | SMALLINT | 目标端口 | ✅ 端口信息 |
| `protocol` | ENUM | 协议类型 | ✅ 协议类型 |
| `username` | VARCHAR(255) | 用户名 | ✅ 用户名 |
| `password` | VARCHAR(255) | 密码 | ✅ 密码 |
| `command` | TEXT | 执行的命令 | ✅ 输入命令 |
| `response` | TEXT | 命令响应 | ✅ 命令响应 |
| `start_time` | DATETIME(6) | 会话开始时间 | ✅ 连接时间 |
| `end_time` | DATETIME(6) | 会话结束时间 | ✅ 会话关闭时间 |
| `duration_seconds` | INT | 会话持续时长 | ✅ 持续时长 |

### 扩展字段（增强功能）

| 分类 | 字段 | 说明 |
|------|------|------|
| **网络信息** | `user_agent`, `client_info`, `fingerprint` | 客户端识别 |
| **HTTP详情** | `request_headers`, `response_headers`, `request_body` | HTTP协议详情 |
| **系统资源** | `cpu_usage`, `memory_usage`, `process_id` | 系统性能监控 |
| **文件操作** | `file_path`, `file_operation` | 文件系统活动 |
| **安全分析** | `severity_score`, `is_malicious`, `detection_rules` | 威胁检测 |
| **地理位置** | `geolocation` | IP地理位置信息 |

## 🔧 Go语言模型定义

```go
// ContainerRuntimeLog 容器运行时日志模型
type ContainerRuntimeLog struct {
    ID                uint       `json:"id" gorm:"primaryKey"`
    LogID             string     `json:"log_id" gorm:"size:36;not null;uniqueIndex;comment:日志唯一ID"`
    ContainerID       string     `json:"container_id" gorm:"size:64;not null;index;comment:Docker容器ID"`
    ContainerName     string     `json:"container_name" gorm:"size:100;comment:容器名称"`
    ImageName         string     `json:"image_name" gorm:"size:200;comment:镜像名称"`
    LogTimestamp      time.Time  `json:"log_timestamp" gorm:"type:datetime(6);not null;index;comment:日志产生时间戳"`
    LogLevel          string     `json:"log_level" gorm:"type:enum('DEBUG','INFO','WARN','ERROR','FATAL');not null;comment:日志级别"`
    EventType         string     `json:"event_type" gorm:"type:enum('connection','authentication','command','response','disconnection','error','system');not null;index;comment:事件类型"`
    SourceIP          string     `json:"source_ip" gorm:"size:45;index;comment:源IP地址"`
    SourcePort        *uint16    `json:"source_port" gorm:"comment:源端口"`
    DestinationIP     string     `json:"destination_ip" gorm:"size:45;comment:目标IP地址"`
    DestinationPort   *uint16    `json:"destination_port" gorm:"comment:目标端口"`
    Protocol          string     `json:"protocol" gorm:"type:enum('tcp','udp','http','https','ssh','telnet','ftp','smtp','mysql','redis','other');not null;index;comment:协议类型"`
    SessionID         string     `json:"session_id" gorm:"size:36;index;comment:会话ID"`
    Username          string     `json:"username" gorm:"size:255;index;comment:用户名"`
    Password          string     `json:"password" gorm:"size:255;comment:密码"`
    PasswordHash      string     `json:"password_hash" gorm:"size:255;comment:密码哈希值"`
    AuthSuccess       *bool      `json:"auth_success" gorm:"comment:认证是否成功"`
    Command           string     `json:"command" gorm:"type:text;comment:执行的命令"`
    CommandArgs       string     `json:"command_args" gorm:"type:text;comment:命令参数"`
    Response          string     `json:"response" gorm:"type:text;comment:命令响应内容"`
    ResponseCode      *int       `json:"response_code" gorm:"comment:响应状态码"`
    ExecutionTimeMs   *int       `json:"execution_time_ms" gorm:"comment:命令执行时间(毫秒)"`
    UserAgent         string     `json:"user_agent" gorm:"size:500;comment:用户代理字符串"`
    ClientInfo        string     `json:"client_info" gorm:"size:255;comment:客户端信息"`
    Fingerprint       string     `json:"fingerprint" gorm:"size:64;comment:客户端指纹"`
    RequestHeaders    string     `json:"request_headers" gorm:"type:json;comment:HTTP请求头"`
    ResponseHeaders   string     `json:"response_headers" gorm:"type:json;comment:HTTP响应头"`
    RequestBody       string     `json:"request_body" gorm:"type:text;comment:请求体内容"`
    ResponseBody      string     `json:"response_body" gorm:"type:text;comment:响应体内容"`
    FilePath          string     `json:"file_path" gorm:"size:500;comment:涉及的文件路径"`
    FileOperation     string     `json:"file_operation" gorm:"type:enum('read','write','create','delete','modify','execute');comment:文件操作类型"`
    ProcessID         *int       `json:"process_id" gorm:"comment:进程ID"`
    ProcessName       string     `json:"process_name" gorm:"size:100;comment:进程名称"`
    ParentProcessID   *int       `json:"parent_process_id" gorm:"comment:父进程ID"`
    CPUUsage          *float64   `json:"cpu_usage" gorm:"type:decimal(5,2);comment:CPU使用率(%)"`
    MemoryUsage       *int64     `json:"memory_usage" gorm:"comment:内存使用量(字节)"`
    NetworkBytesIn    *int64     `json:"network_bytes_in" gorm:"comment:网络入流量(字节)"`
    NetworkBytesOut   *int64     `json:"network_bytes_out" gorm:"comment:网络出流量(字节)"`
    ErrorMessage      string     `json:"error_message" gorm:"type:text;comment:错误信息"`
    StackTrace        string     `json:"stack_trace" gorm:"type:text;comment:错误堆栈"`
    RawLogLine        string     `json:"raw_log_line" gorm:"type:text;not null;comment:原始日志行"`
    ParsedFields      string     `json:"parsed_fields" gorm:"type:json;comment:解析出的额外字段"`
    Tags              string     `json:"tags" gorm:"type:json;comment:标签信息"`
    SeverityScore     *int       `json:"severity_score" gorm:"comment:威胁严重程度评分(1-10)"`
    IsMalicious       bool       `json:"is_malicious" gorm:"default:false;comment:是否为恶意行为"`
    DetectionRules    string     `json:"detection_rules" gorm:"type:json;comment:触发的检测规则"`
    Geolocation       string     `json:"geolocation" gorm:"type:json;comment:IP地理位置信息"`
    CreatedAt         time.Time  `json:"created_at" gorm:"not null;comment:记录创建时间"`
    UpdatedAt         time.Time  `json:"updated_at" gorm:"not null;comment:记录更新时间"`
}

func (ContainerRuntimeLog) TableName() string {
    return "container_runtime_log"
}
```

## 📈 使用场景示例

### 1. SSH连接日志
```json
{
    "log_id": "550e8400-e29b-41d4-a716-446655440001",
    "container_id": "abc123",
    "log_timestamp": "2025-06-29T21:30:15.123456Z",
    "log_level": "INFO",
    "event_type": "connection",
    "source_ip": "192.168.1.100",
    "source_port": 54321,
    "destination_port": 22,
    "protocol": "ssh",
    "session_id": "ssh-session-001",
    "raw_log_line": "2025-06-29 21:30:15 [INFO] SSH connection from 192.168.1.100:54321"
}
```

### 2. 认证尝试日志
```json
{
    "log_id": "550e8400-e29b-41d4-a716-446655440002",
    "container_id": "abc123",
    "log_timestamp": "2025-06-29T21:30:20.456789Z",
    "log_level": "WARN",
    "event_type": "authentication",
    "source_ip": "192.168.1.100",
    "protocol": "ssh",
    "session_id": "ssh-session-001",
    "username": "admin",
    "password": "123456",
    "auth_success": false,
    "raw_log_line": "2025-06-29 21:30:20 [WARN] Failed login attempt for user 'admin'"
}
```

### 3. 命令执行日志
```json
{
    "log_id": "550e8400-e29b-41d4-a716-446655440003",
    "container_id": "abc123",
    "log_timestamp": "2025-06-29T21:30:25.789012Z",
    "log_level": "INFO",
    "event_type": "command",
    "source_ip": "192.168.1.100",
    "protocol": "ssh",
    "session_id": "ssh-session-001",
    "username": "root",
    "command": "ls -la /etc",
    "response": "total 1024\ndrwxr-xr-x 2 root root 4096 Jan 1 12:00 .",
    "execution_time_ms": 45,
    "raw_log_line": "2025-06-29 21:30:25 [INFO] Command executed: ls -la /etc"
}
```

这个容器日志结构完全满足您的项目需求，并提供了丰富的扩展功能用于安全分析和威胁检测。
