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
	"strings"
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
	IP            string            `json:"ip"`                               // 绑定的主机IP（可选，默认 0.0.0.0）
	InterfaceType string            `json:"interface_type"`                   // 接口类型
	PortMappings  map[string]string `json:"port_mappings"`                    // 端口映射（支持"auto"自动分配）
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

	// 3. 使用端口管理服务处理端口映射
	pm := services.GetPortManager()
	var finalPortMappings map[string]string
	var mainPort int

	if len(req.PortMappings) > 0 {
		// 使用端口管理服务自动分配端口映射
		allocatedMappings, err := pm.AutoAllocatePortMapping(containerName, req.PortMappings)
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

		// 准备端口映射（使用分配后的端口）
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

	// 5. 序列化配置（使用分配后的端口映射）
	portMappingsJSON, _ := json.Marshal(finalPortMappings)
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

	// 确保容器ID被正确设置
	if dockerAvailable && containerID == "" {
		fmt.Printf("警告: Docker容器创建成功但容器ID为空\n")
	}

	instance := &repositories.HoneypotInstance{
		Name:          req.Name,
		HoneypotName:  req.HoneypotName,
		ContainerName: containerName,
		ContainerID:   containerID, // 确保容器ID被正确保存
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
		PortMappings:  string(portMappingsJSON),
		Environment:   string(environmentJSON),
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Description:   req.Description,
	}

	// 调试日志
	fmt.Printf("创建容器实例 - 容器ID: %s, 状态: %s, Docker可用: %v\n", containerID, containerStatus, dockerAvailable)

	if err := service.CreateInstance(instance); err != nil {
		// 清理资源
		if dockerAvailable && containerID != "" {
			config.DockerCli.ContainerStop(context.Background(), containerID, container.StopOptions{})
			config.DockerCli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
		}
		// 释放已分配的端口
		pm.ReleasePortsByContainer(containerName)
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("保存实例记录失败: %v", err))
		return
	}

	// 7. 返回创建结果
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
		"port_mappings":           finalPortMappings, // 显示分配后的端口映射
		"ports_pretty":            portsPretty,       // 友好展示：如 12223 ssh
		"requested_port_mappings": req.PortMappings,  // 显示原始请求的端口映射
		"environment":             req.Environment,
		"create_time":             instance.CreateTime,
		"description":             instance.Description,
		"docker_available":        dockerAvailable,
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

	// 同步Docker容器状态
	if config.DockerCli != nil {
		for i, instance := range instances {
			if instance.ContainerID != "" {
				containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
				if err != nil {
					fmt.Printf("获取容器 %s 状态失败: %v\n", instance.ContainerID, err)
					// 如果容器不存在，标记为已删除
					instances[i].Status = "deleted"
				} else {
					// 更新实际状态
					dockerStatus := containerInfo.State.Status
					if dockerStatus != instance.Status {
						fmt.Printf("同步容器状态: %s -> %s\n", instance.Status, dockerStatus)
						instances[i].Status = dockerStatus
						// 异步更新数据库状态
						go func(id uint, status string) {
							service.UpdateInstanceStatus(id, status)
						}(instance.ID, dockerStatus)
					}
				}
			}
		}
	}

	fmt.Printf("返回 %d 个容器实例\n", len(instances))
	// 包装响应，增加友好端口展示
	resp := make([]map[string]interface{}, 0, len(instances))
	for _, inst := range instances {
		portsPretty := []string{}
		if inst.PortMappings != "" {
			var pm map[string]string
			if err := json.Unmarshal([]byte(inst.PortMappings), &pm); err == nil {
				for _, host := range pm {
					portsPretty = append(portsPretty, fmt.Sprintf("%s %s", host, strings.ToLower(inst.Protocol)))
				}
			}
		}
		resp = append(resp, map[string]interface{}{
			"id":             inst.ID,
			"name":           inst.Name,
			"honeypot_name":  inst.HoneypotName,
			"container_name": inst.ContainerName,
			"container_id":   inst.ContainerID,
			"ip":             inst.IP,
			"honeypot_ip":    inst.HoneypotIP,
			"port":           inst.Port,
			"protocol":       inst.Protocol,
			"interface_type": inst.InterfaceType,
			"status":         inst.Status,
			"image_name":     inst.ImageName,
			"image_id":       inst.ImageID,
			"port_mappings":  inst.PortMappings,
			"ports_pretty":   portsPretty,
			"environment":    inst.Environment,
			"create_time":    inst.CreateTime,
			"update_time":    inst.UpdateTime,
			"description":    inst.Description,
		})
	}
	utils.ResponseSuccess(c, resp)
}

// SyncContainerStatus 同步所有容器状态
func SyncContainerStatus(c *gin.Context) {
	service, err := services.NewHoneypotInstanceService()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建服务失败: "+err.Error())
		return
	}

	if config.DockerCli == nil {
		utils.ResponseError(c, http.StatusServiceUnavailable, "Docker服务不可用")
		return
	}

	// 1. 获取数据库中所有实例
	instances, err := service.GetAllInstances()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取容器实例失败: "+err.Error())
		return
	}

	// 建立已知容器ID的映射
	knownContainerIDs := make(map[string]bool)
	for _, instance := range instances {
		if instance.ContainerID != "" {
			knownContainerIDs[instance.ContainerID] = true
		}
	}

	syncResults := make([]map[string]interface{}, 0)

	// 2. 更新已有实例的状态
	for _, instance := range instances {
		result := map[string]interface{}{
			"id":           instance.ID,
			"name":         instance.Name,
			"container_id": instance.ContainerID,
			"old_status":   instance.Status,
		}

		if instance.ContainerID != "" {
			containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
			if err != nil {
				// 容器在Docker中不存在，标记为deleted
				if instance.Status != "deleted" {
					result["new_status"] = "deleted"
					result["error"] = err.Error()
					service.UpdateInstanceStatus(instance.ID, "deleted")
				}
			} else {
				dockerStatus := containerInfo.State.Status
				result["new_status"] = dockerStatus

				if dockerStatus != instance.Status {
					if err := service.UpdateInstanceStatus(instance.ID, dockerStatus); err != nil {
						result["update_error"] = err.Error()
					} else {
						result["updated"] = true
					}
				} else {
					result["updated"] = false
				}
			}
		}
		syncResults = append(syncResults, result)
	}

	// 3. 发现并导入未管理的容器
	dockerContainers, err := config.DockerCli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err == nil {
		for _, dc := range dockerContainers {
			// 如果是未知的容器（不在数据库中）
			if !knownContainerIDs[dc.ID] {
				// 忽略非蜜罐相关的容器（可选：根据标签或命名规则过滤，这里暂时全部导入）
				// 简单过滤：忽略 Exited 的容器，除非明确需要
				// if dc.State == "exited" { continue }

				name := dc.Names[0]
				if strings.HasPrefix(name, "/") {
					name = name[1:]
				}

				// 尝试获取更多信息
				info, err := config.DockerCli.ContainerInspect(context.Background(), dc.ID)
				if err != nil {
					continue
				}

				// 简单的端口推断
				var mainPort int
				var protocol string = "tcp"
				portMappings := make(map[string]string)

				for p, bindings := range info.HostConfig.PortBindings {
					if len(bindings) > 0 {
						portMappings[p.Port()] = bindings[0].HostPort
						if mainPort == 0 {
							mainPort, _ = strconv.Atoi(bindings[0].HostPort)
							if p.Port() == "22" {
								protocol = "ssh"
							}
							if p.Port() == "80" {
								protocol = "http"
							}
							if p.Port() == "3306" {
								protocol = "mysql"
							}
						}
					}
				}

				portJSON, _ := json.Marshal(portMappings)

				newInstance := &repositories.HoneypotInstance{
					Name:          name,
					HoneypotName:  name, // 默认使用容器名
					ContainerName: name,
					ContainerID:   dc.ID,
					IP:            "0.0.0.0",
					Port:          mainPort,
					Protocol:      protocol,
					InterfaceType: "imported",
					Status:        dc.State,
					ImageName:     dc.Image,
					ImageID:       dc.ImageID,
					PortMappings:  string(portJSON),
					Environment:   "{}",
					CreateTime:    time.Unix(dc.Created, 0),
					UpdateTime:    time.Now(),
					Description:   "Auto-imported from Docker",
				}

				if err := service.CreateInstance(newInstance); err == nil {
					syncResults = append(syncResults, map[string]interface{}{
						"name":         name,
						"container_id": dc.ID,
						"action":       "imported",
						"status":       dc.State,
					})
				}
			}
		}
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"message": "容器状态同步完成",
		"results": syncResults,
	})
}

// GetContainerDebugInfo 获取容器调试信息
func GetContainerDebugInfo(c *gin.Context) {
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

	debugInfo := map[string]interface{}{
		"database_info": map[string]interface{}{
			"id":           instance.ID,
			"name":         instance.Name,
			"container_id": instance.ContainerID,
			"status":       instance.Status,
			"image_name":   instance.ImageName,
			"create_time":  instance.CreateTime,
			"update_time":  instance.UpdateTime,
		},
		"docker_available": config.DockerCli != nil,
	}

	// 如果Docker可用且有容器ID，获取Docker信息
	if config.DockerCli != nil && instance.ContainerID != "" {
		containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
		if err != nil {
			debugInfo["docker_error"] = err.Error()
			debugInfo["docker_status"] = "error"
		} else {
			debugInfo["docker_info"] = map[string]interface{}{
				"id":      containerInfo.ID,
				"name":    containerInfo.Name,
				"status":  containerInfo.State.Status,
				"running": containerInfo.State.Running,
				"image":   containerInfo.Config.Image,
				"created": containerInfo.Created,
			}
			debugInfo["docker_status"] = "found"
		}

		// 列出所有容器，查看是否有匹配的
		containers, err := config.DockerCli.ContainerList(context.Background(), container.ListOptions{All: true})
		if err != nil {
			debugInfo["list_error"] = err.Error()
		} else {
			matchingContainers := make([]map[string]interface{}, 0)
			for _, cont := range containers {
				if cont.ID == instance.ContainerID || strings.Contains(cont.Names[0], instance.ContainerName) {
					matchingContainers = append(matchingContainers, map[string]interface{}{
						"id":     cont.ID,
						"names":  cont.Names,
						"image":  cont.Image,
						"status": cont.Status,
						"state":  cont.State,
					})
				}
			}
			debugInfo["matching_containers"] = matchingContainers
		}
	} else {
		debugInfo["docker_status"] = "unavailable"
		if instance.ContainerID == "" {
			debugInfo["container_id_empty"] = true
		}
	}

	utils.ResponseSuccess(c, debugInfo)
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

	// 增加 ports_pretty
	portsPretty := []string{}
	if instance.PortMappings != "" {
		var pm map[string]string
		if err := json.Unmarshal([]byte(instance.PortMappings), &pm); err == nil {
			for _, host := range pm {
				portsPretty = append(portsPretty, fmt.Sprintf("%s %s", host, strings.ToLower(instance.Protocol)))
			}
		}
	}
	result := map[string]interface{}{
		"id":             instance.ID,
		"name":           instance.Name,
		"honeypot_name":  instance.HoneypotName,
		"container_name": instance.ContainerName,
		"container_id":   instance.ContainerID,
		"ip":             instance.IP,
		"honeypot_ip":    instance.HoneypotIP,
		"port":           instance.Port,
		"protocol":       instance.Protocol,
		"interface_type": instance.InterfaceType,
		"status":         instance.Status,
		"image_name":     instance.ImageName,
		"image_id":       instance.ImageID,
		"port_mappings":  instance.PortMappings,
		"ports_pretty":   portsPretty,
		"environment":    instance.Environment,
		"create_time":    instance.CreateTime,
		"update_time":    instance.UpdateTime,
		"description":    instance.Description,
	}
	utils.ResponseSuccess(c, result)
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
		fmt.Printf("尝试启动容器 ID: %s, 名称: %s\n", instance.ContainerID, instance.ContainerName)

		if err := config.DockerCli.ContainerStart(context.Background(), instance.ContainerID, container.StartOptions{}); err != nil {
			fmt.Printf("启动容器失败: %v\n", err)
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("启动容器失败: %v", err))
			return
		}

		// 验证容器是否真的启动了
		containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
		if err != nil {
			fmt.Printf("获取容器状态失败: %v\n", err)
		} else {
			fmt.Printf("容器当前状态: %s\n", containerInfo.State.Status)
		}

		// 更新数据库状态
		if err := service.UpdateInstanceStatus(uint(id), "running"); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("更新状态失败: %v", err))
			return
		}

		fmt.Printf("容器实例 %s 启动成功\n", instance.ContainerName)
	} else {
		fmt.Printf("容器实例 %s 没有有效的容器ID，无法启动\n", instance.ContainerName)
		utils.ResponseError(c, http.StatusBadRequest, "容器实例没有有效的容器ID")
		return
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
		fmt.Printf("尝试停止容器 ID: %s, 名称: %s\n", instance.ContainerID, instance.ContainerName)

		timeout := 30 // 30秒超时
		if err := config.DockerCli.ContainerStop(context.Background(), instance.ContainerID, container.StopOptions{
			Timeout: &timeout,
		}); err != nil {
			fmt.Printf("停止容器失败: %v\n", err)
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("停止容器失败: %v", err))
			return
		}

		// 验证容器是否真的停止了
		containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), instance.ContainerID)
		if err != nil {
			fmt.Printf("获取容器状态失败: %v\n", err)
		} else {
			fmt.Printf("容器当前状态: %s\n", containerInfo.State.Status)
		}

		// 更新数据库状态
		if err := service.UpdateInstanceStatus(uint(id), "stopped"); err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, fmt.Sprintf("更新状态失败: %v", err))
			return
		}

		fmt.Printf("容器实例 %s 停止成功\n", instance.ContainerName)
	} else {
		fmt.Printf("容器实例 %s 没有有效的容器ID，无法停止\n", instance.ContainerName)
		utils.ResponseError(c, http.StatusBadRequest, "容器实例没有有效的容器ID")
		return
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

	// 释放端口
	pm := services.GetPortManager()
	if err := pm.ReleasePortsByContainer(instance.ContainerName); err != nil {
		fmt.Printf("释放端口失败: %v\n", err)
	} else {
		fmt.Printf("容器 %s 的端口已释放\n", instance.ContainerName)
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

	// 11. 友好端口展示
	inferProto := func(p int) string {
		switch p {
		case 22, 2222:
			return "ssh"
		case 21:
			return "ftp"
		case 23:
			return "telnet"
		case 80, 8080:
			return "http"
		case 443, 8443:
			return "https"
		default:
			return "tcp"
		}
	}
	portsPretty := []string{}
	for _, host := range req.PortMappings {
		portsPretty = append(portsPretty, fmt.Sprintf("%s %s", host, inferProto(mainPort)))
	}

	// 12. 返回部署结果
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
		"ports_pretty":   portsPretty,
		"environment":    req.Environment,
		"create_time":    instance.CreateTime,
		"message":        fmt.Sprintf("成功将镜像 %s 部署到容器 %s", req.ImageName, req.ContainerName),
	}

	utils.ResponseSuccess(c, result)
}
