package services

import (
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
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Type      string `json:"type"`
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
	for _, line := range logLines {
		if line == "" {
			continue
		}

		matches := logPattern.FindStringSubmatch(line)

		var segment LogSegmentInfo

		if len(matches) >= 5 {
			// 匹配成功，提取信息
			segment = LogSegmentInfo{
				Timestamp: matches[1],
				Level:     matches[2],
				Component: matches[3],
				Message:   matches[4],
			}
		} else {
			// 无法匹配标准格式，作为原始消息处理
			segment = LogSegmentInfo{
				Timestamp: time.Now().Format(time.RFC3339),
				Level:     "UNKNOWN",
				Component: "system",
				Message:   line,
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
	// 创建数据库服务实例
	dbService, err := NewDatabaseService("mysql")
	if err != nil {
		return fmt.Errorf("创建数据库服务失败: %v", err)
	}

	// 将每个日志片段保存为一条记录
	for _, segment := range segments {
		// 创建数据模型
		data := &repositories.DataModel{
			Name:      containerID,
			Behavior:  segment.Type,
			Data:      fmt.Sprintf("%s [%s] %s", segment.Timestamp, segment.Component, segment.Message),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 保存到数据库
		if err := dbService.CreateData(data); err != nil {
			return fmt.Errorf("保存日志片段失败: %v", err)
		}
	}

	return nil
}
