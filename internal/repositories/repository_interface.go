package repositories

// HoneypotTemplateRepository 蜜罐模板仓库接口
type HoneypotTemplateRepository interface {
	GetAll() ([]HoneypotTemplate, error)
	GetByID(id uint) (*HoneypotTemplate, error)
	Create(template *HoneypotTemplate) error
	Update(template *HoneypotTemplate) error
	Delete(id uint) error
	IncrementDeployCount(id uint) error
	DecrementDeployCount(id uint) error
}

// HoneypotInstanceRepository 蜜罐实例仓库接口
type HoneypotInstanceRepository interface {
	GetAll() ([]HoneypotInstance, error)
	GetByID(id uint) (*HoneypotInstance, error)
	GetByTemplateID(templateID uint) ([]HoneypotInstance, error)
	Create(instance *HoneypotInstance) error
	Update(instance *HoneypotInstance) error
	UpdateStatus(id uint, status string) error
	Delete(id uint) error
}

// HoneypotLogRepository 蜜罐日志仓库接口
type HoneypotLogRepository interface {
	GetAll() ([]HoneypotLog, error)
	GetByID(id uint) (*HoneypotLog, error)
	GetByInstanceID(instanceID uint) ([]HoneypotLog, error)
	Create(log *HoneypotLog) error
	Delete(id uint) error
}

// BaitRepository 诱饵仓库接口
type BaitRepository interface {
	GetAll() ([]Bait, error)
	GetByID(id uint) (*Bait, error)
	GetByInstanceID(instanceID uint) ([]Bait, error)
	Create(bait *Bait) error
	Update(bait *Bait) error
	UpdateDeployStatus(id uint, isDeployed bool) error
	Delete(id uint) error
}

// SecurityRuleRepository 安全规则仓库接口
type SecurityRuleRepository interface {
	GetAll() ([]SecurityRule, error)
	GetEnabled() ([]SecurityRule, error)
	GetByID(id uint) (*SecurityRule, error)
	Create(rule *SecurityRule) error
	Update(rule *SecurityRule) error
	UpdateStatus(id uint, isEnabled bool) error
	Delete(id uint) error
}

// RuleLogRepository 规则日志仓库接口
type RuleLogRepository interface {
	GetAll() ([]RuleLog, error)
	GetByID(id uint) (*RuleLog, error)
	GetByRuleID(ruleID uint) ([]RuleLog, error)
	Create(log *RuleLog) error
	Delete(id uint) error
}
