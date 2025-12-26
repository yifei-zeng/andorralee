package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PullDionaeaLogsRequest 请求体
type PullDionaeaLogsRequest struct {
	ContainerID string `json:"container_id" binding:"required"`
}

// PullDionaeaLogs 从容器拉取 Dionaea 日志
func PullDionaeaLogs(c *gin.Context) {
	var req PullDionaeaLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := svc.PullDionaeaLogs(req.ContainerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "拉取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "Dionaea 日志拉取成功")
}

// GetAllDionaeaLogs 获取全部日志
func GetAllDionaeaLogs(c *gin.Context) {
	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := svc.GetAllLogs()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetDionaeaLogByID 根据ID获取日志
func GetDionaeaLogByID(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logItem, err := svc.GetLogByID(id)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "日志不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logItem)
}

// GetDionaeaLogsByContainer 根据容器ID获取日志
func GetDionaeaLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := svc.GetLogsByContainer(containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetDionaeaLogsBySourceIP 根据源IP获取日志
func GetDionaeaLogsBySourceIP(c *gin.Context) {
	sourceIP := c.Param("source_ip")
	if sourceIP == "" {
		utils.ResponseError(c, http.StatusBadRequest, "源IP不能为空")
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := svc.GetLogsBySourceIP(sourceIP)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetDionaeaLogsByProtocol 根据协议获取日志
func GetDionaeaLogsByProtocol(c *gin.Context) {
	protocol := c.Param("protocol")
	if protocol == "" {
		utils.ResponseError(c, http.StatusBadRequest, "协议不能为空")
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := svc.GetLogsByProtocol(protocol)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetDionaeaLogsByTimeRange 按时间范围获取日志
func GetDionaeaLogsByTimeRange(c *gin.Context) {
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")
	if startStr == "" || endStr == "" {
		utils.ResponseError(c, http.StatusBadRequest, "开始时间和结束时间不能为空")
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "开始时间格式错误: "+err.Error())
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "结束时间格式错误: "+err.Error())
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	logs, err := svc.GetLogsByTimeRange(start, end)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// DeleteDionaeaLogsByContainer 删除指定容器的日志
func DeleteDionaeaLogsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	svc, err := services.NewDionaeaService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if err := svc.DeleteLogsByContainer(containerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "删除成功")
}

// parseUintParam 将字符串参数解析为 uint
func parseUintParam(val string) (uint, error) {
	v, err := strconv.ParseUint(val, 10, 32)
	return uint(v), err
}
