package test

import (
	"andorralee/internal/services"
	"fmt"
	"testing"
)

// TestLogSegmentation 测试容器日志语义分割功能
func TestLogSegmentation(t *testing.T) {
	// 模拟容器日志内容
	logContent := `2023-01-01T12:00:00.000000000Z INFO [app] 应用启动成功
2023-01-01T12:01:00.000000000Z WARN [security] 检测到可疑登录尝试
2023-01-01T12:02:00.000000000Z ERROR [database] 数据库连接失败: connection refused
2023-01-01T12:03:00.000000000Z DEBUG [cache] 缓存刷新完成
未格式化的日志行，应该被识别为未知类型`
	// 调用日志分析函数
	segments, stats := services.AnalyzeContainerLogs(logContent)

	// 验证结果
	if len(segments) != 5 {
		t.Errorf("期望5个日志片段，实际得到%d个", len(segments))
	}

	// 验证统计信息
	expectedStats := map[string]int{
		"error":   1,
		"warning": 1,
		"info":    1,
		"debug":   1,
		"unknown": 1,
	}

	for category, count := range expectedStats {
		if stats[category] != count {
			t.Errorf("类别 %s 期望数量 %d，实际数量 %d", category, count, stats[category])
		}
	}

	// 打印结果
	fmt.Println("日志分析结果:")
	for i, segment := range segments {
		fmt.Printf("[%d] 类型: %s, 级别: %s, 组件: %s, 消息: %s\n",
			i, segment.Type, segment.Level, segment.Component, segment.Message)
	}
}
