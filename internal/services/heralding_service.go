package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"archive/tar"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HeraldingService Heralding认证日志服务
type HeraldingService struct {
	Repo repositories.HeraldingAuthLogRepository
}

// NewHeraldingService 创建Heralding服务
func NewHeraldingService() (*HeraldingService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("MySQL数据库未初始化")
	}

	return &HeraldingService{
		Repo: repositories.NewMySQLHeraldingAuthLogRepo(config.MySQLDB),
	}, nil
}

// PullHeraldingLogs 从容器中拉取heralding认证日志
func (s *HeraldingService) PullHeraldingLogs(containerID string) error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker服务不可用")
	}

	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("容器ID不能为空")
	}

	csvContent, err := s.readHeraldingLogsFromContainer(containerID)
	if err != nil {
		return fmt.Errorf("读取容器日志失败: %w", err)
	}

	if len(csvContent) == 0 {
		log.Printf("容器 %s 未找到可用的Heralding日志", containerID)
		return nil
	}

	var latestTimestamp time.Time
	if latest, err := s.Repo.GetLatestByContainerID(containerID); err == nil && latest != nil {
		latestTimestamp = latest.Timestamp
	}

	logs, err := s.parseCSVLogs(csvContent, containerID, latestTimestamp)
	if err != nil {
		return fmt.Errorf("解析CSV日志失败: %v", err)
	}

	// 批量保存到数据库
	if len(logs) > 0 {
		if err := s.Repo.CreateBatch(logs); err != nil {
			return fmt.Errorf("保存日志到数据库失败: %v", err)
		}
		log.Printf("成功从容器 %s 拉取并保存了 %d 条Heralding认证日志", containerID, len(logs))
	} else {
		log.Printf("容器 %s 没有新的Heralding认证日志", containerID)
	}

	return nil
}

// parseCSVLogs 解析CSV格式的日志
func (s *HeraldingService) parseCSVLogs(csvContent []byte, containerID string, latestTimestamp time.Time) ([]repositories.HeraldingAuthLog, error) {
	reader := csv.NewReader(bytes.NewReader(csvContent))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析CSV失败: %v", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	// determine header
	var dataRecords [][]string
	headers := defaultHeraldingHeader()
	if len(records) > 0 {
		firstLine := strings.ToLower(strings.Join(records[0], ","))
		if strings.Contains(firstLine, "timestamp") || strings.Contains(firstLine, "auth_id") {
			headers = buildHeraldingHeader(records[0])
			dataRecords = records[1:]
		} else {
			dataRecords = records
		}
	}
	if len(dataRecords) == 0 {
		return nil, nil
	}

	var logs []repositories.HeraldingAuthLog
	containerName := s.getContainerName(containerID)
	seen := make(map[string]struct{})
	for idx, record := range dataRecords {
		if len(record) == 0 {
			continue
		}

		timestampStr := getCSVField(record, headers, "timestamp", "time")
		timestamp, err := parseHeraldingTimestamp(timestampStr)
		if err != nil {
			log.Printf("跳过Heralding记录 %d，时间格式无效: %v", idx+1, err)
			continue
		}
		if !latestTimestamp.IsZero() && timestamp.Before(latestTimestamp) {
			continue
		}

		authID := getCSVField(record, headers, "auth_id", "authentication_id")
		if authID == "" {
			authID = uuid.New().String()
		}
		if _, exists := seen[authID]; exists {
			continue
		}
		if existing, _ := s.Repo.GetByAuthID(authID); existing != nil {
			continue
		}

		sessionID := getCSVField(record, headers, "session_id", "session")
		if sessionID == "" {
			sessionID = uuid.New().String()
		}

		sourceIP := getCSVField(record, headers, "source_ip", "src_ip", "remote_ip")
		sourcePort := parseUintField(getCSVField(record, headers, "source_port", "src_port"))
		destinationIP := getCSVField(record, headers, "destination_ip", "dst_ip", "target_ip")
		destinationPort := parseUintField(getCSVField(record, headers, "destination_port", "dst_port", "target_port"))
		protocol := strings.ToLower(getCSVField(record, headers, "protocol", "transport"))
		username := getCSVField(record, headers, "username", "user")
		password := getCSVField(record, headers, "password", "pass")
		passwordHash := getCSVField(record, headers, "password_hash", "pass_hash")

		logEntry := repositories.HeraldingAuthLog{
			Timestamp:       timestamp,
			AuthID:          authID,
			SessionID:       sessionID,
			SourceIP:        sourceIP,
			SourcePort:      sourcePort,
			DestinationIP:   destinationIP,
			DestinationPort: destinationPort,
			Protocol:        protocol,
			Username:        username,
			Password:        password,
			PasswordHash:    passwordHash,
			ContainerID:     containerID,
			ContainerName:   containerName,
		}

		logs = append(logs, logEntry)
		seen[authID] = struct{}{}
	}

	return logs, nil
}

// readHeraldingLogsFromContainer 从容器中读取Heralding日志文件
func (s *HeraldingService) readHeraldingLogsFromContainer(containerID string) ([]byte, error) {
	if config.DockerCli == nil {
		return nil, fmt.Errorf("Docker客户端未初始化")
	}

	var candidatePaths []string
	if custom := strings.TrimSpace(os.Getenv("HERALDING_LOG_PATH")); custom != "" {
		candidatePaths = append(candidatePaths, custom)
	}
	candidatePaths = append(candidatePaths,
		"/var/log/heralding/auth.csv",
		"/var/log/heralding/auth.log",
		"/var/log/heralding/auth.csv",
		"/var/log/heralding/auth.log",
		"/opt/heralding/logs/auth.csv",
		"/opt/heralding/logs/auth.log",
		"/data/logs/heralding/auth.csv",
		"/data/logs/heralding/auth.log",
	)

	var lastErr error
	for _, logPath := range candidatePaths {
		reader, _, err := config.DockerCli.CopyFromContainer(context.Background(), containerID, logPath)
		if err != nil {
			lastErr = err
			continue
		}

		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("读取日志内容失败: %w", err)
		}

		csvData, err := s.extractLatestCSV(content)
		if err != nil {
			lastErr = err
			continue
		}
		if len(csvData) > 0 {
			return csvData, nil
		}
	}

	if lastErr != nil {
		log.Printf("尝试读取Heralding日志失败（容器 %s）: %v", containerID, lastErr)
	}
	return []byte{}, nil
}

// extractLatestCSV 从tar内容中提取最新的CSV/LOG文件
func (s *HeraldingService) extractLatestCSV(tarData []byte) ([]byte, error) {
	reader := tar.NewReader(bytes.NewReader(tarData))
	type fileEntry struct {
		name    string
		modTime time.Time
		data    []byte
	}

	var selected *fileEntry
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		info := header.FileInfo()
		if info == nil || !info.Mode().IsRegular() {
			continue
		}

		name := strings.ToLower(info.Name())
		if !(strings.HasSuffix(name, ".csv") || strings.HasSuffix(name, ".log")) {
			continue
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		if selected == nil || info.ModTime().After(selected.modTime) {
			selected = &fileEntry{name: header.Name, modTime: info.ModTime(), data: data}
		}
	}

	if selected == nil {
		return []byte{}, nil
	}
	return selected.data, nil
}

func buildHeraldingHeader(header []string) map[string]int {
	result := make(map[string]int, len(header))
	for idx, column := range header {
		column = strings.TrimSpace(column)
		column = strings.TrimPrefix(column, "\ufeff")
		column = strings.ToLower(column)
		if column == "" {
			continue
		}
		result[column] = idx
	}
	return result
}

func defaultHeraldingHeader() map[string]int {
	headers := []string{"timestamp", "auth_id", "session_id", "source_ip", "source_port", "destination_ip", "destination_port", "protocol", "username", "password", "password_hash"}
	return buildHeraldingHeader(headers)
}

func getCSVField(record []string, header map[string]int, keys ...string) string {
	for _, key := range keys {
		idx, ok := header[strings.ToLower(key)]
		if !ok {
			continue
		}
		if idx >= 0 && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
	}
	return ""
}

func parseHeraldingTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("空时间字符串")
	}
	layouts := []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02T15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", value)
}

func parseUintField(value string) uint {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if v, err := strconv.ParseUint(value, 10, 32); err == nil {
		return uint(v)
	}
	return 0
}

// getContainerName 获取容器名称
func (s *HeraldingService) getContainerName(containerID string) string {
	if !IsDockerAvailable() {
		return ""
	}

	containerInfo, err := config.DockerCli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return ""
	}

	return strings.TrimPrefix(containerInfo.Name, "/")
}

// GetLogsByContainer 获取指定容器的认证日志
func (s *HeraldingService) GetLogsByContainer(containerID string) ([]repositories.HeraldingAuthLog, error) {
	return s.Repo.GetByContainerID(containerID)
}

// GetLogsBySourceIP 获取指定源IP的认证日志
func (s *HeraldingService) GetLogsBySourceIP(sourceIP string) ([]repositories.HeraldingAuthLog, error) {
	return s.Repo.GetBySourceIP(sourceIP)
}

// GetLogsByProtocol 获取指定协议的认证日志
func (s *HeraldingService) GetLogsByProtocol(protocol string) ([]repositories.HeraldingAuthLog, error) {
	return s.Repo.GetByProtocol(protocol)
}

// GetLogsByTimeRange 获取指定时间范围的认证日志
func (s *HeraldingService) GetLogsByTimeRange(startTime, endTime time.Time) ([]repositories.HeraldingAuthLog, error) {
	return s.Repo.GetByTimeRange(startTime, endTime)
}

// GetStatistics 获取认证统计信息
func (s *HeraldingService) GetStatistics() ([]repositories.HeraldingAuthStatistics, error) {
	return s.Repo.GetStatistics()
}

// GetAttackerIPStatistics 获取攻击者IP统计信息
func (s *HeraldingService) GetAttackerIPStatistics() ([]repositories.AttackerIPStatistics, error) {
	return s.Repo.GetAttackerIPStatistics()
}

// GetTopAttackers 获取前N个攻击者
func (s *HeraldingService) GetTopAttackers(limit int) ([]repositories.AttackerIPStatistics, error) {
	return s.Repo.GetTopAttackers(limit)
}

// GetTopUsernames 获取最常用的用户名
func (s *HeraldingService) GetTopUsernames(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopUsernames(limit)
}

// GetTopPasswords 获取最常用的密码
func (s *HeraldingService) GetTopPasswords(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopPasswords(limit)
}

// DeleteLogsByContainer 删除指定容器的所有认证日志
func (s *HeraldingService) DeleteLogsByContainer(containerID string) error {
	return s.Repo.DeleteByContainerID(containerID)
}

// CreateManualLog 手动创建认证日志（用于测试或手动导入）
func (s *HeraldingService) CreateManualLog(log *repositories.HeraldingAuthLog) error {
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

// GetAllLogs 获取所有认证日志
func (s *HeraldingService) GetAllLogs() ([]repositories.HeraldingAuthLog, error) {
	return s.Repo.List()
}

// GetLogByID 根据ID获取认证日志
func (s *HeraldingService) GetLogByID(id uint) (*repositories.HeraldingAuthLog, error) {
	return s.Repo.GetByID(id)
}

// GetLogByAuthID 根据认证ID获取认证日志
func (s *HeraldingService) GetLogByAuthID(authID string) (*repositories.HeraldingAuthLog, error) {
	return s.Repo.GetByAuthID(authID)
}

// GetLogsBySessionID 获取指定会话的所有认证日志
func (s *HeraldingService) GetLogsBySessionID(sessionID string) ([]repositories.HeraldingAuthLog, error) {
	return s.Repo.GetBySessionID(sessionID)
}
