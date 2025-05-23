package handlers

import (
	"andorralee/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TrafficHandler 流量管理处理器
type TrafficHandler struct {
	trafficService *services.TrafficService
}

// NewTrafficHandler 创建流量管理处理器实例
func NewTrafficHandler() *TrafficHandler {
	return &TrafficHandler{
		trafficService: services.NewTrafficService(),
	}
}

// AddRedirectRule 添加流量重定向规则
func (h *TrafficHandler) AddRedirectRule(c *gin.Context) {
	var req struct {
		SourcePort string `json:"source_port" binding:"required"`
		TargetPort string `json:"target_port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.trafficService.AddRedirectRule(req.SourcePort, req.TargetPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Redirect rule added successfully",
	})
}

// RemoveRedirectRule 删除流量重定向规则
func (h *TrafficHandler) RemoveRedirectRule(c *gin.Context) {
	var req struct {
		SourcePort string `json:"source_port" binding:"required"`
		TargetPort string `json:"target_port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.trafficService.RemoveRedirectRule(req.SourcePort, req.TargetPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Redirect rule removed successfully",
	})
}

// ListRules 列出所有规则
func (h *TrafficHandler) ListRules(c *gin.Context) {
	rules, err := h.trafficService.ListRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
	})
}

// AddFilterRule 添加过滤规则
func (h *TrafficHandler) AddFilterRule(c *gin.Context) {
	var req struct {
		SourceIP   string `json:"source_ip" binding:"required"`
		TargetPort string `json:"target_port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.trafficService.AddFilterRule(req.SourceIP, req.TargetPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Filter rule added successfully",
	})
}

// RemoveFilterRule 删除过滤规则
func (h *TrafficHandler) RemoveFilterRule(c *gin.Context) {
	var req struct {
		SourceIP   string `json:"source_ip" binding:"required"`
		TargetPort string `json:"target_port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.trafficService.RemoveFilterRule(req.SourceIP, req.TargetPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Filter rule removed successfully",
	})
}

// SaveRules 保存规则
func (h *TrafficHandler) SaveRules(c *gin.Context) {
	if err := h.trafficService.SaveRules(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rules saved successfully",
	})
}

// RestoreRules 恢复规则
func (h *TrafficHandler) RestoreRules(c *gin.Context) {
	if err := h.trafficService.RestoreRules(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rules restored successfully",
	})
}
