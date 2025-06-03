package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CowrieService Cowrie蜜罐日志服务
type CowrieService struct {
	Repo repositories.CowrieLogRepository
}

// NewCowrieService 创建Cowrie服务
func NewCowrieService() (*CowrieService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("MySQL数据库未初始化")
	}

	return &CowrieService{
		Repo: repositories.NewMySQLCowrieLogRepo(config.MySQLDB),
	}, nil
}

// PullCowrieLogs 从容器中拉取Cowrie日志
func (s *CowrieService) PullCowrieLogs(containerID string) error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker服务不可用")
	}

	// 模拟JSON格式的Cowrie日志数据
	// 实际项目中应该从容器中读取真实的日志文件
	jsonLogs := []string{
		`{"event_time":"2025-01-15T10:30:45.123456","auth_id":"550e8400-e29b-41d4-a716-446655440001","session_id":"550e8400-e29b-41d4-a716-446655440002","source_ip":"192.168.1.100","source_port":45678,"destination_ip":"172.17.0.2","destination_port":22,"protocol":"ssh","client_info":"SSH-2.0-OpenSSH_8.0","fingerprint":"92:65:ee:de:36:63:d9:f2:24:de:c4:84:ba:14:c3:42","username":"admin","password":"123456","command":"ls -la","command_found":true,"raw_log":"2025-01-15T10:30:45.123456Z [SSHChannel session (0) on SSHService b'ssh-connection' on SSHTransport,1,192.168.1.100] CMD: ls -la"}`,
		`{"event_time":"2025-01-15T10:31:20.654321","auth_id":"550e8400-e29b-41d4-a716-446655440003","session_id":"550e8400-e29b-41d4-a716-446655440002","source_ip":"192.168.1.100","source_port":45678,"destination_ip":"172.17.0.2","destination_port":22,"protocol":"ssh","client_info":"SSH-2.0-OpenSSH_8.0","fingerprint":"92:65:ee:de:36:63:d9:f2:24:de:c4:84:ba:14:c3:42","username":"admin","password":"123456","command":"cat /etc/passwd","command_found":true,"raw_log":"2025-01-15T10:31:20.654321Z [SSHChannel session (0) on SSHService b'ssh-connection' on SSHTransport,1,192.168.1.100] CMD: cat /etc/passwd"}`,
		`{"event_time":"2025-01-15T10:32:15.789012","auth_id":"550e8400-e29b-41d4-a716-446655440004","session_id":"550e8400-e29b-41d4-a716-446655440005","source_ip":"192.168.1.101","source_port":54321,"destination_ip":"172.17.0.2","destination_port":22,"protocol":"ssh","client_info":"SSH-2.0-libssh_0.8.9","fingerprint":"a1:b2:c3:d4:e5:f6:07:08:09:0a:0b:0c:0d:0e:0f:10","username":"root","password":"password","command":"whoami","command_found":true,"raw_log":"2025-01-15T10:32:15.789012Z [SSHChannel session (0) on SSHService b'ssh-connection' on SSHTransport,2,192.168.1.101] CMD: whoami"}`,
	}

	// 解析JSON日志
	logs, err := s.parseJSONLogs(jsonLogs, containerID)
	if err != nil {
		return fmt.Errorf("解析JSON日志失败: %v", err)
	}

	// 批量保存到数据库
	if len(logs) > 0 {
		if err := s.Repo.CreateBatch(logs); err != nil {
			return fmt.Errorf("保存日志到数据库失败: %v", err)
		}
		fmt.Printf("成功从容器 %s 拉取并保存了 %d 条Cowrie日志\n", containerID, len(logs))
	} else {
		fmt.Printf("容器 %s 没有新的Cowrie日志\n", containerID)
	}

	return nil
}

// parseJSONLogs 解析JSON格式的日志
func (s *CowrieService) parseJSONLogs(jsonLogs []string, containerID string) ([]repositories.CowrieLog, error) {
	var logs []repositories.CowrieLog

	for i, jsonLog := range jsonLogs {
		var logData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonLog), &logData); err != nil {
			fmt.Printf("跳过JSON解析失败的记录 %d: %v\n", i+1, err)
			continue
		}

		// 解析时间戳
		eventTimeStr, ok := logData["event_time"].(string)
		if !ok {
			fmt.Printf("跳过时间戳格式错误的记录 %d\n", i+1)
			continue
		}

		eventTime, err := time.Parse("2006-01-02T15:04:05.999999", eventTimeStr)
		if err != nil {
			fmt.Printf("跳过时间戳解析失败的记录 %d: %v\n", i+1, err)
			continue
		}

		// 检查是否已存在（根据auth_id）
		authID := getString(logData, "auth_id")
		if authID == "" {
			authID = uuid.New().String()
		}

		existing, _ := s.Repo.GetByAuthID(authID)
		if existing != nil {
			continue // 跳过已存在的记录
		}

		// 获取容器名称
		containerName := s.getContainerName(containerID)

		// 解析端口号
		sourcePort := uint16(getInt(logData, "source_port"))
		destinationPort := uint16(getInt(logData, "destination_port"))

		// 解析command_found
		var commandFound *bool
		if val, exists := logData["command_found"]; exists {
			if boolVal, ok := val.(bool); ok {
				commandFound = &boolVal
			}
		}

		log := repositories.CowrieLog{
			EventTime:       eventTime,
			AuthID:          authID,
			SessionID:       getString(logData, "session_id"),
			SourceIP:        getString(logData, "source_ip"),
			SourcePort:      sourcePort,
			DestinationIP:   getString(logData, "destination_ip"),
			DestinationPort: destinationPort,
			Protocol:        getString(logData, "protocol"),
			ClientInfo:      getString(logData, "client_info"),
			Fingerprint:     getString(logData, "fingerprint"),
			Username:        getString(logData, "username"),
			Password:        getString(logData, "password"),
			PasswordHash:    getString(logData, "password_hash"),
			Command:         getString(logData, "command"),
			CommandFound:    commandFound,
			RawLog:          getString(logData, "raw_log"),
			ContainerID:     containerID,
			ContainerName:   containerName,
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// 辅助函数：从map中获取字符串值
func getString(data map[string]interface{}, key string) string {
	if val, exists := data[key]; exists {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// 辅助函数：从map中获取整数值
func getInt(data map[string]interface{}, key string) int {
	if val, exists := data[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return 0
}

// getContainerName 获取容器名称
func (s *CowrieService) getContainerName(containerID string) string {
	if !IsDockerAvailable() {
		return ""
	}

	containerInfo, err := GetContainerInfo(containerID)
	if err != nil {
		return ""
	}

	return strings.TrimPrefix(containerInfo.Name, "/")
}

// GetLogsByContainer 获取指定容器的Cowrie日志
func (s *CowrieService) GetLogsByContainer(containerID string) ([]repositories.CowrieLog, error) {
	return s.Repo.GetByContainerID(containerID)
}

// GetLogsBySourceIP 获取指定源IP的Cowrie日志
func (s *CowrieService) GetLogsBySourceIP(sourceIP string) ([]repositories.CowrieLog, error) {
	return s.Repo.GetBySourceIP(sourceIP)
}

// GetLogsByProtocol 获取指定协议的Cowrie日志
func (s *CowrieService) GetLogsByProtocol(protocol string) ([]repositories.CowrieLog, error) {
	return s.Repo.GetByProtocol(protocol)
}

// GetLogsByTimeRange 获取指定时间范围的Cowrie日志
func (s *CowrieService) GetLogsByTimeRange(startTime, endTime time.Time) ([]repositories.CowrieLog, error) {
	return s.Repo.GetByTimeRange(startTime, endTime)
}

// GetLogsByCommand 获取包含指定命令的Cowrie日志
func (s *CowrieService) GetLogsByCommand(command string) ([]repositories.CowrieLog, error) {
	return s.Repo.GetByCommand(command)
}

// GetLogsByCommandFound 获取命令识别状态的Cowrie日志
func (s *CowrieService) GetLogsByCommandFound(found bool) ([]repositories.CowrieLog, error) {
	return s.Repo.GetByCommandFound(found)
}

// GetLogsByUsername 获取指定用户名的Cowrie日志
func (s *CowrieService) GetLogsByUsername(username string) ([]repositories.CowrieLog, error) {
	return s.Repo.GetByUsername(username)
}

// GetStatistics 获取Cowrie统计信息
func (s *CowrieService) GetStatistics() ([]repositories.CowrieStatistics, error) {
	return s.Repo.GetStatistics()
}

// GetAttackerBehavior 获取攻击者行为统计信息
func (s *CowrieService) GetAttackerBehavior() ([]repositories.CowrieAttackerBehavior, error) {
	return s.Repo.GetAttackerBehavior()
}

// GetTopAttackers 获取前N个攻击者
func (s *CowrieService) GetTopAttackers(limit int) ([]repositories.CowrieAttackerBehavior, error) {
	return s.Repo.GetTopAttackers(limit)
}

// GetCommandStatistics 获取命令统计信息
func (s *CowrieService) GetCommandStatistics() ([]repositories.CowrieCommandStatistics, error) {
	return s.Repo.GetCommandStatistics()
}

// GetTopCommands 获取最常用的命令
func (s *CowrieService) GetTopCommands(limit int) ([]repositories.CowrieCommandStatistics, error) {
	return s.Repo.GetTopCommands(limit)
}

// GetTopUsernames 获取最常用的用户名
func (s *CowrieService) GetTopUsernames(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopUsernames(limit)
}

// GetTopPasswords 获取最常用的密码
func (s *CowrieService) GetTopPasswords(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopPasswords(limit)
}

// GetTopFingerprints 获取最常用的指纹
func (s *CowrieService) GetTopFingerprints(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopFingerprints(limit)
}

// DeleteLogsByContainer 删除指定容器的所有Cowrie日志
func (s *CowrieService) DeleteLogsByContainer(containerID string) error {
	return s.Repo.DeleteByContainerID(containerID)
}

// CreateManualLog 手动创建Cowrie日志（用于测试或手动导入）
func (s *CowrieService) CreateManualLog(log *repositories.CowrieLog) error {
	// 如果没有提供AuthID，生成一个
	if log.AuthID == "" {
		log.AuthID = uuid.New().String()
	}

	// 如果没有提供SessionID，生成一个
	if log.SessionID == "" {
		log.SessionID = uuid.New().String()
	}

	return s.Repo.Create(log)
}

// GetAllLogs 获取所有Cowrie日志
func (s *CowrieService) GetAllLogs() ([]repositories.CowrieLog, error) {
	return s.Repo.List()
}

// GetLogByID 根据ID获取Cowrie日志
func (s *CowrieService) GetLogByID(id uint) (*repositories.CowrieLog, error) {
	return s.Repo.GetByID(id)
}

// GetLogByAuthID 根据认证ID获取Cowrie日志
func (s *CowrieService) GetLogByAuthID(authID string) (*repositories.CowrieLog, error) {
	return s.Repo.GetByAuthID(authID)
}

// GetLogsBySessionID 获取指定会话的所有Cowrie日志
func (s *CowrieService) GetLogsBySessionID(sessionID string) ([]repositories.CowrieLog, error) {
	return s.Repo.GetBySessionID(sessionID)
}
