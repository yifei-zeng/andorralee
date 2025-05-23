package services

import (
	"andorralee/internal/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// AlertType 告警类型
type AlertType string

const (
	AlertTypeHoneypot AlertType = "honeypot" // 蜜罐告警
	AlertTypeBait     AlertType = "bait"     // 蜜签告警
	AlertTypeTraffic  AlertType = "traffic"  // 流量告警
	AlertTypeSystem   AlertType = "system"   // 系统告警
)

// Alert 告警信息
type Alert struct {
	ID         string     `json:"id"`
	Type       AlertType  `json:"type"`
	Level      AlertLevel `json:"level"`
	Source     string     `json:"source"`
	Message    string     `json:"message"`
	Details    string     `json:"details"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy string     `json:"resolved_by,omitempty"`
}

// MonitorService 监控服务
type MonitorService struct {
	basePath string
	alerts   []Alert
}

// NewMonitorService 创建监控服务实例
func NewMonitorService(basePath string) *MonitorService {
	return &MonitorService{
		basePath: basePath,
		alerts:   make([]Alert, 0),
	}
}

// CreateAlert 创建告警
func (s *MonitorService) CreateAlert(alertType AlertType, level AlertLevel, source, message, details string) error {
	alert := Alert{
		ID:        utils.GenerateUniqueID(),
		Type:      alertType,
		Level:     level,
		Source:    source,
		Message:   message,
		Details:   details,
		CreatedAt: time.Now(),
	}

	// 保存告警到文件
	alertPath := filepath.Join(s.basePath, "alerts", alert.ID+".json")
	if err := os.MkdirAll(filepath.Dir(alertPath), 0755); err != nil {
		return fmt.Errorf("failed to create alert directory: %v", err)
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %v", err)
	}

	if err := os.WriteFile(alertPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write alert file: %v", err)
	}

	// 添加到内存中的告警列表
	s.alerts = append(s.alerts, alert)

	return nil
}

// ResolveAlert 解决告警
func (s *MonitorService) ResolveAlert(id, resolvedBy string) error {
	alertPath := filepath.Join(s.basePath, "alerts", id+".json")
	data, err := os.ReadFile(alertPath)
	if err != nil {
		return fmt.Errorf("failed to read alert file: %v", err)
	}

	var alert Alert
	if err := json.Unmarshal(data, &alert); err != nil {
		return fmt.Errorf("failed to unmarshal alert: %v", err)
	}

	now := time.Now()
	alert.ResolvedAt = &now
	alert.ResolvedBy = resolvedBy

	// 更新告警文件
	data, err = json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %v", err)
	}

	if err := os.WriteFile(alertPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write alert file: %v", err)
	}

	// 更新内存中的告警列表
	for i, a := range s.alerts {
		if a.ID == id {
			s.alerts[i] = alert
			break
		}
	}

	return nil
}

// ListAlerts 列出所有告警
func (s *MonitorService) ListAlerts() ([]Alert, error) {
	var alerts []Alert
	alertsPath := filepath.Join(s.basePath, "alerts")

	// 遍历告警目录
	err := filepath.Walk(alertsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".json" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var alert Alert
			if err := json.Unmarshal(data, &alert); err != nil {
				return err
			}

			alerts = append(alerts, alert)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %v", err)
	}

	return alerts, nil
}

// GetAlert 获取告警信息
func (s *MonitorService) GetAlert(id string) (*Alert, error) {
	alertPath := filepath.Join(s.basePath, "alerts", id+".json")
	data, err := os.ReadFile(alertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read alert file: %v", err)
	}

	var alert Alert
	if err := json.Unmarshal(data, &alert); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert: %v", err)
	}

	return &alert, nil
}

// MonitorHoneypot 监控蜜罐
func (s *MonitorService) MonitorHoneypot(containerID string) error {
	// 获取蜜罐状态
	honeypotService := NewHoneypotService()
	status, err := honeypotService.GetHoneypotStatus(containerID)
	if err != nil {
		return s.CreateAlert(
			AlertTypeHoneypot,
			AlertLevelError,
			containerID,
			"Failed to get honeypot status",
			err.Error(),
		)
	}

	// 检查容器状态
	if !status.State.Running {
		return s.CreateAlert(
			AlertTypeHoneypot,
			AlertLevelWarning,
			containerID,
			"Honeypot container is not running",
			fmt.Sprintf("Status: %s", status.State.Status),
		)
	}

	return nil
}

// MonitorBait 监控蜜签
func (s *MonitorService) MonitorBait(id string) error {
	baitService := NewBaitService(s.basePath)
	bait, err := baitService.GetBait(id)
	if err != nil {
		return s.CreateAlert(
			AlertTypeBait,
			AlertLevelError,
			id,
			"Failed to get bait status",
			err.Error(),
		)
	}

	// 检查蜜签访问日志
	accessLogPath := filepath.Join(s.basePath, "baits", id, "access.log")
	if _, err := os.Stat(accessLogPath); err == nil {
		return s.CreateAlert(
			AlertTypeBait,
			AlertLevelInfo,
			id,
			"Bait accessed",
			fmt.Sprintf("Bait: %s, Type: %s", bait.Name, bait.Type),
		)
	}

	return nil
}

// MonitorTraffic 监控流量
func (s *MonitorService) MonitorTraffic(sourceIP, targetPort string) error {
	trafficService := NewTrafficService()
	rules, err := trafficService.ListRules()
	if err != nil {
		return s.CreateAlert(
			AlertTypeTraffic,
			AlertLevelError,
			sourceIP,
			"Failed to get traffic rules",
			err.Error(),
		)
	}

	// 检查是否有可疑流量
	if rules != "" {
		return s.CreateAlert(
			AlertTypeTraffic,
			AlertLevelWarning,
			sourceIP,
			"Suspicious traffic detected",
			fmt.Sprintf("Source IP: %s, Target Port: %s", sourceIP, targetPort),
		)
	}

	return nil
}
