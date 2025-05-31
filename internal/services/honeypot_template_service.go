package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"errors"
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
	return s.repo.GetAll()
}

// GetTemplateByID 根据ID获取蜜罐模板
func (s *HoneypotTemplateService) GetTemplateByID(id uint) (*repositories.HoneypotTemplate, error) {
	return s.repo.GetByID(id)
}

// CreateTemplate 创建蜜罐模板
func (s *HoneypotTemplateService) CreateTemplate(template *repositories.HoneypotTemplate) error {
	return s.repo.Create(template)
}

// UpdateTemplate 更新蜜罐模板
func (s *HoneypotTemplateService) UpdateTemplate(template *repositories.HoneypotTemplate) error {
	return s.repo.Update(template)
}

// DeleteTemplate 删除蜜罐模板
func (s *HoneypotTemplateService) DeleteTemplate(id uint) error {
	return s.repo.Delete(id)
}

// DeployTemplate 部署蜜罐模板（创建实例）
func (s *HoneypotTemplateService) DeployTemplate(templateID uint, instance *repositories.HoneypotInstance) error {
	// 检查模板是否存在
	template, err := s.repo.GetByID(templateID)
	if err != nil {
		return err
	}

	// 设置实例的模板ID
	instance.TemplateID = template.ID

	// 创建实例
	instanceService, err := NewHoneypotInstanceService()
	if err != nil {
		return err
	}

	if err := instanceService.CreateInstance(instance); err != nil {
		return err
	}

	// 增加模板部署数量
	return s.repo.IncrementDeployCount(templateID)
}

// ImportTemplate 导入蜜罐模板
func (s *HoneypotTemplateService) ImportTemplate(template *repositories.HoneypotTemplate) error {
	// 这里可以添加导入模板的逻辑，例如从Docker Hub拉取镜像等
	// 暂时简单实现为创建模板
	return s.repo.Create(template)
}
