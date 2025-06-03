package handlers

import (
	"andorralee/internal/repositories"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllTemplates 获取所有蜜罐模板
// @Summary 获取所有蜜罐模板
// @Description 获取所有蜜罐模板信息
// @Tags 蜜罐管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /honeypot/templates [get]
func GetAllTemplates(c *gin.Context) {
	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	templates, err := service.GetAllTemplates()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取模板失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, templates)
}

// GetTemplateByID 根据ID获取蜜罐模板
// @Summary 根据ID获取蜜罐模板
// @Description 根据ID获取蜜罐模板详细信息
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/templates/{id} [get]
func GetTemplateByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	template, err := service.GetTemplateByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取模板失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, template)
}

// CreateTemplate 创建蜜罐模板
// @Summary 创建蜜罐模板
// @Description 创建新的蜜罐模板
// @Tags 蜜罐管理
// @Accept json
// @Produce json
// @Param template body repositories.HoneypotTemplate true "模板信息"
// @Success 200 {object} utils.Response
// @Router /honeypot/templates [post]
func CreateTemplate(c *gin.Context) {
	var template repositories.HoneypotTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.CreateTemplate(&template); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建模板失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, template)
}

// UpdateTemplate 更新蜜罐模板
// @Summary 更新蜜罐模板
// @Description 更新现有的蜜罐模板
// @Tags 蜜罐管理
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param template body repositories.HoneypotTemplate true "模板信息"
// @Success 200 {object} utils.Response
// @Router /honeypot/templates/{id} [put]
func UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	var template repositories.HoneypotTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	template.ID = uint(id)

	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.UpdateTemplate(&template); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "更新模板失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, template)
}

// DeleteTemplate 删除蜜罐模板
// @Summary 删除蜜罐模板
// @Description 删除指定的蜜罐模板
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/templates/{id} [delete]
func DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.DeleteTemplate(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除模板失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "删除成功")
}

// ImportTemplate 导入蜜罐模板
// @Summary 导入蜜罐模板
// @Description 从外部导入蜜罐模板
// @Tags 蜜罐管理
// @Accept json
// @Produce json
// @Param template body repositories.HoneypotTemplate true "模板信息"
// @Success 200 {object} utils.Response
// @Router /honeypot/templates/import [post]
func ImportTemplate(c *gin.Context) {
	var template repositories.HoneypotTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	if err := service.ImportTemplate(template.Name, template.Protocol); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "导入模板失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, template)
}

// DeployTemplate 部署蜜罐模板
// @Summary 部署蜜罐模板
// @Description 部署蜜罐模板创建实例
// @Tags 蜜罐管理
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param instance body repositories.HoneypotInstance true "实例信息"
// @Success 200 {object} utils.Response
// @Router /honeypot/templates/{id}/deploy [post]
func DeployTemplate(c *gin.Context) {
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

	service, err := services.NewHoneypotTemplateService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "服务初始化失败: "+err.Error())
		return
	}

	instanceID, err := service.DeployTemplate(uint(id), instance.Name, instance.IP, instance.Port)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "部署模板失败: "+err.Error())
		return
	}

	instance.ID = instanceID
	utils.ResponseSuccess(c, instance)
}
