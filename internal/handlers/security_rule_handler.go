package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllRules 获取所有安全规则
// @Summary 获取所有安全规则
// @Description 获取所有安全规则信息
// @Tags 安全规则管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /rules [get]
func GetAllRules(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	rules, err := repo.GetAll()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取规则失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, rules)
}

// GetRuleByID 根据ID获取安全规则
// @Summary 根据ID获取安全规则
// @Description 根据ID获取安全规则详情
// @Tags 安全规则管理
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} utils.Response
// @Router /rules/{id} [get]
func GetRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	rule, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取规则失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, rule)
}

// CreateRule 创建安全规则
// @Summary 创建安全规则
// @Description 创建新的安全规则
// @Tags 安全规则管理
// @Accept json
// @Produce json
// @Param rule body repositories.SecurityRule true "规则信息"
// @Success 200 {object} utils.Response
// @Router /rules [post]
func CreateRule(c *gin.Context) {
	var rule repositories.SecurityRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	if err := repo.Create(&rule); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建规则失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, rule)
}

// UpdateRule 更新安全规则
// @Summary 更新安全规则
// @Description 更新现有的安全规则
// @Tags 安全规则管理
// @Accept json
// @Produce json
// @Param id path int true "规则ID"
// @Param rule body repositories.SecurityRule true "规则信息"
// @Success 200 {object} utils.Response
// @Router /rules/{id} [put]
func UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	var rule repositories.SecurityRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	rule.ID = uint(id)

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	if err := repo.Update(&rule); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "更新规则失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, rule)
}

// DeleteRule 删除安全规则
// @Summary 删除安全规则
// @Description 删除指定的安全规则
// @Tags 安全规则管理
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} utils.Response
// @Router /rules/{id} [delete]
func DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	if err := repo.Delete(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除规则失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "删除成功")
}

// EnableRule 启用安全规则
// @Summary 启用安全规则
// @Description 启用指定的安全规则
// @Tags 安全规则管理
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} utils.Response
// @Router /rules/{id}/enable [put]
func EnableRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	if err := repo.UpdateStatus(uint(id), true); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "启用规则失败: "+err.Error())
		return
	}

	rule, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取规则信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, rule)
}

// DisableRule 禁用安全规则
// @Summary 禁用安全规则
// @Description 禁用指定的安全规则
// @Tags 安全规则管理
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} utils.Response
// @Router /rules/{id}/disable [put]
func DisableRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLSecurityRuleRepo(config.MySQLDB)
	if err := repo.UpdateStatus(uint(id), false); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "禁用规则失败: "+err.Error())
		return
	}

	rule, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取规则信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, rule)
}
