package handlers

import (
	"andorralee/pkg/utils"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Session 会话模型
type Session struct {
	ID            string                 `json:"id"`
	SourceIP      string                 `json:"source_ip"`
	ContainerID   string                 `json:"container_id"`
	ContainerName string                 `json:"container_name"`
	Protocol      string                 `json:"protocol"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time"`
	LastActivity  time.Time              `json:"last_activity"`
	IsActive      bool                   `json:"is_active"`
	AuthAttempts  []AuthAttempt          `json:"auth_attempts"`
	Commands      []CommandExecution     `json:"commands"`
	Events        []SessionEvent         `json:"events"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AuthAttempt 认证尝试
type AuthAttempt struct {
	ID        uint      `json:"id"`
	SessionID string    `json:"session_id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"` // password, key, etc.
}

// CommandExecution 命令执行
type CommandExecution struct {
	ID        uint      `json:"id"`
	SessionID string    `json:"session_id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Output    string    `json:"output"`
	ExitCode  int       `json:"exit_code"`
	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"duration_ms"`
}

// SessionEvent 会话事件
type SessionEvent struct {
	ID        uint                   `json:"id"`
	SessionID string                 `json:"session_id"`
	EventType string                 `json:"event_type"` // connect, disconnect, auth, command, file_transfer, etc.
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// SessionHandler 会话处理器
type SessionHandler struct{}

// 内存存储
var (
	sessionStore         = make(map[string]*Session)
	sessionAuthAttempts  = make(map[uint]*AuthAttempt)
	sessionCommands      = make(map[uint]*CommandExecution)
	sessionEventStore    = make(map[uint]*SessionEvent)
	sessionStoreMutex    = sync.RWMutex{}
	nextSessionAuthID    = uint(1)
	nextSessionCommandID = uint(1)
	nextSessionEventID   = uint(1)
)

// NewSessionHandler 创建会话处理器
func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

// GetSessionByID 获取会话基本信息
func (h *SessionHandler) GetSessionByID(c *gin.Context) {
	sessionID := c.Param("id")

	sessionStoreMutex.RLock()
	session, exists := sessionStore[sessionID]
	sessionStoreMutex.RUnlock()

	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "会话不存在")
		return
	}

	utils.ResponseSuccess(c, session)
}

// GetDetailedSessionInfo 获取会话详细信息
func (h *SessionHandler) GetDetailedSessionInfo(c *gin.Context) {
	sessionID := c.Param("id")

	sessionStoreMutex.RLock()
	session, exists := sessionStore[sessionID]
	if !exists {
		sessionStoreMutex.RUnlock()
		utils.ResponseError(c, http.StatusNotFound, "会话不存在")
		return
	}

	// 创建详细信息副本
	detailedSession := *session
	sessionStoreMutex.RUnlock()

	// 添加统计信息
	stats := map[string]interface{}{
		"total_auth_attempts": len(detailedSession.AuthAttempts),
		"successful_auths":    0,
		"failed_auths":        0,
		"total_commands":      len(detailedSession.Commands),
		"total_events":        len(detailedSession.Events),
		"session_duration":    0,
	}

	for _, auth := range detailedSession.AuthAttempts {
		if auth.Success {
			stats["successful_auths"] = stats["successful_auths"].(int) + 1
		} else {
			stats["failed_auths"] = stats["failed_auths"].(int) + 1
		}
	}

	if detailedSession.EndTime != nil {
		stats["session_duration"] = detailedSession.EndTime.Sub(detailedSession.StartTime).Milliseconds()
	} else {
		stats["session_duration"] = time.Since(detailedSession.StartTime).Milliseconds()
	}

	result := map[string]interface{}{
		"session":    detailedSession,
		"statistics": stats,
	}

	utils.ResponseSuccess(c, result)
}

// GetSessionEvents 获取会话事件
func (h *SessionHandler) GetSessionEvents(c *gin.Context) {
	sessionID := c.Param("id")

	sessionStoreMutex.RLock()
	session, exists := sessionStore[sessionID]
	if !exists {
		sessionStoreMutex.RUnlock()
		utils.ResponseError(c, http.StatusNotFound, "会话不存在")
		return
	}

	events := session.Events
	sessionStoreMutex.RUnlock()

	utils.ResponseSuccess(c, events)
}

// CloseSession 关闭会话
func (h *SessionHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("id")

	sessionStoreMutex.Lock()
	session, exists := sessionStore[sessionID]
	if !exists {
		sessionStoreMutex.Unlock()
		utils.ResponseError(c, http.StatusNotFound, "会话不存在")
		return
	}

	if !session.IsActive {
		sessionStoreMutex.Unlock()
		utils.ResponseError(c, http.StatusBadRequest, "会话已关闭")
		return
	}

	now := time.Now()
	session.EndTime = &now
	session.IsActive = false

	// 添加关闭事件
	event := &SessionEvent{
		ID:        nextSessionEventID,
		SessionID: sessionID,
		EventType: "session_closed",
		Data: map[string]interface{}{
			"reason": "manual_close",
		},
		Timestamp: now,
	}
	session.Events = append(session.Events, *event)
	sessionEventStore[nextSessionEventID] = event
	nextSessionEventID++

	sessionStoreMutex.Unlock()

	utils.ResponseSuccess(c, "会话已关闭")
}

// GetSessionsByIP 根据IP获取会话
func (h *SessionHandler) GetSessionsByIP(c *gin.Context) {
	ip := c.Param("ip")

	sessionStoreMutex.RLock()
	var result []*Session
	for _, session := range sessionStore {
		if session.SourceIP == ip {
			result = append(result, session)
		}
	}
	sessionStoreMutex.RUnlock()

	utils.ResponseSuccess(c, result)
}

// GetActiveSessionsByIP 获取IP的活跃会话
func (h *SessionHandler) GetActiveSessionsByIP(c *gin.Context) {
	ip := c.Param("ip")

	sessionStoreMutex.RLock()
	var result []*Session
	for _, session := range sessionStore {
		if session.SourceIP == ip && session.IsActive {
			result = append(result, session)
		}
	}
	sessionStoreMutex.RUnlock()

	utils.ResponseSuccess(c, result)
}

// GetSessionsByContainer 根据容器获取会话
func (h *SessionHandler) GetSessionsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")

	sessionStoreMutex.RLock()
	var result []*Session
	for _, session := range sessionStore {
		if session.ContainerID == containerID {
			result = append(result, session)
		}
	}
	sessionStoreMutex.RUnlock()

	utils.ResponseSuccess(c, result)
}

// GetSessionStatistics 获取会话统计
func (h *SessionHandler) GetSessionStatistics(c *gin.Context) {
	sessionStoreMutex.RLock()

	stats := map[string]interface{}{
		"total_sessions":       len(sessionStore),
		"active_sessions":      0,
		"closed_sessions":      0,
		"total_auth_attempts":  len(sessionAuthAttempts),
		"total_commands":       len(sessionCommands),
		"total_events":         len(sessionEventStore),
		"sessions_by_ip":       make(map[string]int),
		"sessions_by_protocol": make(map[string]int),
		"recent_sessions":      make([]*Session, 0),
	}

	for _, session := range sessionStore {
		if session.IsActive {
			stats["active_sessions"] = stats["active_sessions"].(int) + 1
		} else {
			stats["closed_sessions"] = stats["closed_sessions"].(int) + 1
		}

		stats["sessions_by_ip"].(map[string]int)[session.SourceIP]++
		stats["sessions_by_protocol"].(map[string]int)[session.Protocol]++
	}

	sessionStoreMutex.RUnlock()

	utils.ResponseSuccess(c, stats)
}

// GetSessionsInTimeRange 根据时间范围获取会话
func (h *SessionHandler) GetSessionsInTimeRange(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		utils.ResponseError(c, http.StatusBadRequest, "需要提供start_time和end_time参数")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "start_time格式错误: "+err.Error())
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "end_time格式错误: "+err.Error())
		return
	}

	sessionStoreMutex.RLock()
	var result []*Session
	for _, session := range sessionStore {
		if session.StartTime.After(startTime) && session.StartTime.Before(endTime) {
			result = append(result, session)
		}
	}
	sessionStoreMutex.RUnlock()

	utils.ResponseSuccess(c, result)
}

// TimeoutInactiveSessions 处理超时会话
func (h *SessionHandler) TimeoutInactiveSessions(c *gin.Context) {
	var req struct {
		TimeoutMinutes int `json:"timeout_minutes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if req.TimeoutMinutes <= 0 {
		utils.ResponseError(c, http.StatusBadRequest, "超时时间必须大于0")
		return
	}

	cutoffTime := time.Now().Add(-time.Duration(req.TimeoutMinutes) * time.Minute)
	var timeoutSessions []string

	sessionStoreMutex.Lock()
	for sessionID, session := range sessionStore {
		if session.IsActive && session.LastActivity.Before(cutoffTime) {
			now := time.Now()
			session.EndTime = &now
			session.IsActive = false

			// 添加超时事件
			event := &SessionEvent{
				ID:        nextSessionEventID,
				SessionID: sessionID,
				EventType: "session_timeout",
				Data: map[string]interface{}{
					"timeout_minutes": req.TimeoutMinutes,
					"last_activity":   session.LastActivity,
				},
				Timestamp: now,
			}
			session.Events = append(session.Events, *event)
			sessionEventStore[nextSessionEventID] = event
			nextSessionEventID++

			timeoutSessions = append(timeoutSessions, sessionID)
		}
	}
	sessionStoreMutex.Unlock()

	result := map[string]interface{}{
		"timeout_sessions": timeoutSessions,
		"count":            len(timeoutSessions),
	}

	utils.ResponseSuccess(c, result)
}

// RecordAuthAttempt 记录认证尝试
func (h *SessionHandler) RecordAuthAttempt(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password"`
		Success   bool   `json:"success"`
		Method    string `json:"method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	sessionStoreMutex.Lock()
	session, exists := sessionStore[req.SessionID]
	if !exists {
		// 创建新会话
		session = &Session{
			ID:           req.SessionID,
			SourceIP:     c.ClientIP(),
			StartTime:    time.Now(),
			LastActivity: time.Now(),
			IsActive:     true,
			AuthAttempts: make([]AuthAttempt, 0),
			Commands:     make([]CommandExecution, 0),
			Events:       make([]SessionEvent, 0),
			Metadata:     make(map[string]interface{}),
		}
		sessionStore[req.SessionID] = session
	}

	// 更新最后活动时间
	session.LastActivity = time.Now()

	// 创建认证尝试记录
	authAttempt := &AuthAttempt{
		ID:        nextSessionAuthID,
		SessionID: req.SessionID,
		Username:  req.Username,
		Password:  req.Password,
		Success:   req.Success,
		Timestamp: time.Now(),
		Method:    req.Method,
	}

	if authAttempt.Method == "" {
		authAttempt.Method = "password"
	}

	session.AuthAttempts = append(session.AuthAttempts, *authAttempt)
	sessionAuthAttempts[nextSessionAuthID] = authAttempt
	nextSessionAuthID++

	// 添加认证事件
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
	session.Events = append(session.Events, *event)
	sessionEventStore[nextSessionEventID] = event
	nextSessionEventID++

	sessionStoreMutex.Unlock()

	utils.ResponseSuccess(c, authAttempt)
}

// RecordCommand 记录命令执行
func (h *SessionHandler) RecordCommand(c *gin.Context) {
	var req struct {
		SessionID string   `json:"session_id" binding:"required"`
		Command   string   `json:"command" binding:"required"`
		Args      []string `json:"args"`
		Output    string   `json:"output"`
		ExitCode  int      `json:"exit_code"`
		Duration  int64    `json:"duration_ms"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	sessionStoreMutex.Lock()
	session, exists := sessionStore[req.SessionID]
	if !exists {
		sessionStoreMutex.Unlock()
		utils.ResponseError(c, http.StatusNotFound, "会话不存在")
		return
	}

	// 更新最后活动时间
	session.LastActivity = time.Now()

	// 创建命令执行记录
	command := &CommandExecution{
		ID:        nextSessionCommandID,
		SessionID: req.SessionID,
		Command:   req.Command,
		Args:      req.Args,
		Output:    req.Output,
		ExitCode:  req.ExitCode,
		Timestamp: time.Now(),
		Duration:  req.Duration,
	}

	if command.Args == nil {
		command.Args = make([]string, 0)
	}

	session.Commands = append(session.Commands, *command)
	sessionCommands[nextSessionCommandID] = command
	nextSessionCommandID++

	// 添加命令事件
	event := &SessionEvent{
		ID:        nextSessionEventID,
		SessionID: req.SessionID,
		EventType: "command_execution",
		Data: map[string]interface{}{
			"command":   req.Command,
			"args":      req.Args,
			"exit_code": req.ExitCode,
			"duration":  req.Duration,
		},
		Timestamp: time.Now(),
	}
	session.Events = append(session.Events, *event)
	sessionEventStore[nextSessionEventID] = event
	nextSessionEventID++

	sessionStoreMutex.Unlock()

	utils.ResponseSuccess(c, command)
}
