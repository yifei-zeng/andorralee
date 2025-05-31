package routers

import (
	"andorralee/internal/handlers" // 替换为你的模块路径
	"andorralee/pkg/middleware"    // 中间件包

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 初始化路由
// 返回值 *gin.Engine 是 Gin 框架的核心引擎，用于处理 HTTP 请求
func SetupRouter() *gin.Engine {
	// 1. 创建默认 Gin 引擎（包含日志和恢复中间件）
	r := gin.Default()

	// 2. 添加全局中间件
	// - 跨域处理（允许前端访问）
	r.Use(middleware.Cors())

	// 静态文件服务
	r.Static("/swagger", "./static/swagger")

	// 3. 定义 API 路由分组 `/api/v1`
	api := r.Group("/api/v1")
	{
		// ------------------------------ Docker 操作接口 ------------------------------
		// 拉取镜像
		docker := api.Group("/docker")
		{
			docker.POST("/pull", handlers.PullImage)
			docker.POST("/start", handlers.StartContainer)
			docker.POST("/stop", handlers.StopContainer)
			docker.GET("/images", handlers.ListImages)
			docker.GET("/logs", handlers.GetContainerLogs)
			docker.GET("/containers", handlers.ListContainers)
			docker.GET("/container/:id", handlers.GetContainerInfo)
		}

		// ------------------------------ 蜜罐管理接口 ------------------------------
		honeypot := api.Group("/honeypot")
		{
			// 蜜罐模板管理
			templates := honeypot.Group("/templates")
			{
				templates.GET("", handlers.GetAllTemplates)
				templates.GET("/:id", handlers.GetTemplateByID)
				templates.POST("", handlers.CreateTemplate)
				templates.PUT("/:id", handlers.UpdateTemplate)
				templates.DELETE("/:id", handlers.DeleteTemplate)
				templates.POST("/import", handlers.ImportTemplate)
				templates.POST("/:id/deploy", handlers.DeployTemplate)
			}

			// 蜜罐实例管理
			instances := honeypot.Group("/instances")
			{
				instances.GET("", handlers.GetAllInstances)
				instances.GET("/:id", handlers.GetInstanceByID)
				instances.PUT("/:id", handlers.UpdateInstance)
				instances.DELETE("/:id", handlers.DeleteInstance)
				instances.POST("/:id/deploy", handlers.DeployInstance)
				instances.POST("/:id/stop", handlers.StopInstance)
				instances.GET("/:id/logs", handlers.GetInstanceLogs)
			}

			// 蜜罐日志管理
			logs := honeypot.Group("/logs")
			{
				logs.GET("", handlers.GetAllHoneypotLogs)
				logs.GET("/:id", handlers.GetHoneypotLogByID)
				logs.GET("/instance/:id", handlers.GetLogsByInstanceID)
			}
		}

		// ------------------------------ 诱饵(蜜签)管理接口 ------------------------------
		baits := api.Group("/baits")
		{
			baits.GET("", handlers.GetAllBaits)
			baits.GET("/:id", handlers.GetBaitByID)
			baits.POST("", handlers.CreateBait)
			baits.PUT("/:id", handlers.UpdateBait)
			baits.DELETE("/:id", handlers.DeleteBait)
			baits.POST("/:id/deploy", handlers.DeployBait)
		}

		// ------------------------------ 安全规则管理接口 ------------------------------
		rules := api.Group("/rules")
		{
			rules.GET("", handlers.GetAllRules)
			rules.GET("/:id", handlers.GetRuleByID)
			rules.POST("", handlers.CreateRule)
			rules.PUT("/:id", handlers.UpdateRule)
			rules.DELETE("/:id", handlers.DeleteRule)
			rules.PUT("/:id/enable", handlers.EnableRule)
			rules.PUT("/:id/disable", handlers.DisableRule)

			// 规则日志
			ruleLogs := rules.Group("/logs")
			{
				ruleLogs.GET("", handlers.GetAllRuleLogs)
				ruleLogs.GET("/:id", handlers.GetRuleLogByID)
				ruleLogs.GET("/rule/:id", handlers.GetLogsByRuleID)
			}
		}

		// ------------------------------ 数据库操作接口 ------------------------------
		// 查询数据库字段（支持 MySQL 和达梦）
		data := api.Group("/data")
		{
			data.GET("", handlers.QueryData)
			data.POST("", handlers.CreateData)
			data.PUT("", handlers.UpdateData)
			data.DELETE("", handlers.DeleteData)
			data.GET("/id", handlers.GetDataByID)
			data.GET("/name", handlers.GetDataByName)
		}

		// ------------------------------ AI 功能接口 ------------------------------
		// 语义分割
		ai := api.Group("/ai")
		{
			ai.POST("/semantic-segment", handlers.SemanticSegment)   // 日志语义分割
			ai.POST("/image-segment", handlers.ImageSemanticSegment) // 图像语义分割
		}
	}

	// 4. Swagger 文档路由
	// 访问 http://localhost:8080/swagger/index.html 查看接口文档
	// Swagger 文档路由 - 使用静态文件路径
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 5. 返回路由引擎
	return r
}
