package repositories

import "time"

// HoneypotTemplateRepository 蜜罐模板仓库接口
type HoneypotTemplateRepository interface {
	List() ([]HoneypotTemplate, error)
	GetByID(id uint) (*HoneypotTemplate, error)
	Create(template *HoneypotTemplate) error
	Update(template *HoneypotTemplate) error
	Delete(id uint) error
	IncrementDeployCount(id uint) error
	DecrementDeployCount(id uint) error
}

// HoneypotInstanceRepository 蜜罐实例仓库接口
type HoneypotInstanceRepository interface {
	List() ([]HoneypotInstance, error)
	GetByID(id uint) (*HoneypotInstance, error)
	Create(instance *HoneypotInstance) error
	Update(instance *HoneypotInstance) error
	Delete(id uint) error
	UpdateStatus(id uint, status string) error
	GetByTemplateID(templateID uint) ([]HoneypotInstance, error)
	GetByStatus(status string) ([]HoneypotInstance, error)
	GetByContainerID(containerID string) (*HoneypotInstance, error)
}

// HoneypotLogRepository 蜜罐日志仓库接口
type HoneypotLogRepository interface {
	List() ([]HoneypotLog, error)
	GetByID(id uint) (*HoneypotLog, error)
	GetByInstanceID(instanceID uint) ([]HoneypotLog, error)
	Create(log *HoneypotLog) error
	Delete(id uint) error
}

// SecurityRuleRepository 安全规则仓库接口
type SecurityRuleRepository interface {
	List() ([]SecurityRule, error)
	GetByID(id uint) (*SecurityRule, error)
	Create(rule *SecurityRule) error
	Update(rule *SecurityRule) error
	Delete(id uint) error
	UpdateStatus(id uint, isEnabled bool) error
}

// RuleLogRepository 规则日志仓库接口
type RuleLogRepository interface {
	List() ([]RuleLog, error)
	GetByID(id uint) (*RuleLog, error)
	GetByRuleID(ruleID uint) ([]RuleLog, error)
	Create(log *RuleLog) error
	Delete(id uint) error
}

// DockerImageRepository Docker镜像仓库接口
type DockerImageRepository interface {
	List() ([]DockerImage, error)
	GetByID(id uint) (*DockerImage, error)
	GetByImageID(imageID string) (*DockerImage, error)
	Create(image *DockerImage) error
	Update(image *DockerImage) error
	Delete(id uint) error
	DeleteByImageID(imageID string) error
}

// DockerImageLogRepository Docker镜像日志仓库接口
type DockerImageLogRepository interface {
	List() ([]DockerImageLog, error)
	GetByID(id uint) (*DockerImageLog, error)
	GetByImageID(imageID string) ([]DockerImageLog, error)
	Create(log *DockerImageLog) error
	Delete(id uint) error
}

// ContainerLogSegmentRepository 容器日志分析仓库接口
type ContainerLogSegmentRepository interface {
	List() ([]ContainerLogSegment, error)
	GetByID(id uint) (*ContainerLogSegment, error)
	GetByContainerID(containerID string) ([]ContainerLogSegment, error)
	GetBySegmentType(segmentType string) ([]ContainerLogSegment, error)
	Create(segment *ContainerLogSegment) error
	CreateBatch(segments []ContainerLogSegment) error
	Delete(id uint) error
	DeleteByContainerID(containerID string) error
}

// DockerContainerRepository Docker容器仓库接口
type DockerContainerRepository interface {
	List() ([]DockerContainer, error)
	GetByID(id uint) (*DockerContainer, error)
	GetByContainerID(containerID string) (*DockerContainer, error)
	GetByImageID(imageID string) ([]DockerContainer, error)
	Create(container *DockerContainer) error
	Update(container *DockerContainer) error
	Delete(id uint) error
	DeleteByContainerID(containerID string) error
	UpdateStatus(containerID string, status string) error
}

// HeraldingAuthLogRepository Heralding认证日志仓库接口
type HeraldingAuthLogRepository interface {
	List() ([]HeraldingAuthLog, error)
	GetByID(id uint) (*HeraldingAuthLog, error)
	GetByAuthID(authID string) (*HeraldingAuthLog, error)
	GetBySessionID(sessionID string) ([]HeraldingAuthLog, error)
	GetBySourceIP(sourceIP string) ([]HeraldingAuthLog, error)
	GetByContainerID(containerID string) ([]HeraldingAuthLog, error)
	GetLatestByContainerID(containerID string) (*HeraldingAuthLog, error)
	GetByProtocol(protocol string) ([]HeraldingAuthLog, error)
	GetByTimeRange(startTime, endTime time.Time) ([]HeraldingAuthLog, error)
	Create(log *HeraldingAuthLog) error
	CreateBatch(logs []HeraldingAuthLog) error
	Update(log *HeraldingAuthLog) error
	Delete(id uint) error
	DeleteByContainerID(containerID string) error
	GetStatistics() ([]HeraldingAuthStatistics, error)
	GetAttackerIPStatistics() ([]AttackerIPStatistics, error)
	GetTopAttackers(limit int) ([]AttackerIPStatistics, error)
	GetTopUsernames(limit int) ([]map[string]interface{}, error)
	GetTopPasswords(limit int) ([]map[string]interface{}, error)
}

// HeraldingSessionLogRepository Heralding会话日志仓库接口
type HeraldingSessionLogRepository interface {
	List() ([]HeraldingSessionLog, error)
	GetByID(id uint) (*HeraldingSessionLog, error)
	GetBySessionID(sessionID string) (*HeraldingSessionLog, error)
	GetBySourceIP(sourceIP string) ([]HeraldingSessionLog, error)
	GetByContainerID(containerID string) ([]HeraldingSessionLog, error)
	GetLatestByContainerID(containerID string) (*HeraldingSessionLog, error)
	GetByProtocol(protocol string) ([]HeraldingSessionLog, error)
	GetByTimeRange(startTime, endTime time.Time) ([]HeraldingSessionLog, error)
	Create(log *HeraldingSessionLog) error
	CreateBatch(logs []HeraldingSessionLog) error
	Update(log *HeraldingSessionLog) error
	Delete(id uint) error
	DeleteByContainerID(containerID string) error
}

// CowrieLogRepository Cowrie蜜罐日志仓库接口
type CowrieLogRepository interface {
	List() ([]CowrieLog, error)
	GetByID(id uint) (*CowrieLog, error)
	GetByAuthID(authID string) (*CowrieLog, error)
	GetBySessionID(sessionID string) ([]CowrieLog, error)
	GetBySourceIP(sourceIP string) ([]CowrieLog, error)
	GetByContainerID(containerID string) ([]CowrieLog, error)
	GetLatestByContainerID(containerID string) (*CowrieLog, error)
	GetByProtocol(protocol string) ([]CowrieLog, error)
	GetByTimeRange(startTime, endTime time.Time) ([]CowrieLog, error)
	GetByCommand(command string) ([]CowrieLog, error)
	GetByCommandFound(found bool) ([]CowrieLog, error)
	GetByUsername(username string) ([]CowrieLog, error)
	Create(log *CowrieLog) error
	CreateBatch(logs []CowrieLog) error
	Update(log *CowrieLog) error
	Delete(id uint) error
	DeleteByContainerID(containerID string) error
	GetStatistics() ([]CowrieStatistics, error)
	GetAttackerBehavior() ([]CowrieAttackerBehavior, error)
	GetTopAttackers(limit int) ([]CowrieAttackerBehavior, error)
	GetCommandStatistics() ([]CowrieCommandStatistics, error)
	GetTopCommands(limit int) ([]CowrieCommandStatistics, error)
	GetTopUsernames(limit int) ([]map[string]interface{}, error)
	GetTopPasswords(limit int) ([]map[string]interface{}, error)
	GetTopFingerprints(limit int) ([]map[string]interface{}, error)
}

// MySQLHoneypotLogRepository MySQL蜜罐日志仓库接口
type MySQLHoneypotLogRepository interface {
	List() ([]MySQLHoneypotLog, error)
	GetByID(id uint) (*MySQLHoneypotLog, error)
	GetByEventID(eventID string) (*MySQLHoneypotLog, error)
	GetByContainerID(containerID string) ([]MySQLHoneypotLog, error)
	GetLatestByContainerID(containerID string) (*MySQLHoneypotLog, error)
	GetBySourceIP(sourceIP string) ([]MySQLHoneypotLog, error)
	GetByDestinationIP(destinationIP string) ([]MySQLHoneypotLog, error)
	GetByUsername(username string) ([]MySQLHoneypotLog, error)
	GetByDatabaseName(databaseName string) ([]MySQLHoneypotLog, error)
	GetByTimeRange(startTime, endTime time.Time) ([]MySQLHoneypotLog, error)
	GetByQueryKeyword(keyword string, limit int) ([]MySQLHoneypotLog, error)
	GetByErrorCode(errorCode string) ([]MySQLHoneypotLog, error)
	Search(filter MySQLHoneypotSearchFilter) ([]MySQLHoneypotLog, error)
	Create(log *MySQLHoneypotLog) error
	CreateBatch(logs []MySQLHoneypotLog) error
	Delete(id uint) error
	DeleteByContainerID(containerID string) error
	GetQueryStatistics(limit int) ([]MySQLHoneypotStatistics, error)
}
