package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HoneypotHandler 蜜罐处理器
type HoneypotHandler struct{}

// NewHoneypotHandler 创建蜜罐处理器
func NewHoneypotHandler() *HoneypotHandler {
	return &HoneypotHandler{}
}

// DeployHoneypot 部署蜜罐
func (h *HoneypotHandler) DeployHoneypot(c *gin.Context) {
	// 蜜罐模板功能已移除，返回错误
	utils.ResponseError(c, http.StatusNotImplemented, "蜜罐模板功能已移除，请使用容器实例管理功能")
}

// StopHoneypot 停止蜜罐
func (h *HoneypotHandler) StopHoneypot(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	// 初始化服务
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	// 停止蜜罐
	if err := service.StopInstance(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "停止蜜罐失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "停止成功")
}

// GetHoneypotStatus 获取蜜罐状态
func (h *HoneypotHandler) GetHoneypotStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	// 初始化服务
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	// 获取实例
	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取蜜罐实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"id":     instance.ID,
		"name":   instance.Name,
		"status": instance.Status,
	})
}

// ListHoneypots 列出所有蜜罐
func (h *HoneypotHandler) ListHoneypots(c *gin.Context) {
	// 初始化服务
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	// 获取所有实例
	instances, err := service.GetAllInstances()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取蜜罐实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instances)
}

// GetHoneypotLogs 获取蜜罐日志
func (h *HoneypotHandler) GetHoneypotLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	// 获取日志
	repo := repositories.NewMySQLHoneypotLogRepo(config.MySQLDB)
	logs, err := repo.GetByInstanceID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}
