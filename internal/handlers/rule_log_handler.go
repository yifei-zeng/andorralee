package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllRuleLogs 获取所有规则日志
// @Summary 获取所有规则日志
// @Description 获取所有安全规则执行日志
// @Tags 安全规则管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /rules/logs [get]
func GetAllRuleLogs(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLRuleLogRepo(config.MySQLDB)
	logs, err := repo.GetAll()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetRuleLogByID 根据ID获取规则日志
// @Summary 根据ID获取规则日志
// @Description 根据ID获取规则日志详情
// @Tags 安全规则管理
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /rules/logs/{id} [get]
func GetRuleLogByID(c *gin.Context) {
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

	repo := repositories.NewMySQLRuleLogRepo(config.MySQLDB)
	log, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, log)
}

// GetLogsByRuleID 根据规则ID获取日志
// @Summary 根据规则ID获取日志
// @Description 获取指定安全规则的所有执行日志
// @Tags 安全规则管理
// @Produce json
// @Param id path int true "规则ID"
// @Success 200 {object} utils.Response
// @Router /rules/logs/rule/{id} [get]
func GetLogsByRuleID(c *gin.Context) {
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

	repo := repositories.NewMySQLRuleLogRepo(config.MySQLDB)
	logs, err := repo.GetByRuleID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}
