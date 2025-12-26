package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DionaeaService 负责拉取并存储 Dionaea 日志
// 支持从容器内的 JSON 行日志文件读取，兼容多个候选路径。
type DionaeaService struct {
	Repo repositories.DionaeaLogRepository
}

func NewDionaeaService() (*DionaeaService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	return &DionaeaService{Repo: repositories.NewMySQLDionaeaLogRepo(config.MySQLDB)}, nil
}

// PullDionaeaLogs 从容器读取 Dionaea 日志并落库
func (s *DionaeaService) PullDionaeaLogs(containerID string) error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker服务不可用")
	}
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("容器ID不能为空")
	}

	content, err := s.readDionaeaLogsFromContainer(containerID)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return nil
	}

	containerName := s.getContainerName(containerID)
	var latest time.Time
	if latestLog, err := s.Repo.GetLatestByContainerID(containerID); err == nil && latestLog != nil {
		latest = latestLog.EventTime
	}

	logs, err := s.parseDionaeaJSONLines(content, containerID, containerName, latest)
	if err != nil {
		return err
	}
	if len(logs) == 0 {
		return nil
	}
	if err := s.Repo.CreateBatch(logs); err != nil {
		return fmt.Errorf("保存Dionaea日志失败: %w", err)
	}
	log.Printf("成功从容器 %s 拉取并保存了 %d 条Dionaea日志", containerID, len(logs))
	return nil
}

// readDionaeaLogsFromContainer 复制容器日志文件（支持多个候选路径）
func (s *DionaeaService) readDionaeaLogsFromContainer(containerID string) ([]byte, error) {
	if config.DockerCli == nil {
		return nil, fmt.Errorf("Docker客户端未初始化")
	}
	candidatePaths := []string{
		"/opt/dionaea/var/log/dionaea/dionaea.json",
		"/opt/dionaea/var/log/dionaea/dionaea.log",
		"/opt/dionaea/var/dionaea/dionaea.log",
		"/var/log/dionaea/dionaea.json",
		"/var/log/dionaea/dionaea.log",
		"/data/logs/dionaea/dionaea.json",
		"/data/logs/dionaea/dionaea.log",
	}
	var lastErr error
	for _, path := range candidatePaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		reader, _, err := config.DockerCli.CopyFromContainer(context.Background(), containerID, path)
		if err != nil {
			lastErr = err
			continue
		}
		tr := tar.NewReader(reader)
		_, err = tr.Next()
		if err != nil {
			reader.Close()
			lastErr = err
			continue
		}
		data, err := io.ReadAll(tr)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("读取日志内容失败: %w", err)
		}
		if len(data) > 0 {
			return data, nil
		}
	}
	if lastErr != nil {
		log.Printf("尝试读取Dionaea日志失败（容器 %s）: %v", containerID, lastErr)
	}
	return []byte{}, nil
}

// parseDionaeaJSONLines 解析 JSON 行格式的日志
func (s *DionaeaService) parseDionaeaJSONLines(content []byte, containerID, containerName string, latest time.Time) ([]repositories.DionaeaLog, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	scanner.Buffer(make([]byte, 0, 1024*64), 1024*1024)

	now := time.Now()

	seen := make(map[string]struct{})
	var logs []repositories.DionaeaLog

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			plainTS, plainMsg := parseDionaeaPlainTimestamp(line)
			ts := ensureRecentEventTime(plainTS, now)
			if plainMsg == "" {
				plainMsg = line
			}
			logs = append(logs, repositories.DionaeaLog{
				EventTime:     ts,
				Payload:       plainMsg,
				RawLog:        line,
				ContainerID:   containerID,
				ContainerName: containerName,
				CreatedAt:     time.Now(),
			})
			continue
		}

		// 提取字段
		ts := ensureRecentEventTime(parseTimeAny(getStringMulti(m, "timestamp", "time", "ts")), now)
		if !latest.IsZero() && (ts.Before(latest) || ts.Equal(latest)) {
			continue
		}
		srcIP := getStringMulti(m, "source_ip", "src_ip", "remote", "remote_host")
		dstIP := getStringMulti(m, "destination_ip", "dst_ip", "target", "sensor")
		srcPort := parseUint16(getStringMulti(m, "source_port", "sport", "src_port"))
		dstPort := parseUint16(getStringMulti(m, "destination_port", "dport", "dst_port"))
		proto := strings.ToLower(getStringMulti(m, "protocol", "proto"))
		ctype := strings.ToLower(getStringMulti(m, "connection_type", "connection"))
		url := getStringMulti(m, "url", "uri", "request")
		payloadType := getStringMulti(m, "payload_type", "ptype")
		payload := getStringMulti(m, "payload", "data", "body")

		key := fmt.Sprintf("%s|%s|%d|%s|%d|%s|%s|%s", ts.Format(time.RFC3339Nano), srcIP, srcPort, dstIP, dstPort, proto, payloadType, payload)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		logs = append(logs, repositories.DionaeaLog{
			EventTime:       ts,
			ConnectionType:  ctype,
			Protocol:        proto,
			SourceIP:        srcIP,
			SourcePort:      srcPort,
			DestinationIP:   dstIP,
			DestinationPort: dstPort,
			Sensor:          getStringMulti(m, "sensor", "hostname", "host"),
			URL:             url,
			PayloadType:     payloadType,
			Payload:         payload,
			RawLog:          line,
			ContainerID:     containerID,
			ContainerName:   containerName,
			CreatedAt:       time.Now(),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// parseDionaeaPlainTimestamp extracts timestamps like "[22122025 12:44:45] ...".
func parseDionaeaPlainTimestamp(line string) (time.Time, string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return time.Time{}, ""
	}
	re := regexp.MustCompile(`^\[?(\d{8}\s+\d{2}:\d{2}:\d{2})\]?\s*(.*)$`)
	matches := re.FindStringSubmatch(trimmed)
	if len(matches) < 3 {
		return time.Time{}, ""
	}
	tsStr := matches[1]
	message := strings.TrimSpace(matches[2])
	layouts := []string{
		"02012006 15:04:05",
		"02-01-2006 15:04:05",
		"02/01/2006 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.ParseInLocation(layout, tsStr, time.Local); err == nil {
			return ts, message
		}
	}
	return time.Time{}, message
}

// ensureRecentEventTime 将异常早的时间戳（如老旧镜像里的 2016 年日志）替换为 fallback，避免排序/筛选混乱。
func ensureRecentEventTime(ts time.Time, fallback time.Time) time.Time {
	if ts.IsZero() {
		return fallback
	}
	if ts.Year() < 2023 {
		return fallback
	}
	return ts
}

// getContainerName 获取容器名称
func (s *DionaeaService) getContainerName(containerID string) string {
	if !IsDockerAvailable() {
		return ""
	}
	info, err := config.DockerCli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(info.Name, "/")
}

// 查询相关方法
func (s *DionaeaService) GetAllLogs() ([]repositories.DionaeaLog, error) {
	return s.Repo.List()
}

func (s *DionaeaService) GetLogByID(id uint) (*repositories.DionaeaLog, error) {
	return s.Repo.GetByID(id)
}

func (s *DionaeaService) GetLogsByContainer(containerID string) ([]repositories.DionaeaLog, error) {
	return s.Repo.GetByContainerID(containerID)
}

func (s *DionaeaService) GetLogsBySourceIP(sourceIP string) ([]repositories.DionaeaLog, error) {
	return s.Repo.GetBySourceIP(sourceIP)
}

func (s *DionaeaService) GetLogsByProtocol(protocol string) ([]repositories.DionaeaLog, error) {
	return s.Repo.GetByProtocol(protocol)
}

func (s *DionaeaService) GetLogsByTimeRange(start, end time.Time) ([]repositories.DionaeaLog, error) {
	return s.Repo.GetByTimeRange(start, end)
}

func (s *DionaeaService) DeleteLogsByContainer(containerID string) error {
	return s.Repo.DeleteByContainerID(containerID)
}

// 辅助函数
func getStringMulti(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch vv := v.(type) {
			case string:
				return strings.TrimSpace(vv)
			case float64:
				return fmt.Sprintf("%v", vv)
			}
		}
	}
	return ""
}

func parseUint16(s string) uint16 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 || v > 65535 {
		return 0
	}
	return uint16(v)
}

func parseTimeAny(val string) time.Time {
	val = strings.TrimSpace(val)
	if val == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"02-01-2006 15:04:05",
		"02/01/2006 15:04:05",
		"02012006 15:04:05",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, val); err == nil {
			return t
		}
	}
	return time.Time{}
}
