package services

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// PortManager 端口管理器
type PortManager struct {
	mutex        sync.RWMutex
	allocatedPorts map[int]PortAllocation // 已分配的端口
	portRanges   []PortRange             // 可用端口范围
}

// PortAllocation 端口分配信息
type PortAllocation struct {
	Port        int       `json:"port"`
	ContainerID string    `json:"container_id"`
	ServiceType string    `json:"service_type"` // mysql, ssh, http, etc.
	AllocatedAt time.Time `json:"allocated_at"`
	Description string    `json:"description"`
}

// PortRange 端口范围
type PortRange struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Type  string `json:"type"` // dynamic, reserved, system
}

// 全局端口管理器实例
var (
	globalPortManager *PortManager
	once              sync.Once
)

// GetPortManager 获取全局端口管理器实例
func GetPortManager() *PortManager {
	once.Do(func() {
		globalPortManager = NewPortManager()
	})
	return globalPortManager
}

// NewPortManager 创建新的端口管理器
func NewPortManager() *PortManager {
	pm := &PortManager{
		allocatedPorts: make(map[int]PortAllocation),
		portRanges: []PortRange{
			{Start: 10000, End: 19999, Type: "dynamic"},   // 动态分配端口范围
			{Start: 20000, End: 29999, Type: "reserved"},  // 预留端口范围
			{Start: 30000, End: 39999, Type: "dynamic"},   // 扩展动态端口范围
		},
	}
	
	// 初始化时扫描已占用的端口
	pm.scanOccupiedPorts()
	
	return pm
}

// scanOccupiedPorts 扫描系统已占用的端口
func (pm *PortManager) scanOccupiedPorts() {
	// 扫描常见的系统端口
	systemPorts := []int{22, 80, 443, 3306, 5432, 6379, 8080, 8081, 9000}
	
	for _, port := range systemPorts {
		if pm.isPortOccupied(port) {
			pm.allocatedPorts[port] = PortAllocation{
				Port:        port,
				ContainerID: "system",
				ServiceType: "system",
				AllocatedAt: time.Now(),
				Description: "系统占用端口",
			}
		}
	}
}

// IsPortOccupied 检查端口是否被占用（公开方法）
func (pm *PortManager) IsPortOccupied(port int) bool {
	return pm.isPortOccupied(port)
}

// isPortOccupied 检查端口是否被占用
func (pm *PortManager) isPortOccupied(port int) bool {
	// 尝试监听端口来检查是否被占用
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true // 端口被占用
	}
	listener.Close()
	return false
}

// AllocatePort 分配端口
func (pm *PortManager) AllocatePort(containerID, serviceType, description string) (int, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// 根据服务类型选择合适的端口范围
	var targetRanges []PortRange
	switch serviceType {
	case "mysql":
		// MySQL服务优先使用3306附近的端口
		targetRanges = []PortRange{{Start: 13306, End: 13399, Type: "mysql"}}
	case "ssh":
		// SSH服务优先使用2222附近的端口
		targetRanges = []PortRange{{Start: 12222, End: 12299, Type: "ssh"}}
	case "http", "web":
		// HTTP服务优先使用8000-8999范围
		targetRanges = []PortRange{{Start: 18000, End: 18999, Type: "http"}}
	default:
		// 默认使用动态端口范围
		targetRanges = pm.getDynamicPortRanges()
	}
	
	// 如果特定范围没有可用端口，回退到动态范围
	targetRanges = append(targetRanges, pm.getDynamicPortRanges()...)
	
	for _, portRange := range targetRanges {
		for port := portRange.Start; port <= portRange.End; port++ {
			if pm.isPortAvailable(port) {
				pm.allocatedPorts[port] = PortAllocation{
					Port:        port,
					ContainerID: containerID,
					ServiceType: serviceType,
					AllocatedAt: time.Now(),
					Description: description,
				}
				return port, nil
			}
		}
	}
	
	return 0, fmt.Errorf("没有可用的端口")
}

// AllocateSpecificPort 分配指定端口
func (pm *PortManager) AllocateSpecificPort(port int, containerID, serviceType, description string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if !pm.isPortAvailable(port) {
		return fmt.Errorf("端口 %d 已被占用", port)
	}
	
	pm.allocatedPorts[port] = PortAllocation{
		Port:        port,
		ContainerID: containerID,
		ServiceType: serviceType,
		AllocatedAt: time.Now(),
		Description: description,
	}
	
	return nil
}

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port int) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if _, exists := pm.allocatedPorts[port]; !exists {
		return fmt.Errorf("端口 %d 未被分配", port)
	}
	
	delete(pm.allocatedPorts, port)
	return nil
}

// ReleasePortsByContainer 释放容器的所有端口
func (pm *PortManager) ReleasePortsByContainer(containerID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	var releasedPorts []int
	for port, allocation := range pm.allocatedPorts {
		if allocation.ContainerID == containerID {
			releasedPorts = append(releasedPorts, port)
		}
	}
	
	for _, port := range releasedPorts {
		delete(pm.allocatedPorts, port)
	}
	
	return nil
}

// isPortAvailable 检查端口是否可用
func (pm *PortManager) isPortAvailable(port int) bool {
	// 检查是否已被分配
	if _, exists := pm.allocatedPorts[port]; exists {
		return false
	}

	// 检查是否被系统占用
	return !pm.isPortOccupied(port)
}

// getDynamicPortRanges 获取动态端口范围
func (pm *PortManager) getDynamicPortRanges() []PortRange {
	var dynamicRanges []PortRange
	for _, portRange := range pm.portRanges {
		if portRange.Type == "dynamic" {
			dynamicRanges = append(dynamicRanges, portRange)
		}
	}
	return dynamicRanges
}

// GetPortAllocation 获取端口分配信息
func (pm *PortManager) GetPortAllocation(port int) (PortAllocation, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	allocation, exists := pm.allocatedPorts[port]
	return allocation, exists
}

// GetAllocatedPorts 获取所有已分配的端口
func (pm *PortManager) GetAllocatedPorts() []PortAllocation {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var allocations []PortAllocation
	for _, allocation := range pm.allocatedPorts {
		allocations = append(allocations, allocation)
	}

	// 按端口号排序
	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].Port < allocations[j].Port
	})

	return allocations
}

// GetPortsByContainer 获取容器分配的所有端口
func (pm *PortManager) GetPortsByContainer(containerID string) []PortAllocation {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var allocations []PortAllocation
	for _, allocation := range pm.allocatedPorts {
		if allocation.ContainerID == containerID {
			allocations = append(allocations, allocation)
		}
	}

	// 按端口号排序
	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].Port < allocations[j].Port
	})

	return allocations
}

// GetAvailablePortsInRange 获取指定范围内的可用端口
func (pm *PortManager) GetAvailablePortsInRange(start, end int, limit int) []int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var availablePorts []int
	count := 0

	for port := start; port <= end && count < limit; port++ {
		if pm.isPortAvailable(port) {
			availablePorts = append(availablePorts, port)
			count++
		}
	}

	return availablePorts
}

// GetNextAvailablePort 获取下一个可用端口
func (pm *PortManager) GetNextAvailablePort(startPort int) (int, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 从指定端口开始查找
	for port := startPort; port <= 65535; port++ {
		if pm.isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("从端口 %d 开始没有找到可用端口", startPort)
}

// ValidatePortRange 验证端口范围
func (pm *PortManager) ValidatePortRange(start, end int) error {
	if start < 1 || start > 65535 {
		return fmt.Errorf("起始端口 %d 超出有效范围 (1-65535)", start)
	}

	if end < 1 || end > 65535 {
		return fmt.Errorf("结束端口 %d 超出有效范围 (1-65535)", end)
	}

	if start > end {
		return fmt.Errorf("起始端口 %d 不能大于结束端口 %d", start, end)
	}

	return nil
}

// GetPortStatistics 获取端口统计信息
func (pm *PortManager) GetPortStatistics() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := make(map[string]interface{})

	// 总分配端口数
	stats["total_allocated"] = len(pm.allocatedPorts)

	// 按服务类型统计
	serviceStats := make(map[string]int)
	for _, allocation := range pm.allocatedPorts {
		serviceStats[allocation.ServiceType]++
	}
	stats["by_service_type"] = serviceStats

	// 按端口范围统计可用端口
	rangeStats := make(map[string]interface{})
	for _, portRange := range pm.portRanges {
		available := 0
		total := portRange.End - portRange.Start + 1

		for port := portRange.Start; port <= portRange.End; port++ {
			if pm.isPortAvailable(port) {
				available++
			}
		}

		rangeStats[fmt.Sprintf("%s_%d_%d", portRange.Type, portRange.Start, portRange.End)] = map[string]int{
			"total":     total,
			"available": available,
			"used":      total - available,
		}
	}
	stats["by_range"] = rangeStats

	return stats
}

// AutoAllocatePortMapping 自动分配端口映射
func (pm *PortManager) AutoAllocatePortMapping(containerID string, portMappings map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	for containerPort, hostPort := range portMappings {
		// 验证容器端口
		if _, err := strconv.Atoi(containerPort); err != nil {
			return nil, fmt.Errorf("无效的容器端口: %s", containerPort)
		}

		// 如果主机端口为空或为"auto"，自动分配
		if hostPort == "" || hostPort == "auto" {
			// 根据容器端口推断服务类型
			serviceType := pm.inferServiceType(containerPort)

			allocatedPort, err := pm.AllocatePort(containerID, serviceType, fmt.Sprintf("容器端口 %s 的自动映射", containerPort))
			if err != nil {
				return nil, fmt.Errorf("为容器端口 %s 分配主机端口失败: %v", containerPort, err)
			}

			result[containerPort] = strconv.Itoa(allocatedPort)
		} else {
			// 使用指定的主机端口
			hostPortInt, err := strconv.Atoi(hostPort)
			if err != nil {
				return nil, fmt.Errorf("无效的主机端口: %s", hostPort)
			}

			serviceType := pm.inferServiceType(containerPort)
			err = pm.AllocateSpecificPort(hostPortInt, containerID, serviceType, fmt.Sprintf("容器端口 %s 的指定映射", containerPort))
			if err != nil {
				return nil, fmt.Errorf("分配指定主机端口 %s 失败: %v", hostPort, err)
			}

			result[containerPort] = hostPort
		}
	}

	return result, nil
}

// inferServiceType 根据端口推断服务类型
func (pm *PortManager) inferServiceType(containerPort string) string {
	port, err := strconv.Atoi(containerPort)
	if err != nil {
		return "unknown"
	}

	switch port {
	case 22:
		return "ssh"
	case 80, 8080, 8000, 8888:
		return "http"
	case 443, 8443:
		return "https"
	case 3306:
		return "mysql"
	case 5432:
		return "postgresql"
	case 6379:
		return "redis"
	case 27017:
		return "mongodb"
	case 21:
		return "ftp"
	case 25, 587:
		return "smtp"
	case 53:
		return "dns"
	default:
		return "unknown"
	}
}
