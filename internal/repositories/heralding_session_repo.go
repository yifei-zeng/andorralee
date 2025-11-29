package repositories

import (
	"time"

	"gorm.io/gorm"
)

type mysqlHeraldingSessionLogRepo struct {
	db *gorm.DB
}

// NewMySQLHeraldingSessionLogRepo 创建MySQL Heralding会话日志仓库
func NewMySQLHeraldingSessionLogRepo(db *gorm.DB) HeraldingSessionLogRepository {
	return &mysqlHeraldingSessionLogRepo{db: db}
}

func (r *mysqlHeraldingSessionLogRepo) List() ([]HeraldingSessionLog, error) {
	var logs []HeraldingSessionLog
	err := r.db.Order("timestamp desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlHeraldingSessionLogRepo) GetByID(id uint) (*HeraldingSessionLog, error) {
	var log HeraldingSessionLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *mysqlHeraldingSessionLogRepo) GetBySessionID(sessionID string) (*HeraldingSessionLog, error) {
	var log HeraldingSessionLog
	err := r.db.Where("session_id = ?", sessionID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *mysqlHeraldingSessionLogRepo) GetBySourceIP(sourceIP string) ([]HeraldingSessionLog, error) {
	var logs []HeraldingSessionLog
	err := r.db.Where("source_ip = ?", sourceIP).Order("timestamp desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlHeraldingSessionLogRepo) GetByContainerID(containerID string) ([]HeraldingSessionLog, error) {
	var logs []HeraldingSessionLog
	err := r.db.Where("container_id = ?", containerID).Order("timestamp desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlHeraldingSessionLogRepo) GetLatestByContainerID(containerID string) (*HeraldingSessionLog, error) {
	var log HeraldingSessionLog
	err := r.db.Where("container_id = ?", containerID).Order("timestamp desc").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *mysqlHeraldingSessionLogRepo) GetByProtocol(protocol string) ([]HeraldingSessionLog, error) {
	var logs []HeraldingSessionLog
	err := r.db.Where("protocol = ?", protocol).Order("timestamp desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlHeraldingSessionLogRepo) GetByTimeRange(startTime, endTime time.Time) ([]HeraldingSessionLog, error) {
	var logs []HeraldingSessionLog
	err := r.db.Where("timestamp BETWEEN ? AND ?", startTime, endTime).Order("timestamp desc").Find(&logs).Error
	return logs, err
}

func (r *mysqlHeraldingSessionLogRepo) Create(log *HeraldingSessionLog) error {
	return r.db.Create(log).Error
}

func (r *mysqlHeraldingSessionLogRepo) CreateBatch(logs []HeraldingSessionLog) error {
	return r.db.CreateInBatches(logs, 100).Error
}

func (r *mysqlHeraldingSessionLogRepo) Update(log *HeraldingSessionLog) error {
	return r.db.Save(log).Error
}

func (r *mysqlHeraldingSessionLogRepo) Delete(id uint) error {
	return r.db.Delete(&HeraldingSessionLog{}, id).Error
}

func (r *mysqlHeraldingSessionLogRepo) DeleteByContainerID(containerID string) error {
	return r.db.Where("container_id = ?", containerID).Delete(&HeraldingSessionLog{}).Error
}
