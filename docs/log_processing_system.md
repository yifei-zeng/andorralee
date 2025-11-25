# 📊 日志处理系统技术文档

## 📋 系统概述

日志处理系统是Andorralee蜜罐管理系统的核心分析组件，负责收集、解析、分析和存储来自各种蜜罐容器的日志数据。系统支持多种日志格式，提供实时分析、会话管理和攻击行为识别功能。

## 🏗️ 技术架构

### 核心技术栈
- **编程语言**: Go 1.23.4
- **Web框架**: Gin (RESTful API)
- **存储方式**: 混合存储 (MySQL数据库 + 内存缓存)
- **日志格式**: JSON、CSV、纯文本
- **分析引擎**: 基于规则的语义分析 + 统计分析

### 系统架构图
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   日志收集      │───▶│   解析引擎      │───▶│   存储层        │
│ (多种蜜罐)      │    │ (多格式支持)    │    │ (数据库+内存)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ 会话管理        │    │ 语义分析        │    │ 统计报告        │
│ 生命周期跟踪    │    │ 攻击行为识别    │    │ 实时监控        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 代码结构

### 主要文件位置
```
internal/handlers/container_log_handler.go          # 容器日志语义分析 (数据库存储)
internal/handlers/container_runtime_log_handler.go  # 容器运行时日志 (内存存储)
internal/handlers/session_handler.go                # 会话管理系统
internal/handlers/heralding_handler.go               # Heralding认证日志处理
internal/handlers/cowrie_handler.go                 # Cowrie蜜罐日志处理
internal/services/ai.go                             # 日志语义分析服务 (基于规则匹配)
routers/router.go                                   # API路由配置
```

### 路由配置位置
```
routers/router.go:308-331   # 容器日志分析接口
routers/router.go:343-359   # 会话管理接口  
routers/router.go:130-149   # Heralding认证日志接口
routers/router.go:151-175   # Cowrie蜜罐日志接口
```

## 🔍 日志处理流程

### 1. 日志收集阶段
```
容器日志 → 格式识别 → 内容提取 → 初步解析 → 存储队列
```

### 2. 解析分析阶段  
```
原始日志 → 语义分割 → 事件提取 → 会话关联 → 威胁分析
```

### 3. 存储管理阶段
```
结构化数据 → 数据库存储 → 索引建立 → 缓存更新 → 统计计算
```

## 🧩 核心组件详解

### 1. 容器日志语义分析系统

#### 核心数据结构
```go
// 位置: internal/repositories/container_log_segment.go
type ContainerLogSegment struct {
    ID            uint       `json:"id"`             // 分段ID
    ContainerID   string     `json:"container_id"`   // 容器ID
    ContainerName string     `json:"container_name"` // 容器名称
    SegmentType   string     `json:"segment_type"`   // 分段类型
    Content       string     `json:"content"`        // 日志内容
    LineNumber    int        `json:"line_number"`    // 行号
    Component     string     `json:"component"`      // 组件名称
    SeverityLevel string     `json:"severity_level"` // 严重程度
    Timestamp     *time.Time `json:"timestamp"`      // 时间戳
}
```

#### 日志语义分析引擎
```go
// 位置: internal/services/ai.go:55-130
func AnalyzeContainerLogs(logContent string) ([]LogSegmentInfo, map[string]int) {
    // 按行分割日志
    logLines := strings.Split(logContent, "\n")

    // 定义正则表达式匹配日志格式
    logPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z)\s+(\w+)\s+\[([^\]]+)\]\s+(.+)`)

    // 定义日志类型识别规则
    errorPattern := regexp.MustCompile(`(?i)(error|exception|fail|failed|crash)`)
    warningPattern := regexp.MustCompile(`(?i)(warn|warning|caution)`)
    infoPattern := regexp.MustCompile(`(?i)(info|information|notice)`)
    debugPattern := regexp.MustCompile(`(?i)(debug|trace|verbose)`)

    var segments []LogSegmentInfo
    stats := map[string]int{
        "error": 0, "warning": 0, "info": 0, "debug": 0, "unknown": 0,
    }

    // 分析每行日志
    for lineNumber, line := range logLines {
        if line == "" { continue }

        matches := logPattern.FindStringSubmatch(line)
        var segment LogSegmentInfo

        if len(matches) >= 5 {
            // 匹配成功，提取信息
            segment = LogSegmentInfo{
                Timestamp:  matches[1],
                Level:      matches[2],
                Component:  matches[3],
                Message:    matches[4],
                LineNumber: lineNumber + 1,
            }
        } else {
            // 无法匹配标准格式，作为原始消息处理
            segment = LogSegmentInfo{
                Timestamp:  time.Now().Format(time.RFC3339),
                Level:      "UNKNOWN",
                Component:  "system",
                Message:    line,
                LineNumber: lineNumber + 1,
            }
        }

        // 确定日志类型 (基于规则匹配，非AI)
        if errorPattern.MatchString(segment.Message) {
            segment.Type = "error"
            stats["error"]++
        } else if warningPattern.MatchString(segment.Message) {
            segment.Type = "warning"
            stats["warning"]++
        } else if infoPattern.MatchString(segment.Message) {
            segment.Type = "info"
            stats["info"]++
        } else {
            segment.Type = "debug"
            stats["debug"]++
        }

        segments = append(segments, segment)
    }

    return segments, stats
}
```

#### API接口实现
```go
// 位置: internal/handlers/container_log_handler.go

// 获取所有日志分析结果
// API: GET /api/v1/container-logs/segments
func GetAllContainerLogSegments(c *gin.Context)

// 根据容器ID获取分析结果  
// API: GET /api/v1/container-logs/segments/container/:container_id
func GetLogSegmentsByContainerID(c *gin.Context)

// 根据类型获取分析结果
// API: GET /api/v1/container-logs/segments/type/:type  
func GetLogSegmentsByType(c *gin.Context)

// 删除分析结果
// API: DELETE /api/v1/container-logs/segments/:id
func DeleteContainerLogSegment(c *gin.Context)
```

### 2. 容器运行时日志处理系统

#### 核心数据结构
```go
// 位置: internal/handlers/container_runtime_log_handler.go:16-28
type ContainerRuntimeLog struct {
    ID          uint                   `json:"id"`          // 日志ID
    ContainerID string                 `json:"container_id"` // 容器ID
    SessionID   string                 `json:"session_id"`   // 会话ID
    EventType   string                 `json:"event_type"`   // 事件类型
    SourceIP    string                 `json:"source_ip"`    // 源IP
    SourcePort  uint                   `json:"source_port"`  // 源端口
    Protocol    string                 `json:"protocol"`     // 协议
    Timestamp   time.Time              `json:"timestamp"`    // 时间戳
    Data        map[string]interface{} `json:"data"`         // 数据载荷
    RawLog      string                 `json:"raw_log"`      // 原始日志
    Parsed      bool                   `json:"parsed"`       // 是否已解析
}
```

#### 会话汇总结构
```go
// 位置: internal/handlers/container_runtime_log_handler.go:31-45
type SessionSummary struct {
    ID               uint       `json:"id"`                // 汇总ID
    SessionID        string     `json:"session_id"`        // 会话ID
    ContainerID      string     `json:"container_id"`      // 容器ID
    SourceIP         string     `json:"source_ip"`         // 源IP
    Protocol         string     `json:"protocol"`          // 协议
    StartTime        time.Time  `json:"start_time"`        // 开始时间
    EndTime          *time.Time `json:"end_time"`          // 结束时间
    Duration         int64      `json:"duration_ms"`       // 持续时间(毫秒)
    EventCount       int        `json:"event_count"`       // 事件数量
    AuthAttempts     int        `json:"auth_attempts"`     // 认证尝试次数
    SuccessfulAuths  int        `json:"successful_auths"`  // 成功认证次数
    CommandsExecuted int        `json:"commands_executed"` // 执行命令数
    FilesTransferred int        `json:"files_transferred"` // 传输文件数
    ThreatLevel      string     `json:"threat_level"`      // 威胁等级
    Summary          string     `json:"summary"`           // 汇总描述
}
```

#### 日志解析引擎
```go
// 位置: internal/handlers/container_runtime_log_handler.go:350-390
func (h *ContainerRuntimeLogHandler) parseLogLine(containerID, logLine, logType string) *ContainerRuntimeLog {
    log := &ContainerRuntimeLog{
        ContainerID: containerID,
        RawLog:      logLine,
        Timestamp:   time.Now(),
        Data:        make(map[string]interface{}),
        Parsed:      false,
    }
    
    // 1. JSON格式解析
    var jsonData map[string]interface{}
    if err := json.Unmarshal([]byte(logLine), &jsonData); err == nil {
        log.Data = jsonData
        log.Parsed = true
        
        // 提取标准字段
        if sessionID, ok := jsonData["session_id"].(string); ok {
            log.SessionID = sessionID
        }
        if sourceIP, ok := jsonData["source_ip"].(string); ok {
            log.SourceIP = sourceIP
        }
        if eventType, ok := jsonData["event_type"].(string); ok {
            log.EventType = eventType
        }
    } else {
        // 2. 文本格式解析
        if strings.Contains(logLine, "SSH") {
            log.EventType = "ssh_connection"
            log.Protocol = "ssh"
        } else if strings.Contains(logLine, "HTTP") {
            log.EventType = "http_request"
            log.Protocol = "http"
        }
    }
    
    return log
}
```

#### API接口实现
```go
// 解析容器日志
// API: POST /api/v1/container-logs/parse
func (h *ContainerRuntimeLogHandler) ParseContainerLogs(c *gin.Context)

// 根据容器获取日志
// API: GET /api/v1/container-logs/container/:container_id
func (h *ContainerRuntimeLogHandler) GetLogsByContainer(c *gin.Context)

// 根据时间范围获取日志
// API: GET /api/v1/container-logs/time-range
func (h *ContainerRuntimeLogHandler) GetLogsByTimeRange(c *gin.Context)

// 获取攻击分析
// API: GET /api/v1/container-logs/analysis
func (h *ContainerRuntimeLogHandler) GetAttackAnalysis(c *gin.Context)

// 导出日志
// API: POST /api/v1/container-logs/export
func (h *ContainerRuntimeLogHandler) ExportLogs(c *gin.Context)
```

### 3. 会话管理系统

#### 核心数据结构
```go
// 位置: internal/handlers/session_handler.go:13-27
type Session struct {
    ID            string                 `json:"id"`             // 会话ID
    SourceIP      string                 `json:"source_ip"`      // 源IP地址
    ContainerID   string                 `json:"container_id"`   // 容器ID
    ContainerName string                 `json:"container_name"` // 容器名称
    Protocol      string                 `json:"protocol"`       // 协议类型
    StartTime     time.Time              `json:"start_time"`     // 开始时间
    EndTime       *time.Time             `json:"end_time"`       // 结束时间
    LastActivity  time.Time              `json:"last_activity"`  // 最后活动时间
    IsActive      bool                   `json:"is_active"`      // 是否活跃
    AuthAttempts  []AuthAttempt          `json:"auth_attempts"`  // 认证尝试列表
    Commands      []CommandExecution     `json:"commands"`       // 命令执行列表
    Events        []SessionEvent         `json:"events"`         // 事件列表
    Metadata      map[string]interface{} `json:"metadata"`       // 元数据
}
```

#### 认证尝试结构
```go
// 位置: internal/handlers/session_handler.go:30-38
type AuthAttempt struct {
    ID        uint      `json:"id"`         // 尝试ID
    SessionID string    `json:"session_id"` // 会话ID
    Username  string    `json:"username"`   // 用户名
    Password  string    `json:"password"`   // 密码
    Success   bool      `json:"success"`    // 是否成功
    Timestamp time.Time `json:"timestamp"`  // 时间戳
    Method    string    `json:"method"`     // 认证方法
}
```

#### 命令执行结构
```go
// 位置: internal/handlers/session_handler.go:41-50
type CommandExecution struct {
    ID        uint      `json:"id"`          // 执行ID
    SessionID string    `json:"session_id"`  // 会话ID
    Command   string    `json:"command"`     // 命令
    Args      []string  `json:"args"`        // 参数
    Output    string    `json:"output"`      // 输出
    ExitCode  int       `json:"exit_code"`   // 退出码
    Timestamp time.Time `json:"timestamp"`   // 时间戳
    Duration  int64     `json:"duration_ms"` // 执行时长(毫秒)
}
```

#### 会话事件结构
```go
// 位置: internal/handlers/session_handler.go:53-60
type SessionEvent struct {
    ID        uint                   `json:"id"`         // 事件ID
    SessionID string                 `json:"session_id"` // 会话ID
    EventType string                 `json:"event_type"` // 事件类型
    Data      map[string]interface{} `json:"data"`       // 事件数据
    Timestamp time.Time              `json:"timestamp"`  // 时间戳
}
```

#### 会话生命周期管理
```go
// 位置: internal/handlers/session_handler.go:372-446
func (h *SessionHandler) RecordAuthAttempt(c *gin.Context) {
    // 1. 会话自动创建或获取
    session, exists := sessionStore[req.SessionID]
    if !exists {
        session = &Session{
            ID:           req.SessionID,
            SourceIP:     c.ClientIP(),
            StartTime:    time.Now(),
            LastActivity: time.Now(),
            IsActive:     true,
            // ... 初始化其他字段
        }
        sessionStore[req.SessionID] = session
    }
    
    // 2. 记录认证尝试
    authAttempt := &AuthAttempt{
        ID:        nextSessionAuthID,
        SessionID: req.SessionID,
        Username:  req.Username,
        Password:  req.Password,
        Success:   req.Success,
        Timestamp: time.Now(),
        Method:    req.Method,
    }
    
    // 3. 更新会话活动时间
    session.LastActivity = time.Now()
    
    // 4. 添加会话事件
    event := &SessionEvent{
        ID:        nextSessionEventID,
        SessionID: req.SessionID,
        EventType: "auth_attempt",
        Data: map[string]interface{}{
            "username": req.Username,
            "success":  req.Success,
            "method":   authAttempt.Method,
        },
        Timestamp: time.Now(),
    }
}
```

#### API接口实现
```go
// 获取会话基本信息
// API: GET /api/v1/sessions/:id
func (h *SessionHandler) GetSessionByID(c *gin.Context)

// 获取会话详细信息
// API: GET /api/v1/sessions/:id/details  
func (h *SessionHandler) GetDetailedSessionInfo(c *gin.Context)

// 记录认证尝试
// API: POST /api/v1/sessions/auth
func (h *SessionHandler) RecordAuthAttempt(c *gin.Context)

// 记录命令执行
// API: POST /api/v1/sessions/command
func (h *SessionHandler) RecordCommand(c *gin.Context)

// 处理超时会话
// API: POST /api/v1/sessions/timeout
func (h *SessionHandler) TimeoutInactiveSessions(c *gin.Context)

// 获取会话统计
// API: GET /api/v1/sessions/statistics
func (h *SessionHandler) GetSessionStatistics(c *gin.Context)
```

### 4. 专用蜜罐日志处理

#### Heralding认证日志处理
```go
// 位置: internal/handlers/heralding_handler.go
// 功能: SSH/Telnet认证日志分析
// 格式: CSV格式日志文件
// API路由: /api/v1/heralding/*

// 主要功能:
// - 认证尝试统计
// - 攻击者IP分析  
// - 用户名/密码统计
// - 时间趋势分析
```

#### Cowrie蜜罐日志处理  
```go
// 位置: internal/handlers/cowrie_handler.go
// 功能: SSH蜜罐交互日志分析
// 格式: JSON格式日志文件
// API路由: /api/v1/cowrie/*

// 主要功能:
// - 命令执行分析
// - 攻击者行为模式
// - 文件传输监控
// - 客户端指纹识别
```

#### MySQL 蜜罐日志处理
```go
// 位置: internal/handlers/mysql_honeypot_handler.go
// 功能: 数据库蜜罐SQL/认证日志分析
// 格式: JSON/CSV/纯文本日志 (自动转换)
// API路由: /api/v1/mysql-honeypot/*

// 主要功能:
// - SQL查询及凭据统计
// - IP/用户名溯源
// - 容器级日志拉取 (Docker Copy)
// - 自定义路径(MYSQL_HONEYPOT_LOG_PATH)
```

## 📊 统计分析功能

### 1. 实时统计
```go
// 会话统计
stats := map[string]interface{}{
    "total_sessions":       len(sessionStore),
    "active_sessions":      activeCount,
    "total_auth_attempts":  len(sessionAuthAttempts),
    "total_commands":       len(sessionCommands),
    "sessions_by_ip":       ipStats,
    "sessions_by_protocol": protocolStats,
}
```

### 2. 攻击分析
```go
// 攻击行为分析
analysis := map[string]interface{}{
    "total_attacks":     attackCount,
    "attack_sources":    sourceStats,
    "attack_types":      typeStats,
    "threat_indicators": indicators,
}
```

### 3. 导出功能
```go
// 支持格式: JSON, CSV
// 过滤条件: 时间范围, 容器ID, 事件类型
// 导出内容: 完整日志数据 + 统计信息
```

## 🔒 安全特性

### 1. 数据隔离
- 容器级别的日志隔离
- 会话级别的数据分离
- 多租户支持

### 2. 并发安全
```go
// 使用读写锁保护共享数据
var (
    sessionStoreMutex    = sync.RWMutex{}
    runtimeLogMutex      = sync.RWMutex{}
)
```

### 3. 内存管理
- 自动会话超时清理
- 日志轮转机制
- 内存使用监控

## 🚀 性能优化

### 1. 存储优化
- 数据库索引优化
- 内存缓存策略
- 批量操作支持

### 2. 查询优化
- 时间范围查询优化
- 分页查询支持
- 条件过滤优化

### 3. 并发处理
- 异步日志处理
- 并发解析支持
- 队列缓冲机制

---

**文档版本**: v1.0  
**最后更新**: 2025-01-17  
**维护者**: Andorralee开发团队
