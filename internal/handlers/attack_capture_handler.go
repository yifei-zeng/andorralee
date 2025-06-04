package handlers

import (
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AttackEvent 攻击事件模型
type AttackEvent struct {
	ID            uint      `json:"id"`
	SourceIP      string    `json:"source_ip"`
	SourcePort    int       `json:"source_port"`
	DestIP        string    `json:"dest_ip"`
	DestPort      int       `json:"dest_port"`
	Protocol      string    `json:"protocol"`
	AttackType    string    `json:"attack_type"`
	Payload       string    `json:"payload"`
	Timestamp     time.Time `json:"timestamp"`
	Severity      string    `json:"severity"` // low, medium, high, critical
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	UserAgent     string    `json:"user_agent"`
	SessionID     string    `json:"session_id"`
}

// AttackSession 攻击会话模型
type AttackSession struct {
	ID          uint           `json:"id"`
	SessionID   string         `json:"session_id"`
	SourceIP    string         `json:"source_ip"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	EventCount  int            `json:"event_count"`
	AttackTypes []string       `json:"attack_types"`
	Events      []*AttackEvent `json:"events"`
}

// 内存存储
var (
	attackEvents   = make(map[uint]*AttackEvent)
	attackSessions = make(map[string]*AttackSession)
	attackMutex    = sync.RWMutex{}
	sessionMutex   = sync.RWMutex{}
	nextEventID    = uint(1)
	nextSessionID  = uint(1)
)

// CaptureAttackEvent 捕获攻击事件
func CaptureAttackEvent(c *gin.Context) {
	var req struct {
		SourceIP      string `json:"source_ip" binding:"required"`
		SourcePort    int    `json:"source_port"`
		DestIP        string `json:"dest_ip" binding:"required"`
		DestPort      int    `json:"dest_port" binding:"required"`
		Protocol      string `json:"protocol" binding:"required"`
		AttackType    string `json:"attack_type" binding:"required"`
		Payload       string `json:"payload"`
		ContainerID   string `json:"container_id"`
		ContainerName string `json:"container_name"`
		UserAgent     string `json:"user_agent"`
		SessionID     string `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 分析攻击严重程度
	severity := analyzeSeverity(req.AttackType, req.Payload)

	attackMutex.Lock()
	event := &AttackEvent{
		ID:            nextEventID,
		SourceIP:      req.SourceIP,
		SourcePort:    req.SourcePort,
		DestIP:        req.DestIP,
		DestPort:      req.DestPort,
		Protocol:      req.Protocol,
		AttackType:    req.AttackType,
		Payload:       req.Payload,
		Timestamp:     time.Now(),
		Severity:      severity,
		ContainerID:   req.ContainerID,
		ContainerName: req.ContainerName,
		UserAgent:     req.UserAgent,
		SessionID:     req.SessionID,
	}
	attackEvents[nextEventID] = event
	nextEventID++
	attackMutex.Unlock()

	// 更新攻击会话
	updateAttackSession(event)

	utils.ResponseSuccess(c, event)
}

// analyzeSeverity 分析攻击严重程度
func analyzeSeverity(attackType, payload string) string {
	attackType = strings.ToLower(attackType)
	payload = strings.ToLower(payload)

	// 高危攻击类型
	if strings.Contains(attackType, "sql injection") ||
		strings.Contains(attackType, "command injection") ||
		strings.Contains(attackType, "rce") ||
		strings.Contains(payload, "union select") ||
		strings.Contains(payload, "exec") ||
		strings.Contains(payload, "system") {
		return "critical"
	}

	// 中高危攻击类型
	if strings.Contains(attackType, "xss") ||
		strings.Contains(attackType, "csrf") ||
		strings.Contains(attackType, "brute force") ||
		strings.Contains(payload, "script") ||
		strings.Contains(payload, "alert") {
		return "high"
	}

	// 中危攻击类型
	if strings.Contains(attackType, "scan") ||
		strings.Contains(attackType, "probe") ||
		strings.Contains(attackType, "enumeration") {
		return "medium"
	}

	return "low"
}

// updateAttackSession 更新攻击会话
func updateAttackSession(event *AttackEvent) {
	sessionKey := event.SourceIP
	if event.SessionID != "" {
		sessionKey = event.SessionID
	}

	sessionMutex.Lock()
	session, exists := attackSessions[sessionKey]
	if !exists {
		session = &AttackSession{
			ID:          nextSessionID,
			SessionID:   sessionKey,
			SourceIP:    event.SourceIP,
			StartTime:   event.Timestamp,
			EventCount:  0,
			AttackTypes: make([]string, 0),
			Events:      make([]*AttackEvent, 0),
		}
		attackSessions[sessionKey] = session
		nextSessionID++
	}

	// 更新会话信息
	session.EventCount++
	session.Events = append(session.Events, event)

	// 添加攻击类型（去重）
	found := false
	for _, t := range session.AttackTypes {
		if t == event.AttackType {
			found = true
			break
		}
	}
	if !found {
		session.AttackTypes = append(session.AttackTypes, event.AttackType)
	}

	sessionMutex.Unlock()
}

// GetAllAttackEvents 获取所有攻击事件
func GetAllAttackEvents(c *gin.Context) {
	// 分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	attackMutex.RLock()
	events := make([]*AttackEvent, 0, len(attackEvents))
	for _, event := range attackEvents {
		events = append(events, event)
	}
	attackMutex.RUnlock()

	// 简单分页
	total := len(events)
	start := (page - 1) * limit
	end := start + limit
	
	if start >= total {
		events = []*AttackEvent{}
	} else {
		if end > total {
			end = total
		}
		events = events[start:end]
	}

	result := map[string]interface{}{
		"events": events,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	utils.ResponseSuccess(c, result)
}

// GetAttackEventsByIP 根据源IP获取攻击事件
func GetAttackEventsByIP(c *gin.Context) {
	sourceIP := c.Param("ip")

	attackMutex.RLock()
	events := make([]*AttackEvent, 0)
	for _, event := range attackEvents {
		if event.SourceIP == sourceIP {
			events = append(events, event)
		}
	}
	attackMutex.RUnlock()

	utils.ResponseSuccess(c, events)
}

// GetAttackSessions 获取攻击会话
func GetAttackSessions(c *gin.Context) {
	sessionMutex.RLock()
	sessions := make([]*AttackSession, 0, len(attackSessions))
	for _, session := range attackSessions {
		sessions = append(sessions, session)
	}
	sessionMutex.RUnlock()

	utils.ResponseSuccess(c, sessions)
}

// GetAttackSessionByID 根据会话ID获取攻击会话
func GetAttackSessionByID(c *gin.Context) {
	sessionID := c.Param("session_id")

	sessionMutex.RLock()
	session, exists := attackSessions[sessionID]
	sessionMutex.RUnlock()

	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "攻击会话不存在")
		return
	}

	utils.ResponseSuccess(c, session)
}

// GetAttackStatistics 获取攻击统计信息
func GetAttackStatistics(c *gin.Context) {
	attackMutex.RLock()
	sessionMutex.RLock()

	stats := map[string]interface{}{
		"total_events":      len(attackEvents),
		"total_sessions":    len(attackSessions),
		"events_by_type":    make(map[string]int),
		"events_by_severity": make(map[string]int),
		"top_attackers":     make(map[string]int),
		"recent_events":     make([]*AttackEvent, 0),
	}

	// 统计攻击类型和严重程度
	for _, event := range attackEvents {
		stats["events_by_type"].(map[string]int)[event.AttackType]++
		stats["events_by_severity"].(map[string]int)[event.Severity]++
		stats["top_attackers"].(map[string]int)[event.SourceIP]++
	}

	// 获取最近的攻击事件（最多20条）
	recentEvents := make([]*AttackEvent, 0, 20)
	for _, event := range attackEvents {
		recentEvents = append(recentEvents, event)
		if len(recentEvents) >= 20 {
			break
		}
	}
	stats["recent_events"] = recentEvents

	sessionMutex.RUnlock()
	attackMutex.RUnlock()

	utils.ResponseSuccess(c, stats)
}

// SimulateAttack 模拟攻击（用于测试）
func SimulateAttack(c *gin.Context) {
	var req struct {
		AttackType string `json:"attack_type" binding:"required"`
		TargetIP   string `json:"target_ip"`
		TargetPort int    `json:"target_port"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 模拟不同类型的攻击
	var payload string
	switch strings.ToLower(req.AttackType) {
	case "sql_injection":
		payload = "' OR 1=1--"
	case "xss":
		payload = "<script>alert('xss')</script>"
	case "brute_force":
		payload = "admin:123456"
	case "command_injection":
		payload = "; cat /etc/passwd"
	default:
		payload = "generic attack payload"
	}

	targetIP := req.TargetIP
	if targetIP == "" {
		targetIP = "127.0.0.1"
	}

	targetPort := req.TargetPort
	if targetPort == 0 {
		targetPort = 80
	}

	// 创建攻击事件
	attackMutex.Lock()
	event := &AttackEvent{
		ID:            nextEventID,
		SourceIP:      c.ClientIP(),
		SourcePort:    12345,
		DestIP:        targetIP,
		DestPort:      targetPort,
		Protocol:      "tcp",
		AttackType:    req.AttackType,
		Payload:       payload,
		Timestamp:     time.Now(),
		Severity:      analyzeSeverity(req.AttackType, payload),
		ContainerID:   "simulated",
		ContainerName: "test-container",
		UserAgent:     c.GetHeader("User-Agent"),
		SessionID:     "sim-" + c.ClientIP(),
	}
	attackEvents[nextEventID] = event
	nextEventID++
	attackMutex.Unlock()

	// 更新攻击会话
	updateAttackSession(event)

	utils.ResponseSuccess(c, map[string]interface{}{
		"message": "攻击模拟成功",
		"event":   event,
	})
}
