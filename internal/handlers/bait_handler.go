package handlers

import (
	"andorralee/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaitHandler 蜜签处理器
type BaitHandler struct {
	baitService *services.BaitService
}

// NewBaitHandler 创建蜜签处理器实例
func NewBaitHandler(basePath string) *BaitHandler {
	return &BaitHandler{
		baitService: services.NewBaitService(basePath),
	}
}

// CreateBait 创建蜜签
func (h *BaitHandler) CreateBait(c *gin.Context) {
	var config services.BaitConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.baitService.CreateBait(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bait created successfully",
	})
}

// GetBait 获取蜜签信息
func (h *BaitHandler) GetBait(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bait ID is required"})
		return
	}

	bait, err := h.baitService.GetBait(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bait)
}

// ListBaits 列出所有蜜签
func (h *BaitHandler) ListBaits(c *gin.Context) {
	baits, err := h.baitService.ListBaits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, baits)
}

// DeleteBait 删除蜜签
func (h *BaitHandler) DeleteBait(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bait ID is required"})
		return
	}

	if err := h.baitService.DeleteBait(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bait deleted successfully",
	})
}

// MonitorBait 监控蜜签访问
func (h *BaitHandler) MonitorBait(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bait ID is required"})
		return
	}

	if err := h.baitService.MonitorBait(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bait access monitored successfully",
	})
}
