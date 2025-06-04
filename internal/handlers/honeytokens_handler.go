package handlers

import (
	"andorralee/pkg/utils"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// HoneyToken 蜜签模型
type HoneyToken struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // credential, file, url, email
	Content      string    `json:"content"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
	TriggerCount int       `json:"trigger_count"`
}

// HoneyTokenTrigger 蜜签触发记录
type HoneyTokenTrigger struct {
	ID          uint      `json:"id"`
	TokenID     uint      `json:"token_id"`
	TokenName   string    `json:"token_name"`
	SourceIP    string    `json:"source_ip"`
	UserAgent   string    `json:"user_agent"`
	TriggerTime time.Time `json:"trigger_time"`
	Action      string    `json:"action"`
	Details     string    `json:"details"`
}

// 内存存储
var (
	honeyTokens   = make(map[uint]*HoneyToken)
	tokenTriggers = make(map[uint]*HoneyTokenTrigger)
	tokenMutex    = sync.RWMutex{}
	triggerMutex  = sync.RWMutex{}
	nextTokenID   = uint(1)
	nextTriggerID = uint(1)
)

// CreateHoneyToken 创建蜜签
func CreateHoneyToken(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tokenMutex.Lock()
	token := &HoneyToken{
		ID:           nextTokenID,
		Name:         req.Name,
		Type:         req.Type,
		Content:      req.Content,
		Description:  req.Description,
		IsActive:     true,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
		TriggerCount: 0,
	}
	honeyTokens[nextTokenID] = token
	nextTokenID++
	tokenMutex.Unlock()

	utils.ResponseSuccess(c, token)
}

// GetAllHoneyTokens 获取所有蜜签
func GetAllHoneyTokens(c *gin.Context) {
	tokenMutex.RLock()
	tokens := make([]*HoneyToken, 0, len(honeyTokens))
	for _, token := range honeyTokens {
		tokens = append(tokens, token)
	}
	tokenMutex.RUnlock()

	utils.ResponseSuccess(c, tokens)
}

// GetHoneyTokenByID 根据ID获取蜜签
func GetHoneyTokenByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	tokenMutex.RLock()
	token, exists := honeyTokens[uint(id)]
	tokenMutex.RUnlock()

	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "蜜签不存在")
		return
	}

	utils.ResponseSuccess(c, token)
}

// UpdateHoneyToken 更新蜜签
func UpdateHoneyToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Content     string `json:"content"`
		Description string `json:"description"`
		IsActive    *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tokenMutex.Lock()
	token, exists := honeyTokens[uint(id)]
	if !exists {
		tokenMutex.Unlock()
		utils.ResponseError(c, http.StatusNotFound, "蜜签不存在")
		return
	}

	// 更新字段
	if req.Name != "" {
		token.Name = req.Name
	}
	if req.Type != "" {
		token.Type = req.Type
	}
	if req.Content != "" {
		token.Content = req.Content
	}
	if req.Description != "" {
		token.Description = req.Description
	}
	if req.IsActive != nil {
		token.IsActive = *req.IsActive
	}
	token.UpdateTime = time.Now()
	tokenMutex.Unlock()

	utils.ResponseSuccess(c, token)
}

// DeleteHoneyToken 删除蜜签
func DeleteHoneyToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	tokenMutex.Lock()
	_, exists := honeyTokens[uint(id)]
	if !exists {
		tokenMutex.Unlock()
		utils.ResponseError(c, http.StatusNotFound, "蜜签不存在")
		return
	}
	delete(honeyTokens, uint(id))
	tokenMutex.Unlock()

	utils.ResponseSuccess(c, "蜜签删除成功")
}

// TriggerHoneyToken 触发蜜签（模拟攻击者触发）
func TriggerHoneyToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	var req struct {
		Action  string `json:"action" binding:"required"`
		Details string `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	tokenMutex.Lock()
	token, exists := honeyTokens[uint(id)]
	if !exists {
		tokenMutex.Unlock()
		utils.ResponseError(c, http.StatusNotFound, "蜜签不存在")
		return
	}

	if !token.IsActive {
		tokenMutex.Unlock()
		utils.ResponseError(c, http.StatusBadRequest, "蜜签未激活")
		return
	}

	// 增加触发次数
	token.TriggerCount++
	tokenMutex.Unlock()

	// 记录触发事件
	triggerMutex.Lock()
	trigger := &HoneyTokenTrigger{
		ID:          nextTriggerID,
		TokenID:     uint(id),
		TokenName:   token.Name,
		SourceIP:    c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		TriggerTime: time.Now(),
		Action:      req.Action,
		Details:     req.Details,
	}
	tokenTriggers[nextTriggerID] = trigger
	nextTriggerID++
	triggerMutex.Unlock()

	// 记录日志
	fmt.Printf("🚨 蜜签触发警报: %s (ID:%d) 被 %s 触发，动作: %s\n",
		token.Name, token.ID, c.ClientIP(), req.Action)

	utils.ResponseSuccess(c, map[string]interface{}{
		"message":      "蜜签触发记录成功",
		"token_name":   token.Name,
		"trigger_id":   trigger.ID,
		"trigger_time": trigger.TriggerTime,
	})
}

// GetHoneyTokenTriggers 获取蜜签触发记录
func GetHoneyTokenTriggers(c *gin.Context) {
	tokenIDStr := c.Query("token_id")

	triggerMutex.RLock()
	triggers := make([]*HoneyTokenTrigger, 0)

	if tokenIDStr != "" {
		// 获取特定蜜签的触发记录
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			triggerMutex.RUnlock()
			utils.ResponseError(c, http.StatusBadRequest, "无效的token_id: "+err.Error())
			return
		}

		for _, trigger := range tokenTriggers {
			if trigger.TokenID == uint(tokenID) {
				triggers = append(triggers, trigger)
			}
		}
	} else {
		// 获取所有触发记录
		for _, trigger := range tokenTriggers {
			triggers = append(triggers, trigger)
		}
	}
	triggerMutex.RUnlock()

	utils.ResponseSuccess(c, triggers)
}

// GetHoneyTokenStatistics 获取蜜签统计信息
func GetHoneyTokenStatistics(c *gin.Context) {
	tokenMutex.RLock()
	triggerMutex.RLock()

	stats := map[string]interface{}{
		"total_tokens":    len(honeyTokens),
		"active_tokens":   0,
		"total_triggers":  len(tokenTriggers),
		"tokens_by_type":  make(map[string]int),
		"recent_triggers": make([]*HoneyTokenTrigger, 0),
	}

	// 统计活跃蜜签和类型分布
	for _, token := range honeyTokens {
		if token.IsActive {
			stats["active_tokens"] = stats["active_tokens"].(int) + 1
		}
		stats["tokens_by_type"].(map[string]int)[token.Type]++
	}

	// 获取最近的触发记录（最多10条）
	recentTriggers := make([]*HoneyTokenTrigger, 0, 10)
	for _, trigger := range tokenTriggers {
		recentTriggers = append(recentTriggers, trigger)
		if len(recentTriggers) >= 10 {
			break
		}
	}
	stats["recent_triggers"] = recentTriggers

	triggerMutex.RUnlock()
	tokenMutex.RUnlock()

	utils.ResponseSuccess(c, stats)
}

// CreateDefaultHoneyTokens 创建默认蜜签
func CreateDefaultHoneyTokens() {
	defaultTokens := []*HoneyToken{
		{
			ID:          1,
			Name:        "管理员凭证",
			Type:        "credential",
			Content:     "admin:123456",
			Description: "虚假的管理员账号密码",
			IsActive:    true,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		},
		{
			ID:          2,
			Name:        "敏感文件路径",
			Type:        "file",
			Content:     "/etc/passwd",
			Description: "系统敏感文件路径",
			IsActive:    true,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		},
		{
			ID:          3,
			Name:        "数据库连接字符串",
			Type:        "credential",
			Content:     "mysql://root:password@localhost:3306/production",
			Description: "虚假的数据库连接信息",
			IsActive:    true,
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		},
	}

	tokenMutex.Lock()
	for _, token := range defaultTokens {
		honeyTokens[token.ID] = token
		if token.ID >= nextTokenID {
			nextTokenID = token.ID + 1
		}
	}
	tokenMutex.Unlock()
}
