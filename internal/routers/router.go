package routers

import (
	"andorralee/internal/handlers"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 创建处理器实例
	honeypotHandler := handlers.NewHoneypotHandler()
	trafficHandler := handlers.NewTrafficHandler()
	baitHandler := handlers.NewBaitHandler(filepath.Join("data", "baits"))
	monitorHandler := handlers.NewMonitorHandler(filepath.Join("data", "monitor"))

	// 蜜罐相关路由
	honeypot := r.Group("/api/honeypot")
	{
		honeypot.POST("/deploy", honeypotHandler.DeployHoneypot)
		honeypot.POST("/stop/:id", honeypotHandler.StopHoneypot)
		honeypot.GET("/status/:id", honeypotHandler.GetHoneypotStatus)
		honeypot.GET("/list", honeypotHandler.ListHoneypots)
		honeypot.GET("/logs/:id", honeypotHandler.GetHoneypotLogs)
	}

	// 流量管理相关路由
	traffic := r.Group("/api/traffic")
	{
		traffic.POST("/redirect/add", trafficHandler.AddRedirectRule)
		traffic.POST("/redirect/remove", trafficHandler.RemoveRedirectRule)
		traffic.POST("/filter/add", trafficHandler.AddFilterRule)
		traffic.POST("/filter/remove", trafficHandler.RemoveFilterRule)
		traffic.GET("/rules", trafficHandler.ListRules)
		traffic.POST("/rules/save", trafficHandler.SaveRules)
		traffic.POST("/rules/restore", trafficHandler.RestoreRules)
	}

	// 蜜签相关路由
	bait := r.Group("/api/bait")
	{
		bait.POST("/create", baitHandler.CreateBait)
		bait.GET("/:id", baitHandler.GetBait)
		bait.GET("/list", baitHandler.ListBaits)
		bait.DELETE("/:id", baitHandler.DeleteBait)
		bait.POST("/:id/monitor", baitHandler.MonitorBait)
	}

	// 监控和告警相关路由
	monitor := r.Group("/api/monitor")
	{
		monitor.POST("/alert", monitorHandler.CreateAlert)
		monitor.POST("/alert/:id/resolve", monitorHandler.ResolveAlert)
		monitor.GET("/alerts", monitorHandler.ListAlerts)
		monitor.GET("/alert/:id", monitorHandler.GetAlert)
		monitor.POST("/honeypot/:id", monitorHandler.MonitorHoneypot)
		monitor.POST("/bait/:id", monitorHandler.MonitorBait)
		monitor.POST("/traffic", monitorHandler.MonitorTraffic)
	}

	return r
}
