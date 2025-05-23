package services

import (
	"andorralee/internal/config"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// HoneypotService 蜜罐服务
type HoneypotService struct {
	cli *client.Client
}

// NewHoneypotService 创建蜜罐服务实例
func NewHoneypotService() *HoneypotService {
	return &HoneypotService{
		cli: config.DockerCli,
	}
}

// DeployHoneypot 部署蜜罐
func (s *HoneypotService) DeployHoneypot(honeypotType config.HoneypotType) (string, error) {
	// 获取蜜罐配置
	honeypotConfig, exists := config.DefaultHoneypotConfigs[honeypotType]
	if !exists {
		return "", fmt.Errorf("unsupported honeypot type: %s", honeypotType)
	}

	// 创建容器配置
	containerConfig := &container.Config{
		Image: honeypotConfig.Image,
		Env:   convertEnvMapToList(honeypotConfig.Environment),
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		PortBindings: parsePortBindings(honeypotConfig.Ports),
		Resources: container.Resources{
			CPUQuota:   parseCPUQuota(honeypotConfig.Resources.CPULimit),
			Memory:     parseMemoryLimit(honeypotConfig.Resources.MemoryLimit),
			MemorySwap: -1, // 禁用swap
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}

	// 创建容器
	resp, err := s.cli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		&network.NetworkingConfig{},
		nil,
		honeypotConfig.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	// 启动容器
	if err := s.cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	return resp.ID, nil
}

// StopHoneypot 停止蜜罐
func (s *HoneypotService) StopHoneypot(containerID string) error {
	timeout := 10
	return s.cli.ContainerStop(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
}

// GetHoneypotStatus 获取蜜罐状态
func (s *HoneypotService) GetHoneypotStatus(containerID string) (*types.ContainerJSON, error) {
	containerJSON, err := s.cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, err
	}
	return &containerJSON, nil
}

// ListHoneypots 列出所有蜜罐
func (s *HoneypotService) ListHoneypots() ([]types.Container, error) {
	return s.cli.ContainerList(context.Background(), container.ListOptions{All: true})
}

// GetHoneypotLogs 获取蜜罐日志
func (s *HoneypotService) GetHoneypotLogs(containerID string) (string, error) {
	reader, err := s.cli.ContainerLogs(
		context.Background(),
		containerID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
			Timestamps: true,
		},
	)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// 辅助函数

// convertEnvMapToList 将环境变量map转换为列表
func convertEnvMapToList(envMap map[string]string) []string {
	var envList []string
	for k, v := range envMap {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	return envList
}

// parsePortBindings 解析端口绑定
// func parsePortBindings(portMap map[string]string) nat.PortMap {
// 	bindings := nat.PortMap{}
// 	for containerPort, hostPort := range portMap {
// 		port, err := nat.NewPort("tcp", containerPort)
// 		if err != nil {
// 			continue
// 		}
// 		bindings[port] = []nat.PortBinding{
// 			{HostPort: hostPort},
// 		}
// 	}
// 	return bindings
// }

// parseCPUQuota 解析CPU限制
func parseCPUQuota(cpuLimit string) int64 {
	// 这里简单实现，实际应该根据CPU核心数计算
	return 100000 // 默认限制为1个CPU核心
}

// parseMemoryLimit 解析内存限制
func parseMemoryLimit(memoryLimit string) int64 {
	// 这里简单实现，实际应该解析单位（m, g等）
	return 512 * 1024 * 1024 // 默认512MB
}
