package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// LogSegmentResult 日志分割结果
type LogSegmentResult struct {
	Success  bool             `json:"success"`
	Segments []LogSegmentInfo `json:"segments"`
	Stats    map[string]int   `json:"stats"`
}

// LogSegmentInfo 日志片段信息
type LogSegmentInfo struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	Component  string `json:"component"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	LineNumber int    `json:"line_number"`
}

// SemanticSegment 对Docker容器日志进行语义分割
func SemanticSegment(containerID string) (map[string]interface{}, error) {
	// 获取容器日志
	logs, err := GetContainerLogs(containerID)
	if err != nil {
		return nil, fmt.Errorf("获取容器日志失败: %v", err)
	}

	// 分析日志内容
	segments, stats := AnalyzeContainerLogs(logs)

	// 将分析结果存入数据库
	if err := saveLogSegmentsToDatabase(containerID, segments); err != nil {
		return nil, fmt.Errorf("保存日志分析结果失败: %v", err)
	}

	// 构建响应
	result := map[string]interface{}{
		"success":  true,
		"segments": segments,
		"stats":    stats,
	}

	return result, nil
}

// AnalyzeContainerLogs 分析容器日志内容
func AnalyzeContainerLogs(logContent string) ([]LogSegmentInfo, map[string]int) {
	// 按行分割日志
	logLines := strings.Split(logContent, "\n")

	// 定义正则表达式匹配日志格式
	// 假设日志格式为: 2023-01-01T12:00:00.000000000Z LEVEL [COMPONENT] Message
	logPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z)\s+(\w+)\s+\[([^\]]+)\]\s+(.+)`)

	// 定义日志类型识别规则
	errorPattern := regexp.MustCompile(`(?i)(error|exception|fail|failed|crash)`)
	warningPattern := regexp.MustCompile(`(?i)(warn|warning|caution)`)
	infoPattern := regexp.MustCompile(`(?i)(info|information|notice)`)
	debugPattern := regexp.MustCompile(`(?i)(debug|trace|verbose)`)

	// 存储分析结果
	var segments []LogSegmentInfo
	stats := map[string]int{
		"error":   0,
		"warning": 0,
		"info":    0,
		"debug":   0,
		"unknown": 0,
	}

	// 分析每行日志
	for lineNumber, line := range logLines {
		if line == "" {
			continue
		}

		matches := logPattern.FindStringSubmatch(line)

		var segment LogSegmentInfo

		if len(matches) >= 5 {
			// 匹配成功，提取信息
			segment = LogSegmentInfo{
				Timestamp:  matches[1],
				Level:      matches[2],
				Component:  matches[3],
				Message:    matches[4],
				LineNumber: lineNumber + 1, // 行号从1开始
			}
		} else {
			// 无法匹配标准格式，作为原始消息处理
			segment = LogSegmentInfo{
				Timestamp:  time.Now().Format(time.RFC3339),
				Level:      "UNKNOWN",
				Component:  "system",
				Message:    line,
				LineNumber: lineNumber + 1, // 行号从1开始
			}
		}

		// 确定日志类型
		if errorPattern.MatchString(segment.Message) || strings.ToUpper(segment.Level) == "ERROR" {
			segment.Type = "error"
			stats["error"]++
		} else if warningPattern.MatchString(segment.Message) || strings.ToUpper(segment.Level) == "WARN" {
			segment.Type = "warning"
			stats["warning"]++
		} else if infoPattern.MatchString(segment.Message) || strings.ToUpper(segment.Level) == "INFO" {
			segment.Type = "info"
			stats["info"]++
		} else if debugPattern.MatchString(segment.Message) || strings.ToUpper(segment.Level) == "DEBUG" {
			segment.Type = "debug"
			stats["debug"]++
		} else {
			segment.Type = "unknown"
			stats["unknown"]++
		}

		segments = append(segments, segment)
	}

	return segments, stats
}

// saveLogSegmentsToDatabase 将日志分析结果保存到数据库
func saveLogSegmentsToDatabase(containerID string, segments []LogSegmentInfo) error {
	if config.MySQLDB == nil {
		fmt.Printf("MySQL数据库未初始化，无法保存日志分析结果\n")
		return nil
	}

	// 创建仓库实例
	segmentRepo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	containerRepo := repositories.NewMySQLDockerContainerRepo(config.MySQLDB)

	// 获取容器信息
	var containerName string
	if container, err := containerRepo.GetByContainerID(containerID); err == nil {
		containerName = container.ContainerName
	}

	// 转换为数据库模型
	var dbSegments []repositories.ContainerLogSegment
	for _, segment := range segments {
		dbSegment := repositories.ContainerLogSegment{
			ContainerID:   containerID,
			ContainerName: containerName,
			SegmentType:   segment.Type,
			Content:       segment.Message,
			LineNumber:    segment.LineNumber,
			Component:     segment.Component,
			SeverityLevel: determineSeverityLevel(segment.Type),
		}

		// 解析时间戳
		if segment.Timestamp != "" {
			if timestamp, err := time.Parse(time.RFC3339Nano, segment.Timestamp); err == nil {
				dbSegment.Timestamp = &timestamp
			}
		}

		dbSegments = append(dbSegments, dbSegment)
	}

	// 批量保存到数据库
	if err := segmentRepo.CreateBatch(dbSegments); err != nil {
		return fmt.Errorf("保存日志分析结果到数据库失败: %v", err)
	}

	fmt.Printf("成功保存容器 %s 的 %d 个日志段到数据库\n", containerID, len(segments))
	return nil
}

// determineSeverityLevel 根据日志类型确定严重程度
func determineSeverityLevel(segmentType string) string {
	switch strings.ToLower(segmentType) {
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "info":
		return "low"
	case "debug":
		return "low"
	default:
		return "unknown"
	}
}
