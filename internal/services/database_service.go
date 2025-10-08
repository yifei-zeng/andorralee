package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DatabaseService 数据库服务接口
type DatabaseService interface {
	// 病毒检测相关
	SaveMalwareSignature(signature *repositories.MalwareSignature) error
	GetMalwareSignatures() ([]repositories.MalwareSignature, error)
	SaveScanResult(result *repositories.ScanResult) error
	SaveDetectionResult(detection *repositories.DetectionResult) error
	GetScanResult(fileHash string) (*repositories.ScanResult, error)
	GetScanHistory(limit int) ([]repositories.ScanResult, error)
	// 新增：按任意哈希（SHA256或MD5）查询扫描结果
	GetScanResultByAnyHash(hash string) (*repositories.ScanResult, error)
	// 新增：按ID获取与更新扫描结果
	GetScanByID(id uint) (*repositories.ScanResult, error)
	UpdateScanResult(id uint, updates map[string]interface{}) error

	// 攻击会话相关
	SaveAttackSession(session *repositories.AttackSession) error
	UpdateAttackSession(sessionID string, updates map[string]interface{}) error
	GetAttackSession(sessionID string) (*repositories.AttackSession, error)
	SaveAttackEvent(event *repositories.AttackEvent) error
	GetAttackEvents(sessionID string) ([]repositories.AttackEvent, error)

	// 威胁情报相关
	SaveThreatIntelligence(threat *repositories.ThreatIntelligence) error
	GetThreatIntelligence(indicatorType, indicatorValue string) (*repositories.ThreatIntelligence, error)

	// 蜜签事件相关
	SaveHoneytokenEvent(event *repositories.HoneytokenEvent) error
	GetHoneytokenEvents(limit int) ([]repositories.HoneytokenEvent, error)

	// 统计查询
	GetAttackStatistics() (map[string]interface{}, error)
	GetThreatStatistics() (map[string]interface{}, error)
}

// MySQLService MySQL数据库服务实现
type MySQLService struct {
	db *gorm.DB
}

// DamengService 达梦数据库服务实现
type DamengService struct {
	db *gorm.DB
}

// NewMySQLService 创建MySQL服务实例
func NewMySQLService() DatabaseService {
	return &MySQLService{db: config.MySQLDB}
}

// NewDamengService 创建达梦数据库服务实例
func NewDamengService() DatabaseService {
	return &DamengService{db: config.DamengDB}
}

// GetDatabaseService 根据配置获取数据库服务
func GetDatabaseService() DatabaseService {
	// 优先使用MySQL，如果不可用则使用达梦
	if config.MySQLDB != nil {
		return NewMySQLService()
	}
	if config.DamengDB != nil {
		return NewDamengService()
	}
	return nil
}

// MySQL服务实现

func (s *MySQLService) SaveMalwareSignature(signature *repositories.MalwareSignature) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	signature.CreateTime = time.Now()
	signature.UpdateTime = time.Now()
	return s.db.Create(signature).Error
}

func (s *MySQLService) GetMalwareSignatures() ([]repositories.MalwareSignature, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var signatures []repositories.MalwareSignature
	err := s.db.Where("is_active = ?", true).Find(&signatures).Error
	return signatures, err
}

func (s *MySQLService) SaveScanResult(result *repositories.ScanResult) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	result.ScanTime = time.Now()
	return s.db.Create(result).Error
}

func (s *MySQLService) SaveDetectionResult(detection *repositories.DetectionResult) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	detection.CreatedAt = time.Now()
	return s.db.Create(detection).Error
}

func (s *MySQLService) GetScanResult(fileHash string) (*repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var result repositories.ScanResult
	err := s.db.Where("file_hash = ?", fileHash).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MySQLService) GetScanHistory(limit int) ([]repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var results []repositories.ScanResult
	err := s.db.Order("scan_time DESC").Limit(limit).Find(&results).Error
	return results, err
}

func (s *MySQLService) GetScanResultByAnyHash(hash string) (*repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var result repositories.ScanResult
	err := s.db.Where("file_hash = ? OR md5_hash = ?", hash, hash).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MySQLService) GetScanByID(id uint) (*repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var result repositories.ScanResult
	err := s.db.First(&result, id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MySQLService) UpdateScanResult(id uint, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	return s.db.Model(&repositories.ScanResult{}).Where("id = ?", id).Updates(updates).Error
}

func (s *MySQLService) SaveAttackSession(session *repositories.AttackSession) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	session.StartTime = time.Now()
	return s.db.Create(session).Error
}

func (s *MySQLService) UpdateAttackSession(sessionID string, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	return s.db.Model(&repositories.AttackSession{}).Where("session_id = ?", sessionID).Updates(updates).Error
}

func (s *MySQLService) GetAttackSession(sessionID string) (*repositories.AttackSession, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var session repositories.AttackSession
	err := s.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *MySQLService) SaveAttackEvent(event *repositories.AttackEvent) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	event.EventTime = time.Now()
	return s.db.Create(event).Error
}

func (s *MySQLService) GetAttackEvents(sessionID string) ([]repositories.AttackEvent, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var events []repositories.AttackEvent
	err := s.db.Where("session_id = ?", sessionID).Order("event_time ASC").Find(&events).Error
	return events, err
}

func (s *MySQLService) SaveThreatIntelligence(threat *repositories.ThreatIntelligence) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	threat.CreatedAt = time.Now()
	threat.UpdatedAt = time.Now()
	return s.db.Create(threat).Error
}

func (s *MySQLService) GetThreatIntelligence(indicatorType, indicatorValue string) (*repositories.ThreatIntelligence, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var threat repositories.ThreatIntelligence
	err := s.db.Where("indicator_type = ? AND indicator_value = ? AND is_active = ?",
		indicatorType, indicatorValue, true).First(&threat).Error
	if err != nil {
		return nil, err
	}
	return &threat, nil
}

func (s *MySQLService) SaveHoneytokenEvent(event *repositories.HoneytokenEvent) error {
	if s.db == nil {
		return errors.New("MySQL数据库未初始化")
	}
	event.TriggerTime = time.Now()
	return s.db.Create(event).Error
}

func (s *MySQLService) GetHoneytokenEvents(limit int) ([]repositories.HoneytokenEvent, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}
	var events []repositories.HoneytokenEvent
	err := s.db.Order("trigger_time DESC").Limit(limit).Find(&events).Error
	return events, err
}

func (s *MySQLService) GetAttackStatistics() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}

	stats := make(map[string]interface{})

	// 总攻击会话数
	var totalSessions int64
	s.db.Model(&repositories.AttackSession{}).Count(&totalSessions)
	stats["total_sessions"] = totalSessions

	// 活跃会话数
	var activeSessions int64
	s.db.Model(&repositories.AttackSession{}).Where("status = ?", "active").Count(&activeSessions)
	stats["active_sessions"] = activeSessions

	// 今日新增会话
	var todaySessions int64
	today := time.Now().Format("2006-01-02")
	s.db.Model(&repositories.AttackSession{}).Where("DATE(start_time) = ?", today).Count(&todaySessions)
	stats["today_sessions"] = todaySessions

	// 按IP统计
	var ipStats []map[string]interface{}
	s.db.Model(&repositories.AttackSession{}).
		Select("source_ip, COUNT(*) as session_count, MAX(start_time) as last_seen").
		Group("source_ip").
		Order("session_count DESC").
		Limit(10).
		Scan(&ipStats)
	stats["top_attackers"] = ipStats

	return stats, nil
}

func (s *MySQLService) GetThreatStatistics() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, errors.New("MySQL数据库未初始化")
	}

	stats := make(map[string]interface{})

	// 扫描统计
	var totalScans int64
	s.db.Model(&repositories.ScanResult{}).Count(&totalScans)
	stats["total_scans"] = totalScans

	// 感染文件数
	var infectedFiles int64
	s.db.Model(&repositories.ScanResult{}).Where("is_infected = ?", true).Count(&infectedFiles)
	stats["infected_files"] = infectedFiles

	// 威胁等级分布
	var threatDistribution []map[string]interface{}
	s.db.Model(&repositories.ScanResult{}).
		Select("threat_level, COUNT(*) as count").
		Where("is_infected = ?", true).
		Group("threat_level").
		Scan(&threatDistribution)
	stats["threat_distribution"] = threatDistribution

	// 蜜签触发统计
	var honeytokenTriggers int64
	s.db.Model(&repositories.HoneytokenEvent{}).Count(&honeytokenTriggers)
	stats["honeytoken_triggers"] = honeytokenTriggers

	return stats, nil
}

// 达梦数据库服务实现 (与MySQL实现类似，但针对达梦数据库优化)

func (s *DamengService) SaveMalwareSignature(signature *repositories.MalwareSignature) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	signature.CreateTime = time.Now()
	signature.UpdateTime = time.Now()
	return s.db.Create(signature).Error
}

func (s *DamengService) GetMalwareSignatures() ([]repositories.MalwareSignature, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var signatures []repositories.MalwareSignature
	err := s.db.Where("is_active = ?", true).Find(&signatures).Error
	return signatures, err
}

func (s *DamengService) SaveScanResult(result *repositories.ScanResult) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	result.ScanTime = time.Now()
	return s.db.Create(result).Error
}

func (s *DamengService) SaveDetectionResult(detection *repositories.DetectionResult) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	detection.CreatedAt = time.Now()
	return s.db.Create(detection).Error
}

func (s *DamengService) GetScanResult(fileHash string) (*repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var result repositories.ScanResult
	err := s.db.Where("file_hash = ?", fileHash).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DamengService) GetScanHistory(limit int) ([]repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var results []repositories.ScanResult
	err := s.db.Order("scan_time DESC").Limit(limit).Find(&results).Error
	return results, err
}

func (s *DamengService) GetScanResultByAnyHash(hash string) (*repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var result repositories.ScanResult
	err := s.db.Where("file_hash = ? OR md5_hash = ?", hash, hash).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DamengService) GetScanByID(id uint) (*repositories.ScanResult, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var result repositories.ScanResult
	err := s.db.First(&result, id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DamengService) UpdateScanResult(id uint, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	return s.db.Model(&repositories.ScanResult{}).Where("id = ?", id).Updates(updates).Error
}

func (s *DamengService) SaveAttackSession(session *repositories.AttackSession) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	session.StartTime = time.Now()
	return s.db.Create(session).Error
}

func (s *DamengService) UpdateAttackSession(sessionID string, updates map[string]interface{}) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	return s.db.Model(&repositories.AttackSession{}).Where("session_id = ?", sessionID).Updates(updates).Error
}

func (s *DamengService) GetAttackSession(sessionID string) (*repositories.AttackSession, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var session repositories.AttackSession
	err := s.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *DamengService) SaveAttackEvent(event *repositories.AttackEvent) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	event.EventTime = time.Now()
	return s.db.Create(event).Error
}

func (s *DamengService) GetAttackEvents(sessionID string) ([]repositories.AttackEvent, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var events []repositories.AttackEvent
	err := s.db.Where("session_id = ?", sessionID).Order("event_time ASC").Find(&events).Error
	return events, err
}

func (s *DamengService) SaveThreatIntelligence(threat *repositories.ThreatIntelligence) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	threat.CreatedAt = time.Now()
	threat.UpdatedAt = time.Now()
	return s.db.Create(threat).Error
}

func (s *DamengService) GetThreatIntelligence(indicatorType, indicatorValue string) (*repositories.ThreatIntelligence, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var threat repositories.ThreatIntelligence
	err := s.db.Where("indicator_type = ? AND indicator_value = ? AND is_active = ?",
		indicatorType, indicatorValue, true).First(&threat).Error
	if err != nil {
		return nil, err
	}
	return &threat, nil
}

func (s *DamengService) SaveHoneytokenEvent(event *repositories.HoneytokenEvent) error {
	if s.db == nil {
		return errors.New("达梦数据库未初始化")
	}
	event.TriggerTime = time.Now()
	return s.db.Create(event).Error
}

func (s *DamengService) GetHoneytokenEvents(limit int) ([]repositories.HoneytokenEvent, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}
	var events []repositories.HoneytokenEvent
	err := s.db.Order("trigger_time DESC").Limit(limit).Find(&events).Error
	return events, err
}

func (s *DamengService) GetAttackStatistics() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}

	stats := make(map[string]interface{})

	// 总攻击会话数
	var totalSessions int64
	s.db.Model(&repositories.AttackSession{}).Count(&totalSessions)
	stats["total_sessions"] = totalSessions

	// 活跃会话数
	var activeSessions int64
	s.db.Model(&repositories.AttackSession{}).Where("status = ?", "active").Count(&activeSessions)
	stats["active_sessions"] = activeSessions

	// 今日新增会话 (达梦数据库日期函数)
	var todaySessions int64
	s.db.Model(&repositories.AttackSession{}).Where("TO_CHAR(start_time, 'YYYY-MM-DD') = TO_CHAR(SYSDATE, 'YYYY-MM-DD')").Count(&todaySessions)
	stats["today_sessions"] = todaySessions

	// 按IP统计
	var ipStats []map[string]interface{}
	s.db.Model(&repositories.AttackSession{}).
		Select("source_ip, COUNT(*) as session_count, MAX(start_time) as last_seen").
		Group("source_ip").
		Order("session_count DESC").
		Limit(10).
		Scan(&ipStats)
	stats["top_attackers"] = ipStats

	return stats, nil
}

func (s *DamengService) GetThreatStatistics() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, errors.New("达梦数据库未初始化")
	}

	stats := make(map[string]interface{})

	// 扫描统计
	var totalScans int64
	s.db.Model(&repositories.ScanResult{}).Count(&totalScans)
	stats["total_scans"] = totalScans

	// 感染文件数
	var infectedFiles int64
	s.db.Model(&repositories.ScanResult{}).Where("is_infected = ?", true).Count(&infectedFiles)
	stats["infected_files"] = infectedFiles

	// 威胁等级分布
	var threatDistribution []map[string]interface{}
	s.db.Model(&repositories.ScanResult{}).
		Select("threat_level, COUNT(*) as count").
		Where("is_infected = ?", true).
		Group("threat_level").
		Scan(&threatDistribution)
	stats["threat_distribution"] = threatDistribution

	// 蜜签触发统计
	var honeytokenTriggers int64
	s.db.Model(&repositories.HoneytokenEvent{}).Count(&honeytokenTriggers)
	stats["honeytoken_triggers"] = honeytokenTriggers

	return stats, nil
}

// InitDefaultMalwareSignatures 初始化默认的恶意软件特征
func InitDefaultMalwareSignatures(dbService DatabaseService) error {
	// EICAR测试签名
	eicarSignature := &repositories.MalwareSignature{
		Name:        "EICAR Test String",
		Pattern:     "X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*",
		Type:        "string",
		Severity:    "low",
		Description: "EICAR反病毒测试文件标准签名",
		IsActive:    true,
	}

	// 常见恶意软件哈希
	malwareHashes := []repositories.MalwareSignature{
		{
			Name:        "WannaCry Ransomware",
			Pattern:     "ed01ebfbc9eb5bbea545af4d01bf5f1071661840480439c6e5babe8e080e41aa",
			Type:        "hash",
			Severity:    "critical",
			Description: "WannaCry勒索软件SHA256哈希",
			IsActive:    true,
		},
		{
			Name:        "Mimikatz",
			Pattern:     "b42bd5b1056e3ee9e2e66c11f6270e2b",
			Type:        "hash",
			Severity:    "high",
			Description: "Mimikatz密码提取工具MD5哈希",
			IsActive:    true,
		},
	}

	// 保存EICAR签名
	if err := dbService.SaveMalwareSignature(eicarSignature); err != nil {
		fmt.Printf("保存EICAR签名失败: %v\n", err)
	}

	// 保存其他签名
	for _, sig := range malwareHashes {
		if err := dbService.SaveMalwareSignature(&sig); err != nil {
			fmt.Printf("保存恶意软件签名失败: %v\n", err)
		}
	}

	fmt.Println("默认恶意软件特征初始化完成")
	return nil
}
