package handlers

import (
	"andorralee/internal/repositories"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllInstances 获取所有蜜罐实例
// @Summary 获取所有蜜罐实例
// @Description 获取所有蜜罐实例信息
// @Tags 蜜罐管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /honeypot/instances [get]
func GetAllInstances(c *gin.Context) {
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	instances, err := service.GetAllInstances()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instances)
}

// GetInstanceByID 根据ID获取蜜罐实例
// @Summary 根据ID获取蜜罐实例
// @Description 根据ID获取蜜罐实例详细信息
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "实例ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/instances/{id} [get]
func GetInstanceByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instance)
}

// UpdateInstance 更新蜜罐实例
// @Summary 更新蜜罐实例
// @Description 更新现有的蜜罐实例
// @Tags 蜜罐管理
// @Accept json
// @Produce json
// @Param id path int true "实例ID"
// @Param instance body repositories.HoneypotInstance true "实例信息"
// @Success 200 {object} utils.Response
// @Router /honeypot/instances/{id} [put]
func UpdateInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	var instance repositories.HoneypotInstance
	if err := c.ShouldBindJSON(&instance); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	instance.ID = uint(id)

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.UpdateInstance(&instance); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "更新实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instance)
}

// DeleteInstance 删除蜜罐实例
// @Summary 删除蜜罐实例
// @Description 删除指定的蜜罐实例
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "实例ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/instances/{id} [delete]
func DeleteInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.DeleteInstance(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "删除成功")
}

// DeployInstance 部署蜜罐实例
// @Summary 部署蜜罐实例
// @Description 部署指定的蜜罐实例（启动Docker容器）
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "实例ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/instances/{id}/deploy [post]
func DeployInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.DeployInstance(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "部署实例失败: "+err.Error())
		return
	}

	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取实例信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instance)
}

// StopInstance 停止蜜罐实例
// @Summary 停止蜜罐实例
// @Description 停止指定的蜜罐实例（停止Docker容器）
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "实例ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/instances/{id}/stop [post]
func StopInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.StopInstance(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "停止实例失败: "+err.Error())
		return
	}

	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取实例信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instance)
}

// GetInstanceLogs 获取蜜罐实例日志
// @Summary 获取蜜罐实例日志
// @Description 获取指定蜜罐实例的运行日志
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "实例ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/instances/{id}/logs [get]
func GetInstanceLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	logs, err := service.GetInstanceLogs(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}
