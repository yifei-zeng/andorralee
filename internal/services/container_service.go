package services

import (
	"andorralee/internal/config"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// ContainerInfo 容器信息结构体
type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	Ports   map[string]string `json:"ports"`
	Created string            `json:"created"`
}

// StartContainer 启动容器
func StartContainer(image, name string, portMap, envVars map[string]string) (string, error) {
	cli := config.DockerCli

	// 转换端口映射
	hostConfig := &container.HostConfig{}
	if portMap != nil {
		hostConfig.PortBindings = parsePortBindings(portMap)
	}

	// 转换环境变量
	var envList []string
	for k, v := range envVars {
		envList = append(envList, k+"="+v)
	}

	// 创建容器配置
	resp, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: image,
			Env:   envList,
		},
		hostConfig,
		&network.NetworkingConfig{},
		nil,
		name,
	)
	if err != nil {
		return "", err
	}

	// 启动容器
	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

// StopContainer 停止容器
func StopContainer(containerID string) error {
	cli := config.DockerCli
	timeout := 10
	return cli.ContainerStop(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
}

// GetContainerLogs 获取容器日志
func GetContainerLogs(containerID string) (string, error) {
	cli := config.DockerCli
	reader, err := cli.ContainerLogs(
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

// GetContainerInfo 获取容器详细信息
func GetContainerInfo(containerID string) (*ContainerInfo, error) {
	cli := config.DockerCli
	info, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, err
	}

	ports := make(map[string]string)
	for containerPort, bindings := range info.HostConfig.PortBindings {
		if len(bindings) > 0 {
			ports[containerPort.Port()] = bindings[0].HostPort
		}
	}

	return &ContainerInfo{
		ID:      info.ID,
		Name:    info.Name,
		Image:   info.Config.Image,
		Status:  info.State.Status,
		Ports:   ports,
		Created: info.Created,
	}, nil
}

// ListContainers 列出所有容器
func ListContainers() ([]ContainerInfo, error) {
	cli := config.DockerCli
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, c := range containers {
		info, err := GetContainerInfo(c.ID)
		if err != nil {
			continue
		}
		result = append(result, *info)
	}

	return result, nil
}

// parsePortBindings 转换端口映射格式
func parsePortBindings(portMap map[string]string) nat.PortMap {
	bindings := nat.PortMap{}
	for containerPort, hostPort := range portMap {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			continue
		}
		bindings[port] = []nat.PortBinding{
			{HostPort: hostPort},
		}
	}
	return bindings
}
