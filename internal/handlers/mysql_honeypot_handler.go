package handlers

import (
	"andorralee/internal/repositories"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// PullMySQLHoneypotLogsRequest 用于触发日志拉取
type PullMySQLHoneypotLogsRequest struct {
	ContainerID string `json:"container_id" binding:"required"`
}

// MySQLHoneypotSearchRequest 综合搜索请求
type MySQLHoneypotSearchRequest struct {
	ContainerID   string `json:"container_id"`
	SourceIP      string `json:"source_ip"`
	DestinationIP string `json:"destination_ip"`
	Username      string `json:"username"`
	DatabaseName  string `json:"database_name"`
	QueryKeyword  string `json:"query_keyword"`
	ErrorCode     string `json:"error_code"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Limit         int    `json:"limit"`
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

// GetMySQLHoneypotLogsByDestinationIP 根据目标IP获取日志
func GetMySQLHoneypotLogsByDestinationIP(c *gin.Context) {
	destinationIP := c.Param("destination_ip")
	if destinationIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "目标IP不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByDestinationIP(destinationIP)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogsByDatabaseName 根据数据库名获取日志
func GetMySQLHoneypotLogsByDatabaseName(c *gin.Context) {
	databaseName := c.Param("database_name")
	if databaseName == "" {
		utils.ResponseError(c, http.StatusBadRequest, "数据库名不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByDatabaseName(databaseName)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogsByErrorCode 根据错误码获取日志
func GetMySQLHoneypotLogsByErrorCode(c *gin.Context) {
	errorCode := c.Param("error_code")
	if errorCode == "" {
		utils.ResponseError(c, http.StatusBadRequest, "错误码不能为空")
		return
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByErrorCode(errorCode)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetMySQLHoneypotLogsByQueryKeyword 通过SQL关键字模糊查询
func GetMySQLHoneypotLogsByQueryKeyword(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		utils.ResponseError(c, http.StatusBadRequest, "keyword 不能为空")
		return
	}
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByQueryKeyword(keyword, limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// SearchMySQLHoneypotLogs 综合搜索接口
func SearchMySQLHoneypotLogs(c *gin.Context) {
	var req MySQLHoneypotSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	filter := repositories.MySQLHoneypotSearchFilter{
		ContainerID:   strings.TrimSpace(req.ContainerID),
		SourceIP:      strings.TrimSpace(req.SourceIP),
		DestinationIP: strings.TrimSpace(req.DestinationIP),
		Username:      strings.TrimSpace(req.Username),
		DatabaseName:  strings.TrimSpace(req.DatabaseName),
		QueryKeyword:  strings.TrimSpace(req.QueryKeyword),
		ErrorCode:     strings.TrimSpace(req.ErrorCode),
		Limit:         req.Limit,
	}

	if req.StartTime != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.StartTime))
		if err != nil {
			utils.ResponseError(c, http.StatusBadRequest, "开始时间格式错误: "+err.Error())
			return
		}
		filter.StartTime = &parsed
	}
	if req.EndTime != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndTime))
		if err != nil {
			utils.ResponseError(c, http.StatusBadRequest, "结束时间格式错误: "+err.Error())
			return
		}
		filter.EndTime = &parsed
	}

	service, err := services.NewMySQLHoneypotService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.SearchLogs(filter)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "搜索日志失败: "+err.Error())
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
