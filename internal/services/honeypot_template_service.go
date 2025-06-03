package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"errors"
	"time"
)

// 错误定义
var (
	ErrTemplateInUse = errors.New("蜜罐模板正在使用中")
)

// HoneypotTemplateService 蜜罐模板服务
type HoneypotTemplateService struct {
	repo repositories.HoneypotTemplateRepository
}

// NewHoneypotTemplateService 创建蜜罐模板服务
func NewHoneypotTemplateService() (*HoneypotTemplateService, error) {
	if config.MySQLDB == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}

	repo := repositories.NewMySQLHoneypotTemplateRepo(config.MySQLDB)
	return &HoneypotTemplateService{repo: repo}, nil
}

// GetAllTemplates 获取所有蜜罐模板
func (s *HoneypotTemplateService) GetAllTemplates() ([]repositories.HoneypotTemplate, error) {
	return s.repo.List()
}

// GetTemplateByID 根据ID获取蜜罐模板
func (s *HoneypotTemplateService) GetTemplateByID(id uint) (*repositories.HoneypotTemplate, error) {
	return s.repo.GetByID(id)
}

// CreateTemplate 创建蜜罐模板
func (s *HoneypotTemplateService) CreateTemplate(template *repositories.HoneypotTemplate) error {
	template.ImportTime = time.Now()
	template.DeployCount = 0
	return s.repo.Create(template)
}

// UpdateTemplate 更新蜜罐模板
func (s *HoneypotTemplateService) UpdateTemplate(template *repositories.HoneypotTemplate) error {
	return s.repo.Update(template)
}

// DeleteTemplate 删除蜜罐模板
func (s *HoneypotTemplateService) DeleteTemplate(id uint) error {
	// 检查是否有实例在使用该模板
	instanceRepo := repositories.NewMySQLHoneypotInstanceRepo(config.MySQLDB)
	instances, err := instanceRepo.GetByTemplateID(id)
	if err != nil {
		return err
	}

	if len(instances) > 0 {
		return ErrTemplateInUse
	}

	return s.repo.Delete(id)
}

// DeployTemplate 部署蜜罐模板
func (s *HoneypotTemplateService) DeployTemplate(templateID uint, instanceName, ip string, port int) (uint, error) {
	// 获取模板
	template, err := s.GetTemplateByID(templateID)
	if err != nil {
		return 0, err
	}

	// 创建实例
	instance := &repositories.HoneypotInstance{
		Name:          instanceName,
		ContainerName: "honeypot_" + instanceName,
		IP:            ip,
		Port:          port,
		Protocol:      template.Protocol,
		TemplateID:    templateID,
		Status:        "已部署",
		CreateTime:    time.Now(),
	}

	// 保存实例
	instanceRepo := repositories.NewMySQLHoneypotInstanceRepo(config.MySQLDB)
	if err := instanceRepo.Create(instance); err != nil {
		return 0, err
	}

	// 增加模板部署数量
	if err := s.repo.IncrementDeployCount(templateID); err != nil {
		return 0, err
	}

	return instance.ID, nil
}

// ImportTemplate 导入蜜罐模板
func (s *HoneypotTemplateService) ImportTemplate(name, protocol string) error {
	// 创建模板
	template := &repositories.HoneypotTemplate{
		Name:     name,
		Protocol: protocol,
	}

	return s.CreateTemplate(template)
}
