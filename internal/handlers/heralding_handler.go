package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PullHeraldingLogsRequest 拉取heralding日志请求参数
type PullHeraldingLogsRequest struct {
	ContainerID string `json:"container_id" binding:"required"` // 容器ID
}

// HeraldingLogQueryRequest 查询heralding日志请求参数
type HeraldingLogQueryRequest struct {
	ContainerID string `form:"container_id"` // 容器ID
	SourceIP    string `form:"source_ip"`    // 源IP
	Protocol    string `form:"protocol"`     // 协议
	StartTime   string `form:"start_time"`   // 开始时间
	EndTime     string `form:"end_time"`     // 结束时间
	Limit       int    `form:"limit"`        // 限制数量
}

// PullHeraldingLogs 拉取heralding认证日志
// @Summary 拉取heralding认证日志
// @Description 从指定容器中拉取heralding认证日志并保存到数据库
// @Tags Heralding认证日志
// @Accept json
// @Produce json
// @Param payload body PullHeraldingLogsRequest true "拉取参数"
// @Success 200 {object} utils.Response
// @Router /heralding/pull-logs [post]
func PullHeraldingLogs(c *gin.Context) {
	var req PullHeraldingLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := service.PullHeraldingLogs(req.ContainerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "拉取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "heralding认证日志拉取成功")
}

// GetAllHeraldingLogs 获取所有heralding认证日志
// @Summary 获取所有heralding认证日志
// @Description 获取所有heralding认证日志记录
// @Tags Heralding认证日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /heralding/logs [get]
func GetAllHeraldingLogs(c *gin.Context) {
	service, err := services.NewHeraldingService()
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

// GetHeraldingLogByID 根据ID获取heralding认证日志
// @Summary 根据ID获取heralding认证日志
// @Description 根据ID获取指定的heralding认证日志记录
// @Tags Heralding认证日志
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /heralding/logs/{id} [get]
func GetHeraldingLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	log, err := service.GetLogByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "日志不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, log)
}

// GetHeraldingLogsByContainer 根据容器ID获取认证日志
// @Summary 根据容器ID获取认证日志
// @Description 获取指定容器的所有heralding认证日志
// @Tags Heralding认证日志
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /heralding/logs/container/{container_id} [get]
func GetHeraldingLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	service, err := services.NewHeraldingService()
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

// GetHeraldingLogsBySourceIP 根据源IP获取认证日志
// @Summary 根据源IP获取认证日志
// @Description 获取指定源IP的所有heralding认证日志
// @Tags Heralding认证日志
// @Produce json
// @Param source_ip path string true "源IP地址"
// @Success 200 {object} utils.Response
// @Router /heralding/logs/source-ip/{source_ip} [get]
func GetHeraldingLogsBySourceIP(c *gin.Context) {
	sourceIP := c.Param("source_ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "源IP不能为空")
		return
	}

	service, err := services.NewHeraldingService()
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

// GetHeraldingLogsByProtocol 根据协议获取认证日志
// @Summary 根据协议获取认证日志
// @Description 获取指定协议的所有heralding认证日志
// @Tags Heralding认证日志
// @Produce json
// @Param protocol path string true "协议类型"
// @Success 200 {object} utils.Response
// @Router /heralding/logs/protocol/{protocol} [get]
func GetHeraldingLogsByProtocol(c *gin.Context) {
	protocol := c.Param("protocol")
	if protocol == "" {
		utils.ResponseError(c, http.StatusBadRequest, "协议不能为空")
		return
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByProtocol(protocol)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetHeraldingLogsByTimeRange 根据时间范围获取认证日志
// @Summary 根据时间范围获取认证日志
// @Description 获取指定时间范围内的heralding认证日志
// @Tags Heralding认证日志
// @Produce json
// @Param start_time query string true "开始时间(RFC3339格式)"
// @Param end_time query string true "结束时间(RFC3339格式)"
// @Success 200 {object} utils.Response
// @Router /heralding/logs/time-range [get]
func GetHeraldingLogsByTimeRange(c *gin.Context) {
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

	service, err := services.NewHeraldingService()
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

// GetHeraldingStatistics 获取heralding认证统计信息
// @Summary 获取heralding认证统计信息
// @Description 获取heralding认证日志的统计信息
// @Tags Heralding认证日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /heralding/statistics [get]
func GetHeraldingStatistics(c *gin.Context) {
	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	stats, err := service.GetStatistics()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取统计信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, stats)
}

// GetAttackerIPStatistics 获取攻击者IP统计信息
// @Summary 获取攻击者IP统计信息
// @Description 获取攻击者IP的详细统计信息
// @Tags Heralding认证日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /heralding/attacker-statistics [get]
func GetAttackerIPStatistics(c *gin.Context) {
	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	stats, err := service.GetAttackerIPStatistics()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取攻击者统计信息失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, stats)
}

// GetTopAttackers 获取前N个攻击者
// @Summary 获取前N个攻击者
// @Description 获取攻击次数最多的前N个攻击者
// @Tags Heralding认证日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /heralding/top-attackers [get]
func GetTopAttackers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	attackers, err := service.GetTopAttackers(limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取顶级攻击者失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, attackers)
}

// GetTopUsernames 获取最常用的用户名
// @Summary 获取最常用的用户名
// @Description 获取使用频率最高的前N个用户名
// @Tags Heralding认证日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /heralding/top-usernames [get]
func GetTopUsernames(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	usernames, err := service.GetTopUsernames(limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取常用用户名失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, usernames)
}

// GetTopPasswords 获取最常用的密码
// @Summary 获取最常用的密码
// @Description 获取使用频率最高的前N个密码
// @Tags Heralding认证日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /heralding/top-passwords [get]
func GetTopPasswords(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	passwords, err := service.GetTopPasswords(limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取常用密码失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, passwords)
}

// DeleteHeraldingLogsByContainer 删除指定容器的认证日志
// @Summary 删除指定容器的认证日志
// @Description 删除指定容器的所有heralding认证日志
// @Tags Heralding认证日志
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /heralding/logs/container/{container_id} [delete]
func DeleteHeraldingLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	service, err := services.NewHeraldingService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := service.DeleteLogsByContainer(containerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "容器认证日志删除成功")
}
