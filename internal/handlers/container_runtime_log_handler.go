package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ContainerRuntimeLogHandler 容器运行时日志处理器
type ContainerRuntimeLogHandler struct {
	logService *services.ContainerLogService
}

// NewContainerRuntimeLogHandler 创建容器运行时日志处理器
func NewContainerRuntimeLogHandler() *ContainerRuntimeLogHandler {
	return &ContainerRuntimeLogHandler{
		logService: services.NewContainerLogService(),
	}
}

// ParseContainerLogs 解析并存储容器日志
func (h *ContainerRuntimeLogHandler) ParseContainerLogs(c *gin.Context) {
	type ParseRequest struct {
		ContainerID string `json:"container_id" binding:"required"`
	}

	var req ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	err := h.logService.ParseAndStoreContainerLogs(req.ContainerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "解析容器日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"message":      "容器日志解析成功",
		"container_id": req.ContainerID,
		"parsed_at":    time.Now(),
	})
}

// GetLogsByContainer 根据容器ID获取日志
func (h *ContainerRuntimeLogHandler) GetLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	logs, err := h.logService.GetLogsByContainer(containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"logs":         logs,
		"count":        len(logs),
		"container_id": containerID,
	})
}

// GetLogsByTimeRange 根据时间范围获取日志
func (h *ContainerRuntimeLogHandler) GetLogsByTimeRange(c *gin.Context) {
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

	logs, err := h.logService.GetLogsByTimeRange(startTime, endTime)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"logs":       logs,
		"count":      len(logs),
		"start_time": startTime,
		"end_time":   endTime,
	})
}

// GetLogsByEventType 根据事件类型获取日志
func (h *ContainerRuntimeLogHandler) GetLogsByEventType(c *gin.Context) {
	eventType := c.Param("event_type")
	if eventType == "" {
		utils.ResponseError(c, http.StatusBadRequest, "事件类型不能为空")
		return
	}

	logs, err := h.logService.GetLogsByEventType(eventType)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"logs":       logs,
		"count":      len(logs),
		"event_type": eventType,
	})
}

// GetLogsBySourceIP 根据源IP获取日志
func (h *ContainerRuntimeLogHandler) GetLogsBySourceIP(c *gin.Context) {
	sourceIP := c.Param("source_ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "源IP不能为空")
		return
	}

	logs, err := h.logService.GetLogsBySourceIP(sourceIP)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"logs":      logs,
		"count":     len(logs),
		"source_ip": sourceIP,
	})
}

// GetSessionSummary 获取会话汇总信息
func (h *ContainerRuntimeLogHandler) GetSessionSummary(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	summary, err := h.logService.GetSessionSummary(sessionID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "获取会话汇总失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, summary)
}

// CreateSessionSummary 创建会话汇总
func (h *ContainerRuntimeLogHandler) CreateSessionSummary(c *gin.Context) {
	type SummaryRequest struct {
		SessionID   string `json:"session_id" binding:"required"`
		ContainerID string `json:"container_id" binding:"required"`
	}

	var req SummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	err := h.logService.CreateSessionSummary(req.SessionID, req.ContainerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建会话汇总失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"message":      "会话汇总创建成功",
		"session_id":   req.SessionID,
		"container_id": req.ContainerID,
		"created_at":   time.Now(),
	})
}

// GetLogStatistics 获取日志统计信息
func (h *ContainerRuntimeLogHandler) GetLogStatistics(c *gin.Context) {
	// 可以添加查询参数来过滤统计范围
	containerID := c.Query("container_id")
	
	// 这里可以根据需要扩展统计逻辑
	_ = containerID // 暂时未使用
	
	// 获取基础统计信息（这里需要在service中实现）
	stats := gin.H{
		"message": "日志统计功能",
		"note":    "可以根据需要扩展统计维度",
	}

	utils.ResponseSuccess(c, stats)
}

// GetAttackAnalysis 获取攻击分析报告
func (h *ContainerRuntimeLogHandler) GetAttackAnalysis(c *gin.Context) {
	containerID := c.Query("container_id")
	sourceIP := c.Query("source_ip")
	hoursStr := c.DefaultQuery("hours", "24")
	
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "时间范围格式错误: "+err.Error())
		return
	}

	// 计算时间范围
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	// 获取时间范围内的日志
	logs, err := h.logService.GetLogsByTimeRange(startTime, endTime)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	// 过滤日志
	var filteredLogs []interface{}
	for _, log := range logs {
		include := true
		if containerID != "" && log.ContainerID != containerID {
			include = false
		}
		if sourceIP != "" && log.SourceIP != sourceIP {
			include = false
		}
		if include {
			filteredLogs = append(filteredLogs, log)
		}
	}

	// 简单的攻击分析
	analysis := gin.H{
		"time_range": gin.H{
			"start_time": startTime,
			"end_time":   endTime,
			"hours":      hours,
		},
		"filters": gin.H{
			"container_id": containerID,
			"source_ip":    sourceIP,
		},
		"summary": gin.H{
			"total_events":     len(filteredLogs),
			"analysis_time":    time.Now(),
		},
		"events": filteredLogs,
	}

	utils.ResponseSuccess(c, analysis)
}

// ExportLogs 导出日志
func (h *ContainerRuntimeLogHandler) ExportLogs(c *gin.Context) {
	type ExportRequest struct {
		ContainerID string    `json:"container_id"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		EventType   string    `json:"event_type"`
		Format      string    `json:"format"` // csv, json
	}

	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 获取日志数据
	logs, err := h.logService.GetLogsByTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	// 过滤日志
	var filteredLogs []interface{}
	for _, log := range logs {
		include := true
		if req.ContainerID != "" && log.ContainerID != req.ContainerID {
			include = false
		}
		if req.EventType != "" && log.EventType != req.EventType {
			include = false
		}
		if include {
			filteredLogs = append(filteredLogs, log)
		}
	}

	// 根据格式返回数据
	if req.Format == "csv" {
		// 这里可以实现CSV格式导出
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=container_logs.csv")
		utils.ResponseSuccess(c, gin.H{
			"message": "CSV导出功能待实现",
			"count":   len(filteredLogs),
		})
	} else {
		// 默认JSON格式
		utils.ResponseSuccess(c, gin.H{
			"logs":   filteredLogs,
			"count":  len(filteredLogs),
			"format": "json",
		})
	}
}
