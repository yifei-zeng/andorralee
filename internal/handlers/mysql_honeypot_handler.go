package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PullMySQLHoneypotLogsRequest 用于触发日志拉取
type PullMySQLHoneypotLogsRequest struct {
	ContainerID string `json:"container_id" binding:"required"`
}

// PullMySQLHoneypotLogs 拉取MySQL蜜罐日志
func PullMySQLHoneypotLogs(c *gin.Context) {
	var req PullMySQLHoneypotLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := service.PullMySQLHoneypotLogs(req.ContainerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "拉取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "MySQL蜜罐日志拉取成功")
}

// GetAllMySQLHoneypotLogs 获取全部MySQL蜜罐日志
func GetAllMySQLHoneypotLogs(c *gin.Context) {
	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetAllLogs()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogByID 根据ID获取MySQL蜜罐日志
func GetMySQLHoneypotLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logEntry, err := service.GetLogByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "日志不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logEntry)
}

// GetMySQLHoneypotLogsByContainer 根据容器ID获取日志
func GetMySQLHoneypotLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByContainer(containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogsBySourceIP 根据源IP获取日志
func GetMySQLHoneypotLogsBySourceIP(c *gin.Context) {
	sourceIP := c.Param("source_ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "源IP不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsBySourceIP(sourceIP)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogsByUsername 根据用户名获取日志
func GetMySQLHoneypotLogsByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.ResponseError(c, http.StatusBadRequest, "用户名不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByUsername(username)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogsByTimeRange 根据时间范围获取日志
func GetMySQLHoneypotLogsByTimeRange(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	if startTimeStr == "" || endTimeStr == "" {
		utils.ResponseError(c, http.StatusBadRequest, "开始时间和结束时间不能为空")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "开始时间格式错误: "+err.Error())
		return
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "结束时间格式错误: "+err.Error())
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByTimeRange(startTime, endTime)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// DeleteMySQLHoneypotLogsByContainer 删除指定容器的日志
func DeleteMySQLHoneypotLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := service.DeleteLogsByContainer(containerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "容器MySQL蜜罐日志删除成功")
}

// GetMySQLHoneypotQueryStatistics 获取SQL查询统计
func GetMySQLHoneypotQueryStatistics(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	stats, err := service.GetQueryStatistics(limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取统计信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, stats)
}
