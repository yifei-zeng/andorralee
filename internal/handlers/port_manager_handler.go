package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AllocatePortRequest 分配端口请求
type AllocatePortRequest struct {
	ContainerID string `json:"container_id" binding:"required"`
	ServiceType string `json:"service_type"`
	Description string `json:"description"`
}

// AllocateSpecificPortRequest 分配指定端口请求
type AllocateSpecificPortRequest struct {
	Port        int    `json:"port" binding:"required"`
	ContainerID string `json:"container_id" binding:"required"`
	ServiceType string `json:"service_type"`
	Description string `json:"description"`
}

// AutoAllocatePortMappingRequest 自动分配端口映射请求
type AutoAllocatePortMappingRequest struct {
	ContainerID  string            `json:"container_id" binding:"required"`
	PortMappings map[string]string `json:"port_mappings" binding:"required"`
}

// GetAvailablePortsRequest 获取可用端口请求
type GetAvailablePortsRequest struct {
	Start int `json:"start" binding:"required"`
	End   int `json:"end" binding:"required"`
	Limit int `json:"limit"`
}

// AllocatePort 分配端口
// @Summary 自动分配端口
// @Description 为容器自动分配一个可用端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param payload body AllocatePortRequest true "分配端口请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/allocate [post]
func AllocatePort(c *gin.Context) {
	var req AllocatePortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	pm := services.GetPortManager()
	port, err := pm.AllocatePort(req.ContainerID, req.ServiceType, req.Description)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "分配端口失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"port":         port,
		"container_id": req.ContainerID,
		"service_type": req.ServiceType,
		"description":  req.Description,
	})
}

// AllocateSpecificPort 分配指定端口
// @Summary 分配指定端口
// @Description 为容器分配指定的端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param payload body AllocateSpecificPortRequest true "分配指定端口请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/allocate-specific [post]
func AllocateSpecificPort(c *gin.Context) {
	var req AllocateSpecificPortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	pm := services.GetPortManager()
	err := pm.AllocateSpecificPort(req.Port, req.ContainerID, req.ServiceType, req.Description)
	if err != nil {
		utils.ResponseError(c, http.StatusConflict, "分配指定端口失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"port":         req.Port,
		"container_id": req.ContainerID,
		"service_type": req.ServiceType,
		"description":  req.Description,
	})
}

// ReleasePort 释放端口
// @Summary 释放端口
// @Description 释放指定的端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param port path int true "端口号"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/{port}/release [delete]
func ReleasePort(c *gin.Context) {
	portStr := c.Param("port")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的端口号: "+portStr)
		return
	}

	pm := services.GetPortManager()
	err = pm.ReleasePort(port)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "释放端口失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"port":    port,
		"message": "端口释放成功",
	})
}

// ReleasePortsByContainer 释放容器的所有端口
// @Summary 释放容器的所有端口
// @Description 释放指定容器的所有端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/container/{container_id}/release [delete]
func ReleasePortsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	pm := services.GetPortManager()
	
	// 获取容器的端口列表
	ports := pm.GetPortsByContainer(containerID)
	
	err := pm.ReleasePortsByContainer(containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "释放容器端口失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"container_id":    containerID,
		"released_ports":  ports,
		"message":         "容器端口释放成功",
	})
}

// GetPortAllocation 获取端口分配信息
// @Summary 获取端口分配信息
// @Description 获取指定端口的分配信息
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param port path int true "端口号"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/{port} [get]
func GetPortAllocation(c *gin.Context) {
	portStr := c.Param("port")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的端口号: "+portStr)
		return
	}

	pm := services.GetPortManager()
	allocation, exists := pm.GetPortAllocation(port)
	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "端口未被分配")
		return
	}

	utils.ResponseSuccess(c, allocation)
}

// GetAllocatedPorts 获取所有已分配的端口
// @Summary 获取所有已分配的端口
// @Description 获取系统中所有已分配的端口列表
// @Tags 端口管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/allocated [get]
func GetAllocatedPorts(c *gin.Context) {
	pm := services.GetPortManager()
	allocations := pm.GetAllocatedPorts()

	utils.ResponseSuccess(c, gin.H{
		"total":       len(allocations),
		"allocations": allocations,
	})
}

// GetPortsByContainer 获取容器分配的端口
// @Summary 获取容器分配的端口
// @Description 获取指定容器分配的所有端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/container/{container_id} [get]
func GetPortsByContainer(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	pm := services.GetPortManager()
	allocations := pm.GetPortsByContainer(containerID)

	utils.ResponseSuccess(c, gin.H{
		"container_id": containerID,
		"total":        len(allocations),
		"allocations":  allocations,
	})
}

// GetAvailablePorts 获取可用端口
// @Summary 获取可用端口
// @Description 获取指定范围内的可用端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param payload body GetAvailablePortsRequest true "获取可用端口请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/available [post]
func GetAvailablePorts(c *gin.Context) {
	var req GetAvailablePortsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认限制
	if req.Limit <= 0 {
		req.Limit = 100
	}

	pm := services.GetPortManager()
	
	// 验证端口范围
	if err := pm.ValidatePortRange(req.Start, req.End); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "端口范围无效: "+err.Error())
		return
	}

	availablePorts := pm.GetAvailablePortsInRange(req.Start, req.End, req.Limit)

	utils.ResponseSuccess(c, gin.H{
		"start":           req.Start,
		"end":             req.End,
		"limit":           req.Limit,
		"available_count": len(availablePorts),
		"available_ports": availablePorts,
	})
}

// GetNextAvailablePort 获取下一个可用端口
// @Summary 获取下一个可用端口
// @Description 从指定端口开始获取下一个可用端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param start_port query int true "起始端口"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/next-available [get]
func GetNextAvailablePort(c *gin.Context) {
	startPortStr := c.Query("start_port")
	if startPortStr == "" {
		utils.ResponseError(c, http.StatusBadRequest, "起始端口参数不能为空")
		return
	}

	startPort, err := strconv.Atoi(startPortStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的起始端口: "+startPortStr)
		return
	}

	pm := services.GetPortManager()
	nextPort, err := pm.GetNextAvailablePort(startPort)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"start_port": startPort,
		"next_port":  nextPort,
	})
}

// GetPortStatistics 获取端口统计信息
// @Summary 获取端口统计信息
// @Description 获取端口使用统计信息
// @Tags 端口管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/statistics [get]
func GetPortStatistics(c *gin.Context) {
	pm := services.GetPortManager()
	stats := pm.GetPortStatistics()

	utils.ResponseSuccess(c, stats)
}

// AutoAllocatePortMapping 自动分配端口映射
// @Summary 自动分配端口映射
// @Description 为容器自动分配端口映射，支持自动分配和指定端口
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param payload body AutoAllocatePortMappingRequest true "自动分配端口映射请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/auto-allocate-mapping [post]
func AutoAllocatePortMapping(c *gin.Context) {
	var req AutoAllocatePortMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	pm := services.GetPortManager()
	result, err := pm.AutoAllocatePortMapping(req.ContainerID, req.PortMappings)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "自动分配端口映射失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"container_id":   req.ContainerID,
		"port_mappings":  result,
		"message":        "端口映射分配成功",
	})
}

// CheckPortAvailability 检查端口可用性
// @Summary 检查端口可用性
// @Description 检查指定端口是否可用
// @Tags 端口管理
// @Accept json
// @Produce json
// @Param port path int true "端口号"
// @Success 200 {object} utils.Response
// @Router /api/v1/ports/{port}/check [get]
func CheckPortAvailability(c *gin.Context) {
	portStr := c.Param("port")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的端口号: "+portStr)
		return
	}

	pm := services.GetPortManager()

	// 检查端口是否已被分配
	allocation, allocated := pm.GetPortAllocation(port)

	// 检查端口是否被系统占用
	occupied := pm.IsPortOccupied(port)

	utils.ResponseSuccess(c, gin.H{
		"port":       port,
		"available":  !allocated && !occupied,
		"allocated":  allocated,
		"occupied":   occupied,
		"allocation": allocation,
	})
}

// 注意：这个方法需要在PortManager中公开isPortOccupied方法
// 或者在这里重新实现端口占用检查逻辑
