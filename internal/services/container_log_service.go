package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"context"
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContainerLogService 容器日志服务
type ContainerLogService struct {
	logRepo     repositories.ContainerRuntimeLogRepository
	sessionRepo repositories.ContainerSessionSummaryRepository
}

// NewContainerLogService 创建容器日志服务
func NewContainerLogService() *ContainerLogService {
	return &ContainerLogService{
		logRepo:     repositories.NewMySQLContainerRuntimeLogRepo(config.MySQLDB),
		sessionRepo: repositories.NewMySQLContainerSessionSummaryRepo(config.MySQLDB),
	}
}

// ParseAndStoreContainerLogs 解析并存储容器日志
func (s *ContainerLogService) ParseAndStoreContainerLogs(containerID string) error {
	// 获取容器原始日志
	rawLogs, err := GetContainerLogs(containerID)
	if err != nil {
		return fmt.Errorf("获取容器日志失败: %v", err)
	}

	// 获取容器信息
	containerInfo, err := s.getContainerInfo(containerID)
	if err != nil {
		return fmt.Errorf("获取容器信息失败: %v", err)
	}

	// 解析日志行
	logLines := strings.Split(rawLogs, "\n")
	var parsedLogs []repositories.ContainerRuntimeLog

	for lineNum, line := range logLines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parsedLog := s.parseLogLine(line, lineNum+1, containerID, containerInfo)
		if parsedLog != nil {
			parsedLogs = append(parsedLogs, *parsedLog)
		}
	}

	// 批量保存到数据库
	if len(parsedLogs) > 0 {
		if err := s.logRepo.CreateBatch(parsedLogs); err != nil {
			return fmt.Errorf("保存日志失败: %v", err)
		}
		fmt.Printf("成功解析并保存了 %d 条容器日志\n", len(parsedLogs))
	}

	return nil
}

// parseLogLine 解析单行日志
func (s *ContainerLogService) parseLogLine(line string, lineNum int, containerID string, containerInfo map[string]string) *repositories.ContainerRuntimeLog {
	logID := uuid.New().String()
	now := time.Now()

	// 基础日志结构
	log := &repositories.ContainerRuntimeLog{
		LogID:         logID,
		ContainerID:   containerID,
		ContainerName: containerInfo["name"],
		ImageName:     containerInfo["image"],
		LogTimestamp:  now,
		LogLevel:      "INFO",
		EventType:     "system",
		Protocol:      "other",
		RawLogLine:    line,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 尝试解析不同格式的日志
	if s.parseSSHLog(line, log) {
		return log
	}
	if s.parseHTTPLog(line, log) {
		return log
	}
	if s.parseMySQLLog(line, log) {
		return log
	}
	if s.parseGenericLog(line, log) {
		return log
	}

	// 如果无法解析，返回基础日志
	return log
}

// parseSSHLog 解析SSH日志
func (s *ContainerLogService) parseSSHLog(line string, log *repositories.ContainerRuntimeLog) bool {
	// SSH连接模式: "2025-06-29 21:30:15 [INFO] SSH connection from 192.168.1.100:54321"
	sshConnPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+SSH connection from\s+([0-9.]+):(\d+)`)
	if matches := sshConnPattern.FindStringSubmatch(line); len(matches) == 5 {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", matches[1])
		log.LogTimestamp = timestamp
		log.LogLevel = matches[2]
		log.EventType = "connection"
		log.SourceIP = matches[3]
		if port, err := strconv.Atoi(matches[4]); err == nil {
			sourcePort := uint16(port)
			log.SourcePort = &sourcePort
		}
		log.Protocol = "ssh"
		destPort := uint16(22)
		log.DestinationPort = &destPort
		return true
	}

	// SSH认证模式: "2025-06-29 21:30:20 [WARN] Failed login attempt for user 'admin' with password '123456'"
	sshAuthPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+(Failed|Successful) login attempt for user '([^']+)'(?:\s+with password '([^']+)')?`)
	if matches := sshAuthPattern.FindStringSubmatch(line); len(matches) >= 5 {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", matches[1])
		log.LogTimestamp = timestamp
		log.LogLevel = matches[2]
		log.EventType = "authentication"
		log.Protocol = "ssh"
		log.Username = matches[4]
		if len(matches) > 5 && matches[5] != "" {
			log.Password = matches[5]
			log.PasswordHash = fmt.Sprintf("%x", md5.Sum([]byte(matches[5])))
		}
		success := matches[3] == "Successful"
		log.AuthSuccess = &success
		return true
	}

	// SSH命令模式: "2025-06-29 21:30:25 [INFO] Command executed: ls -la /etc"
	sshCmdPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+Command executed:\s+(.+)`)
	if matches := sshCmdPattern.FindStringSubmatch(line); len(matches) == 4 {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", matches[1])
		log.LogTimestamp = timestamp
		log.LogLevel = matches[2]
		log.EventType = "command"
		log.Protocol = "ssh"
		log.Command = matches[3]
		return true
	}

	return false
}

// parseHTTPLog 解析HTTP日志
func (s *ContainerLogService) parseHTTPLog(line string, log *repositories.ContainerRuntimeLog) bool {
	// HTTP访问日志模式: "192.168.1.100 - - [29/Jun/2025:21:30:15 +0800] "GET /login HTTP/1.1" 200 1234"
	httpPattern := regexp.MustCompile(`([0-9.]+)\s+-\s+-\s+\[([^\]]+)\]\s+"(\w+)\s+([^\s]+)\s+HTTP/[\d.]+"\s+(\d+)\s+(\d+)`)
	if matches := httpPattern.FindStringSubmatch(line); len(matches) == 7 {
		timestamp, _ := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
		log.LogTimestamp = timestamp
		log.LogLevel = "INFO"
		log.EventType = "connection"
		log.SourceIP = matches[1]
		log.Protocol = "http"

		// 解析HTTP方法和路径
		method := matches[3]
		path := matches[4]
		log.Command = fmt.Sprintf("%s %s", method, path)

		// 解析响应码
		if code, err := strconv.Atoi(matches[5]); err == nil {
			log.ResponseCode = &code
		}

		destPort := uint16(80)
		log.DestinationPort = &destPort
		return true
	}

	// HTTP认证日志: "2025-06-29 21:30:20 [WARN] HTTP Basic Auth failed for user 'admin'"
	httpAuthPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+HTTP Basic Auth (failed|succeeded) for user '([^']+)'`)
	if matches := httpAuthPattern.FindStringSubmatch(line); len(matches) == 5 {
		timestamp, _ := time.Parse("2006-01-02 15:04:05", matches[1])
		log.LogTimestamp = timestamp
		log.LogLevel = matches[2]
		log.EventType = "authentication"
		log.Protocol = "http"
		log.Username = matches[4]
		success := matches[3] == "succeeded"
		log.AuthSuccess = &success
		return true
	}

	return false
}

// parseMySQLLog 解析MySQL日志
func (s *ContainerLogService) parseMySQLLog(line string, log *repositories.ContainerRuntimeLog) bool {
	// MySQL连接日志: "2025-06-29T21:30:15.123456Z [Note] [MY-010914] [Server] A new connection was established as id 8."
	mysqlConnPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+\[(\w+)\].*new connection.*id\s+(\d+)`)
	if matches := mysqlConnPattern.FindStringSubmatch(line); len(matches) == 4 {
		timestamp, _ := time.Parse(time.RFC3339Nano, matches[1])
		log.LogTimestamp = timestamp
		log.LogLevel = strings.ToUpper(matches[2])
		log.EventType = "connection"
		log.Protocol = "mysql"
		destPort := uint16(3306)
		log.DestinationPort = &destPort
		return true
	}

	// MySQL认证日志: "2025-06-29T21:30:20.456789Z [Warning] [MY-010914] Access denied for user 'root'@'192.168.1.100'"
	mysqlAuthPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+\[(\w+)\].*Access denied for user '([^']+)'@'([^']+)'`)
	if matches := mysqlAuthPattern.FindStringSubmatch(line); len(matches) == 5 {
		timestamp, _ := time.Parse(time.RFC3339Nano, matches[1])
		log.LogTimestamp = timestamp
		log.LogLevel = strings.ToUpper(matches[2])
		log.EventType = "authentication"
		log.Protocol = "mysql"
		log.Username = matches[3]
		log.SourceIP = matches[4]
		success := false
		log.AuthSuccess = &success
		return true
	}

	return false
}

// parseGenericLog 解析通用日志格式
func (s *ContainerLogService) parseGenericLog(line string, log *repositories.ContainerRuntimeLog) bool {
	// 通用时间戳模式: "2025-06-29T21:30:15.123456Z LEVEL [COMPONENT] Message"
	genericPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?)\s+(\w+)\s+(?:\[([^\]]+)\])?\s*(.*)`)
	if matches := genericPattern.FindStringSubmatch(line); len(matches) >= 4 {
		// 尝试解析时间戳
		timeFormats := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05.000000Z",
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
		}

		var timestamp time.Time
		for _, format := range timeFormats {
			if t, err := time.Parse(format, matches[1]); err == nil {
				timestamp = t
				break
			}
		}

		if !timestamp.IsZero() {
			log.LogTimestamp = timestamp
		}

		log.LogLevel = strings.ToUpper(matches[2])
		if len(matches) > 3 && matches[3] != "" {
			log.ProcessName = matches[3]
		}

		// 根据关键词判断事件类型
		message := strings.ToLower(matches[4])
		if strings.Contains(message, "connect") || strings.Contains(message, "connection") {
			log.EventType = "connection"
		} else if strings.Contains(message, "auth") || strings.Contains(message, "login") {
			log.EventType = "authentication"
		} else if strings.Contains(message, "command") || strings.Contains(message, "exec") {
			log.EventType = "command"
		} else if strings.Contains(message, "error") || strings.Contains(message, "fail") {
			log.EventType = "error"
			log.ErrorMessage = matches[4]
		}

		return true
	}

	return false
}

// getContainerInfo 获取容器信息
func (s *ContainerLogService) getContainerInfo(containerID string) (map[string]string, error) {
	info := map[string]string{
		"name":  "unknown",
		"image": "unknown",
	}

	if !IsDockerAvailable() {
		return info, nil
	}

	containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return info, err
	}

	info["name"] = strings.TrimPrefix(containerInfo.Name, "/")
	info["image"] = containerInfo.Config.Image

	return info, nil
}

// GetLogsByContainer 获取指定容器的日志
func (s *ContainerLogService) GetLogsByContainer(containerID string) ([]repositories.ContainerRuntimeLog, error) {
	return s.logRepo.GetByContainerID(containerID)
}

// GetLogsByTimeRange 根据时间范围获取日志
func (s *ContainerLogService) GetLogsByTimeRange(startTime, endTime time.Time) ([]repositories.ContainerRuntimeLog, error) {
	return s.logRepo.GetByTimeRange(startTime, endTime)
}

// GetLogsByEventType 根据事件类型获取日志
func (s *ContainerLogService) GetLogsByEventType(eventType string) ([]repositories.ContainerRuntimeLog, error) {
	return s.logRepo.GetByEventType(eventType)
}

// GetLogsBySourceIP 根据源IP获取日志
func (s *ContainerLogService) GetLogsBySourceIP(sourceIP string) ([]repositories.ContainerRuntimeLog, error) {
	return s.logRepo.GetBySourceIP(sourceIP)
}

// GetSessionSummary 获取会话汇总信息
func (s *ContainerLogService) GetSessionSummary(sessionID string) (*repositories.ContainerSessionSummary, error) {
	return s.sessionRepo.GetBySessionID(sessionID)
}

// CreateSessionSummary 创建会话汇总
func (s *ContainerLogService) CreateSessionSummary(sessionID string, containerID string) error {
	// 获取该会话的所有日志
	logs, err := s.logRepo.GetBySessionID(sessionID)
	if err != nil {
		return fmt.Errorf("获取会话日志失败: %v", err)
	}

	if len(logs) == 0 {
		return fmt.Errorf("会话 %s 没有找到日志记录", sessionID)
	}

	// 分析日志并生成汇总
	summary := s.analyzeSessionLogs(logs, sessionID, containerID)

	// 保存到数据库
	return s.sessionRepo.Create(summary)
}

// analyzeSessionLogs 分析会话日志生成汇总
func (s *ContainerLogService) analyzeSessionLogs(logs []repositories.ContainerRuntimeLog, sessionID, containerID string) *repositories.ContainerSessionSummary {
	if len(logs) == 0 {
		return nil
	}

	firstLog := logs[0]
	lastLog := logs[len(logs)-1]

	summary := &repositories.ContainerSessionSummary{
		SessionID:       sessionID,
		ContainerID:     containerID,
		ContainerName:   firstLog.ContainerName,
		SourceIP:        firstLog.SourceIP,
		SourcePort:      firstLog.SourcePort,
		DestinationIP:   firstLog.DestinationIP,
		DestinationPort: firstLog.DestinationPort,
		Protocol:        firstLog.Protocol,
		StartTime:       firstLog.LogTimestamp,
		TotalEvents:     len(logs),
		SessionStatus:   "completed",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 计算持续时长
	if !lastLog.LogTimestamp.Equal(firstLog.LogTimestamp) {
		summary.EndTime = &lastLog.LogTimestamp
		duration := int(lastLog.LogTimestamp.Sub(firstLog.LogTimestamp).Seconds())
		summary.DurationSeconds = &duration
	}

	// 统计各类事件
	usernames := make(map[string]bool)
	passwords := make(map[string]bool)
	commands := make(map[string]bool)

	for _, log := range logs {
		switch log.EventType {
		case "connection":
			summary.ConnectionEvents++
		case "authentication":
			summary.AuthAttempts++
			if log.AuthSuccess != nil && *log.AuthSuccess {
				summary.SuccessfulAuths++
			} else {
				summary.FailedAuths++
			}
			if log.Username != "" {
				usernames[log.Username] = true
			}
			if log.Password != "" {
				passwords[log.Password] = true
			}
		case "command":
			summary.CommandExecutions++
			if log.Command != "" {
				commands[log.Command] = true
			}
		case "error":
			summary.ErrorEvents++
		}

		// 收集客户端信息
		if log.ClientInfo != "" {
			summary.ClientInfo = log.ClientInfo
		}
		if log.UserAgent != "" {
			summary.UserAgent = log.UserAgent
		}
		if log.Fingerprint != "" {
			summary.Fingerprint = log.Fingerprint
		}
	}

	summary.UniqueUsernames = len(usernames)
	summary.UniquePasswords = len(passwords)
	summary.UniqueCommands = len(commands)

	// 评估威胁等级
	if summary.SuccessfulAuths > 0 || summary.CommandExecutions > 10 {
		summary.ThreatLevel = "high"
		summary.IsSuccessfulBreach = summary.SuccessfulAuths > 0
	} else if summary.AuthAttempts > 5 {
		summary.ThreatLevel = "medium"
	} else {
		summary.ThreatLevel = "low"
	}

	return summary
}
