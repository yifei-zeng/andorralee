package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PullCowrieLogsRequest 拉取Cowrie日志请求参数
type PullCowrieLogsRequest struct {
	ContainerID string `json:"container_id" binding:"required"` // 容器ID
}

// CowrieLogQueryRequest 查询Cowrie日志请求参数
type CowrieLogQueryRequest struct {
	ContainerID  string `form:"container_id"`  // 容器ID
	SourceIP     string `form:"source_ip"`     // 源IP
	Protocol     string `form:"protocol"`      // 协议
	Command      string `form:"command"`       // 命令
	Username     string `form:"username"`      // 用户名
	CommandFound string `form:"command_found"` // 命令是否被识别
	StartTime    string `form:"start_time"`    // 开始时间
	EndTime      string `form:"end_time"`      // 结束时间
	Limit        int    `form:"limit"`         // 限制数量
}

// PullCowrieLogs 拉取Cowrie蜜罐日志
// @Summary 拉取Cowrie蜜罐日志
// @Description 从指定容器中拉取Cowrie蜜罐日志并保存到数据库
// @Tags Cowrie蜜罐日志
// @Accept json
// @Produce json
// @Param payload body PullCowrieLogsRequest true "拉取参数"
// @Success 200 {object} utils.Response
// @Router /cowrie/pull-logs [post]
func PullCowrieLogs(c *gin.Context) {
	var req PullCowrieLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := service.PullCowrieLogs(req.ContainerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "拉取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "Cowrie蜜罐日志拉取成功")
}

// GetAllCowrieLogs 获取所有Cowrie蜜罐日志
// @Summary 获取所有Cowrie蜜罐日志
// @Description 获取所有Cowrie蜜罐日志记录
// @Tags Cowrie蜜罐日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /cowrie/logs [get]
func GetAllCowrieLogs(c *gin.Context) {
	service, err := services.NewCowrieService()
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

// GetCowrieLogByID 根据ID获取Cowrie蜜罐日志
// @Summary 根据ID获取Cowrie蜜罐日志
// @Description 根据ID获取指定的Cowrie蜜罐日志记录
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/{id} [get]
func GetCowrieLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewCowrieService()
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

// GetCowrieLogsByContainer 根据容器ID获取Cowrie日志
// @Summary 根据容器ID获取Cowrie日志
// @Description 获取指定容器的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/container/{container_id} [get]
func GetCowrieLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	service, err := services.NewCowrieService()
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

// GetCowrieLogsBySourceIP 根据源IP获取Cowrie日志
// @Summary 根据源IP获取Cowrie日志
// @Description 获取指定源IP的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param source_ip path string true "源IP地址"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/source-ip/{source_ip} [get]
func GetCowrieLogsBySourceIP(c *gin.Context) {
	sourceIP := c.Param("source_ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "源IP不能为空")
		return
	}

	service, err := services.NewCowrieService()
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

// GetCowrieLogsByProtocol 根据协议获取Cowrie日志
// @Summary 根据协议获取Cowrie日志
// @Description 获取指定协议的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param protocol path string true "协议类型"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/protocol/{protocol} [get]
func GetCowrieLogsByProtocol(c *gin.Context) {
	protocol := c.Param("protocol")
	if protocol == "" {
		utils.ResponseError(c, http.StatusBadRequest, "协议不能为空")
		return
	}

	service, err := services.NewCowrieService()
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

// GetCowrieLogsByCommand 根据命令获取Cowrie日志
// @Summary 根据命令获取Cowrie日志
// @Description 获取包含指定命令的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param command path string true "命令内容"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/command/{command} [get]
func GetCowrieLogsByCommand(c *gin.Context) {
	command := c.Param("command")
	if command == "" {
		utils.ResponseError(c, http.StatusBadRequest, "命令不能为空")
		return
	}

	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByCommand(command)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetCowrieLogsByUsername 根据用户名获取Cowrie日志
// @Summary 根据用户名获取Cowrie日志
// @Description 获取指定用户名的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param username path string true "用户名"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/username/{username} [get]
func GetCowrieLogsByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.ResponseError(c, http.StatusBadRequest, "用户名不能为空")
		return
	}

	service, err := services.NewCowrieService()
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

// GetCowrieLogsByCommandFound 根据命令识别状态获取Cowrie日志
// @Summary 根据命令识别状态获取Cowrie日志
// @Description 获取命令识别状态的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param found path bool true "命令是否被识别"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/command-found/{found} [get]
func GetCowrieLogsByCommandFound(c *gin.Context) {
	foundStr := c.Param("found")
	found, err := strconv.ParseBool(foundStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的布尔值: "+err.Error())
		return
	}

	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := service.GetLogsByCommandFound(found)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetCowrieLogsByTimeRange 根据时间范围获取Cowrie日志
// @Summary 根据时间范围获取Cowrie日志
// @Description 获取指定时间范围内的Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param start_time query string true "开始时间(RFC3339格式)"
// @Param end_time query string true "结束时间(RFC3339格式)"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/time-range [get]
func GetCowrieLogsByTimeRange(c *gin.Context) {
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

	service, err := services.NewCowrieService()
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

// GetCowrieStatistics 获取Cowrie统计信息
// @Summary 获取Cowrie统计信息
// @Description 获取Cowrie蜜罐日志的统计信息
// @Tags Cowrie蜜罐日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /cowrie/statistics [get]
func GetCowrieStatistics(c *gin.Context) {
	service, err := services.NewCowrieService()
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

// GetCowrieAttackerBehavior 获取攻击者行为统计信息
// @Summary 获取攻击者行为统计信息
// @Description 获取攻击者行为的详细统计信息
// @Tags Cowrie蜜罐日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /cowrie/attacker-behavior [get]
func GetCowrieAttackerBehavior(c *gin.Context) {
	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	behavior, err := service.GetAttackerBehavior()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取攻击者行为统计失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, behavior)
}

// GetCowrieTopAttackers 获取前N个攻击者
// @Summary 获取前N个攻击者
// @Description 获取攻击活动最频繁的前N个攻击者
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /cowrie/top-attackers [get]
func GetCowrieTopAttackers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewCowrieService()
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

// GetCowrieTopCommands 获取最常用的命令
// @Summary 获取最常用的命令
// @Description 获取使用频率最高的前N个命令
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /cowrie/top-commands [get]
func GetCowrieTopCommands(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	commands, err := service.GetTopCommands(limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取常用命令失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, commands)
}

// GetCowrieTopUsernames 获取最常用的用户名
// @Summary 获取最常用的用户名
// @Description 获取使用频率最高的前N个用户名
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /cowrie/top-usernames [get]
func GetCowrieTopUsernames(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewCowrieService()
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

// GetCowrieTopPasswords 获取最常用的密码
// @Summary 获取最常用的密码
// @Description 获取使用频率最高的前N个密码
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /cowrie/top-passwords [get]
func GetCowrieTopPasswords(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewCowrieService()
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

// GetCowrieTopFingerprints 获取最常用的指纹
// @Summary 获取最常用的指纹
// @Description 获取使用频率最高的前N个客户端指纹
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} utils.Response
// @Router /cowrie/top-fingerprints [get]
func GetCowrieTopFingerprints(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	fingerprints, err := service.GetTopFingerprints(limit)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取常用指纹失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, fingerprints)
}

// DeleteCowrieLogsByContainer 删除指定容器的Cowrie日志
// @Summary 删除指定容器的Cowrie日志
// @Description 删除指定容器的所有Cowrie蜜罐日志
// @Tags Cowrie蜜罐日志
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /cowrie/logs/container/{container_id} [delete]
func DeleteCowrieLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	service, err := services.NewCowrieService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := service.DeleteLogsByContainer(containerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "容器Cowrie日志删除成功")
}
