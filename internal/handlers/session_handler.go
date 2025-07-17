package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionHandler 会话管理处理器
type SessionHandler struct {
	sessionService *services.SessionService
}

// NewSessionHandler 创建会话管理处理器
func NewSessionHandler() *SessionHandler {
	return &SessionHandler{
		sessionService: services.NewSessionService(),
	}
}

// GetSessionByID 根据ID获取会话详情
func (h *SessionHandler) GetSessionByID(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "会话不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, session)
}

// GetDetailedSessionInfo 获取详细会话信息（包含事件）
func (h *SessionHandler) GetDetailedSessionInfo(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	detailedInfo, err := h.sessionService.GetDetailedSessionInfo(sessionID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "获取会话详情失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, detailedInfo)
}

// GetSessionsByIP 根据IP获取会话列表
func (h *SessionHandler) GetSessionsByIP(c *gin.Context) {
	sourceIP := c.Param("ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "IP地址不能为空")
		return
	}

	sessions, err := h.sessionService.GetSessionsByIP(sourceIP)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取会话列表失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
		"ip":       sourceIP,
	})
}

// GetSessionsByContainer 根据容器ID获取会话列表
func (h *SessionHandler) GetSessionsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	sessions, err := h.sessionService.GetSessionsByContainer(containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取会话列表失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"sessions":     sessions,
		"count":        len(sessions),
		"container_id": containerID,
	})
}

// GetActiveSessionsByIP 获取指定IP的活跃会话
func (h *SessionHandler) GetActiveSessionsByIP(c *gin.Context) {
	sourceIP := c.Param("ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "IP地址不能为空")
		return
	}

	sessions, err := h.sessionService.GetActiveSessionsByIP(sourceIP)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取活跃会话失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"active_sessions": sessions,
		"count":           len(sessions),
		"ip":              sourceIP,
	})
}

// GetSessionEvents 获取会话事件
func (h *SessionHandler) GetSessionEvents(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	events, err := h.sessionService.GetSessionEvents(sessionID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取会话事件失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"events":     events,
		"count":      len(events),
		"session_id": sessionID,
	})
}

// GetSessionStatistics 获取会话统计信息
func (h *SessionHandler) GetSessionStatistics(c *gin.Context) {
	stats, err := h.sessionService.GetSessionStatistics()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取会话统计失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, stats)
}

// GetSessionsInTimeRange 根据时间范围获取会话
func (h *SessionHandler) GetSessionsInTimeRange(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		utils.ResponseError(c, http.StatusBadRequest, "开始时间和结束时间不能为空")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "开始时间格式错误: "+err.Error())
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "结束时间格式错误: "+err.Error())
		return
	}

	sessions, err := h.sessionService.GetSessionsInTimeRange(startTime, endTime)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取会话列表失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"sessions":   sessions,
		"count":      len(sessions),
		"start_time": startTime,
		"end_time":   endTime,
	})
}

// CloseSession 手动关闭会话
func (h *SessionHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	err := h.sessionService.CloseSession(sessionID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "关闭会话失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"message":    "会话关闭成功",
		"session_id": sessionID,
		"closed_at":  time.Now(),
	})
}

// TimeoutInactiveSessions 超时处理不活跃的会话
func (h *SessionHandler) TimeoutInactiveSessions(c *gin.Context) {
	timeoutMinutesStr := c.DefaultQuery("timeout_minutes", "30")
	timeoutMinutes, err := strconv.Atoi(timeoutMinutesStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "超时时间格式错误: "+err.Error())
		return
	}

	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute
	err = h.sessionService.TimeoutInactiveSessions(timeoutDuration)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "处理超时会话失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"message":         "超时会话处理完成",
		"timeout_minutes": timeoutMinutes,
		"processed_at":    time.Now(),
	})
}

// RecordAuthAttempt 记录认证尝试（用于手动记录）
func (h *SessionHandler) RecordAuthAttempt(c *gin.Context) {
	type AuthAttemptRequest struct {
		SessionID string `json:"session_id" binding:"required"`
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required"`
		Success   bool   `json:"success"`
	}

	var req AuthAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	err := h.sessionService.RecordAuthAttempt(req.SessionID, req.Username, req.Password, req.Success)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "记录认证尝试失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"message":    "认证尝试记录成功",
		"session_id": req.SessionID,
		"username":   req.Username,
		"success":    req.Success,
		"recorded_at": time.Now(),
	})
}

// RecordCommand 记录命令执行（用于手动记录）
func (h *SessionHandler) RecordCommand(c *gin.Context) {
	type CommandRequest struct {
		SessionID string `json:"session_id" binding:"required"`
		Command   string `json:"command" binding:"required"`
		Response  string `json:"response"`
		Success   bool   `json:"success"`
	}

	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	err := h.sessionService.RecordCommand(req.SessionID, req.Command, req.Response, req.Success)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "记录命令执行失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"message":    "命令执行记录成功",
		"session_id": req.SessionID,
		"command":    req.Command,
		"success":    req.Success,
		"recorded_at": time.Now(),
	})
}
