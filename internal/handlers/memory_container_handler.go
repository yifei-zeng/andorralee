package handlers

import (
	"andorralee/internal/config"
	"andorralee/pkg/utils"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MemoryContainerInstance 内存中的容器实例
type MemoryContainerInstance struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	HoneypotName  string            `json:"honeypot_name"`
	ContainerName string            `json:"container_name"`
	ContainerID   string            `json:"container_id"`
	IP            string            `json:"ip"`
	HoneypotIP    string            `json:"honeypot_ip"`
	Port          int               `json:"port"`
	Protocol      string            `json:"protocol"`
	InterfaceType string            `json:"interface_type"`
	Status        string            `json:"status"`
	ImageName     string            `json:"image_name"`
	ImageID       string            `json:"image_id"`
	PortMappings  map[string]string `json:"port_mappings"`
	Environment   map[string]string `json:"environment"`
	CreateTime    time.Time         `json:"create_time"`
	UpdateTime    time.Time         `json:"update_time"`
	Description   string            `json:"description"`
}

// 内存存储
var (
	memoryInstances = make(map[uint]*MemoryContainerInstance)
	instanceMutex   = sync.RWMutex{}
	nextID          = uint(1)
)

// CreateMemoryContainerInstance 创建内存容器实例
func CreateMemoryContainerInstance(c *gin.Context) {
	var req CreateContainerInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查Docker是否可用
	dockerAvailable := config.DockerCli != nil
	if !dockerAvailable {
		fmt.Printf("警告: Docker服务不可用，将创建内存记录但不会创建实际容器\n")
	}

	var containerID string
	var containerInfo types.ContainerJSON
	containerStatus := "created"
	var containerIP string

	// 生成容器名称
	containerName := fmt.Sprintf("%s-%s", req.HoneypotName, uuid.New().String()[:8])

	// 准备端口映射
	var mainPort int
	for _, hostPort := range req.PortMappings {
		if p, err := strconv.Atoi(hostPort); err == nil {
			mainPort = p
			break
		}
	}

	// 如果Docker可用，创建真实容器
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

	// 获取镜像ID
	var imageID string
	if dockerAvailable && containerInfo.Image != "" {
		imageID = containerInfo.Image
	} else {
		imageID = fmt.Sprintf("mock-%s", uuid.New().String()[:8])
	}

	// 创建内存记录
	instanceMutex.Lock()
	instance := &MemoryContainerInstance{
		ID:            nextID,
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
		PortMappings:  req.PortMappings,
		Environment:   req.Environment,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Description:   req.Description,
	}
	memoryInstances[nextID] = instance
	nextID++
	instanceMutex.Unlock()

	// 返回创建结果
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
		"storage_type":     "memory",
	}

	utils.ResponseSuccess(c, result)
}

// GetAllMemoryContainerInstances 获取所有内存容器实例
func GetAllMemoryContainerInstances(c *gin.Context) {
	instanceMutex.RLock()
	instances := make([]*MemoryContainerInstance, 0, len(memoryInstances))
	for _, instance := range memoryInstances {
		instances = append(instances, instance)
	}
	instanceMutex.RUnlock()

	utils.ResponseSuccess(c, instances)
}

// GetMemoryContainerInstanceByID 根据ID获取内存容器实例
func GetMemoryContainerInstanceByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	instanceMutex.RLock()
	instance, exists := memoryInstances[uint(id)]
	instanceMutex.RUnlock()

	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在")
		return
	}

	utils.ResponseSuccess(c, instance)
}

// DeleteMemoryContainerInstance 删除内存容器实例
func DeleteMemoryContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	instanceMutex.Lock()
	instance, exists := memoryInstances[uint(id)]
	if !exists {
		instanceMutex.Unlock()
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在")
		return
	}
	delete(memoryInstances, uint(id))
	instanceMutex.Unlock()

	// 如果有真实容器，删除它
	if config.DockerCli != nil && instance.ContainerID != "" && !strings.HasPrefix(instance.ContainerID, "mock") {
		// 先停止容器
		timeout := 10
		config.DockerCli.ContainerStop(context.Background(), instance.ContainerID, container.StopOptions{
			Timeout: &timeout,
		})

		// 删除容器
		if err := config.DockerCli.ContainerRemove(context.Background(), instance.ContainerID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			fmt.Printf("删除容器失败: %v\n", err)
		} else {
			fmt.Printf("容器 %s 删除成功\n", instance.ContainerName)
		}
	}

	utils.ResponseSuccess(c, "容器实例删除成功")
}
