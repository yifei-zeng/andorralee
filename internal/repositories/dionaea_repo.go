package repositories

import (
	"time"

	"gorm.io/gorm"
)

// DionaeaLogRepository 定义 Dionaea 日志仓库接口
// 提供基础增删查与按容器/源IP/时间范围查询

type DionaeaLogRepository interface {
	Create(log *DionaeaLog) error
	CreateBatch(logs []DionaeaLog) error
	List() ([]DionaeaLog, error)
	GetByID(id uint) (*DionaeaLog, error)
	GetByContainerID(containerID string) ([]DionaeaLog, error)
	GetBySourceIP(sourceIP string) ([]DionaeaLog, error)
	GetByProtocol(protocol string) ([]DionaeaLog, error)
	GetByTimeRange(start, end time.Time) ([]DionaeaLog, error)
	GetLatestByContainerID(containerID string) (*DionaeaLog, error)
	DeleteByContainerID(containerID string) error
}

type mysqlDionaeaLogRepo struct {
	db *gorm.DB
}

func NewMySQLDionaeaLogRepo(db *gorm.DB) DionaeaLogRepository {
	return &mysqlDionaeaLogRepo{db: db}
}

func (r *mysqlDionaeaLogRepo) Create(log *DionaeaLog) error {
	return r.db.Create(log).Error
}

func (r *mysqlDionaeaLogRepo) CreateBatch(logs []DionaeaLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(logs, 200).Error
}

func (r *mysqlDionaeaLogRepo) List() ([]DionaeaLog, error) {
	var logs []DionaeaLog
	err := r.db.Order("event_time desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlDionaeaLogRepo) GetByID(id uint) (*DionaeaLog, error) {
	var log DionaeaLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *mysqlDionaeaLogRepo) GetByContainerID(containerID string) ([]DionaeaLog, error) {
	var logs []DionaeaLog
	err := r.db.Where("container_id = ?", containerID).Order("event_time desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlDionaeaLogRepo) GetBySourceIP(sourceIP string) ([]DionaeaLog, error) {
	var logs []DionaeaLog
	err := r.db.Where("source_ip = ?", sourceIP).Order("event_time desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlDionaeaLogRepo) GetByProtocol(protocol string) ([]DionaeaLog, error) {
	var logs []DionaeaLog
	err := r.db.Where("protocol = ?", protocol).Order("event_time desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlDionaeaLogRepo) GetByTimeRange(start, end time.Time) ([]DionaeaLog, error) {
	var logs []DionaeaLog
	err := r.db.Where("event_time BETWEEN ? AND ?", start, end).Order("event_time desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlDionaeaLogRepo) GetLatestByContainerID(containerID string) (*DionaeaLog, error) {
	var log DionaeaLog
	err := r.db.Where("container_id = ?", containerID).Order("event_time desc").Limit(1).Find(&log).Error
	if err != nil {
		return nil, err
	}
	if log.ID == 0 {
		return nil, nil
	}
	return &log, nil
}

func (r *mysqlDionaeaLogRepo) DeleteByContainerID(containerID string) error {
	return r.db.Where("container_id = ?", containerID).Delete(&DionaeaLog{}).Error
}
