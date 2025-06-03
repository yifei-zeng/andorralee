package repositories

import (
	"time"

	"gorm.io/gorm"
)

// -------------------- 蜜罐模板仓库 --------------------

// MySQLHoneypotTemplateRepo 蜜罐模板MySQL仓库
type MySQLHoneypotTemplateRepo struct {
	DB *gorm.DB
}

// NewMySQLHoneypotTemplateRepo 创建蜜罐模板MySQL仓库
func NewMySQLHoneypotTemplateRepo(db *gorm.DB) HoneypotTemplateRepository {
	return &MySQLHoneypotTemplateRepo{DB: db}
}

// List 获取所有蜜罐模板
func (r *MySQLHoneypotTemplateRepo) List() ([]HoneypotTemplate, error) {
	var templates []HoneypotTemplate
	result := r.DB.Find(&templates)
	return templates, result.Error
}

// GetByID 根据ID获取蜜罐模板
func (r *MySQLHoneypotTemplateRepo) GetByID(id uint) (*HoneypotTemplate, error) {
	var template HoneypotTemplate
	result := r.DB.First(&template, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &template, nil
}

// Create 创建蜜罐模板
func (r *MySQLHoneypotTemplateRepo) Create(template *HoneypotTemplate) error {
	template.ImportTime = time.Now()
	return r.DB.Create(template).Error
}

// Update 更新蜜罐模板
func (r *MySQLHoneypotTemplateRepo) Update(template *HoneypotTemplate) error {
	return r.DB.Save(template).Error
}

// Delete 删除蜜罐模板
func (r *MySQLHoneypotTemplateRepo) Delete(id uint) error {
	return r.DB.Delete(&HoneypotTemplate{}, id).Error
}

// IncrementDeployCount 增加部署数量
func (r *MySQLHoneypotTemplateRepo) IncrementDeployCount(id uint) error {
	return r.DB.Model(&HoneypotTemplate{}).Where("id = ?", id).
		UpdateColumn("deploy_count", gorm.Expr("deploy_count + ?", 1)).Error
}

// DecrementDeployCount 减少部署数量
func (r *MySQLHoneypotTemplateRepo) DecrementDeployCount(id uint) error {
	return r.DB.Model(&HoneypotTemplate{}).Where("id = ? AND deploy_count > 0", id).
		UpdateColumn("deploy_count", gorm.Expr("deploy_count - ?", 1)).Error
}

// -------------------- 蜜罐实例仓库 --------------------

// MySQLHoneypotInstanceRepo 蜜罐实例MySQL仓库
type MySQLHoneypotInstanceRepo struct {
	DB *gorm.DB
}

// NewMySQLHoneypotInstanceRepo 创建蜜罐实例MySQL仓库
func NewMySQLHoneypotInstanceRepo(db *gorm.DB) HoneypotInstanceRepository {
	return &MySQLHoneypotInstanceRepo{DB: db}
}

// List 获取所有蜜罐实例
func (r *MySQLHoneypotInstanceRepo) List() ([]HoneypotInstance, error) {
	var instances []HoneypotInstance
	result := r.DB.Preload("Template").Find(&instances)
	return instances, result.Error
}

// GetByID 根据ID获取蜜罐实例
func (r *MySQLHoneypotInstanceRepo) GetByID(id uint) (*HoneypotInstance, error) {
	var instance HoneypotInstance
	result := r.DB.Preload("Template").First(&instance, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &instance, nil
}

// GetByTemplateID 根据模板ID获取蜜罐实例
func (r *MySQLHoneypotInstanceRepo) GetByTemplateID(templateID uint) ([]HoneypotInstance, error) {
	var instances []HoneypotInstance
	result := r.DB.Where("template_id = ?", templateID).Find(&instances)
	return instances, result.Error
}

// Create 创建蜜罐实例
func (r *MySQLHoneypotInstanceRepo) Create(instance *HoneypotInstance) error {
	instance.CreateTime = time.Now()
	instance.Status = "已部署"
	return r.DB.Create(instance).Error
}

// Update 更新蜜罐实例
func (r *MySQLHoneypotInstanceRepo) Update(instance *HoneypotInstance) error {
	return r.DB.Save(instance).Error
}

// Delete 删除蜜罐实例
func (r *MySQLHoneypotInstanceRepo) Delete(id uint) error {
	return r.DB.Delete(&HoneypotInstance{}, id).Error
}

// UpdateStatus 更新蜜罐实例状态
func (r *MySQLHoneypotInstanceRepo) UpdateStatus(id uint, status string) error {
	return r.DB.Model(&HoneypotInstance{}).Where("id = ?", id).
		Update("status", status).Error
}

// GetByStatus 根据状态获取蜜罐实例
func (r *MySQLHoneypotInstanceRepo) GetByStatus(status string) ([]HoneypotInstance, error) {
	var instances []HoneypotInstance
	result := r.DB.Where("status = ?", status).Preload("Template").Find(&instances)
	return instances, result.Error
}

// GetByContainerID 根据容器ID获取蜜罐实例
func (r *MySQLHoneypotInstanceRepo) GetByContainerID(containerID string) (*HoneypotInstance, error) {
	var instance HoneypotInstance
	result := r.DB.Where("container_id = ?", containerID).Preload("Template").First(&instance)
	if result.Error != nil {
		return nil, result.Error
	}
	return &instance, nil
}

// -------------------- 蜜罐日志仓库 --------------------

// MySQLHoneypotLogRepo 蜜罐日志MySQL仓库
type MySQLHoneypotLogRepo struct {
	DB *gorm.DB
}

// NewMySQLHoneypotLogRepo 创建蜜罐日志MySQL仓库
func NewMySQLHoneypotLogRepo(db *gorm.DB) HoneypotLogRepository {
	return &MySQLHoneypotLogRepo{DB: db}
}

// GetAll 获取所有蜜罐日志
func (r *MySQLHoneypotLogRepo) GetAll() ([]HoneypotLog, error) {
	var logs []HoneypotLog
	result := r.DB.Find(&logs)
	return logs, result.Error
}

// List 获取所有蜜罐日志
func (r *MySQLHoneypotLogRepo) List() ([]HoneypotLog, error) {
	return r.GetAll()
}

// GetByID 根据ID获取蜜罐日志
func (r *MySQLHoneypotLogRepo) GetByID(id uint) (*HoneypotLog, error) {
	var log HoneypotLog
	result := r.DB.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetByInstanceID 根据实例ID获取蜜罐日志
func (r *MySQLHoneypotLogRepo) GetByInstanceID(instanceID uint) ([]HoneypotLog, error) {
	var logs []HoneypotLog
	result := r.DB.Where("instance_id = ?", instanceID).Find(&logs)
	return logs, result.Error
}

// Create 创建蜜罐日志
func (r *MySQLHoneypotLogRepo) Create(log *HoneypotLog) error {
	return r.DB.Create(log).Error
}

// Delete 删除蜜罐日志
func (r *MySQLHoneypotLogRepo) Delete(id uint) error {
	return r.DB.Delete(&HoneypotLog{}, id).Error
}

// -------------------- 诱饵仓库 --------------------

// MySQLBaitRepo 诱饵MySQL仓库
type MySQLBaitRepo struct {
	DB *gorm.DB
}

// NewMySQLBaitRepo 创建诱饵MySQL仓库
func NewMySQLBaitRepo(db *gorm.DB) BaitRepository {
	return &MySQLBaitRepo{DB: db}
}

// List 获取所有诱饵
func (r *MySQLBaitRepo) List() ([]Bait, error) {
	var baits []Bait
	result := r.DB.Preload("Instance").Find(&baits)
	return baits, result.Error
}

// GetByID 根据ID获取诱饵
func (r *MySQLBaitRepo) GetByID(id uint) (*Bait, error) {
	var bait Bait
	result := r.DB.Preload("Instance").First(&bait, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &bait, nil
}

// Create 创建诱饵
func (r *MySQLBaitRepo) Create(bait *Bait) error {
	bait.CreateTime = time.Now()
	return r.DB.Create(bait).Error
}

// Update 更新诱饵
func (r *MySQLBaitRepo) Update(bait *Bait) error {
	return r.DB.Save(bait).Error
}

// Delete 删除诱饵
func (r *MySQLBaitRepo) Delete(id uint) error {
	return r.DB.Delete(&Bait{}, id).Error
}

// UpdateDeployStatus 更新诱饵部署状态
func (r *MySQLBaitRepo) UpdateDeployStatus(id uint, isDeployed bool) error {
	return r.DB.Model(&Bait{}).Where("id = ?", id).
		Update("is_deployed", isDeployed).Error
}

// -------------------- 安全规则仓库 --------------------

// MySQLSecurityRuleRepo 安全规则MySQL仓库
type MySQLSecurityRuleRepo struct {
	DB *gorm.DB
}

// NewMySQLSecurityRuleRepo 创建安全规则MySQL仓库
func NewMySQLSecurityRuleRepo(db *gorm.DB) SecurityRuleRepository {
	return &MySQLSecurityRuleRepo{DB: db}
}

// List 获取所有安全规则
func (r *MySQLSecurityRuleRepo) List() ([]SecurityRule, error) {
	var rules []SecurityRule
	result := r.DB.Find(&rules)
	return rules, result.Error
}

// GetByID 根据ID获取安全规则
func (r *MySQLSecurityRuleRepo) GetByID(id uint) (*SecurityRule, error) {
	var rule SecurityRule
	result := r.DB.First(&rule, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &rule, nil
}

// Create 创建安全规则
func (r *MySQLSecurityRuleRepo) Create(rule *SecurityRule) error {
	return r.DB.Create(rule).Error
}

// Update 更新安全规则
func (r *MySQLSecurityRuleRepo) Update(rule *SecurityRule) error {
	return r.DB.Save(rule).Error
}

// Delete 删除安全规则
func (r *MySQLSecurityRuleRepo) Delete(id uint) error {
	return r.DB.Delete(&SecurityRule{}, id).Error
}

// UpdateStatus 更新安全规则状态
func (r *MySQLSecurityRuleRepo) UpdateStatus(id uint, isEnabled bool) error {
	return r.DB.Model(&SecurityRule{}).Where("id = ?", id).
		Update("is_enabled", isEnabled).Error
}

// -------------------- 规则日志仓库 --------------------

// MySQLRuleLogRepo 规则日志MySQL仓库
type MySQLRuleLogRepo struct {
	DB *gorm.DB
}

// NewMySQLRuleLogRepo 创建规则日志MySQL仓库
func NewMySQLRuleLogRepo(db *gorm.DB) RuleLogRepository {
	return &MySQLRuleLogRepo{DB: db}
}

// GetAll 获取所有规则日志
func (r *MySQLRuleLogRepo) GetAll() ([]RuleLog, error) {
	var logs []RuleLog
	result := r.DB.Find(&logs)
	return logs, result.Error
}

// List 获取所有规则日志
func (r *MySQLRuleLogRepo) List() ([]RuleLog, error) {
	return r.GetAll()
}

// GetByID 根据ID获取规则日志
func (r *MySQLRuleLogRepo) GetByID(id uint) (*RuleLog, error) {
	var log RuleLog
	result := r.DB.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetByRuleID 根据规则ID获取规则日志
func (r *MySQLRuleLogRepo) GetByRuleID(ruleID uint) ([]RuleLog, error) {
	var logs []RuleLog
	result := r.DB.Where("rule_id = ?", ruleID).Find(&logs)
	return logs, result.Error
}

// Create 创建规则日志
func (r *MySQLRuleLogRepo) Create(log *RuleLog) error {
	return r.DB.Create(log).Error
}

// Delete 删除规则日志
func (r *MySQLRuleLogRepo) Delete(id uint) error {
	return r.DB.Delete(&RuleLog{}, id).Error
}

// -------------------- Docker镜像仓库 --------------------

// MySQLDockerImageRepo Docker镜像MySQL仓库
type MySQLDockerImageRepo struct {
	DB *gorm.DB
}

// NewMySQLDockerImageRepo 创建Docker镜像MySQL仓库
func NewMySQLDockerImageRepo(db *gorm.DB) DockerImageRepository {
	return &MySQLDockerImageRepo{DB: db}
}

// List 获取所有Docker镜像
func (r *MySQLDockerImageRepo) List() ([]DockerImage, error) {
	var images []DockerImage
	result := r.DB.Find(&images)
	return images, result.Error
}

// GetByID 根据ID获取Docker镜像
func (r *MySQLDockerImageRepo) GetByID(id uint) (*DockerImage, error) {
	var image DockerImage
	result := r.DB.First(&image, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &image, nil
}

// GetByImageID 根据镜像ID获取Docker镜像
func (r *MySQLDockerImageRepo) GetByImageID(imageID string) (*DockerImage, error) {
	var image DockerImage
	result := r.DB.Where("image_id = ?", imageID).First(&image)
	if result.Error != nil {
		return nil, result.Error
	}
	return &image, nil
}

// Create 创建Docker镜像记录
func (r *MySQLDockerImageRepo) Create(image *DockerImage) error {
	image.CreatedAt = time.Now()
	image.UpdatedAt = time.Now()
	return r.DB.Create(image).Error
}

// Update 更新Docker镜像记录
func (r *MySQLDockerImageRepo) Update(image *DockerImage) error {
	image.UpdatedAt = time.Now()
	return r.DB.Save(image).Error
}

// Delete 删除Docker镜像记录
func (r *MySQLDockerImageRepo) Delete(id uint) error {
	return r.DB.Delete(&DockerImage{}, id).Error
}

// DeleteByImageID 根据镜像ID删除Docker镜像记录
func (r *MySQLDockerImageRepo) DeleteByImageID(imageID string) error {
	return r.DB.Where("image_id = ?", imageID).Delete(&DockerImage{}).Error
}

// -------------------- Docker镜像日志仓库 --------------------

// MySQLDockerImageLogRepo Docker镜像日志MySQL仓库
type MySQLDockerImageLogRepo struct {
	DB *gorm.DB
}

// NewMySQLDockerImageLogRepo 创建Docker镜像日志MySQL仓库
func NewMySQLDockerImageLogRepo(db *gorm.DB) DockerImageLogRepository {
	return &MySQLDockerImageLogRepo{DB: db}
}

// List 获取所有Docker镜像日志
func (r *MySQLDockerImageLogRepo) List() ([]DockerImageLog, error) {
	var logs []DockerImageLog
	result := r.DB.Order("created_at DESC").Find(&logs)
	return logs, result.Error
}

// GetByID 根据ID获取Docker镜像日志
func (r *MySQLDockerImageLogRepo) GetByID(id uint) (*DockerImageLog, error) {
	var log DockerImageLog
	result := r.DB.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetByImageID 根据镜像ID获取Docker镜像日志
func (r *MySQLDockerImageLogRepo) GetByImageID(imageID string) ([]DockerImageLog, error) {
	var logs []DockerImageLog
	result := r.DB.Where("image_id = ?", imageID).Order("created_at DESC").Find(&logs)
	return logs, result.Error
}

// Create 创建Docker镜像日志
func (r *MySQLDockerImageLogRepo) Create(log *DockerImageLog) error {
	log.CreatedAt = time.Now()
	return r.DB.Create(log).Error
}

// Delete 删除Docker镜像日志
func (r *MySQLDockerImageLogRepo) Delete(id uint) error {
	return r.DB.Delete(&DockerImageLog{}, id).Error
}

// -------------------- 容器日志分析仓库 --------------------

// MySQLContainerLogSegmentRepo 容器日志分析MySQL仓库
type MySQLContainerLogSegmentRepo struct {
	DB *gorm.DB
}

// NewMySQLContainerLogSegmentRepo 创建容器日志分析MySQL仓库
func NewMySQLContainerLogSegmentRepo(db *gorm.DB) ContainerLogSegmentRepository {
	return &MySQLContainerLogSegmentRepo{DB: db}
}

// List 获取所有容器日志分析结果
func (r *MySQLContainerLogSegmentRepo) List() ([]ContainerLogSegment, error) {
	var segments []ContainerLogSegment
	result := r.DB.Order("created_at DESC").Find(&segments)
	return segments, result.Error
}

// GetByID 根据ID获取容器日志分析结果
func (r *MySQLContainerLogSegmentRepo) GetByID(id uint) (*ContainerLogSegment, error) {
	var segment ContainerLogSegment
	result := r.DB.First(&segment, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &segment, nil
}

// GetByContainerID 根据容器ID获取日志分析结果
func (r *MySQLContainerLogSegmentRepo) GetByContainerID(containerID string) ([]ContainerLogSegment, error) {
	var segments []ContainerLogSegment
	result := r.DB.Where("container_id = ?", containerID).Order("timestamp DESC, line_number ASC").Find(&segments)
	return segments, result.Error
}

// GetBySegmentType 根据日志类型获取分析结果
func (r *MySQLContainerLogSegmentRepo) GetBySegmentType(segmentType string) ([]ContainerLogSegment, error) {
	var segments []ContainerLogSegment
	result := r.DB.Where("segment_type = ?", segmentType).Order("created_at DESC").Find(&segments)
	return segments, result.Error
}

// Create 创建容器日志分析结果
func (r *MySQLContainerLogSegmentRepo) Create(segment *ContainerLogSegment) error {
	segment.CreatedAt = time.Now()
	return r.DB.Create(segment).Error
}

// CreateBatch 批量创建容器日志分析结果
func (r *MySQLContainerLogSegmentRepo) CreateBatch(segments []ContainerLogSegment) error {
	now := time.Now()
	for i := range segments {
		segments[i].CreatedAt = now
	}
	return r.DB.CreateInBatches(segments, 100).Error
}

// Delete 删除容器日志分析结果
func (r *MySQLContainerLogSegmentRepo) Delete(id uint) error {
	return r.DB.Delete(&ContainerLogSegment{}, id).Error
}

// DeleteByContainerID 根据容器ID删除所有相关日志分析结果
func (r *MySQLContainerLogSegmentRepo) DeleteByContainerID(containerID string) error {
	return r.DB.Where("container_id = ?", containerID).Delete(&ContainerLogSegment{}).Error
}

// -------------------- Docker容器仓库 --------------------

// MySQLDockerContainerRepo Docker容器MySQL仓库
type MySQLDockerContainerRepo struct {
	DB *gorm.DB
}

// NewMySQLDockerContainerRepo 创建Docker容器MySQL仓库
func NewMySQLDockerContainerRepo(db *gorm.DB) DockerContainerRepository {
	return &MySQLDockerContainerRepo{DB: db}
}

// List 获取所有Docker容器
func (r *MySQLDockerContainerRepo) List() ([]DockerContainer, error) {
	var containers []DockerContainer
	result := r.DB.Order("created_at DESC").Find(&containers)
	return containers, result.Error
}

// GetByID 根据ID获取Docker容器
func (r *MySQLDockerContainerRepo) GetByID(id uint) (*DockerContainer, error) {
	var container DockerContainer
	result := r.DB.First(&container, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &container, nil
}

// GetByContainerID 根据容器ID获取Docker容器
func (r *MySQLDockerContainerRepo) GetByContainerID(containerID string) (*DockerContainer, error) {
	var container DockerContainer
	result := r.DB.Where("container_id = ?", containerID).First(&container)
	if result.Error != nil {
		return nil, result.Error
	}
	return &container, nil
}

// GetByImageID 根据镜像ID获取Docker容器列表
func (r *MySQLDockerContainerRepo) GetByImageID(imageID string) ([]DockerContainer, error) {
	var containers []DockerContainer
	result := r.DB.Where("image_id = ?", imageID).Find(&containers)
	return containers, result.Error
}

// Create 创建Docker容器记录
func (r *MySQLDockerContainerRepo) Create(container *DockerContainer) error {
	container.CreatedAt = time.Now()
	container.UpdatedAt = time.Now()
	return r.DB.Create(container).Error
}

// Update 更新Docker容器记录
func (r *MySQLDockerContainerRepo) Update(container *DockerContainer) error {
	container.UpdatedAt = time.Now()
	return r.DB.Save(container).Error
}

// Delete 删除Docker容器记录
func (r *MySQLDockerContainerRepo) Delete(id uint) error {
	return r.DB.Delete(&DockerContainer{}, id).Error
}

// DeleteByContainerID 根据容器ID删除Docker容器记录
func (r *MySQLDockerContainerRepo) DeleteByContainerID(containerID string) error {
	return r.DB.Where("container_id = ?", containerID).Delete(&DockerContainer{}).Error
}

// UpdateStatus 更新容器状态
func (r *MySQLDockerContainerRepo) UpdateStatus(containerID string, status string) error {
	return r.DB.Model(&DockerContainer{}).Where("container_id = ?", containerID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// -------------------- Headling认证日志仓库 --------------------

// MySQLHeadlingAuthLogRepo Headling认证日志MySQL仓库
type MySQLHeadlingAuthLogRepo struct {
	DB *gorm.DB
}

// NewMySQLHeadlingAuthLogRepo 创建Headling认证日志MySQL仓库
func NewMySQLHeadlingAuthLogRepo(db *gorm.DB) HeadlingAuthLogRepository {
	return &MySQLHeadlingAuthLogRepo{DB: db}
}

// List 获取所有Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) List() ([]HeadlingAuthLog, error) {
	var logs []HeadlingAuthLog
	result := r.DB.Order("timestamp DESC").Find(&logs)
	return logs, result.Error
}

// GetByID 根据ID获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetByID(id uint) (*HeadlingAuthLog, error) {
	var log HeadlingAuthLog
	result := r.DB.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetByAuthID 根据认证ID获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetByAuthID(authID string) (*HeadlingAuthLog, error) {
	var log HeadlingAuthLog
	result := r.DB.Where("auth_id = ?", authID).First(&log)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetBySessionID 根据会话ID获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetBySessionID(sessionID string) ([]HeadlingAuthLog, error) {
	var logs []HeadlingAuthLog
	result := r.DB.Where("session_id = ?", sessionID).Order("timestamp ASC").Find(&logs)
	return logs, result.Error
}

// GetBySourceIP 根据源IP获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetBySourceIP(sourceIP string) ([]HeadlingAuthLog, error) {
	var logs []HeadlingAuthLog
	result := r.DB.Where("source_ip = ?", sourceIP).Order("timestamp DESC").Find(&logs)
	return logs, result.Error
}

// GetByContainerID 根据容器ID获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetByContainerID(containerID string) ([]HeadlingAuthLog, error) {
	var logs []HeadlingAuthLog
	result := r.DB.Where("container_id = ?", containerID).Order("timestamp DESC").Find(&logs)
	return logs, result.Error
}

// GetByProtocol 根据协议获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetByProtocol(protocol string) ([]HeadlingAuthLog, error) {
	var logs []HeadlingAuthLog
	result := r.DB.Where("protocol = ?", protocol).Order("timestamp DESC").Find(&logs)
	return logs, result.Error
}

// GetByTimeRange 根据时间范围获取Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) GetByTimeRange(startTime, endTime time.Time) ([]HeadlingAuthLog, error) {
	var logs []HeadlingAuthLog
	result := r.DB.Where("timestamp BETWEEN ? AND ?", startTime, endTime).Order("timestamp DESC").Find(&logs)
	return logs, result.Error
}

// Create 创建Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) Create(log *HeadlingAuthLog) error {
	log.CreatedAt = time.Now()
	return r.DB.Create(log).Error
}

// CreateBatch 批量创建Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) CreateBatch(logs []HeadlingAuthLog) error {
	now := time.Now()
	for i := range logs {
		logs[i].CreatedAt = now
	}
	return r.DB.CreateInBatches(logs, 100).Error
}

// Update 更新Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) Update(log *HeadlingAuthLog) error {
	return r.DB.Save(log).Error
}

// Delete 删除Headling认证日志
func (r *MySQLHeadlingAuthLogRepo) Delete(id uint) error {
	return r.DB.Delete(&HeadlingAuthLog{}, id).Error
}

// DeleteByContainerID 根据容器ID删除所有相关认证日志
func (r *MySQLHeadlingAuthLogRepo) DeleteByContainerID(containerID string) error {
	return r.DB.Where("container_id = ?", containerID).Delete(&HeadlingAuthLog{}).Error
}

// GetStatistics 获取认证统计信息
func (r *MySQLHeadlingAuthLogRepo) GetStatistics() ([]HeadlingAuthStatistics, error) {
	var stats []HeadlingAuthStatistics
	result := r.DB.Table("v_headling_auth_statistics").Find(&stats)
	return stats, result.Error
}

// GetAttackerIPStatistics 获取攻击者IP统计信息
func (r *MySQLHeadlingAuthLogRepo) GetAttackerIPStatistics() ([]AttackerIPStatistics, error) {
	var stats []AttackerIPStatistics
	result := r.DB.Table("v_attacker_ip_statistics").Find(&stats)
	return stats, result.Error
}

// GetTopAttackers 获取前N个攻击者
func (r *MySQLHeadlingAuthLogRepo) GetTopAttackers(limit int) ([]AttackerIPStatistics, error) {
	var stats []AttackerIPStatistics
	result := r.DB.Table("v_attacker_ip_statistics").Limit(limit).Find(&stats)
	return stats, result.Error
}

// GetTopUsernames 获取最常用的用户名
func (r *MySQLHeadlingAuthLogRepo) GetTopUsernames(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	result := r.DB.Table("headling_auth_log").
		Select("username, COUNT(*) as count, COUNT(DISTINCT source_ip) as unique_ips").
		Group("username").
		Order("count DESC").
		Limit(limit).
		Find(&results)
	return results, result.Error
}

// GetTopPasswords 获取最常用的密码
func (r *MySQLHeadlingAuthLogRepo) GetTopPasswords(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	result := r.DB.Table("headling_auth_log").
		Select("password, COUNT(*) as count, COUNT(DISTINCT source_ip) as unique_ips").
		Group("password").
		Order("count DESC").
		Limit(limit).
		Find(&results)
	return results, result.Error
}

// -------------------- Cowrie日志仓库 --------------------

// MySQLCowrieLogRepo Cowrie日志MySQL仓库
type MySQLCowrieLogRepo struct {
	DB *gorm.DB
}

// NewMySQLCowrieLogRepo 创建Cowrie日志MySQL仓库
func NewMySQLCowrieLogRepo(db *gorm.DB) CowrieLogRepository {
	return &MySQLCowrieLogRepo{DB: db}
}

// List 获取所有Cowrie日志
func (r *MySQLCowrieLogRepo) List() ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByID 根据ID获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByID(id uint) (*CowrieLog, error) {
	var log CowrieLog
	result := r.DB.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetByAuthID 根据认证ID获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByAuthID(authID string) (*CowrieLog, error) {
	var log CowrieLog
	result := r.DB.Where("auth_id = ?", authID).First(&log)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// GetBySessionID 根据会话ID获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetBySessionID(sessionID string) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("session_id = ?", sessionID).Order("event_time ASC").Find(&logs)
	return logs, result.Error
}

// GetBySourceIP 根据源IP获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetBySourceIP(sourceIP string) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("source_ip = ?", sourceIP).Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByContainerID 根据容器ID获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByContainerID(containerID string) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("container_id = ?", containerID).Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByProtocol 根据协议获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByProtocol(protocol string) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("protocol = ?", protocol).Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByTimeRange 根据时间范围获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByTimeRange(startTime, endTime time.Time) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("event_time BETWEEN ? AND ?", startTime, endTime).Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByCommand 根据命令获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByCommand(command string) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("command LIKE ?", "%"+command+"%").Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByCommandFound 根据命令是否被识别获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByCommandFound(found bool) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("command_found = ?", found).Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// GetByUsername 根据用户名获取Cowrie日志
func (r *MySQLCowrieLogRepo) GetByUsername(username string) ([]CowrieLog, error) {
	var logs []CowrieLog
	result := r.DB.Where("username = ?", username).Order("event_time DESC").Find(&logs)
	return logs, result.Error
}

// Create 创建Cowrie日志
func (r *MySQLCowrieLogRepo) Create(log *CowrieLog) error {
	log.CreatedAt = time.Now()
	return r.DB.Create(log).Error
}

// CreateBatch 批量创建Cowrie日志
func (r *MySQLCowrieLogRepo) CreateBatch(logs []CowrieLog) error {
	now := time.Now()
	for i := range logs {
		logs[i].CreatedAt = now
	}
	return r.DB.CreateInBatches(logs, 100).Error
}

// Update 更新Cowrie日志
func (r *MySQLCowrieLogRepo) Update(log *CowrieLog) error {
	return r.DB.Save(log).Error
}

// Delete 删除Cowrie日志
func (r *MySQLCowrieLogRepo) Delete(id uint) error {
	return r.DB.Delete(&CowrieLog{}, id).Error
}

// DeleteByContainerID 根据容器ID删除所有相关Cowrie日志
func (r *MySQLCowrieLogRepo) DeleteByContainerID(containerID string) error {
	return r.DB.Where("container_id = ?", containerID).Delete(&CowrieLog{}).Error
}

// GetStatistics 获取Cowrie统计信息
func (r *MySQLCowrieLogRepo) GetStatistics() ([]CowrieStatistics, error) {
	var stats []CowrieStatistics
	result := r.DB.Table("v_cowrie_statistics").Find(&stats)
	return stats, result.Error
}

// GetAttackerBehavior 获取攻击者行为统计信息
func (r *MySQLCowrieLogRepo) GetAttackerBehavior() ([]CowrieAttackerBehavior, error) {
	var behavior []CowrieAttackerBehavior
	result := r.DB.Table("v_cowrie_attacker_behavior").Find(&behavior)
	return behavior, result.Error
}

// GetTopAttackers 获取前N个攻击者
func (r *MySQLCowrieLogRepo) GetTopAttackers(limit int) ([]CowrieAttackerBehavior, error) {
	var attackers []CowrieAttackerBehavior
	result := r.DB.Table("v_cowrie_attacker_behavior").Limit(limit).Find(&attackers)
	return attackers, result.Error
}

// GetCommandStatistics 获取命令统计信息
func (r *MySQLCowrieLogRepo) GetCommandStatistics() ([]CowrieCommandStatistics, error) {
	var stats []CowrieCommandStatistics
	result := r.DB.Table("v_cowrie_command_statistics").Find(&stats)
	return stats, result.Error
}

// GetTopCommands 获取最常用的命令
func (r *MySQLCowrieLogRepo) GetTopCommands(limit int) ([]CowrieCommandStatistics, error) {
	var commands []CowrieCommandStatistics
	result := r.DB.Table("v_cowrie_command_statistics").Limit(limit).Find(&commands)
	return commands, result.Error
}

// GetTopUsernames 获取最常用的用户名
func (r *MySQLCowrieLogRepo) GetTopUsernames(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	result := r.DB.Table("cowrie_log").
		Select("username, COUNT(*) as count, COUNT(DISTINCT source_ip) as unique_ips").
		Where("username IS NOT NULL AND username != ''").
		Group("username").
		Order("count DESC").
		Limit(limit).
		Find(&results)
	return results, result.Error
}

// GetTopPasswords 获取最常用的密码
func (r *MySQLCowrieLogRepo) GetTopPasswords(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	result := r.DB.Table("cowrie_log").
		Select("password, COUNT(*) as count, COUNT(DISTINCT source_ip) as unique_ips").
		Where("password IS NOT NULL AND password != ''").
		Group("password").
		Order("count DESC").
		Limit(limit).
		Find(&results)
	return results, result.Error
}

// GetTopFingerprints 获取最常用的指纹
func (r *MySQLCowrieLogRepo) GetTopFingerprints(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	result := r.DB.Table("cowrie_log").
		Select("fingerprint, COUNT(*) as count, COUNT(DISTINCT source_ip) as unique_ips").
		Where("fingerprint IS NOT NULL AND fingerprint != ''").
		Group("fingerprint").
		Order("count DESC").
		Limit(limit).
		Find(&results)
	return results, result.Error
}
