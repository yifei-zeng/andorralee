package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateContainerInstanceRequest 创建容器实例请求
type CreateContainerInstanceRequest struct {
	Name          string            `json:"name" binding:"required"`          // 实例名称
	HoneypotName  string            `json:"honeypot_name" binding:"required"` // 蜜罐名称
	ImageName     string            `json:"image_name" binding:"required"`    // Docker镜像名称
	Protocol      string            `json:"protocol" binding:"required"`      // 协议类型
	InterfaceType string            `json:"interface_type"`                   // 接口类型
	PortMappings  map[string]string `json:"port_mappings"`                    // 端口映射
	Environment   map[string]string `json:"environment"`                      // 环境变量
	Description   string            `json:"description"`                      // 描述
	AutoStart     bool              `json:"auto_start"`                       // 是否自动启动
}

// CreateContainerInstance 创建容器实例
func CreateContainerInstance(c *gin.Context) {
	var req CreateContainerInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查Docker是否可用
	dockerAvailable := config.DockerCli != nil
	if !dockerAvailable {
		fmt.Printf("警告: Docker服务不可用，将创建数据库记录但不会创建实际容器\n")
	}

	var containerID string
	var containerInfo types.ContainerJSON
	containerStatus := "created"
	var containerIP string

	// 2. 生成容器名称
	containerName := fmt.Sprintf("%s-%s", req.HoneypotName, uuid.New().String()[:8])

	// 3. 准备端口映射
	var mainPort int
	for _, hostPort := range req.PortMappings {
		if p, err := strconv.Atoi(hostPort); err == nil {
			mainPort = p
			break
		}
	}

	// 4. 如果Docker可用，创建真实容器
	if dockerAvailable {
		// 检查镜像是否存在，不存在则拉取
		_, _, err := config.DockerCli.ImageInspectWithRaw(context.Background(), req.ImageName)
		if err != nil {
			fmt.Printf("镜像 %s 不存在，正在拉取...\n", req.ImageName)

			pullResp, err := config.DockerCli.ImagePull(context.Background(), req.ImageName, image.PullOptions{})
			if err != nil {
				utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("拉取镜像失败: %v", err))
				return
			}
			defer pullResp.Close()

			// 读取拉取进度
			io.Copy(io.Discard, pullResp)
			fmt.Printf("镜像 %s 拉取完成\n", req.ImageName)
		}

		// 准备端口映射
		portBindings := nat.PortMap{}
		exposedPorts := nat.PortSet{}

		for containerPort, hostPort := range req.PortMappings {
			port, err := nat.NewPort("tcp", containerPort)
			if err != nil {
				utils.ResponseError(c, http.StatusBadRequest, fmt.Sprintf("无效的容器端口 %s: %v", containerPort, err))
				return
			}

			exposedPorts[port] = struct{}{}
			portBindings[port] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			}
		}

		// 准备环境变量
		var envVars []string
		for key, value := range req.Environment {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}

		// 创建容器配置
		containerConfig := &container.Config{
			Image:        req.ImageName,
			ExposedPorts: exposedPorts,
			Env:          envVars,
		}

		hostConfig := &container.HostConfig{
			PortBindings: portBindings,
			RestartPolicy: container.RestartPolicy{
				Name: "unless-stopped",
			},
		}

		networkConfig := &network.NetworkingConfig{}

		// 创建容器
		resp, err := config.DockerCli.ContainerCreate(
			context.Background(),
			containerConfig,
			hostConfig,
			networkConfig,
			nil,
			containerName,
		)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("创建容器失败: %v", err))
			return
		}

		containerID = resp.ID

		// 如果设置了自动启动，则启动容器
		if req.AutoStart {
			if err := config.DockerCli.ContainerStart(context.Background(), containerID, container.StartOptions{}); err != nil {
				config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
				utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("启动容器失败: %v", err))
				return
			}
			containerStatus = "running"
			fmt.Printf("容器 %s 启动成功\n", containerName)
		}

		// 获取容器信息
		containerInfo, err = config.DockerCli.ContainerInspect(context.Background(), containerID)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取容器信息失败: %v", err))
			return
		}

		// 解析容器IP
		if containerInfo.NetworkSettings != nil && containerInfo.NetworkSettings.IPAddress != "" {
			containerIP = containerInfo.NetworkSettings.IPAddress
		}
	} else {
		// Docker不可用时，生成模拟的容器ID
		containerID = fmt.Sprintf("mock-%s", uuid.New().String())
		containerStatus = "mock-created"
		fmt.Printf("模拟创建容器 %s (Docker不可用)\n", containerName)
	}

	// 5. 序列化配置
	portMappingsJSON, _ := json.Marshal(req.PortMappings)
	environmentJSON, _ := json.Marshal(req.Environment)

	// 6. 创建数据库记录
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		if dockerAvailable && containerID != "" {
			config.DockerCli.ContainerStop(context.Background(), containerID, container.StopOptions{})
			config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
		}
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("创建数据库服务失败: %v", err))
		return
	}

	// 获取镜像ID
	var imageID string
	if dockerAvailable && containerInfo.Image != "" {
		imageID = containerInfo.Image
	} else {
		// 生成一个短的模拟镜像ID（不超过64字符）
		imageID = fmt.Sprintf("mock-%s", uuid.New().String()[:8])
	}

	instance := &repositories.HoneypotInstance{
		Name:          req.Name,
		HoneypotName:  req.HoneypotName,
		ContainerName: containerName,
		ContainerID:   containerID,
		IP:            "0.0.0.0",
		HoneypotIP:    containerIP,
		Port:          mainPort,
		Protocol:      req.Protocol,
		InterfaceType: req.InterfaceType,
		Status:        containerStatus,
		ImageName:     req.ImageName,
		ImageID:       imageID,
		PortMappings:  string(portMappingsJSON),
		Environment:   string(environmentJSON),
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Description:   req.Description,
	}

	if err := service.CreateInstance(instance); err != nil {
		if dockerAvailable && containerID != "" {
			config.DockerCli.ContainerStop(context.Background(), containerID, container.StopOptions{})
			config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
		}
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("保存实例记录失败: %v", err))
		return
	}

	// 7. 返回创建结果
	result := map[string]interface{}{
		"id":               instance.ID,
		"name":             instance.Name,
		"honeypot_name":    instance.HoneypotName,
		"container_name":   instance.ContainerName,
		"container_id":     instance.ContainerID,
		"ip":               instance.IP,
		"honeypot_ip":      instance.HoneypotIP,
		"port":             instance.Port,
		"protocol":         instance.Protocol,
		"interface_type":   instance.InterfaceType,
		"status":           instance.Status,
		"image_name":       instance.ImageName,
		"image_id":         instance.ImageID,
		"port_mappings":    req.PortMappings,
		"environment":      req.Environment,
		"create_time":      instance.CreateTime,
		"description":      instance.Description,
		"docker_available": dockerAvailable,
	}

	utils.ResponseSuccess(c, result)
}

// GetAllContainerInstances 获取所有容器实例
func GetAllContainerInstances(c *gin.Context) {
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	instances, err := service.GetAllInstances()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器实例失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instances)
}

// GetContainerInstanceByID 根据ID获取容器实例
func GetContainerInstanceByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, instance)
}

// StartContainerInstance 启动容器实例
func StartContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	// 获取实例信息
	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在: "+err.Error())
		return
	}

	// 检查Docker是否可用
	if config.DockerCli == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Docker服务不可用")
		return
	}

	// 启动Docker容器
	if instance.ContainerID != "" {
		if err := config.DockerCli.ContainerStart(context.Background(), instance.ContainerID, container.StartOptions{}); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("启动容器失败: %v", err))
			return
		}

		// 更新数据库状态
		if err := service.UpdateInstanceStatus(uint(id), "running"); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("更新状态失败: %v", err))
			return
		}

		fmt.Printf("容器实例 %s 启动成功\n", instance.ContainerName)
	}

	utils.ResponseSuccess(c, "容器实例启动成功")
}

// StopContainerInstance 停止容器实例
func StopContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	// 获取实例信息
	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在: "+err.Error())
		return
	}

	// 检查Docker是否可用
	if config.DockerCli == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Docker服务不可用")
		return
	}

	// 停止Docker容器
	if instance.ContainerID != "" {
		timeout := 30 // 30秒超时
		if err := config.DockerCli.ContainerStop(context.Background(), instance.ContainerID, container.StopOptions{
			Timeout: &timeout,
		}); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("停止容器失败: %v", err))
			return
		}

		// 更新数据库状态
		if err := service.UpdateInstanceStatus(uint(id), "stopped"); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("更新状态失败: %v", err))
			return
		}

		fmt.Printf("容器实例 %s 停止成功\n", instance.ContainerName)
	}

	utils.ResponseSuccess(c, "容器实例停止成功")
}

// DeleteContainerInstance 删除容器实例
func DeleteContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	// 获取实例信息
	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在: "+err.Error())
		return
	}

	// 检查Docker是否可用并删除容器
	if config.DockerCli != nil && instance.ContainerID != "" {
		// 先停止容器
		timeout := 10
		config.DockerCli.ContainerStop(context.Background(), instance.ContainerID, container.StopOptions{
			Timeout: &timeout,
		})

		// 删除容器
		if err := config.DockerCli.ContainerRemove(context.Background(), instance.ContainerID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			// 即使删除容器失败，也继续删除数据库记录
			fmt.Printf("删除容器失败: %v\n", err)
		} else {
			fmt.Printf("容器 %s 删除成功\n", instance.ContainerName)
		}
	}

	// 删除数据库记录
	if err := service.DeleteInstance(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("删除实例记录失败: %v", err))
		return
	}

	utils.ResponseSuccess(c, "容器实例删除成功")
}

// RestartContainerInstance 重启容器实例
func RestartContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	// 获取实例信息
	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在: "+err.Error())
		return
	}

	// 检查Docker是否可用
	if config.DockerCli == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Docker服务不可用")
		return
	}

	// 重启Docker容器
	if instance.ContainerID != "" {
		timeout := 30 // 30秒超时
		if err := config.DockerCli.ContainerRestart(context.Background(), instance.ContainerID, container.StopOptions{
			Timeout: &timeout,
		}); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("重启容器失败: %v", err))
			return
		}

		// 更新数据库状态
		if err := service.UpdateInstanceStatus(uint(id), "running"); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("更新状态失败: %v", err))
			return
		}

		fmt.Printf("容器实例 %s 重启成功\n", instance.ContainerName)
	}

	utils.ResponseSuccess(c, "容器实例重启成功")
}

// GetContainerInstanceStatus 获取容器实例状态
func GetContainerInstanceStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	instance, err := service.GetInstanceByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器实例失败: "+err.Error())
		return
	}

	// 如果有容器ID，尝试获取实时状态
	if config.DockerCli != nil && instance.ContainerID != "" {
		containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
		if err == nil {
			// 更新状态到数据库
			realStatus := containerInfo.State.Status
			if realStatus != instance.Status {
				service.UpdateInstanceStatus(uint(id), realStatus)
				instance.Status = realStatus
			}
		}
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"id":     id,
		"status": instance.Status,
	})
}

// GetContainerInstancesByStatus 根据状态获取容器实例
func GetContainerInstancesByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		utils.ResponseError(c, http.StatusBadRequest, "状态参数不能为空")
		return
	}

	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	// 获取所有实例然后过滤
	allInstances, err := service.GetAllInstances()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器实例失败: "+err.Error())
		return
	}

	var instances []repositories.HoneypotInstance
	for _, instance := range allInstances {
		if instance.Status == status {
			instances = append(instances, instance)
		}
	}

	utils.ResponseSuccess(c, instances)
}

// SyncAllContainerInstancesStatus 同步所有容器实例状态
func SyncAllContainerInstancesStatus(c *gin.Context) {
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if config.DockerCli == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Docker服务不可用")
		return
	}

	// 获取所有实例
	instances, err := service.GetAllInstances()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器实例失败: "+err.Error())
		return
	}

	syncCount := 0
	for _, instance := range instances {
		if instance.ContainerID != "" {
			// 获取容器实时状态
			containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
			if err == nil {
				realStatus := containerInfo.State.Status
				if realStatus != instance.Status {
					// 更新状态
					if err := service.UpdateInstanceStatus(instance.ID, realStatus); err == nil {
						syncCount++
					}
				}
			}
		}
	}

	utils.ResponseSuccess(c, fmt.Sprintf("同步完成，更新了 %d 个容器实例状态", syncCount))
}

// DeployImageToContainerRequest 将镜像部署到容器实例请求
type DeployImageToContainerRequest struct {
	ImageName     string            `json:"image_name" binding:"required"`     // Docker镜像名称
	ContainerName string            `json:"container_name" binding:"required"` // 容器名称
	PortMappings  map[string]string `json:"port_mappings"`                     // 端口映射
	Environment   map[string]string `json:"environment"`                       // 环境变量
	AutoStart     bool              `json:"auto_start"`                        // 是否自动启动
}

// DeployImageToContainer 将指定镜像部署到新的容器实例
func DeployImageToContainer(c *gin.Context) {
	var req DeployImageToContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查Docker是否可用
	if config.DockerCli == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "Docker服务不可用，无法部署镜像")
		return
	}

	// 1. 检查镜像是否存在
	imageInfo, _, err := config.DockerCli.ImageInspectWithRaw(context.Background(), req.ImageName)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Sprintf("镜像 %s 不存在，请先拉取镜像", req.ImageName))
		return
	}

	// 2. 准备端口映射
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	var mainPort int

	for containerPort, hostPort := range req.PortMappings {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			utils.ResponseError(c, http.StatusBadRequest, fmt.Sprintf("无效的容器端口 %s: %v", containerPort, err))
			return
		}

		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}

		if mainPort == 0 {
			if p, err := strconv.Atoi(hostPort); err == nil {
				mainPort = p
			}
		}
	}

	// 3. 准备环境变量
	var envVars []string
	for key, value := range req.Environment {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	// 4. 创建容器配置
	containerConfig := &container.Config{
		Image:        req.ImageName,
		ExposedPorts: exposedPorts,
		Env:          envVars,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkConfig := &network.NetworkingConfig{}

	// 5. 创建容器
	resp, err := config.DockerCli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		req.ContainerName,
	)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("创建容器失败: %v", err))
		return
	}

	containerID := resp.ID
	containerStatus := "created"

	// 6. 如果设置了自动启动，则启动容器
	if req.AutoStart {
		if err := config.DockerCli.ContainerStart(context.Background(), containerID, container.StartOptions{}); err != nil {
			config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("启动容器失败: %v", err))
			return
		}
		containerStatus = "running"
		fmt.Printf("容器 %s 启动成功\n", req.ContainerName)
	}

	// 7. 获取容器信息
	containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("获取容器信息失败: %v", err))
		return
	}

	// 8. 解析容器IP
	containerIP := ""
	if containerInfo.NetworkSettings != nil && containerInfo.NetworkSettings.IPAddress != "" {
		containerIP = containerInfo.NetworkSettings.IPAddress
	}

	// 9. 序列化配置
	portMappingsJSON, _ := json.Marshal(req.PortMappings)
	environmentJSON, _ := json.Marshal(req.Environment)

	// 10. 创建数据库记录
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		config.DockerCli.ContainerStop(context.Background(), containerID, container.StopOptions{})
		config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("创建数据库服务失败: %v", err))
		return
	}

	instance := &repositories.HoneypotInstance{
		Name:          req.ContainerName,
		HoneypotName:  req.ContainerName,
		ContainerName: req.ContainerName,
		ContainerID:   containerID,
		IP:            "0.0.0.0",
		HoneypotIP:    containerIP,
		Port:          mainPort,
		Protocol:      "auto-detected",
		InterfaceType: "docker",
		Status:        containerStatus,
		ImageName:     req.ImageName,
		ImageID:       imageInfo.ID,
		PortMappings:  string(portMappingsJSON),
		Environment:   string(environmentJSON),
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Description:   fmt.Sprintf("从镜像 %s 部署的容器实例", req.ImageName),
	}

	if err := service.CreateInstance(instance); err != nil {
		config.DockerCli.ContainerStop(context.Background(), containerID, container.StopOptions{})
		config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("保存实例记录失败: %v", err))
		return
	}

	// 11. 返回部署结果
	result := map[string]interface{}{
		"id":             instance.ID,
		"name":           instance.Name,
		"container_name": instance.ContainerName,
		"container_id":   instance.ContainerID,
		"ip":             instance.IP,
		"honeypot_ip":    instance.HoneypotIP,
		"port":           instance.Port,
		"status":         instance.Status,
		"image_name":     instance.ImageName,
		"image_id":       instance.ImageID,
		"port_mappings":  req.PortMappings,
		"environment":    req.Environment,
		"create_time":    instance.CreateTime,
		"message":        fmt.Sprintf("成功将镜像 %s 部署到容器 %s", req.ImageName, req.ContainerName),
	}

	utils.ResponseSuccess(c, result)
}
