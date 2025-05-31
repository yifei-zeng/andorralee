package repositories

import (
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

// GetAll 获取所有蜜罐模板
func (r *MySQLHoneypotTemplateRepo) GetAll() ([]HoneypotTemplate, error) {
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

// GetAll 获取所有蜜罐实例
func (r *MySQLHoneypotInstanceRepo) GetAll() ([]HoneypotInstance, error) {
	var instances []HoneypotInstance
	result := r.DB.Find(&instances)
	return instances, result.Error
}

// GetByID 根据ID获取蜜罐实例
func (r *MySQLHoneypotInstanceRepo) GetByID(id uint) (*HoneypotInstance, error) {
	var instance HoneypotInstance
	result := r.DB.First(&instance, id)
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
	return r.DB.Create(instance).Error
}

// Update 更新蜜罐实例
func (r *MySQLHoneypotInstanceRepo) Update(instance *HoneypotInstance) error {
	return r.DB.Save(instance).Error
}

// UpdateStatus 更新蜜罐实例状态
func (r *MySQLHoneypotInstanceRepo) UpdateStatus(id uint, status string) error {
	return r.DB.Model(&HoneypotInstance{}).Where("id = ?", id).
		Update("status", status).Error
}

// Delete 删除蜜罐实例
func (r *MySQLHoneypotInstanceRepo) Delete(id uint) error {
	return r.DB.Delete(&HoneypotInstance{}, id).Error
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

// GetAll 获取所有诱饵
func (r *MySQLBaitRepo) GetAll() ([]Bait, error) {
	var baits []Bait
	result := r.DB.Find(&baits)
	return baits, result.Error
}

// GetByID 根据ID获取诱饵
func (r *MySQLBaitRepo) GetByID(id uint) (*Bait, error) {
	var bait Bait
	result := r.DB.First(&bait, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &bait, nil
}

// GetByInstanceID 根据实例ID获取诱饵
func (r *MySQLBaitRepo) GetByInstanceID(instanceID uint) ([]Bait, error) {
	var baits []Bait
	result := r.DB.Where("instance_id = ?", instanceID).Find(&baits)
	return baits, result.Error
}

// Create 创建诱饵
func (r *MySQLBaitRepo) Create(bait *Bait) error {
	return r.DB.Create(bait).Error
}

// Update 更新诱饵
func (r *MySQLBaitRepo) Update(bait *Bait) error {
	return r.DB.Save(bait).Error
}

// UpdateDeployStatus 更新诱饵部署状态
func (r *MySQLBaitRepo) UpdateDeployStatus(id uint, isDeployed bool) error {
	return r.DB.Model(&Bait{}).Where("id = ?", id).
		Update("is_deployed", isDeployed).Error
}

// Delete 删除诱饵
func (r *MySQLBaitRepo) Delete(id uint) error {
	return r.DB.Delete(&Bait{}, id).Error
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

// GetAll 获取所有安全规则
func (r *MySQLSecurityRuleRepo) GetAll() ([]SecurityRule, error) {
	var rules []SecurityRule
	result := r.DB.Find(&rules)
	return rules, result.Error
}

// GetEnabled 获取所有启用的安全规则
func (r *MySQLSecurityRuleRepo) GetEnabled() ([]SecurityRule, error) {
	var rules []SecurityRule
	result := r.DB.Where("is_enabled = ?", true).Find(&rules)
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

// UpdateStatus 更新安全规则状态
func (r *MySQLSecurityRuleRepo) UpdateStatus(id uint, isEnabled bool) error {
	return r.DB.Model(&SecurityRule{}).Where("id = ?", id).
		Update("is_enabled", isEnabled).Error
}

// Delete 删除安全规则
func (r *MySQLSecurityRuleRepo) Delete(id uint) error {
	return r.DB.Delete(&SecurityRule{}, id).Error
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
