package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HeadlingService Headling认证日志服务
type HeadlingService struct {
	Repo repositories.HeadlingAuthLogRepository
}

// NewHeadlingService 创建Headling服务
func NewHeadlingService() (*HeadlingService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("MySQL数据库未初始化")
	}

	return &HeadlingService{
		Repo: repositories.NewMySQLHeadlingAuthLogRepo(config.MySQLDB),
	}, nil
}

// PullHeadlingLogs 从容器中拉取headling认证日志
func (s *HeadlingService) PullHeadlingLogs(containerID string) error {
	if !IsDockerAvailable() {
		return fmt.Errorf("Docker服务不可用")
	}

	// 暂时使用模拟数据，实际项目中需要实现从容器中读取CSV文件的逻辑
	// 这里可以通过以下方式实现：
	// 1. 使用 docker cp 命令复制文件
	// 2. 使用 docker exec 执行命令读取文件内容
	// 3. 通过挂载卷的方式直接读取文件

	// 模拟CSV内容（实际应该从容器中读取）
	csvContent := `timestamp,auth_id,session_id,source_ip,source_port,destination_ip,destination_port,protocol,username,password,password_hash
2025-06-03 11:45:51.239525,e69fc485-66f8-440b-8a79-27f764ef83b9,b344a630-a024-41fa-9aaa-91b7653ad49c,172.17.0.1,44390,172.17.0.3,80,http,123,123,
2025-06-03 11:45:59.990922,ea9cae68-b2f5-4d9a-a8d7-08f69e8238d7,0853267f-01b0-4b0c-b66a-1bd2686a9393,172.17.0.1,38324,172.17.0.3,80,http,hhh,hhh,`

	// 解析CSV内容
	logs, err := s.parseCSVLogs(csvContent, containerID)
	if err != nil {
		return fmt.Errorf("解析CSV日志失败: %v", err)
	}

	// 批量保存到数据库
	if len(logs) > 0 {
		if err := s.Repo.CreateBatch(logs); err != nil {
			return fmt.Errorf("保存日志到数据库失败: %v", err)
		}
		fmt.Printf("成功从容器 %s 拉取并保存了 %d 条认证日志\n", containerID, len(logs))
	} else {
		fmt.Printf("容器 %s 没有新的认证日志\n", containerID)
	}

	return nil
}

// parseCSVLogs 解析CSV格式的日志
func (s *HeadlingService) parseCSVLogs(csvContent, containerID string) ([]repositories.HeadlingAuthLog, error) {
	reader := csv.NewReader(strings.NewReader(csvContent))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析CSV失败: %v", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	// 跳过标题行
	if len(records) > 0 && records[0][0] == "timestamp" {
		records = records[1:]
	}

	var logs []repositories.HeadlingAuthLog
	for i, record := range records {
		if len(record) < 11 {
			fmt.Printf("跳过格式不正确的记录 %d: %v\n", i+1, record)
			continue
		}

		// 解析时间戳
		timestamp, err := time.Parse("2006-01-02 15:04:05.999999", record[0])
		if err != nil {
			fmt.Printf("跳过时间戳解析失败的记录 %d: %v\n", i+1, err)
			continue
		}

		// 解析端口号
		sourcePort, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			fmt.Printf("跳过源端口解析失败的记录 %d: %v\n", i+1, err)
			continue
		}

		destinationPort, err := strconv.ParseUint(record[6], 10, 32)
		if err != nil {
			fmt.Printf("跳过目标端口解析失败的记录 %d: %v\n", i+1, err)
			continue
		}

		// 检查是否已存在（根据auth_id）
		existing, _ := s.Repo.GetByAuthID(record[1])
		if existing != nil {
			continue // 跳过已存在的记录
		}

		// 获取容器名称
		containerName := s.getContainerName(containerID)

		log := repositories.HeadlingAuthLog{
			Timestamp:       timestamp,
			AuthID:          record[1],
			SessionID:       record[2],
			SourceIP:        record[3],
			SourcePort:      uint(sourcePort),
			DestinationIP:   record[5],
			DestinationPort: uint(destinationPort),
			Protocol:        record[7],
			Username:        record[8],
			Password:        record[9],
			PasswordHash:    record[10],
			ContainerID:     containerID,
			ContainerName:   containerName,
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// getContainerName 获取容器名称
func (s *HeadlingService) getContainerName(containerID string) string {
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
func (s *HeadlingService) GetLogsByContainer(containerID string) ([]repositories.HeadlingAuthLog, error) {
	return s.Repo.GetByContainerID(containerID)
}

// GetLogsBySourceIP 获取指定源IP的认证日志
func (s *HeadlingService) GetLogsBySourceIP(sourceIP string) ([]repositories.HeadlingAuthLog, error) {
	return s.Repo.GetBySourceIP(sourceIP)
}

// GetLogsByProtocol 获取指定协议的认证日志
func (s *HeadlingService) GetLogsByProtocol(protocol string) ([]repositories.HeadlingAuthLog, error) {
	return s.Repo.GetByProtocol(protocol)
}

// GetLogsByTimeRange 获取指定时间范围的认证日志
func (s *HeadlingService) GetLogsByTimeRange(startTime, endTime time.Time) ([]repositories.HeadlingAuthLog, error) {
	return s.Repo.GetByTimeRange(startTime, endTime)
}

// GetStatistics 获取认证统计信息
func (s *HeadlingService) GetStatistics() ([]repositories.HeadlingAuthStatistics, error) {
	return s.Repo.GetStatistics()
}

// GetAttackerIPStatistics 获取攻击者IP统计信息
func (s *HeadlingService) GetAttackerIPStatistics() ([]repositories.AttackerIPStatistics, error) {
	return s.Repo.GetAttackerIPStatistics()
}

// GetTopAttackers 获取前N个攻击者
func (s *HeadlingService) GetTopAttackers(limit int) ([]repositories.AttackerIPStatistics, error) {
	return s.Repo.GetTopAttackers(limit)
}

// GetTopUsernames 获取最常用的用户名
func (s *HeadlingService) GetTopUsernames(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopUsernames(limit)
}

// GetTopPasswords 获取最常用的密码
func (s *HeadlingService) GetTopPasswords(limit int) ([]map[string]interface{}, error) {
	return s.Repo.GetTopPasswords(limit)
}

// DeleteLogsByContainer 删除指定容器的所有认证日志
func (s *HeadlingService) DeleteLogsByContainer(containerID string) error {
	return s.Repo.DeleteByContainerID(containerID)
}

// CreateManualLog 手动创建认证日志（用于测试或手动导入）
func (s *HeadlingService) CreateManualLog(log *repositories.HeadlingAuthLog) error {
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
func (s *HeadlingService) GetAllLogs() ([]repositories.HeadlingAuthLog, error) {
	return s.Repo.List()
}

// GetLogByID 根据ID获取认证日志
func (s *HeadlingService) GetLogByID(id uint) (*repositories.HeadlingAuthLog, error) {
	return s.Repo.GetByID(id)
}

// GetLogByAuthID 根据认证ID获取认证日志
func (s *HeadlingService) GetLogByAuthID(authID string) (*repositories.HeadlingAuthLog, error) {
	return s.Repo.GetByAuthID(authID)
}

// GetLogsBySessionID 获取指定会话的所有认证日志
func (s *HeadlingService) GetLogsBySessionID(sessionID string) ([]repositories.HeadlingAuthLog, error) {
	return s.Repo.GetBySessionID(sessionID)
}
