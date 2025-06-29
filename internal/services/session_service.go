package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionService 会话管理服务
type SessionService struct {
	sessionRepo repositories.HoneypotSessionRepository
	eventRepo   repositories.SessionEventRepository
}

// NewSessionService 创建会话管理服务
func NewSessionService() *SessionService {
	return &SessionService{
		sessionRepo: repositories.NewMySQLHoneypotSessionRepo(config.MySQLDB),
		eventRepo:   repositories.NewMySQLSessionEventRepo(config.MySQLDB),
	}
}

// StartSession 开始新会话
func (s *SessionService) StartSession(sourceIP string, sourcePort uint16, destinationIP string, destinationPort uint16, protocol string, containerID string, containerName string, clientInfo string, fingerprint string) (*repositories.HoneypotSession, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &repositories.HoneypotSession{
		SessionID:       sessionID,
		SourceIP:        sourceIP,
		SourcePort:      sourcePort,
		DestinationIP:   destinationIP,
		DestinationPort: destinationPort,
		Protocol:        protocol,
		ContainerID:     containerID,
		ContainerName:   containerName,
		StartTime:       now,
		Status:          "active",
		EventCount:      0,
		AuthAttempts:    0,
		CommandCount:    0,
		LastActivity:    now,
		ClientInfo:      clientInfo,
		Fingerprint:     fingerprint,
	}

	err := s.sessionRepo.Create(session)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}

	// 记录连接事件
	connectEvent := &repositories.SessionEvent{
		SessionID: sessionID,
		EventType: "connect",
		EventTime: now,
		Details:   fmt.Sprintf("客户端 %s:%d 连接到 %s:%d", sourceIP, sourcePort, destinationIP, destinationPort),
	}

	s.eventRepo.Create(connectEvent)

	return session, nil
}

// CloseSession 关闭会话
func (s *SessionService) CloseSession(sessionID string) error {
	now := time.Now()

	// 更新会话状态
	err := s.sessionRepo.CloseSession(sessionID, now)
	if err != nil {
		return fmt.Errorf("关闭会话失败: %v", err)
	}

	// 记录断开连接事件
	disconnectEvent := &repositories.SessionEvent{
		SessionID: sessionID,
		EventType: "disconnect",
		EventTime: now,
		Details:   "会话正常关闭",
	}

	s.eventRepo.Create(disconnectEvent)

	return nil
}

// RecordAuthAttempt 记录认证尝试
func (s *SessionService) RecordAuthAttempt(sessionID string, username string, password string, success bool) error {
	now := time.Now()

	// 获取或创建会话
	session, err := s.sessionRepo.GetBySessionID(sessionID)
	if err != nil {
		// 如果会话不存在，创建一个新会话
		session = &repositories.HoneypotSession{
			SessionID:       sessionID,
			SourceIP:        "unknown",
			SourcePort:      0,
			DestinationIP:   "unknown",
			DestinationPort: 0,
			Protocol:        "unknown",
			ContainerID:     "manual",
			ContainerName:   "手动记录",
			StartTime:       now,
			Status:          "active",
			EventCount:      0,
			AuthAttempts:    0,
			CommandCount:    0,
			LastActivity:    now,
		}

		err = s.sessionRepo.Create(session)
		if err != nil {
			return fmt.Errorf("创建会话失败: %v", err)
		}
	}

	session.AuthAttempts++
	session.LastActivity = now
	s.sessionRepo.Update(session)

	// 记录认证事件
	authEvent := &repositories.SessionEvent{
		SessionID: sessionID,
		EventType: "auth",
		EventTime: now,
		Username:  username,
		Password:  password,
		Success:   &success,
		Details:   fmt.Sprintf("认证尝试: 用户名=%s, 密码=%s, 结果=%t", username, password, success),
	}

	return s.eventRepo.Create(authEvent)
}

// RecordCommand 记录命令执行
func (s *SessionService) RecordCommand(sessionID string, command string, response string, success bool) error {
	now := time.Now()

	// 获取或创建会话
	session, err := s.sessionRepo.GetBySessionID(sessionID)
	if err != nil {
		// 如果会话不存在，创建一个新会话
		session = &repositories.HoneypotSession{
			SessionID:       sessionID,
			SourceIP:        "unknown",
			SourcePort:      0,
			DestinationIP:   "unknown",
			DestinationPort: 0,
			Protocol:        "unknown",
			ContainerID:     "manual",
			ContainerName:   "手动记录",
			StartTime:       now,
			Status:          "active",
			EventCount:      0,
			AuthAttempts:    0,
			CommandCount:    0,
			LastActivity:    now,
		}

		err = s.sessionRepo.Create(session)
		if err != nil {
			return fmt.Errorf("创建会话失败: %v", err)
		}
	}

	session.CommandCount++
	session.LastActivity = now
	s.sessionRepo.Update(session)

	// 记录命令事件
	commandEvent := &repositories.SessionEvent{
		SessionID: sessionID,
		EventType: "command",
		EventTime: now,
		Command:   command,
		Response:  response,
		Success:   &success,
		Details:   fmt.Sprintf("命令执行: %s", command),
	}

	return s.eventRepo.Create(commandEvent)
}

// RecordError 记录错误事件
func (s *SessionService) RecordError(sessionID string, errorMsg string, details string) error {
	now := time.Now()

	// 更新最后活动时间
	s.sessionRepo.UpdateLastActivity(sessionID, now)

	// 记录错误事件
	errorEvent := &repositories.SessionEvent{
		SessionID: sessionID,
		EventType: "error",
		EventTime: now,
		ErrorMsg:  errorMsg,
		Details:   details,
	}

	return s.eventRepo.Create(errorEvent)
}

// GetSessionByID 根据ID获取会话
func (s *SessionService) GetSessionByID(sessionID string) (*repositories.HoneypotSession, error) {
	return s.sessionRepo.GetBySessionID(sessionID)
}

// GetSessionsByIP 根据IP获取会话列表
func (s *SessionService) GetSessionsByIP(sourceIP string) ([]repositories.HoneypotSession, error) {
	return s.sessionRepo.GetBySourceIP(sourceIP)
}

// GetSessionsByContainer 根据容器ID获取会话列表
func (s *SessionService) GetSessionsByContainer(containerID string) ([]repositories.HoneypotSession, error) {
	return s.sessionRepo.GetByContainerID(containerID)
}

// GetActiveSessionsByIP 获取指定IP的活跃会话
func (s *SessionService) GetActiveSessionsByIP(sourceIP string) ([]repositories.HoneypotSession, error) {
	return s.sessionRepo.GetActiveSessionsByIP(sourceIP)
}

// GetSessionEvents 获取会话事件
func (s *SessionService) GetSessionEvents(sessionID string) ([]repositories.SessionEvent, error) {
	return s.eventRepo.GetBySessionID(sessionID)
}

// GetSessionStatistics 获取会话统计信息
func (s *SessionService) GetSessionStatistics() (map[string]interface{}, error) {
	return s.sessionRepo.GetSessionStatistics()
}

// GetSessionsInTimeRange 根据时间范围获取会话
func (s *SessionService) GetSessionsInTimeRange(startTime, endTime time.Time) ([]repositories.HoneypotSession, error) {
	return s.sessionRepo.GetByTimeRange(startTime, endTime)
}

// UpdateLastActivity 更新会话最后活动时间
func (s *SessionService) UpdateLastActivity(sessionID string) error {
	return s.sessionRepo.UpdateLastActivity(sessionID, time.Now())
}

// GetDetailedSessionInfo 获取详细的会话信息（包含事件）
func (s *SessionService) GetDetailedSessionInfo(sessionID string) (map[string]interface{}, error) {
	// 获取会话基本信息
	session, err := s.sessionRepo.GetBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %v", err)
	}

	// 获取会话事件
	events, err := s.eventRepo.GetBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话事件失败: %v", err)
	}

	// 构建详细信息
	result := map[string]interface{}{
		"session":     session,
		"events":      events,
		"event_count": len(events),
	}

	// 计算会话持续时长
	if session.EndTime != nil {
		duration := session.EndTime.Sub(session.StartTime)
		result["duration_formatted"] = formatDuration(duration)
	} else {
		duration := time.Since(session.StartTime)
		result["duration_formatted"] = formatDuration(duration)
		result["is_active"] = true
	}

	return result, nil
}

// formatDuration 格式化持续时长
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}

// TimeoutInactiveSessions 超时处理不活跃的会话
func (s *SessionService) TimeoutInactiveSessions(timeoutDuration time.Duration) error {
	// 这个方法可以定期调用来清理不活跃的会话
	cutoffTime := time.Now().Add(-timeoutDuration)

	// 获取所有活跃会话
	sessions, err := s.sessionRepo.List()
	if err != nil {
		return fmt.Errorf("获取会话列表失败: %v", err)
	}

	for _, session := range sessions {
		if session.Status == "active" && session.LastActivity.Before(cutoffTime) {
			// 标记为超时
			session.Status = "timeout"
			endTime := session.LastActivity.Add(timeoutDuration)
			session.EndTime = &endTime

			if session.StartTime.Before(endTime) {
				duration := int(endTime.Sub(session.StartTime).Seconds())
				session.DurationSeconds = &duration
			}

			s.sessionRepo.Update(&session)

			// 记录超时事件
			timeoutEvent := &repositories.SessionEvent{
				SessionID: session.SessionID,
				EventType: "disconnect",
				EventTime: endTime,
				Details:   "会话因不活跃而超时",
			}
			s.eventRepo.Create(timeoutEvent)
		}
	}

	return nil
}
