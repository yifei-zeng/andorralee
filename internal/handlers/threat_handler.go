package handlers

import (
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SaveThreatIntelligence 保存威胁情报
func SaveThreatIntelligence(c *gin.Context) {
	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	var threat repositories.ThreatIntelligence
	if err := c.ShouldBindJSON(&threat); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if threat.IndicatorType == "" || threat.IndicatorValue == "" || threat.ThreatType == "" {
		utils.ResponseError(c, http.StatusBadRequest, "指标类型、指标值和威胁类型为必填字段")
		return
	}

	// 设置默认值
	if threat.Severity == "" {
		threat.Severity = "medium"
	}
	if threat.Confidence == 0 {
		threat.Confidence = 50
	}
	threat.IsActive = true

	if err := dbService.SaveThreatIntelligence(&threat); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "保存威胁情报失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, threat)
}

// GetThreatIntelligence 查询威胁情报
func GetThreatIntelligence(c *gin.Context) {
	indicatorType := c.Query("type")
	indicatorValue := c.Query("value")

	if indicatorType == "" || indicatorValue == "" {
		utils.ResponseError(c, http.StatusBadRequest, "指标类型和指标值不能为空")
		return
	}

	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	threat, err := dbService.GetThreatIntelligence(indicatorType, indicatorValue)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "未找到威胁情报")
		return
	}

	utils.ResponseSuccess(c, threat)
}

// StartAttackSession 开始攻击会话跟踪
func StartAttackSession(c *gin.Context) {
	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	var request struct {
		SourceIP        string `json:"source_ip" binding:"required"`
		DestinationIP   string `json:"destination_ip"`
		DestinationPort uint   `json:"destination_port"`
		Protocol        string `json:"protocol"`
		Description     string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 生成会话ID
	sessionID, err := generateSessionID()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "生成会话ID失败: "+err.Error())
		return
	}

	// 创建攻击会话
	session := &repositories.AttackSession{
		SessionID:       sessionID,
		SourceIP:        request.SourceIP,
		DestinationIP:   request.DestinationIP,
		DestinationPort: request.DestinationPort,
		Protocol:        request.Protocol,
		Status:          "active",
		StartTime:       time.Now(),
	}

	if err := dbService.SaveAttackSession(session); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "保存攻击会话失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"session_id": sessionID,
		"session":    session,
	})
}

// EndAttackSession 结束攻击会话
func EndAttackSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	// 更新会话状态
	updates := map[string]interface{}{
		"status":   "completed",
		"end_time": time.Now(),
	}

	if err := dbService.UpdateAttackSession(sessionID, updates); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "更新攻击会话失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"message":    "攻击会话已结束",
		"session_id": sessionID,
	})
}

// AddAttackEvent 添加攻击事件
func AddAttackEvent(c *gin.Context) {
	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	var event repositories.AttackEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if event.SessionID == "" || event.EventType == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID和事件类型为必填字段")
		return
	}

	// 设置默认值
	if event.ThreatLevel == "" {
		event.ThreatLevel = "medium"
	}

	if err := dbService.SaveAttackEvent(&event); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "保存攻击事件失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, event)
}

// GetAttackSession 获取攻击会话详情
func GetAttackSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "会话ID不能为空")
		return
	}

	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	session, err := dbService.GetAttackSession(sessionID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "未找到攻击会话")
		return
	}

	// 获取会话事件
	events, err := dbService.GetAttackEvents(sessionID)
	if err != nil {
		events = []repositories.AttackEvent{} // 如果获取失败，返回空数组
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"session": session,
		"events":  events,
	})
}

// GetHoneytokenEvents 获取蜜签事件历史

// GetThreatAssessment 获取威胁评估
func GetThreatAssessment(c *gin.Context) {
	if dbService == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库服务未初始化")
		return
	}

	// 获取攻击统计
	attackStats, err := dbService.GetAttackStatistics()
	if err != nil {
		attackStats = make(map[string]interface{})
	}

	// 获取威胁统计
	threatStats, err := dbService.GetThreatStatistics()
	if err != nil {
		threatStats = make(map[string]interface{})
	}

	// 计算威胁评估分数
	assessment := calculateThreatScore(attackStats, threatStats)

	utils.ResponseSuccess(c, map[string]interface{}{
		"threat_level":    assessment["threat_level"],
		"threat_score":    assessment["threat_score"],
		"risk_factors":    assessment["risk_factors"],
		"recommendations": assessment["recommendations"],
		"attack_stats":    attackStats,
		"threat_stats":    threatStats,
		"assessment_time": time.Now(),
	})
}

// generateSessionID 生成唯一的会话ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// calculateThreatScore 计算威胁评估分数
func calculateThreatScore(attackStats, threatStats map[string]interface{}) map[string]interface{} {
	score := 0
	riskFactors := []string{}
	recommendations := []string{}

	// 检查活跃会话数
	if activeSessions, ok := attackStats["active_sessions"].(int64); ok && activeSessions > 0 {
		score += int(activeSessions) * 10
		riskFactors = append(riskFactors, fmt.Sprintf("当前有%d个活跃攻击会话", activeSessions))
		recommendations = append(recommendations, "监控活跃攻击会话，必要时阻断恶意连接")
	}

	// 检查感染文件数
	if infectedFiles, ok := threatStats["infected_files"].(int64); ok && infectedFiles > 0 {
		score += int(infectedFiles) * 20
		riskFactors = append(riskFactors, fmt.Sprintf("检测到%d个感染文件", infectedFiles))
		recommendations = append(recommendations, "隔离感染文件，更新防病毒规则")
	}

	// 确定威胁等级
	var threatLevel string
	switch {
	case score >= 100:
		threatLevel = "critical"
	case score >= 50:
		threatLevel = "high"
	case score >= 20:
		threatLevel = "medium"
	case score > 0:
		threatLevel = "low"
	default:
		threatLevel = "clean"
	}

	if len(riskFactors) == 0 {
		riskFactors = append(riskFactors, "当前没有检测到明显的安全威胁")
		recommendations = append(recommendations, "继续保持安全监控")
	}

	return map[string]interface{}{
		"threat_level":    threatLevel,
		"threat_score":    score,
		"risk_factors":    riskFactors,
		"recommendations": recommendations,
	}
}
