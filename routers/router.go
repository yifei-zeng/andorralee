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
			ai.POST("/semantic-segment", handlers.SemanticSegment)
		}
	}

	// 4. Swagger 文档路由
	// 访问 http://localhost:8080/swagger/index.html 查看接口文档
	// Swagger 文档路由 - 使用静态文件路径
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 5. 返回路由引擎
	return r
}
