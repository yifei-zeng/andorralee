package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"archive/tar"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MySQLHoneypotService 负责MySQL蜜罐日志的拉取与查询
type MySQLHoneypotService struct {
	Repo repositories.MySQLHoneypotLogRepository
}

// NewMySQLHoneypotService 创建MySQL蜜罐日志服务
func NewMySQLHoneypotService() (*MySQLHoneypotService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("MySQL数据库未初始化")
	}

	return &MySQLHoneypotService{
		Repo: repositories.NewMySQLMySQLHoneypotLogRepo(config.MySQLDB),
	}, nil
}

// PullMySQLHoneypotLogs 从容器中拉取MySQL蜜罐日志
func (s *MySQLHoneypotService) PullMySQLHoneypotLogs(containerID string) error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker服务不可用")
	}

	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("容器ID不能为空")
	}

	logLines, err := s.readMySQLLogsFromContainer(containerID)
	if err != nil {
		return fmt.Errorf("读取容器日志失败: %w", err)
	}
	if len(logLines) == 0 {
		log.Printf("容器 %s 未找到MySQL蜜罐日志", containerID)
		return nil
	}

	var latest time.Time
	if latestLog, err := s.Repo.GetLatestByContainerID(containerID); err == nil && latestLog != nil {
		latest = latestLog.EventTime
	}

	records, err := s.parseMySQLHoneypotLines(logLines, containerID, latest)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		log.Printf("容器 %s 没有新的MySQL蜜罐日志", containerID)
		return nil
	}

	if err := s.Repo.CreateBatch(records); err != nil {
		return fmt.Errorf("保存MySQL蜜罐日志失败: %w", err)
	}

	log.Printf("成功从容器 %s 拉取并保存了 %d 条MySQL蜜罐日志", containerID, len(records))
	return nil
}

// readMySQLLogsFromContainer 读取容器中的日志文件
func (s *MySQLHoneypotService) readMySQLLogsFromContainer(containerID string) ([]string, error) {
	if config.DockerCli == nil {
		return nil, fmt.Errorf("Docker客户端未初始化")
	}

	var candidatePaths []string
	if custom := strings.TrimSpace(os.Getenv("MYSQL_HONEYPOT_LOG_PATH")); custom != "" {
		candidatePaths = append(candidatePaths, custom)
	}
	candidatePaths = append(candidatePaths,
		"/var/log/mysql-honeypot.log",
		"/var/log/mysql-honeypot/mysql-honeypot.log",
		"/var/log/mysqlpot/mysqlpot.log",
		"/opt/mysql-honeypot/logs/mysql-honeypot.log",
		"/opt/mysqlpot/logs/mysqlpot.log",
		"/data/logs/mysql-honeypot.log",
		"/logs/mysql-honeypot.log",
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

		lines, err := s.extractMySQLLogLines(content)
		if err != nil {
			lastErr = err
			continue
		}
		if len(lines) > 0 {
			return lines, nil
		}
	}

	if lastErr != nil {
		log.Printf("读取MySQL蜜罐日志失败（容器 %s）: %v", containerID, lastErr)
	}
	return []string{}, nil
}

// extractMySQLLogLines 从tar内容中提取最新的日志文件并按行拆分
func (s *MySQLHoneypotService) extractMySQLLogLines(tarData []byte) ([]string, error) {
	reader := tar.NewReader(bytes.NewReader(tarData))
	type fileEntry struct {
		modTime time.Time
		content []byte
	}

	var latest *fileEntry
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
		if !(strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".txt")) {
			continue
		}
		if strings.HasSuffix(name, ".gz") || strings.HasSuffix(name, ".xz") {
			continue
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		if latest == nil || info.ModTime().After(latest.modTime) {
			latest = &fileEntry{modTime: info.ModTime(), content: data}
		}
	}

	if latest == nil {
		return []string{}, nil
	}

	text := strings.TrimSpace(string(latest.content))
	if text == "" {
		return []string{}, nil
	}

	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// parseMySQLHoneypotLines 将日志文本解析为结构化记录
func (s *MySQLHoneypotService) parseMySQLHoneypotLines(lines []string, containerID string, latest time.Time) ([]repositories.MySQLHoneypotLog, error) {
	var containerName, containerIP string
	if info, err := GetContainerInfo(containerID); err == nil && info != nil {
		containerName = strings.TrimPrefix(info.Name, "/")
		containerIP = info.IP
	}

	defaults := mysqlLogDefaults{
		containerID:     containerID,
		containerName:   containerName,
		destinationIP:   containerIP,
		destinationPort: 3306,
	}

	seen := make(map[string]struct{})
	var records []repositories.MySQLHoneypotLog

	for idx, line := range lines {
		entry, err := s.decodeMySQLLogLine(line, defaults)
		if err != nil {
			log.Printf("解析MySQL蜜罐日志失败 (容器 %s, 行 %d): %v", containerID, idx+1, err)
			continue
		}

		if !latest.IsZero() && (entry.EventTime.Before(latest) || entry.EventTime.Equal(latest)) {
			continue
		}

		if entry.EventID != "" {
			if _, exists := seen[entry.EventID]; exists {
				continue
			}
			if existing, _ := s.Repo.GetByEventID(entry.EventID); existing != nil {
				continue
			}
			seen[entry.EventID] = struct{}{}
		}

		records = append(records, *entry)
	}

	return records, nil
}

// decodeMySQLLogLine 尝试解析单行日志
func (s *MySQLHoneypotService) decodeMySQLLogLine(line string, defaults mysqlLogDefaults) (*repositories.MySQLHoneypotLog, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("空日志行")
	}

	if strings.HasPrefix(line, "{") {
		if entry, err := parseMySQLJSONLine(line); err == nil {
			return applyMySQLDefaults(entry, defaults, line), nil
		}
	}

	if strings.Contains(line, ",") {
		if entry, err := parseMySQLCSVLine(line); err == nil {
			return applyMySQLDefaults(entry, defaults, line), nil
		}
	}

	return nil, fmt.Errorf("无法解析日志行")
}

type mysqlLogDefaults struct {
	containerID     string
	containerName   string
	destinationIP   string
	destinationPort uint
}

func applyMySQLDefaults(entry *repositories.MySQLHoneypotLog, defaults mysqlLogDefaults, raw string) *repositories.MySQLHoneypotLog {
	if entry.EventID == "" {
		entry.EventID = uuid.New().String()
	}
	if entry.EventTime.IsZero() {
		entry.EventTime = time.Now()
	}
	if entry.DestinationIP == "" {
		entry.DestinationIP = defaults.destinationIP
	}
	if entry.DestinationPort == 0 {
		entry.DestinationPort = defaults.destinationPort
	}
	entry.ContainerID = defaults.containerID
	entry.ContainerName = defaults.containerName
	entry.RawLog = raw
	return entry
}

func parseMySQLJSONLine(line string) (*repositories.MySQLHoneypotLog, error) {
	decoder := json.NewDecoder(strings.NewReader(line))
	decoder.UseNumber()
	var data map[string]interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	eventTime, err := parseMySQLTimestamp(stringFromMap(data, "timestamp", "event_time", "time"))
	if err != nil {
		eventTime = time.Now()
	}

	return &repositories.MySQLHoneypotLog{
		EventID:         stringFromMap(data, "event_id", "id", "uid"),
		EventTime:       eventTime,
		ContainerID:     "",
		ContainerName:   "",
		SourceIP:        stringFromMap(data, "src_ip", "source_ip", "remote_ip"),
		SourcePort:      uintFromMap(data, "src_port", "source_port", "remote_port"),
		DestinationIP:   stringFromMap(data, "dst_ip", "destination_ip", "local_ip"),
		DestinationPort: uintFromMap(data, "dst_port", "destination_port", "local_port"),
		Username:        stringFromMap(data, "username", "user"),
		Password:        stringFromMap(data, "password", "pass"),
		DatabaseName:    stringFromMap(data, "database", "schema", "db"),
		Query:           stringFromMap(data, "query", "sql", "statement"),
		ErrorCode:       stringFromMap(data, "error_code", "error", "code"),
	}, nil
}

func parseMySQLCSVLine(line string) (*repositories.MySQLHoneypotLog, error) {
	reader := csv.NewReader(strings.NewReader(line))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	record, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if len(record) < 5 {
		return nil, fmt.Errorf("CSV字段不足")
	}

	eventTime, err := parseMySQLTimestamp(record[0])
	if err != nil {
		eventTime = time.Now()
	}

	entry := &repositories.MySQLHoneypotLog{
		EventID:         "",
		EventTime:       eventTime,
		SourceIP:        safeGet(record, 1),
		SourcePort:      parseUintValue(safeGet(record, 2)),
		DestinationIP:   safeGet(record, 3),
		DestinationPort: parseUintValue(safeGet(record, 4)),
	}

	if len(record) > 5 {
		entry.Username = safeGet(record, 5)
	}
	if len(record) > 6 {
		entry.Password = safeGet(record, 6)
	}
	if len(record) > 7 {
		entry.DatabaseName = safeGet(record, 7)
	}
	if len(record) > 8 {
		entry.Query = safeGet(record, 8)
	}
	if len(record) > 9 {
		entry.ErrorCode = safeGet(record, 9)
	}

	return entry, nil
}

func parseMySQLTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("空时间字符串")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000000",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", value)
}

func stringFromMap(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			switch v := val.(type) {
			case string:
				return strings.TrimSpace(v)
			case json.Number:
				return v.String()
			case float64:
				return strconv.FormatFloat(v, 'f', -1, 64)
			}
		}
	}
	return ""
}

func uintFromMap(data map[string]interface{}, keys ...string) uint {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			switch v := val.(type) {
			case json.Number:
				if i, err := v.Int64(); err == nil {
					return uint(i)
				}
			case float64:
				return uint(v)
			case string:
				if parsed, err := strconv.ParseUint(strings.TrimSpace(v), 10, 32); err == nil {
					return uint(parsed)
				}
			}
		}
	}
	return 0
}

func safeGet(record []string, idx int) string {
	if idx >= 0 && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

func parseUintValue(value string) uint {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if v, err := strconv.ParseUint(value, 10, 32); err == nil {
		return uint(v)
	}
	return 0
}

// 数据访问方法封装

func (s *MySQLHoneypotService) GetAllLogs() ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.List()
}

func (s *MySQLHoneypotService) GetLogByID(id uint) (*repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByID(id)
}

func (s *MySQLHoneypotService) GetLogsByContainer(containerID string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByContainerID(containerID)
}

func (s *MySQLHoneypotService) GetLogsBySourceIP(sourceIP string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetBySourceIP(sourceIP)
}

func (s *MySQLHoneypotService) GetLogsByUsername(username string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByUsername(username)
}

func (s *MySQLHoneypotService) GetLogsByTimeRange(startTime, endTime time.Time) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByTimeRange(startTime, endTime)
}

func (s *MySQLHoneypotService) DeleteLogsByContainer(containerID string) error {
	return s.Repo.DeleteByContainerID(containerID)
}

func (s *MySQLHoneypotService) GetQueryStatistics(limit int) ([]repositories.MySQLHoneypotStatistics, error) {
	return s.Repo.GetQueryStatistics(limit)
}
