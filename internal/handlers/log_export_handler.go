package handlers

import (
	"andorralee/pkg/utils"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogExportRequest 日志导出请求
type LogExportRequest struct {
	LogType   string    `json:"log_type"` // headling, cowrie, attack, honeytokens
	Format    string    `json:"format"`   // json, csv
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	SourceIP  string    `json:"source_ip"`
	Protocol  string    `json:"protocol"`
}

// ExportLogs 导出日志
func ExportLogs(c *gin.Context) {
	var req LogExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证参数
	if req.LogType == "" {
		req.LogType = "all"
	}
	if req.Format == "" {
		req.Format = "json"
	}
	if req.Format != "json" && req.Format != "csv" {
		utils.ResponseError(c, http.StatusBadRequest, "不支持的导出格式，仅支持 json 或 csv")
		return
	}

	// 收集日志数据
	var logs []map[string]interface{}
	var err error

	switch req.LogType {
	case "attack":
		logs, err = exportAttackLogs(req)
	case "honeytokens":
		logs, err = exportHoneyTokenLogs(req)
	case "containers":
		logs, err = exportContainerLogs(req)
	case "all":
		logs, err = exportAllLogs(req)
	default:
		utils.ResponseError(c, http.StatusBadRequest, "不支持的日志类型")
		return
	}

	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "导出日志失败: "+err.Error())
		return
	}

	// 脱敏处理
	sanitizedLogs := sanitizeLogs(logs)

	// 根据格式导出
	if req.Format == "csv" {
		exportCSV(c, sanitizedLogs, req.LogType)
	} else {
		exportJSON(c, sanitizedLogs, req.LogType)
	}
}

// exportAttackLogs 导出攻击日志
func exportAttackLogs(req LogExportRequest) ([]map[string]interface{}, error) {
	attackMutex.RLock()
	defer attackMutex.RUnlock()

	var logs []map[string]interface{}
	for _, event := range attackEvents {
		// 时间过滤
		if !req.StartTime.IsZero() && event.Timestamp.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && event.Timestamp.After(req.EndTime) {
			continue
		}

		// IP过滤
		if req.SourceIP != "" && event.SourceIP != req.SourceIP {
			continue
		}

		// 协议过滤
		if req.Protocol != "" && event.Protocol != req.Protocol {
			continue
		}

		log := map[string]interface{}{
			"id":             event.ID,
			"timestamp":      event.Timestamp.Format(time.RFC3339),
			"source_ip":      event.SourceIP,
			"source_port":    event.SourcePort,
			"dest_ip":        event.DestIP,
			"dest_port":      event.DestPort,
			"protocol":       event.Protocol,
			"attack_type":    event.AttackType,
			"payload":        event.Payload,
			"severity":       event.Severity,
			"container_id":   event.ContainerID,
			"container_name": event.ContainerName,
			"user_agent":     event.UserAgent,
			"session_id":     event.SessionID,
			"log_type":       "attack",
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// exportHoneyTokenLogs 导出蜜签日志
func exportHoneyTokenLogs(req LogExportRequest) ([]map[string]interface{}, error) {
	triggerMutex.RLock()
	defer triggerMutex.RUnlock()

	var logs []map[string]interface{}
	for _, trigger := range tokenTriggers {
		// 时间过滤
		if !req.StartTime.IsZero() && trigger.TriggerTime.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && trigger.TriggerTime.After(req.EndTime) {
			continue
		}

		// IP过滤
		if req.SourceIP != "" && trigger.SourceIP != req.SourceIP {
			continue
		}

		log := map[string]interface{}{
			"id":         trigger.ID,
			"timestamp":  trigger.TriggerTime.Format(time.RFC3339),
			"token_id":   trigger.TokenID,
			"token_name": trigger.TokenName,
			"source_ip":  trigger.SourceIP,
			"user_agent": trigger.UserAgent,
			"action":     trigger.Action,
			"details":    trigger.Details,
			"log_type":   "honeytoken",
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// exportContainerLogs 导出容器日志
func exportContainerLogs(req LogExportRequest) ([]map[string]interface{}, error) {
	instanceMutex.RLock()
	defer instanceMutex.RUnlock()

	var logs []map[string]interface{}
	for _, instance := range memoryInstances {
		// 时间过滤
		if !req.StartTime.IsZero() && instance.CreateTime.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && instance.CreateTime.After(req.EndTime) {
			continue
		}

		// 协议过滤
		if req.Protocol != "" && instance.Protocol != req.Protocol {
			continue
		}

		log := map[string]interface{}{
			"id":             instance.ID,
			"timestamp":      instance.CreateTime.Format(time.RFC3339),
			"name":           instance.Name,
			"honeypot_name":  instance.HoneypotName,
			"container_name": instance.ContainerName,
			"container_id":   instance.ContainerID,
			"ip":             instance.IP,
			"honeypot_ip":    instance.HoneypotIP,
			"port":           instance.Port,
			"protocol":       instance.Protocol,
			"interface_type": instance.InterfaceType,
			"status":         instance.Status,
			"image_name":     instance.ImageName,
			"image_id":       instance.ImageID,
			"description":    instance.Description,
			"log_type":       "container",
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// exportAllLogs 导出所有日志
func exportAllLogs(req LogExportRequest) ([]map[string]interface{}, error) {
	var allLogs []map[string]interface{}

	// 攻击日志
	attackLogs, err := exportAttackLogs(req)
	if err != nil {
		return nil, err
	}
	allLogs = append(allLogs, attackLogs...)

	// 蜜签日志
	tokenLogs, err := exportHoneyTokenLogs(req)
	if err != nil {
		return nil, err
	}
	allLogs = append(allLogs, tokenLogs...)

	// 容器日志
	containerLogs, err := exportContainerLogs(req)
	if err != nil {
		return nil, err
	}
	allLogs = append(allLogs, containerLogs...)

	return allLogs, nil
}

// sanitizeLogs 脱敏处理日志
func sanitizeLogs(logs []map[string]interface{}) []map[string]interface{} {
	for _, log := range logs {
		// 脱敏密码相关信息
		if payload, exists := log["payload"]; exists {
			if payloadStr, ok := payload.(string); ok {
				log["payload"] = sanitizePayload(payloadStr)
			}
		}

		// 脱敏用户代理
		if userAgent, exists := log["user_agent"]; exists {
			if userAgentStr, ok := userAgent.(string); ok {
				log["user_agent"] = sanitizeUserAgent(userAgentStr)
			}
		}

		// 脱敏详细信息
		if details, exists := log["details"]; exists {
			if detailsStr, ok := details.(string); ok {
				log["details"] = sanitizeDetails(detailsStr)
			}
		}
	}
	return logs
}

// sanitizePayload 脱敏攻击载荷
func sanitizePayload(payload string) string {
	// 替换可能的密码
	payload = strings.ReplaceAll(payload, "password=", "password=***")
	payload = strings.ReplaceAll(payload, "pwd=", "pwd=***")
	payload = strings.ReplaceAll(payload, "pass=", "pass=***")

	// 替换可能的用户名密码组合
	if strings.Contains(payload, ":") && len(payload) < 50 {
		parts := strings.Split(payload, ":")
		if len(parts) == 2 {
			payload = parts[0] + ":***"
		}
	}

	return payload
}

// sanitizeUserAgent 脱敏用户代理
func sanitizeUserAgent(userAgent string) string {
	// 保留主要信息，移除详细版本号
	if len(userAgent) > 100 {
		return userAgent[:100] + "..."
	}
	return userAgent
}

// sanitizeDetails 脱敏详细信息
func sanitizeDetails(details string) string {
	// 替换可能的敏感信息
	details = strings.ReplaceAll(details, "password", "***")
	details = strings.ReplaceAll(details, "token", "***")
	details = strings.ReplaceAll(details, "secret", "***")
	return details
}

// exportJSON 导出JSON格式
func exportJSON(c *gin.Context, logs []map[string]interface{}, logType string) {
	filename := fmt.Sprintf("%s_logs_%s.json", logType, time.Now().Format("20060102_150405"))

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	result := map[string]interface{}{
		"export_time": time.Now().Format(time.RFC3339),
		"log_type":    logType,
		"total_count": len(logs),
		"logs":        logs,
	}

	c.JSON(http.StatusOK, result)
}

// exportCSV 导出CSV格式
func exportCSV(c *gin.Context, logs []map[string]interface{}, logType string) {
	filename := fmt.Sprintf("%s_logs_%s.csv", logType, time.Now().Format("20060102_150405"))

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	if len(logs) == 0 {
		writer.Write([]string{"No data"})
		return
	}

	// 写入表头
	var headers []string
	for key := range logs[0] {
		headers = append(headers, key)
	}
	writer.Write(headers)

	// 写入数据
	for _, log := range logs {
		var row []string
		for _, header := range headers {
			if value, exists := log[header]; exists {
				row = append(row, fmt.Sprintf("%v", value))
			} else {
				row = append(row, "")
			}
		}
		writer.Write(row)
	}
}

// GetLogStatistics 获取日志统计信息
func GetLogStatistics(c *gin.Context) {
	stats := map[string]interface{}{
		"attack_logs":     len(attackEvents),
		"honeytoken_logs": len(tokenTriggers),
		"container_logs":  len(memoryInstances),
		"total_logs":      len(attackEvents) + len(tokenTriggers) + len(memoryInstances),
		"last_updated":    time.Now().Format(time.RFC3339),
	}

	utils.ResponseSuccess(c, stats)
}
