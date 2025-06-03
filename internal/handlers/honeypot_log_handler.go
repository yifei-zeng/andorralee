package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllHoneypotLogs 获取所有蜜罐日志
// @Summary 获取所有蜜罐日志
// @Description 获取所有蜜罐运行日志
// @Tags 蜜罐管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /honeypot/logs [get]
func GetAllHoneypotLogs(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLHoneypotLogRepo(config.MySQLDB)
	logs, err := repo.List()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetHoneypotLogByID 根据ID获取蜜罐日志
// @Summary 根据ID获取蜜罐日志
// @Description 根据ID获取蜜罐日志详情
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/logs/{id} [get]
func GetHoneypotLogByID(c *gin.Context) {
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

	repo := repositories.NewMySQLHoneypotLogRepo(config.MySQLDB)
	log, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, log)
}

// GetLogsByInstanceID 根据实例ID获取蜜罐日志
// @Summary 根据实例ID获取蜜罐日志
// @Description 获取指定蜜罐实例的所有日志
// @Tags 蜜罐管理
// @Produce json
// @Param id path int true "实例ID"
// @Success 200 {object} utils.Response
// @Router /honeypot/logs/instance/{id} [get]
func GetLogsByInstanceID(c *gin.Context) {
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

	repo := repositories.NewMySQLHoneypotLogRepo(config.MySQLDB)
	logs, err := repo.GetByInstanceID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}
