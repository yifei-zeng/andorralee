package handlers

import (
	"andorralee/internal/config"
	"andorralee/pkg/utils"
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthStatus 健康状态结构
type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Uptime      string                 `json:"uptime"`
	System      SystemInfo             `json:"system"`
	Services    map[string]ServiceInfo `json:"services"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

var startTime = time.Now()

// HealthCheck 健康检查端点
// @Summary 健康检查
// @Description 检查系统健康状态，包括数据库连接、Docker服务等
// @Tags 系统监控
// @Produce json
// @Success 200 {object} HealthStatus
// @Failure 503 {object} utils.Response
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	health := HealthStatus{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     "1.0.0",
		Environment: getEnvironment(),
		Uptime:      time.Since(startTime).String(),
		System: SystemInfo{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Services: make(map[string]ServiceInfo),
	}

	// 检查MySQL数据库连接
	mysqlStatus := checkMySQLHealth()
	health.Services["mysql"] = mysqlStatus

	// 检查达梦数据库连接
	damengStatus := checkDamengHealth()
	health.Services["dameng"] = damengStatus

	// 检查Docker服务
	dockerStatus := checkDockerHealth()
	health.Services["docker"] = dockerStatus

	// 确定整体健康状态
	overallStatus := determineOverallStatus(health.Services)
	health.Status = overallStatus

	// 根据健康状态返回相应的HTTP状态码
	if overallStatus == "healthy" {
		utils.ResponseSuccess(c, health)
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "服务不健康",
			"data":    health,
		})
	}
}

// getEnvironment 获取运行环境
func getEnvironment() string {
	if config.MySQLDB != nil || config.DamengDB != nil {
		return "production"
	}
	return "development"
}

// checkMySQLHealth 检查MySQL数据库健康状态
func checkMySQLHealth() ServiceInfo {
	if config.MySQLDB == nil {
		return ServiceInfo{
			Status:  "unavailable",
			Message: "MySQL数据库未初始化",
		}
	}

	// 尝试ping数据库
	sqlDB, err := config.MySQLDB.DB()
	if err != nil {
		return ServiceInfo{
			Status:  "error",
			Message: "无法获取MySQL数据库连接: " + err.Error(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return ServiceInfo{
			Status:  "error",
			Message: "MySQL数据库连接失败: " + err.Error(),
		}
	}

	return ServiceInfo{
		Status:  "healthy",
		Message: "MySQL数据库连接正常",
	}
}

// checkDamengHealth 检查达梦数据库健康状态
func checkDamengHealth() ServiceInfo {
	if config.DamengDB == nil {
		return ServiceInfo{
			Status:  "unavailable",
			Message: "达梦数据库未初始化",
		}
	}

	// 尝试ping数据库
	sqlDB, err := config.DamengDB.DB()
	if err != nil {
		return ServiceInfo{
			Status:  "error",
			Message: "无法获取达梦数据库连接: " + err.Error(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return ServiceInfo{
			Status:  "error",
			Message: "达梦数据库连接失败: " + err.Error(),
		}
	}

	return ServiceInfo{
		Status:  "healthy",
		Message: "达梦数据库连接正常",
	}
}

// checkDockerHealth 检查Docker服务健康状态
func checkDockerHealth() ServiceInfo {
	if config.DockerCli == nil {
		return ServiceInfo{
			Status:  "unavailable",
			Message: "Docker客户端未初始化",
		}
	}

	// 尝试ping Docker守护进程
	ctx := context.Background()
	_, err := config.DockerCli.Ping(ctx)
	if err != nil {
		return ServiceInfo{
			Status:  "error",
			Message: "Docker服务连接失败: " + err.Error(),
		}
	}

	return ServiceInfo{
		Status:  "healthy",
		Message: "Docker服务连接正常",
	}
}

// determineOverallStatus 确定整体健康状态
func determineOverallStatus(services map[string]ServiceInfo) string {
	hasError := false
	hasHealthy := false

	for _, service := range services {
		switch service.Status {
		case "healthy":
			hasHealthy = true
		case "error":
			hasError = true
		case "unavailable":
			// 服务不可用不影响整体状态，只要有其他服务正常即可
		}
	}

	// 如果有错误，整体状态为不健康
	if hasError {
		return "unhealthy"
	}

	// 如果有健康的服务，整体状态为健康
	if hasHealthy {
		return "healthy"
	}

	// 如果所有服务都不可用，整体状态为降级
	return "degraded"
}

// SimpleHealthCheck 简单健康检查端点（用于Docker健康检查）
// @Summary 简单健康检查
// @Description 简单的健康检查端点，仅返回基本状态
// @Tags 系统监控
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func SimpleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "andorralee",
	})
}

// ReadinessCheck 就绪检查端点
// @Summary 就绪检查
// @Description 检查服务是否准备好接收请求
// @Tags 系统监控
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} utils.Response
// @Router /api/v1/ready [get]
func ReadinessCheck(c *gin.Context) {
	// 检查关键服务是否就绪
	ready := true
	services := make(map[string]bool)

	// 检查数据库连接（至少一个数据库可用）
	mysqlReady := config.MySQLDB != nil
	damengReady := config.DamengDB != nil

	services["mysql"] = mysqlReady
	services["dameng"] = damengReady

	// 至少需要一个数据库可用
	if !mysqlReady && !damengReady {
		ready = false
	}

	response := gin.H{
		"ready":     ready,
		"timestamp": time.Now().Format(time.RFC3339),
		"services":  services,
	}

	if ready {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// LivenessCheck 存活检查端点
// @Summary 存活检查
// @Description 检查服务是否存活
// @Tags 系统监控
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/live [get]
func LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive":     true,
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}
