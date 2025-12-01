package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
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
	deletedIDs      = make([]uint, 0) // 存储已删除的ID，用于重用
)

// getNextAvailableID 获取下一个可用的ID，优先重用已删除的ID
func getNextAvailableID() uint {
	// 如果有已删除的ID，优先重用最小的
	if len(deletedIDs) > 0 {
		// 排序以获取最小的ID
		sort.Slice(deletedIDs, func(i, j int) bool {
			return deletedIDs[i] < deletedIDs[j]
		})

		// 取出最小的ID
		reusedID := deletedIDs[0]
		deletedIDs = deletedIDs[1:]
		return reusedID
	}

	// 没有可重用的ID，使用下一个新ID
	currentID := nextID
	nextID++
	return currentID
}

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

	// 使用端口管理服务处理端口映射
	pm := services.GetPortManager()
	var finalPortMappings map[string]string
	var mainPort int

	// 如果未指定端口映射且镜像是 heralding，使用 heralding 的常见容器端口并标记为 auto
	if len(req.PortMappings) == 0 && strings.Contains(strings.ToLower(req.ImageName), "heralding") {
		req.PortMappings = map[string]string{
			"22":   "auto",
			"23":   "auto",
			"80":   "auto",
			"110":  "auto",
			"143":  "auto",
			"443":  "auto",
			"3306": "auto",
			"3389": "auto",
			"5900": "auto",
			"995":  "auto",
			"993":  "auto",
		}
	}

	if len(req.PortMappings) > 0 {
		normalizedMappings, normalized := normalizeRequestPortMappings(c, req.Description, req.PortMappings)
		if normalized {
			log.Printf("检测到 host-first 端口映射输入，已为 %s 进行格式转换", req.Name)
		}
		if len(normalizedMappings) == 0 {
			utils.ResponseError(c, http.StatusBadRequest, "端口映射格式无效，缺少容器端口")
			return
		}
		// 使用端口管理服务自动分配端口映射
		allocatedMappings, err := pm.AutoAllocatePortMapping(containerName, normalizedMappings)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, "端口分配失败: "+err.Error())
			return
		}
		finalPortMappings = allocatedMappings

		// 获取主端口
		for _, hostPort := range finalPortMappings {
			if p, err := strconv.Atoi(hostPort); err == nil {
				mainPort = p
				break
			}
		}
	} else {
		finalPortMappings = make(map[string]string)
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

		for containerPort, hostPort := range finalPortMappings {
			port, err := nat.NewPort("tcp", containerPort)
			if err != nil {
				// 如果端口分配失败，释放已分配的端口
				pm.ReleasePortsByContainer(containerName)
				utils.ResponseError(c, http.StatusBadRequest, fmt.Sprintf("无效的容器端口 %s: %v", containerPort, err))
				return
			}

			exposedPorts[port] = struct{}{}
			portBindings[port] = []nat.PortBinding{
				{
					HostIP: func() string {
						if req.IP != "" {
							return req.IP
						}
						return "0.0.0.0"
					}(),
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
	availableID := getNextAvailableID()
	instance := &MemoryContainerInstance{
		ID:            availableID,
		Name:          req.Name,
		HoneypotName:  req.HoneypotName,
		ContainerName: containerName,
		ContainerID:   containerID,
		IP: func() string {
			if req.IP != "" {
				return req.IP
			}
			return "0.0.0.0"
		}(),
		HoneypotIP:    containerIP,
		Port:          mainPort,
		Protocol:      req.Protocol,
		InterfaceType: req.InterfaceType,
		Status:        containerStatus,
		ImageName:     req.ImageName,
		ImageID:       imageID,
		PortMappings:  finalPortMappings,
		Environment:   req.Environment,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Description:   req.Description,
	}
	memoryInstances[availableID] = instance
	instanceMutex.Unlock()

	// 返回创建结果
	// 生成友好的端口展示（hostPort protocol）
	portsPretty := make([]string, 0, len(finalPortMappings))
	for _, hostPort := range finalPortMappings {
		portsPretty = append(portsPretty, fmt.Sprintf("%s %s", hostPort, strings.ToLower(req.Protocol)))
	}

	result := map[string]interface{}{
		"id":                      instance.ID,
		"name":                    instance.Name,
		"honeypot_name":           instance.HoneypotName,
		"container_name":          instance.ContainerName,
		"container_id":            instance.ContainerID,
		"ip":                      instance.IP,
		"honeypot_ip":             instance.HoneypotIP,
		"port":                    instance.Port,
		"protocol":                instance.Protocol,
		"interface_type":          instance.InterfaceType,
		"status":                  instance.Status,
		"image_name":              instance.ImageName,
		"image_id":                instance.ImageID,
		"port_mappings":           finalPortMappings,
		"requested_port_mappings": req.PortMappings,
		"ports_pretty":            portsPretty,
		"environment":             req.Environment,
		"create_time":             instance.CreateTime,
		"description":             instance.Description,
		"docker_available":        dockerAvailable,
		"storage_type":            "memory",
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
	// 将删除的ID添加到重用列表
	deletedIDs = append(deletedIDs, uint(id))
	instanceMutex.Unlock()

	// 释放端口
	pm := services.GetPortManager()
	if err := pm.ReleasePortsByContainer(instance.ContainerName); err != nil {
		fmt.Printf("释放端口失败: %v\n", err)
	} else {
		fmt.Printf("容器 %s 的端口已释放\n", instance.ContainerName)
	}

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

// GetContainerIDStatus 获取容器ID使用状态
func GetContainerIDStatus(c *gin.Context) {
	instanceMutex.RLock()
	defer instanceMutex.RUnlock()

	// 统计ID使用情况
	usedIDs := make([]uint, 0)
	for id := range memoryInstances {
		usedIDs = append(usedIDs, id)
	}

	// 排序已使用的ID
	sort.Slice(usedIDs, func(i, j int) bool {
		return usedIDs[i] < usedIDs[j]
	})

	// 排序可重用的ID
	availableIDs := make([]uint, len(deletedIDs))
	copy(availableIDs, deletedIDs)
	sort.Slice(availableIDs, func(i, j int) bool {
		return availableIDs[i] < availableIDs[j]
	})

	result := map[string]interface{}{
		"next_new_id":     nextID,
		"used_ids":        usedIDs,
		"available_ids":   availableIDs,
		"total_instances": len(memoryInstances),
		"reusable_count":  len(deletedIDs),
	}

	utils.ResponseSuccess(c, result)
}

// StartMemoryContainerInstance 启动内存容器实例
func StartMemoryContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	// 优先尝试按数字实例ID处理；如果不是数字，则按容器ID（哈希）处理
	if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
		instanceMutex.RLock()
		instance, exists := memoryInstances[uint(id)]
		instanceMutex.RUnlock()

		if !exists {
			utils.ResponseError(c, http.StatusNotFound, "容器实例不存在")
			return
		}

		// 启动Docker容器
		ctx := context.Background()
		if err := config.DockerCli.ContainerStart(ctx, instance.ContainerID, container.StartOptions{}); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, "启动容器失败: "+err.Error())
			return
		}

		// 更新状态
		instanceMutex.Lock()
		instance.Status = "running"
		instance.UpdateTime = time.Now()
		instanceMutex.Unlock()

		utils.ResponseSuccess(c, "容器启动成功")
		return
	}

	// 非数字：视为 Docker 容器ID
	ctx := context.Background()
	if err := config.DockerCli.ContainerStart(ctx, idStr, container.StartOptions{}); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "启动容器失败: "+err.Error())
		return
	}
	// 若能匹配到内存实例，则同步状态
	instanceMutex.Lock()
	for _, inst := range memoryInstances {
		if inst.ContainerID == idStr {
			inst.Status = "running"
			inst.UpdateTime = time.Now()
			break
		}
	}
	instanceMutex.Unlock()
	utils.ResponseSuccess(c, "容器启动成功")
}

// StopMemoryContainerInstance 停止内存容器实例
func StopMemoryContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	// 数字实例ID优先
	if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
		instanceMutex.RLock()
		instance, exists := memoryInstances[uint(id)]
		instanceMutex.RUnlock()
		if !exists {
			utils.ResponseError(c, http.StatusNotFound, "容器实例不存在")
			return
		}
		ctx := context.Background()
		timeout := 30
		if err := config.DockerCli.ContainerStop(ctx, instance.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, "停止容器失败: "+err.Error())
			return
		}
		instanceMutex.Lock()
		instance.Status = "stopped"
		instance.UpdateTime = time.Now()
		instanceMutex.Unlock()
		utils.ResponseSuccess(c, "容器停止成功")
		return
	}

	// 非数字：视为 Docker 容器ID
	ctx := context.Background()
	timeout := 30
	if err := config.DockerCli.ContainerStop(ctx, idStr, container.StopOptions{Timeout: &timeout}); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "停止容器失败: "+err.Error())
		return
	}
	instanceMutex.Lock()
	for _, inst := range memoryInstances {
		if inst.ContainerID == idStr {
			inst.Status = "stopped"
			inst.UpdateTime = time.Now()
			break
		}
	}
	instanceMutex.Unlock()
	utils.ResponseSuccess(c, "容器停止成功")
}

// RestartMemoryContainerInstance 重启内存容器实例
func RestartMemoryContainerInstance(c *gin.Context) {
	idStr := c.Param("id")
	// 数字实例ID优先
	if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
		instanceMutex.RLock()
		instance, exists := memoryInstances[uint(id)]
		instanceMutex.RUnlock()
		if !exists {
			utils.ResponseError(c, http.StatusNotFound, "容器实例不存在")
			return
		}
		ctx := context.Background()
		timeout := 30
		if err := config.DockerCli.ContainerRestart(ctx, instance.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, "重启容器失败: "+err.Error())
			return
		}
		instanceMutex.Lock()
		instance.Status = "running"
		instance.UpdateTime = time.Now()
		instanceMutex.Unlock()
		utils.ResponseSuccess(c, "容器重启成功")
		return
	}

	// 非数字：视为 Docker 容器ID
	ctx := context.Background()
	timeout := 30
	if err := config.DockerCli.ContainerRestart(ctx, idStr, container.StopOptions{Timeout: &timeout}); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "重启容器失败: "+err.Error())
		return
	}
	instanceMutex.Lock()
	for _, inst := range memoryInstances {
		if inst.ContainerID == idStr {
			inst.Status = "running"
			inst.UpdateTime = time.Now()
			break
		}
	}
	instanceMutex.Unlock()
	utils.ResponseSuccess(c, "容器重启成功")
}

// SyncMemoryContainerStatus 同步内存容器状态
func SyncMemoryContainerStatus(c *gin.Context) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	ctx := context.Background()
	syncCount := 0
	errorCount := 0
	importCount := 0

	// 1. 同步已知容器状态
	for _, instance := range memoryInstances {
		// 获取容器状态
		containerJSON, err := config.DockerCli.ContainerInspect(ctx, instance.ContainerID)
		if err != nil {
			errorCount++
			continue
		}

		// 更新状态
		oldStatus := instance.Status
		if containerJSON.State.Running {
			instance.Status = "running"
		} else if containerJSON.State.Dead {
			instance.Status = "dead"
		} else {
			instance.Status = "stopped"
		}

		if oldStatus != instance.Status {
			instance.UpdateTime = time.Now()
			syncCount++
		}
	}

	// 2. 发现并导入未管理的容器
	knownIDs := make(map[string]bool)
	for _, inst := range memoryInstances {
		if inst.ContainerID != "" {
			knownIDs[inst.ContainerID] = true
		}
	}

	dockerContainers, err := config.DockerCli.ContainerList(ctx, container.ListOptions{All: true})
	if err == nil {
		for _, dc := range dockerContainers {
			if knownIDs[dc.ID] {
				continue // 已存在，跳过
			}

			name := dc.Names[0]
			if strings.HasPrefix(name, "/") {
				name = name[1:]
			}

			info, err := config.DockerCli.ContainerInspect(ctx, dc.ID)
			if err != nil {
				continue
			}

			// 解析端口映射
			var mainPort int
			protocol := "tcp"
			portMappings := make(map[string]string)
			for p, bindings := range info.HostConfig.PortBindings {
				if len(bindings) > 0 {
					portMappings[p.Port()] = bindings[0].HostPort
					if mainPort == 0 {
						mainPort, _ = strconv.Atoi(bindings[0].HostPort)
						switch p.Port() {
						case "22":
							protocol = "ssh"
						case "80":
							protocol = "http"
						case "3306":
							protocol = "mysql"
						}
					}
				}
			}

			// 导入为内存实例
			importedInstance := &MemoryContainerInstance{
				ID:            getNextAvailableID(),
				Name:          name,
				HoneypotName:  name,
				ContainerName: name,
				ContainerID:   dc.ID,
				IP:            "0.0.0.0",
				Port:          mainPort,
				Protocol:      protocol,
				InterfaceType: "imported",
				Status:        dc.State,
				ImageName:     dc.Image,
				ImageID:       dc.ImageID,
				PortMappings:  portMappings,
				Environment:   make(map[string]string),
				CreateTime:    time.Unix(dc.Created, 0),
				UpdateTime:    time.Now(),
				Description:   "Auto-imported from Docker",
			}
			memoryInstances[importedInstance.ID] = importedInstance
			importCount++
		}
	}

	result := map[string]interface{}{
		"synced_count":  syncCount,
		"error_count":   errorCount,
		"import_count":  importCount,
		"total_managed": len(memoryInstances),
	}

	utils.ResponseSuccess(c, result)
}
