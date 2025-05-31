package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// HoneypotInstanceService 蜜罐实例服务
type HoneypotInstanceService struct {
	repo repositories.HoneypotInstanceRepository
}

// NewHoneypotInstanceService 创建蜜罐实例服务
func NewHoneypotInstanceService() (*HoneypotInstanceService, error) {
	if config.MySQLDB == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}

	repo := repositories.NewMySQLHoneypotInstanceRepo(config.MySQLDB)
	return &HoneypotInstanceService{repo: repo}, nil
}

// GetAllInstances 获取所有蜜罐实例
func (s *HoneypotInstanceService) GetAllInstances() ([]repositories.HoneypotInstance, error) {
	return s.repo.GetAll()
}

// GetInstanceByID 根据ID获取蜜罐实例
func (s *HoneypotInstanceService) GetInstanceByID(id uint) (*repositories.HoneypotInstance, error) {
	return s.repo.GetByID(id)
}

// GetInstancesByTemplateID 根据模板ID获取蜜罐实例
func (s *HoneypotInstanceService) GetInstancesByTemplateID(templateID uint) ([]repositories.HoneypotInstance, error) {
	return s.repo.GetByTemplateID(templateID)
}

// CreateInstance 创建蜜罐实例
func (s *HoneypotInstanceService) CreateInstance(instance *repositories.HoneypotInstance) error {
	// 设置初始状态
	instance.Status = "created"
	instance.CreateTime = time.Now()

	return s.repo.Create(instance)
}

// UpdateInstance 更新蜜罐实例
func (s *HoneypotInstanceService) UpdateInstance(instance *repositories.HoneypotInstance) error {
	return s.repo.Update(instance)
}

// UpdateInstanceStatus 更新蜜罐实例状态
func (s *HoneypotInstanceService) UpdateInstanceStatus(id uint, status string) error {
	return s.repo.UpdateStatus(id, status)
}

// DeleteInstance 删除蜜罐实例
func (s *HoneypotInstanceService) DeleteInstance(id uint) error {
	// 获取实例信息
	instance, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 如果实例已经部署，先停止容器
	if instance.Status == "running" {
		if err := s.StopInstance(id); err != nil {
			return err
		}
	}

	// 删除实例记录
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// 减少模板部署数量
	templateService, err := NewHoneypotTemplateService()
	if err != nil {
		return err
	}

	return templateService.repo.DecrementDeployCount(instance.TemplateID)
}

// DeployInstance 部署蜜罐实例（启动Docker容器）
func (s *HoneypotInstanceService) DeployInstance(id uint) error {
	// 获取实例信息
	instance, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 获取模板信息
	templateService, err := NewHoneypotTemplateService()
	if err != nil {
		return err
	}

	template, err := templateService.GetTemplateByID(instance.TemplateID)
	if err != nil {
		return err
	}

	// 检查Docker客户端是否初始化
	if config.DockerCli == nil {
		return errors.New("Docker客户端未初始化")
	}

	// 创建端口映射
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	if instance.Port > 0 {
		// 创建端口映射
		containerPort := nat.Port(fmt.Sprintf("%d/tcp", instance.Port))
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(instance.Port),
			},
		}
	}

	// 创建容器配置
	containerConfig := &container.Config{
		Image:        template.Name,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// 创建容器
	resp, err := config.DockerCli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		&network.NetworkingConfig{},
		nil,
		instance.Name,
	)
	if err != nil {
		return err
	}

	// 更新实例信息
	instance.ContainerName = instance.Name
	instance.Status = "created"
	if err := s.repo.Update(instance); err != nil {
		return err
	}

	// 启动容器
	if err := config.DockerCli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	// 更新实例状态
	return s.repo.UpdateStatus(id, "running")
}

// StopInstance 停止蜜罐实例
func (s *HoneypotInstanceService) StopInstance(id uint) error {
	// 获取实例信息
	instance, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 检查Docker客户端是否初始化
	if config.DockerCli == nil {
		return errors.New("Docker客户端未初始化")
	}

	// 停止容器
	timeout := 10 // 超时时间（秒）
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}
	if err := config.DockerCli.ContainerStop(context.Background(), instance.ContainerName, stopOptions); err != nil {
		return err
	}

	// 更新实例状态
	return s.repo.UpdateStatus(id, "stopped")
}

// GetInstanceLogs 获取蜜罐实例日志
func (s *HoneypotInstanceService) GetInstanceLogs(id uint) ([]repositories.HoneypotLog, error) {
	// 获取实例信息
	instance, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 获取日志
	logRepo := repositories.NewMySQLHoneypotLogRepo(config.MySQLDB)
	return logRepo.GetByInstanceID(instance.ID)
}
