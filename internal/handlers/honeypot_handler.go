package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HoneypotHandler 蜜罐处理器
type HoneypotHandler struct {
	honeypotService *services.HoneypotService
}

// NewHoneypotHandler 创建蜜罐处理器实例
func NewHoneypotHandler() *HoneypotHandler {
	return &HoneypotHandler{
		honeypotService: services.NewHoneypotService(),
	}
}

// DeployHoneypot 部署蜜罐
func (h *HoneypotHandler) DeployHoneypot(c *gin.Context) {
	var req struct {
		Type config.HoneypotType `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerID, err := h.honeypotService.DeployHoneypot(req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"container_id": containerID,
		"message":      "Honeypot deployed successfully",
	})
}

// StopHoneypot 停止蜜罐
func (h *HoneypotHandler) StopHoneypot(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	if err := h.honeypotService.StopHoneypot(containerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Honeypot stopped successfully",
	})
}

// GetHoneypotStatus 获取蜜罐状态
func (h *HoneypotHandler) GetHoneypotStatus(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	status, err := h.honeypotService.GetHoneypotStatus(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListHoneypots 列出所有蜜罐
func (h *HoneypotHandler) ListHoneypots(c *gin.Context) {
	containers, err := h.honeypotService.ListHoneypots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, containers)
}

// GetHoneypotLogs 获取蜜罐日志
func (h *HoneypotHandler) GetHoneypotLogs(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	logs, err := h.honeypotService.GetHoneypotLogs(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}
