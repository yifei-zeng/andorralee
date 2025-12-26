package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
)

// MySQLHoneypotService 负责MySQL蜜罐日志的拉取与查询
type MySQLHoneypotService struct {
	Repo repositories.MySQLHoneypotLogRepository
}

// NewMySQLHoneypotService 创建MySQL蜜罐日志服务
func NewMySQLHoneypotService() (*MySQLHoneypotService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
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

// readMySQLLogsFromContainer 从容器的 docker logs 读取日志
func (s *MySQLHoneypotService) readMySQLLogsFromContainer(containerID string) ([]string, error) {
	if config.DockerCli == nil {
		return nil, fmt.Errorf("Docker客户端未初始化")
	}

	log.Printf("[DEBUG] 开始从容器 %s 读取 docker logs", containerID)
	logsReader, err := config.DockerCli.ContainerLogs(
		context.Background(),
		containerID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
			Tail:       "2000", // 控制读取量，防止一次性取太多
			Timestamps: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("读取 docker logs 失败: %w", err)
	}
	defer logsReader.Close()

	log.Printf("[DEBUG] docker logs 读取成功，开始解析")

	// 直接读取全部内容，不使用 stdcopy（因为容器可能没有 TTY 复用）
	content, err := io.ReadAll(logsReader)
	if err != nil {
		return nil, fmt.Errorf("读取日志内容失败: %w", err)
	}

	log.Printf("[DEBUG] 读取到 %d 字节日志内容", len(content))

	lines := splitMySQLLogLines(string(content))
	log.Printf("[DEBUG] 解析得到 %d 行日志", len(lines))

	if len(lines) > 0 {
		log.Printf("[DEBUG] 前 3 行示例: %v", lines[:min(3, len(lines))])
	}

	return lines, nil
}

// splitMySQLLogLines 将 docker logs 文本拆分成非空行
func splitMySQLLogLines(text string) []string {
	rawLines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
			log.Printf("解析MySQL蜜罐日志失败 (容器 %s, 行 %d): %v, 内容: %s", containerID, idx+1, err, line)
			continue
		}

		log.Printf("[DEBUG] 解析成功 行%d: EventType=%s, SourceIP=%s, Username=%s", idx+1, entry.EventType, entry.SourceIP, entry.Username)

		if !latest.IsZero() && (entry.EventTime.Before(latest) || entry.EventTime.Equal(latest)) {
			log.Printf("[DEBUG] 跳过旧日志 (行 %d): %v <= %v", idx+1, entry.EventTime, latest)
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

	if entry, err := parseMySQLPlainLine(line); err == nil {
		return applyMySQLDefaults(entry, defaults, line), nil
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
	if entry.EventType == "" {
		entry.EventType = "other"
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

var (
	plainLinePrefix   = regexp.MustCompile(`^(?P<ts>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\s+[^:]+:\s+(?P<body>.+)$`)
	plainAccessDenied = regexp.MustCompile(`Access denied for user '([^']+)' from ([\d\.]+):(\d+) to ([\d\.]+):(\d+) \(using password: ([^;\)]+); authentication plugin: ([^\)]+)\)`)
	plainNewConn      = regexp.MustCompile(`New connection from ([\d\.]+):(\d+) \[[^\]]*\] to ([\d\.]+):(\d+)`)
	plainClosing      = regexp.MustCompile(`Closing connection for ([\d\.]+):(\d+)`)
	plainSignal       = regexp.MustCompile(`Got signal (\d+), shutting down`)
)

// parseMySQLPlainLine 解析 docker logs 输出的明文格式
func parseMySQLPlainLine(line string) (*repositories.MySQLHoneypotLog, error) {
	matches := plainLinePrefix.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil, fmt.Errorf("不符合前缀格式")
	}

	ts := matches[1]
	body := matches[2]
	eventTime, err := parseMySQLTimestamp(ts)
	if err != nil {
		eventTime = time.Now()
	}

	entry := &repositories.MySQLHoneypotLog{
		EventTime: eventTime,
		Message:   body,
	}

	switch {
	case plainAccessDenied.MatchString(body):
		s := plainAccessDenied.FindStringSubmatch(body)
		entry.EventType = "access_denied"
		entry.Username = strings.TrimSpace(s[1])
		entry.SourceIP, entry.SourcePort = parseIPPort(s[2], s[3])
		entry.DestinationIP, entry.DestinationPort = parseIPPort(s[4], s[5])
		entry.PasswordUsed = strings.TrimSpace(s[6])
		entry.AuthPlugin = strings.TrimSpace(s[7])
	case plainNewConn.MatchString(body):
		s := plainNewConn.FindStringSubmatch(body)
		entry.EventType = "new_connection"
		entry.SourceIP, entry.SourcePort = parseIPPort(s[1], s[2])
		entry.DestinationIP, entry.DestinationPort = parseIPPort(s[3], s[4])
	case plainClosing.MatchString(body):
		s := plainClosing.FindStringSubmatch(body)
		entry.EventType = "connection_close"
		entry.SourceIP, entry.SourcePort = parseIPPort(s[1], s[2])
	case plainSignal.MatchString(body):
		entry.EventType = "signal"
	default:
		entry.EventType = "other"
	}

	return entry, nil
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

func parseIPPort(ipStr, portStr string) (string, uint) {
	ip := strings.TrimSpace(ipStr)
	port := parseUintValue(strings.TrimSpace(portStr))
	return ip, port
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

func (s *MySQLHoneypotService) GetLogsByDestinationIP(destinationIP string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByDestinationIP(destinationIP)
}

func (s *MySQLHoneypotService) GetLogsByUsername(username string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByUsername(username)
}

func (s *MySQLHoneypotService) GetLogsByDatabaseName(databaseName string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByDatabaseName(databaseName)
}

func (s *MySQLHoneypotService) GetLogsByTimeRange(startTime, endTime time.Time) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByTimeRange(startTime, endTime)
}

func (s *MySQLHoneypotService) GetLogsByQueryKeyword(keyword string, limit int) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByQueryKeyword(keyword, limit)
}

func (s *MySQLHoneypotService) GetLogsByErrorCode(code string) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.GetByErrorCode(code)
}

func (s *MySQLHoneypotService) DeleteLogsByContainer(containerID string) error {
	return s.Repo.DeleteByContainerID(containerID)
}

func (s *MySQLHoneypotService) GetQueryStatistics(limit int) ([]repositories.MySQLHoneypotStatistics, error) {
	return s.Repo.GetQueryStatistics(limit)
}

func (s *MySQLHoneypotService) SearchLogs(filter repositories.MySQLHoneypotSearchFilter) ([]repositories.MySQLHoneypotLog, error) {
	return s.Repo.Search(filter)
}
