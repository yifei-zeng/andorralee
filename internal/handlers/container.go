package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"

	"github.com/gin-gonic/gin"
)

// StartContainerRequest 启动容器请求参数
type StartContainerRequest struct {
	Image   string            `json:"image" binding:"required"` // 镜像名称（如 andorralee/dm8:v0.1）
	Name    string            `json:"name"`                     // 容器名称
	PortMap map[string]string `json:"port_map"`                 // 端口映射（如 {"80/tcp": "8080"}）
	EnvVars map[string]string `json:"env_vars"`                 // 环境变量
}

// StartContainer 启动容器
// @Summary 启动 Docker 容器
// @Description 根据配置启动容器
// @Tags Docker
// @Accept json
// @Produce json
// @Param   payload  body   StartContainerRequest  true  "容器配置"
// @Success 200 {object} utils.Response
// @Router /docker/start [post]
func StartContainer(c *gin.Context) {
	var req StartContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	containerID, err := services.StartContainer(req.Image, req.Name, req.PortMap, req.EnvVars)
	if err != nil {
		utils.ResponseError(c, 500, "启动容器失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{"container_id": containerID})
}

// StopContainer 停止容器
// @Summary 停止 Docker 容器
// @Description 根据容器 ID 停止容器
// @Tags Docker
// @Produce json
// @Param   container_id  query  string  true  "容器 ID"
// @Success 200 {object} utils.Response
// @Router /docker/stop [post]
func StopContainer(c *gin.Context) {
	containerID := c.Query("container_id")
	if containerID == "" {
		utils.ResponseError(c, 400, "container_id 不能为空")
		return
	}

	if err := services.StopContainer(containerID); err != nil {
		utils.ResponseError(c, 500, "停止容器失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "容器已停止")
}

// GetContainerLogs 获取容器日志
// @Summary 获取容器日志
// @Description 实时获取容器标准输出和错误日志
// @Tags Docker
// @Produce json
// @Param   container_id  query  string  true  "容器 ID"
// @Success 200 {object} utils.Response
// @Router /docker/logs [get]
func GetContainerLogs(c *gin.Context) {
	containerID := c.Query("container_id")
	if containerID == "" {
		utils.ResponseError(c, 400, "container_id 不能为空")
		return
	}

	logs, err := services.GetContainerLogs(containerID)
	if err != nil {
		utils.ResponseError(c, 500, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// ListContainers 列出所有容器
// @Summary 列出所有 Docker 容器
// @Description 获取所有容器的基本信息
// @Tags Docker
// @Produce json
// @Success 200 {object} utils.Response
// @Router /docker/list [get]
func ListContainers(c *gin.Context) {
	containers, err := services.ListContainers()
	if err != nil {
		utils.ResponseError(c, 500, "获取容器列表失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, containers)
}

// GetContainerInfo 获取容器详细信息
// @Summary 获取容器详细信息
// @Description 根据容器 ID 获取详细信息
// @Tags Docker
// @Produce json
// @Param   container_id  query  string  true  "容器 ID"
// @Success 200 {object} utils.Response
// @Router /docker/info [get]
func GetContainerInfo(c *gin.Context) {
	containerID := c.Query("container_id")
	if containerID == "" {
		utils.ResponseError(c, 400, "container_id 不能为空")
		return
	}
	info, err := services.GetContainerInfo(containerID)
	if err != nil {
		utils.ResponseError(c, 500, "获取容器信息失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, info)
}
