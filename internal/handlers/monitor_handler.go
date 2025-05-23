package handlers

import (
	"andorralee/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MonitorHandler 监控处理器
type MonitorHandler struct {
	monitorService *services.MonitorService
}

// NewMonitorHandler 创建监控处理器实例
func NewMonitorHandler(basePath string) *MonitorHandler {
	return &MonitorHandler{
		monitorService: services.NewMonitorService(basePath),
	}
}

// CreateAlert 创建告警
func (h *MonitorHandler) CreateAlert(c *gin.Context) {
	var req struct {
		Type    services.AlertType  `json:"type" binding:"required"`
		Level   services.AlertLevel `json:"level" binding:"required"`
		Source  string              `json:"source" binding:"required"`
		Message string              `json:"message" binding:"required"`
		Details string              `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.monitorService.CreateAlert(req.Type, req.Level, req.Source, req.Message, req.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert created successfully",
	})
}

// ResolveAlert 解决告警
func (h *MonitorHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert ID is required"})
		return
	}

	var req struct {
		ResolvedBy string `json:"resolved_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.monitorService.ResolveAlert(id, req.ResolvedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert resolved successfully",
	})
}

// ListAlerts 列出所有告警
func (h *MonitorHandler) ListAlerts(c *gin.Context) {
	alerts, err := h.monitorService.ListAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetAlert 获取告警信息
func (h *MonitorHandler) GetAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert ID is required"})
		return
	}

	alert, err := h.monitorService.GetAlert(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// MonitorHoneypot 监控蜜罐
func (h *MonitorHandler) MonitorHoneypot(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	if err := h.monitorService.MonitorHoneypot(containerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Honeypot monitoring completed",
	})
}

// MonitorBait 监控蜜签
func (h *MonitorHandler) MonitorBait(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bait ID is required"})
		return
	}

	if err := h.monitorService.MonitorBait(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bait monitoring completed",
	})
}

// MonitorTraffic 监控流量
func (h *MonitorHandler) MonitorTraffic(c *gin.Context) {
	var req struct {
		SourceIP   string `json:"source_ip" binding:"required"`
		TargetPort string `json:"target_port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.monitorService.MonitorTraffic(req.SourceIP, req.TargetPort); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Traffic monitoring completed",
	})
}
