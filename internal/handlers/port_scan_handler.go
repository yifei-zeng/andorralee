package handlers

import (
	"andorralee/pkg/utils"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PortScanResult 端口扫描结果
type PortScanResult struct {
	IP       string      `json:"ip"`
	Port     int         `json:"port"`
	Protocol string      `json:"protocol"`
	Status   string      `json:"status"` // open, closed, filtered
	Service  string      `json:"service"`
	Banner   string      `json:"banner"`
	ScanTime time.Time   `json:"scan_time"`
	Duration int64       `json:"duration_ms"`
}

// PortScanRequest 端口扫描请求
type PortScanRequest struct {
	Target   string `json:"target" binding:"required"`   // IP地址或主机名
	Ports    string `json:"ports"`                       // 端口范围，如 "22,80,443" 或 "1-1000"
	Protocol string `json:"protocol"`                    // tcp, udp
	Timeout  int    `json:"timeout"`                     // 超时时间（秒）
}

// ScanPorts 扫描端口
func ScanPorts(c *gin.Context) {
	var req PortScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Protocol == "" {
		req.Protocol = "tcp"
	}
	if req.Timeout == 0 {
		req.Timeout = 3
	}
	if req.Ports == "" {
		req.Ports = "22,80,443,3306,3389,8080"
	}

	// 解析端口列表
	ports, err := parsePorts(req.Ports)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "端口格式错误: "+err.Error())
		return
	}

	// 限制端口数量
	if len(ports) > 1000 {
		utils.ResponseError(c, http.StatusBadRequest, "端口数量过多，最多支持1000个端口")
		return
	}

	// 执行扫描
	results := scanPortsConcurrent(req.Target, ports, req.Protocol, req.Timeout)

	response := map[string]interface{}{
		"target":      req.Target,
		"protocol":    req.Protocol,
		"total_ports": len(ports),
		"open_ports":  countOpenPorts(results),
		"scan_time":   time.Now().Format(time.RFC3339),
		"results":     results,
	}

	utils.ResponseSuccess(c, response)
}

// parsePorts 解析端口字符串
func parsePorts(portStr string) ([]int, error) {
	var ports []int
	
	parts := strings.Split(portStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		if strings.Contains(part, "-") {
			// 端口范围
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("无效的端口范围: %s", part)
			}
			
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("无效的起始端口: %s", rangeParts[0])
			}
			
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("无效的结束端口: %s", rangeParts[1])
			}
			
			if start > end || start < 1 || end > 65535 {
				return nil, fmt.Errorf("无效的端口范围: %d-%d", start, end)
			}
			
			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			// 单个端口
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("无效的端口: %s", part)
			}
			
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("端口超出范围: %d", port)
			}
			
			ports = append(ports, port)
		}
	}
	
	return ports, nil
}

// scanPortsConcurrent 并发扫描端口
func scanPortsConcurrent(target string, ports []int, protocol string, timeout int) []PortScanResult {
	var results []PortScanResult
	var mutex sync.Mutex
	var wg sync.WaitGroup
	
	// 限制并发数
	semaphore := make(chan struct{}, 50)
	
	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := scanSinglePort(target, p, protocol, timeout)
			
			mutex.Lock()
			results = append(results, result)
			mutex.Unlock()
		}(port)
	}
	
	wg.Wait()
	return results
}

// scanSinglePort 扫描单个端口
func scanSinglePort(target string, port int, protocol string, timeout int) PortScanResult {
	startTime := time.Now()
	
	result := PortScanResult{
		IP:       target,
		Port:     port,
		Protocol: protocol,
		Status:   "closed",
		ScanTime: startTime,
	}
	
	address := fmt.Sprintf("%s:%d", target, port)
	
	conn, err := net.DialTimeout(protocol, address, time.Duration(timeout)*time.Second)
	if err != nil {
		// 端口关闭或过滤
		if strings.Contains(err.Error(), "refused") {
			result.Status = "closed"
		} else {
			result.Status = "filtered"
		}
	} else {
		// 端口开放
		result.Status = "open"
		result.Service = getServiceName(port)
		
		// 尝试获取banner
		if banner := getBanner(conn, timeout); banner != "" {
			result.Banner = banner
		}
		
		conn.Close()
	}
	
	result.Duration = time.Since(startTime).Milliseconds()
	return result
}

// getServiceName 根据端口号获取服务名称
func getServiceName(port int) string {
	services := map[int]string{
		21:    "ftp",
		22:    "ssh",
		23:    "telnet",
		25:    "smtp",
		53:    "dns",
		80:    "http",
		110:   "pop3",
		143:   "imap",
		443:   "https",
		993:   "imaps",
		995:   "pop3s",
		1433:  "mssql",
		3306:  "mysql",
		3389:  "rdp",
		5432:  "postgresql",
		6379:  "redis",
		8080:  "http-alt",
		8443:  "https-alt",
		27017: "mongodb",
	}
	
	if service, exists := services[port]; exists {
		return service
	}
	return "unknown"
}

// getBanner 获取服务banner
func getBanner(conn net.Conn, timeout int) string {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return ""
	}
	
	banner := strings.TrimSpace(string(buffer[:n]))
	// 限制banner长度
	if len(banner) > 200 {
		banner = banner[:200] + "..."
	}
	
	return banner
}

// countOpenPorts 统计开放端口数量
func countOpenPorts(results []PortScanResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "open" {
			count++
		}
	}
	return count
}

// ScanContainerPorts 扫描容器端口
func ScanContainerPorts(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	// 获取容器实例
	instanceMutex.RLock()
	instance, exists := memoryInstances[uint(id)]
	instanceMutex.RUnlock()

	if !exists {
		utils.ResponseError(c, http.StatusNotFound, "容器实例不存在")
		return
	}

	// 构建端口列表
	var ports []string
	for _, hostPort := range instance.PortMappings {
		ports = append(ports, hostPort)
	}

	if len(ports) == 0 {
		utils.ResponseError(c, http.StatusBadRequest, "容器没有映射端口")
		return
	}

	// 扫描端口
	target := "127.0.0.1" // 扫描本地主机
	if instance.HoneypotIP != "" {
		target = instance.HoneypotIP
	}

	portList, err := parsePorts(strings.Join(ports, ","))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "解析端口失败: "+err.Error())
		return
	}

	results := scanPortsConcurrent(target, portList, "tcp", 3)

	response := map[string]interface{}{
		"container_id":   instance.ID,
		"container_name": instance.ContainerName,
		"target":         target,
		"total_ports":    len(portList),
		"open_ports":     countOpenPorts(results),
		"scan_time":      time.Now().Format(time.RFC3339),
		"results":        results,
	}

	utils.ResponseSuccess(c, response)
}

// GetPortScanHistory 获取端口扫描历史（简单实现）
func GetPortScanHistory(c *gin.Context) {
	// 这里可以实现扫描历史记录功能
	// 目前返回空列表
	utils.ResponseSuccess(c, []interface{}{})
}
