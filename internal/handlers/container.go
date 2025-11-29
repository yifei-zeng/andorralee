package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StartContainerRequest 启动容器请求参数
type StartContainerRequest struct {
	Image   string            `json:"image" binding:"required"` // 镜像名称（如 andorralee/cowrie:v0.1）
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
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
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
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
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
	// 兼容 path 参数 /docker/container/:id
	id := c.Param("id")
	containerID := id
	if containerID == "" {
		containerID = c.Query("container_id")
	}
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "container_id 不能为空")
		return
	}
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
		return
	}
	// 先检查容器信息是否可获取
	if _, err := services.GetContainerInfo(containerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器信息失败: "+err.Error())
		return
	}
	logs, err := services.GetContainerLogs(containerID)
	if err != nil {
		utils.ResponseError(c, 500, "获取日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
	return
}

// StartContainerByID 通过容器ID启动已存在的容器
func StartContainerByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
		return
	}
	if err := services.StartExistingContainer(id); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "启动容器失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, "容器启动成功")
}

// StopContainerByID 通过容器ID停止容器
func StopContainerByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
		return
	}
	if err := services.StopContainer(id); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "停止容器失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, "容器停止成功")
}

// RestartContainerByID 通过容器ID重启容器
func RestartContainerByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
		return
	}
	if err := services.RestartContainer(id, 10); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "重启容器失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, "容器重启成功")
}

// ListContainers 列出所有容器
// @Summary 列出所有 Docker 容器
// @Description 获取所有容器的基本信息
// @Tags Docker
// @Produce json
// @Success 200 {object} utils.Response
// @Router /docker/list [get]
func ListContainers(c *gin.Context) {
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
		return
	}
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
	// 兼容 path 参数 /docker/container/:id
	containerID := c.Param("id")
	if containerID == "" {
		containerID = c.Query("container_id")
	}
	if containerID == "" {
		utils.ResponseError(c, 400, "container_id 不能为空")
		return
	}
	if !services.IsDockerAvailable() {
		utils.ResponseError(c, 503, "Docker服务不可用")
		return
	}
	info, err := services.GetContainerInfo(containerID)
	if err != nil {
		utils.ResponseError(c, 500, "获取容器信息失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, info)
}
