package handlers

import (
	"andorralee/pkg/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ContainerRuntimeLog 容器运行时日志
type ContainerRuntimeLog struct {
	ID          uint                   `json:"id"`
	ContainerID string                 `json:"container_id"`
	SessionID   string                 `json:"session_id"`
	EventType   string                 `json:"event_type"` // connect, disconnect, auth, command, file_transfer
	SourceIP    string                 `json:"source_ip"`
	SourcePort  uint                   `json:"source_port"`
	Protocol    string                 `json:"protocol"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	RawLog      string                 `json:"raw_log"`
	Parsed      bool                   `json:"parsed"`
}

// SessionSummary 会话汇总
type SessionSummary struct {
	ID               uint      `json:"id"`
	SessionID        string    `json:"session_id"`
	ContainerID      string    `json:"container_id"`
	SourceIP         string    `json:"source_ip"`
	Protocol         string    `json:"protocol"`
	StartTime        time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	Duration         int64     `json:"duration_ms"`
	EventCount       int       `json:"event_count"`
	AuthAttempts     int       `json:"auth_attempts"`
	SuccessfulAuths  int       `json:"successful_auths"`
	CommandsExecuted int       `json:"commands_executed"`
	FilesTransferred int       `json:"files_transferred"`
	ThreatLevel      string    `json:"threat_level"`
	Summary          string    `json:"summary"`
	CreateTime       time.Time `json:"create_time"`
}

// ContainerRuntimeLogHandler 容器运行时日志处理器
type ContainerRuntimeLogHandler struct{}

// 内存存储
var (
	runtimeLogs     = make(map[uint]*ContainerRuntimeLog)
	sessionSummaries = make(map[string]*SessionSummary)
	runtimeLogMutex = sync.RWMutex{}
	nextRuntimeLogID = uint(1)
	nextSummaryID   = uint(1)
)

// NewContainerRuntimeLogHandler 创建容器运行时日志处理器
func NewContainerRuntimeLogHandler() *ContainerRuntimeLogHandler {
	return &ContainerRuntimeLogHandler{}
}

// ParseContainerLogs 解析容器日志
func (h *ContainerRuntimeLogHandler) ParseContainerLogs(c *gin.Context) {
	var req struct {
		ContainerID string   `json:"container_id" binding:"required"`
		LogLines    []string `json:"log_lines" binding:"required"`
		LogType     string   `json:"log_type"` // ssh, http, ftp, etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	var parsedLogs []*ContainerRuntimeLog
	
	runtimeLogMutex.Lock()
	for _, logLine := range req.LogLines {
		log := h.parseLogLine(req.ContainerID, logLine, req.LogType)
		if log != nil {
			runtimeLogs[nextRuntimeLogID] = log
			log.ID = nextRuntimeLogID
			nextRuntimeLogID++
			parsedLogs = append(parsedLogs, log)
		}
	}
	runtimeLogMutex.Unlock()

	result := map[string]interface{}{
		"container_id":  req.ContainerID,
		"total_lines":   len(req.LogLines),
		"parsed_logs":   len(parsedLogs),
		"logs":          parsedLogs,
	}

	utils.ResponseSuccess(c, result)
}

// GetLogsByContainer 根据容器获取日志
func (h *ContainerRuntimeLogHandler) GetLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")

	runtimeLogMutex.RLock()
	var logs []*ContainerRuntimeLog
	for _, log := range runtimeLogs {
		if log.ContainerID == containerID {
			logs = append(logs, log)
		}
	}
	runtimeLogMutex.RUnlock()

	utils.ResponseSuccess(c, logs)
}

// GetLogsByTimeRange 根据时间范围获取日志
func (h *ContainerRuntimeLogHandler) GetLogsByTimeRange(c *gin.Context) {
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

	runtimeLogMutex.RLock()
	var logs []*ContainerRuntimeLog
	for _, log := range runtimeLogs {
		if log.Timestamp.After(startTime) && log.Timestamp.Before(endTime) {
			logs = append(logs, log)
		}
	}
	runtimeLogMutex.RUnlock()

	utils.ResponseSuccess(c, logs)
}

// GetLogsByEventType 根据事件类型获取日志
func (h *ContainerRuntimeLogHandler) GetLogsByEventType(c *gin.Context) {
	eventType := c.Param("event_type")

	runtimeLogMutex.RLock()
	var logs []*ContainerRuntimeLog
	for _, log := range runtimeLogs {
		if log.EventType == eventType {
			logs = append(logs, log)
		}
	}
	runtimeLogMutex.RUnlock()

	utils.ResponseSuccess(c, logs)
}

// GetLogsBySourceIP 根据源IP获取日志
func (h *ContainerRuntimeLogHandler) GetLogsBySourceIP(c *gin.Context) {
	sourceIP := c.Param("source_ip")

	runtimeLogMutex.RLock()
	var logs []*ContainerRuntimeLog
	for _, log := range runtimeLogs {
		if log.SourceIP == sourceIP {
			logs = append(logs, log)
		}
	}
	runtimeLogMutex.RUnlock()

	utils.ResponseSuccess(c, logs)
}

// GetSessionSummary 获取会话汇总
func (h *ContainerRuntimeLogHandler) GetSessionSummary(c *gin.Context) {
	sessionID := c.Param("session_id")

	runtimeLogMutex.RLock()
	summary, exists := sessionSummaries[sessionID]
	runtimeLogMutex.RUnlock()

	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "会话汇总不存在")
		return
	}

	utils.ResponseSuccess(c, summary)
}

// CreateSessionSummary 创建会话汇总
func (h *ContainerRuntimeLogHandler) CreateSessionSummary(c *gin.Context) {
	var req struct {
		SessionID   string `json:"session_id" binding:"required"`
		ContainerID string `json:"container_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 收集会话相关的日志
	runtimeLogMutex.RLock()
	var sessionLogs []*ContainerRuntimeLog
	for _, log := range runtimeLogs {
		if log.SessionID == req.SessionID && log.ContainerID == req.ContainerID {
			sessionLogs = append(sessionLogs, log)
		}
	}
	runtimeLogMutex.RUnlock()

	if len(sessionLogs) == 0 {
		utils.ResponseError(c, http.StatusNotFound, "未找到相关日志")
		return
	}

	// 分析会话数据
	summary := h.analyzeSession(req.SessionID, req.ContainerID, sessionLogs)

	runtimeLogMutex.Lock()
	sessionSummaries[req.SessionID] = summary
	runtimeLogMutex.Unlock()

	utils.ResponseSuccess(c, summary)
}

// GetLogStatistics 获取日志统计
func (h *ContainerRuntimeLogHandler) GetLogStatistics(c *gin.Context) {
	runtimeLogMutex.RLock()
	
	stats := map[string]interface{}{
		"total_logs":        len(runtimeLogs),
		"total_summaries":   len(sessionSummaries),
		"logs_by_type":      make(map[string]int),
		"logs_by_container": make(map[string]int),
		"logs_by_ip":        make(map[string]int),
		"recent_logs":       make([]*ContainerRuntimeLog, 0),
	}

	for _, log := range runtimeLogs {
		stats["logs_by_type"].(map[string]int)[log.EventType]++
		stats["logs_by_container"].(map[string]int)[log.ContainerID]++
		stats["logs_by_ip"].(map[string]int)[log.SourceIP]++
	}

	runtimeLogMutex.RUnlock()

	utils.ResponseSuccess(c, stats)
}

// GetAttackAnalysis 获取攻击分析
func (h *ContainerRuntimeLogHandler) GetAttackAnalysis(c *gin.Context) {
	runtimeLogMutex.RLock()
	
	analysis := map[string]interface{}{
		"total_attacks":     0,
		"attack_sources":    make(map[string]int),
		"attack_types":      make(map[string]int),
		"attack_timeline":   make([]map[string]interface{}, 0),
		"threat_indicators": make([]string, 0),
	}

	for _, log := range runtimeLogs {
		if h.isAttackEvent(log) {
			analysis["total_attacks"] = analysis["total_attacks"].(int) + 1
			analysis["attack_sources"].(map[string]int)[log.SourceIP]++
			analysis["attack_types"].(map[string]int)[log.EventType]++
		}
	}

	runtimeLogMutex.RUnlock()

	utils.ResponseSuccess(c, analysis)
}

// ExportLogs 导出日志
func (h *ContainerRuntimeLogHandler) ExportLogs(c *gin.Context) {
	var req struct {
		ContainerID string    `json:"container_id"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		EventTypes  []string  `json:"event_types"`
		Format      string    `json:"format"` // json, csv
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	runtimeLogMutex.RLock()
	var exportLogs []*ContainerRuntimeLog
	for _, log := range runtimeLogs {
		// 过滤条件
		if req.ContainerID != "" && log.ContainerID != req.ContainerID {
			continue
		}
		if !req.StartTime.IsZero() && log.Timestamp.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && log.Timestamp.After(req.EndTime) {
			continue
		}
		if len(req.EventTypes) > 0 {
			found := false
			for _, eventType := range req.EventTypes {
				if log.EventType == eventType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		exportLogs = append(exportLogs, log)
	}
	runtimeLogMutex.RUnlock()

	if req.Format == "csv" {
		h.exportCSV(c, exportLogs)
	} else {
		h.exportJSON(c, exportLogs)
	}
}

// parseLogLine 解析日志行
func (h *ContainerRuntimeLogHandler) parseLogLine(containerID, logLine, logType string) *ContainerRuntimeLog {
	// 简单的日志解析实现
	log := &ContainerRuntimeLog{
		ContainerID: containerID,
		RawLog:      logLine,
		Timestamp:   time.Now(),
		Data:        make(map[string]interface{}),
		Parsed:      false,
	}

	// 尝试解析JSON格式的日志
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(logLine), &jsonData); err == nil {
		log.Data = jsonData
		log.Parsed = true
		
		// 提取常见字段
		if sessionID, ok := jsonData["session_id"].(string); ok {
			log.SessionID = sessionID
		}
		if sourceIP, ok := jsonData["source_ip"].(string); ok {
			log.SourceIP = sourceIP
		}
		if eventType, ok := jsonData["event_type"].(string); ok {
			log.EventType = eventType
		}
		if protocol, ok := jsonData["protocol"].(string); ok {
			log.Protocol = protocol
		}
	} else {
		// 简单的文本解析
		if strings.Contains(logLine, "SSH") {
			log.EventType = "ssh_connection"
			log.Protocol = "ssh"
		} else if strings.Contains(logLine, "HTTP") {
			log.EventType = "http_request"
			log.Protocol = "http"
		} else {
			log.EventType = "unknown"
		}
	}

	return log
}

// analyzeSession 分析会话
func (h *ContainerRuntimeLogHandler) analyzeSession(sessionID, containerID string, logs []*ContainerRuntimeLog) *SessionSummary {
	summary := &SessionSummary{
		ID:          nextSummaryID,
		SessionID:   sessionID,
		ContainerID: containerID,
		CreateTime:  time.Now(),
	}
	nextSummaryID++

	if len(logs) == 0 {
		return summary
	}

	// 分析时间范围
	summary.StartTime = logs[0].Timestamp
	endTime := logs[len(logs)-1].Timestamp
	summary.EndTime = &endTime
	summary.Duration = endTime.Sub(summary.StartTime).Milliseconds()

	// 统计事件
	summary.EventCount = len(logs)
	for _, log := range logs {
		if log.SourceIP != "" && summary.SourceIP == "" {
			summary.SourceIP = log.SourceIP
		}
		if log.Protocol != "" && summary.Protocol == "" {
			summary.Protocol = log.Protocol
		}

		switch log.EventType {
		case "auth_attempt":
			summary.AuthAttempts++
		case "auth_success":
			summary.SuccessfulAuths++
		case "command_execution":
			summary.CommandsExecuted++
		case "file_transfer":
			summary.FilesTransferred++
		}
	}

	// 评估威胁等级
	if summary.SuccessfulAuths > 0 && summary.CommandsExecuted > 10 {
		summary.ThreatLevel = "high"
	} else if summary.AuthAttempts > 5 {
		summary.ThreatLevel = "medium"
	} else {
		summary.ThreatLevel = "low"
	}

	summary.Summary = h.generateSummaryText(summary)
	return summary
}

// isAttackEvent 判断是否为攻击事件
func (h *ContainerRuntimeLogHandler) isAttackEvent(log *ContainerRuntimeLog) bool {
	attackTypes := []string{"brute_force", "sql_injection", "xss", "command_injection", "file_upload"}
	for _, attackType := range attackTypes {
		if log.EventType == attackType {
			return true
		}
	}
	return false
}

// generateSummaryText 生成汇总文本
func (h *ContainerRuntimeLogHandler) generateSummaryText(summary *SessionSummary) string {
	return "会话分析完成，包含 " + strconv.Itoa(summary.EventCount) + " 个事件"
}

// exportJSON 导出JSON格式
func (h *ContainerRuntimeLogHandler) exportJSON(c *gin.Context, logs []*ContainerRuntimeLog) {
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=container_logs.json")
	c.JSON(http.StatusOK, logs)
}

// exportCSV 导出CSV格式
func (h *ContainerRuntimeLogHandler) exportCSV(c *gin.Context, logs []*ContainerRuntimeLog) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=container_logs.csv")
	
	// 简单的CSV实现
	csvContent := "ID,ContainerID,SessionID,EventType,SourceIP,Protocol,Timestamp\n"
	for _, log := range logs {
		csvContent += strconv.Itoa(int(log.ID)) + "," +
			log.ContainerID + "," +
			log.SessionID + "," +
			log.EventType + "," +
			log.SourceIP + "," +
			log.Protocol + "," +
			log.Timestamp.Format(time.RFC3339) + "\n"
	}
	
	c.String(http.StatusOK, csvContent)
}
