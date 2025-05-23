package services

import (
	"fmt"
	"os/exec"
)

// TrafficService 流量管理服务
type TrafficService struct{}

// NewTrafficService 创建流量管理服务实例
func NewTrafficService() *TrafficService {
	return &TrafficService{}
}

// AddRedirectRule 添加流量重定向规则
func (s *TrafficService) AddRedirectRule(sourcePort, targetPort string) error {
	// 添加DNAT规则
	cmd := exec.Command("iptables",
		"-t", "nat",
		"-A", "PREROUTING",
		"-p", "tcp",
		"--dport", sourcePort,
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("127.0.0.1:%s", targetPort),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add DNAT rule: %v", err)
	}

	// 添加SNAT规则
	cmd = exec.Command("iptables",
		"-t", "nat",
		"-A", "POSTROUTING",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", targetPort,
		"-j", "SNAT",
		"--to-source", "127.0.0.1",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add SNAT rule: %v", err)
	}

	return nil
}

// RemoveRedirectRule 删除流量重定向规则
func (s *TrafficService) RemoveRedirectRule(sourcePort, targetPort string) error {
	// 删除DNAT规则
	cmd := exec.Command("iptables",
		"-t", "nat",
		"-D", "PREROUTING",
		"-p", "tcp",
		"--dport", sourcePort,
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("127.0.0.1:%s", targetPort),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove DNAT rule: %v", err)
	}

	// 删除SNAT规则
	cmd = exec.Command("iptables",
		"-t", "nat",
		"-D", "POSTROUTING",
		"-p", "tcp",
		"-d", "127.0.0.1",
		"--dport", targetPort,
		"-j", "SNAT",
		"--to-source", "127.0.0.1",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove SNAT rule: %v", err)
	}

	return nil
}

// ListRules 列出所有规则
func (s *TrafficService) ListRules() (string, error) {
	cmd := exec.Command("iptables", "-t", "nat", "-L", "-n", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list rules: %v", err)
	}

	return string(output), nil
}

// AddFilterRule 添加过滤规则
func (s *TrafficService) AddFilterRule(sourceIP, targetPort string) error {
	cmd := exec.Command("iptables",
		"-A", "INPUT",
		"-s", sourceIP,
		"-p", "tcp",
		"--dport", targetPort,
		"-j", "DROP",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add filter rule: %v", err)
	}

	return nil
}

// RemoveFilterRule 删除过滤规则
func (s *TrafficService) RemoveFilterRule(sourceIP, targetPort string) error {
	cmd := exec.Command("iptables",
		"-D", "INPUT",
		"-s", sourceIP,
		"-p", "tcp",
		"--dport", targetPort,
		"-j", "DROP",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove filter rule: %v", err)
	}

	return nil
}

// SaveRules 保存规则
func (s *TrafficService) SaveRules() error {
	cmd := exec.Command("iptables-save")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to save rules: %v", err)
	}

	return nil
}

// RestoreRules 恢复规则
func (s *TrafficService) RestoreRules() error {
	cmd := exec.Command("iptables-restore")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore rules: %v", err)
	}

	return nil
}
