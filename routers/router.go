package routers

import (
	"andorralee/internal/handlers" // 替换为你的模块路径
	"andorralee/pkg/middleware"    // 中间件包

	"github.com/gin-gonic/gin"
	// 暂时禁用 swagger 相关导入
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 初始化路由
// 返回值 *gin.Engine 是 Gin 框架的核心引擎，用于处理 HTTP 请求
func SetupRouter() *gin.Engine {
	// 1. 创建默认 Gin 引擎（包含日志和恢复中间件）
	r := gin.Default()

	// 2. 添加全局中间件
	// - 跨域处理（允许前端访问）
	r.Use(middleware.Cors())

	// 删除静态文件路由，避免冲突
	// r.Static("/swagger", "./static/swagger")

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
			docker.GET("/images/:id", handlers.GetImageByID)
			docker.DELETE("/images/:id", handlers.DeleteImage)
			docker.POST("/images/:id/tag", handlers.TagImage)
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

		// ------------------------------ Headling认证日志接口 ------------------------------
		headling := api.Group("/headling")
		{
			// 日志拉取和管理
			headling.POST("/pull-logs", handlers.PullHeadlingLogs)                                   // 拉取认证日志
			headling.GET("/logs", handlers.GetAllHeadlingLogs)                                       // 获取所有日志
			headling.GET("/logs/:id", handlers.GetHeadlingLogByID)                                   // 根据ID获取日志
			headling.GET("/logs/container/:container_id", handlers.GetHeadlingLogsByContainer)       // 根据容器ID获取日志
			headling.GET("/logs/source-ip/:source_ip", handlers.GetHeadlingLogsBySourceIP)           // 根据源IP获取日志
			headling.GET("/logs/protocol/:protocol", handlers.GetHeadlingLogsByProtocol)             // 根据协议获取日志
			headling.GET("/logs/time-range", handlers.GetHeadlingLogsByTimeRange)                    // 根据时间范围获取日志
			headling.DELETE("/logs/container/:container_id", handlers.DeleteHeadlingLogsByContainer) // 删除容器相关日志

			// 统计和分析
			headling.GET("/statistics", handlers.GetHeadlingStatistics)            // 获取统计信息
			headling.GET("/attacker-statistics", handlers.GetAttackerIPStatistics) // 获取攻击者IP统计
			headling.GET("/top-attackers", handlers.GetTopAttackers)               // 获取顶级攻击者
			headling.GET("/top-usernames", handlers.GetTopUsernames)               // 获取常用用户名
			headling.GET("/top-passwords", handlers.GetTopPasswords)               // 获取常用密码
		}

		// ------------------------------ Cowrie蜜罐日志接口 ------------------------------
		cowrie := api.Group("/cowrie")
		{
			// 日志拉取和管理
			cowrie.POST("/pull-logs", handlers.PullCowrieLogs)                                   // 拉取蜜罐日志
			cowrie.GET("/logs", handlers.GetAllCowrieLogs)                                       // 获取所有日志
			cowrie.GET("/logs/:id", handlers.GetCowrieLogByID)                                   // 根据ID获取日志
			cowrie.GET("/logs/container/:container_id", handlers.GetCowrieLogsByContainer)       // 根据容器ID获取日志
			cowrie.GET("/logs/source-ip/:source_ip", handlers.GetCowrieLogsBySourceIP)           // 根据源IP获取日志
			cowrie.GET("/logs/protocol/:protocol", handlers.GetCowrieLogsByProtocol)             // 根据协议获取日志
			cowrie.GET("/logs/command/:command", handlers.GetCowrieLogsByCommand)                // 根据命令获取日志
			cowrie.GET("/logs/username/:username", handlers.GetCowrieLogsByUsername)             // 根据用户名获取日志
			cowrie.GET("/logs/command-found/:found", handlers.GetCowrieLogsByCommandFound)       // 根据命令识别状态获取日志
			cowrie.GET("/logs/time-range", handlers.GetCowrieLogsByTimeRange)                    // 根据时间范围获取日志
			cowrie.DELETE("/logs/container/:container_id", handlers.DeleteCowrieLogsByContainer) // 删除容器相关日志

			// 统计和分析
			cowrie.GET("/statistics", handlers.GetCowrieStatistics)              // 获取统计信息
			cowrie.GET("/attacker-behavior", handlers.GetCowrieAttackerBehavior) // 获取攻击者行为统计
			cowrie.GET("/top-attackers", handlers.GetCowrieTopAttackers)         // 获取顶级攻击者
			cowrie.GET("/top-commands", handlers.GetCowrieTopCommands)           // 获取常用命令
			cowrie.GET("/top-usernames", handlers.GetCowrieTopUsernames)         // 获取常用用户名
			cowrie.GET("/top-passwords", handlers.GetCowrieTopPasswords)         // 获取常用密码
			cowrie.GET("/top-fingerprints", handlers.GetCowrieTopFingerprints)   // 获取常用指纹
		}

		// ------------------------------ 容器实例管理接口 ------------------------------
		containerInstances := api.Group("/container-instances")
		{
			// 实例管理
			containerInstances.POST("", handlers.CreateContainerInstance)       // 创建容器实例
			containerInstances.GET("", handlers.GetAllContainerInstances)       // 获取所有容器实例
			containerInstances.GET("/:id", handlers.GetContainerInstanceByID)   // 根据ID获取容器实例
			containerInstances.DELETE("/:id", handlers.DeleteContainerInstance) // 删除容器实例

			// 实例控制
			containerInstances.POST("/:id/start", handlers.StartContainerInstance)     // 启动容器实例
			containerInstances.POST("/:id/stop", handlers.StopContainerInstance)       // 停止容器实例
			containerInstances.POST("/:id/restart", handlers.RestartContainerInstance) // 重启容器实例

			// 状态管理
			containerInstances.GET("/:id/status", handlers.GetContainerInstanceStatus)        // 获取容器实例状态
			containerInstances.GET("/status/:status", handlers.GetContainerInstancesByStatus) // 根据状态获取容器实例
			containerInstances.POST("/sync-status", handlers.SyncAllContainerInstancesStatus) // 同步所有容器实例状态
		}

		// ------------------------------ 容器日志分析接口 ------------------------------
		containerLogs := api.Group("/container-logs")
		{
			containerLogs.GET("/segments", handlers.GetAllContainerLogSegments)                                // 获取所有日志分析结果
			containerLogs.GET("/segments/:id", handlers.GetContainerLogSegmentByID)                            // 根据ID获取分析结果
			containerLogs.GET("/segments/container/:container_id", handlers.GetLogSegmentsByContainerID)       // 根据容器ID获取分析结果
			containerLogs.GET("/segments/type/:type", handlers.GetLogSegmentsByType)                           // 根据类型获取分析结果
			containerLogs.DELETE("/segments/:id", handlers.DeleteContainerLogSegment)                          // 删除分析结果
			containerLogs.DELETE("/segments/container/:container_id", handlers.DeleteLogSegmentsByContainerID) // 删除容器相关分析结果
		}

		// ------------------------------ Docker镜像日志接口 ------------------------------
		docker.GET("/image-logs", handlers.GetAllDockerImageLogs)                       // 获取所有镜像操作日志
		docker.GET("/image-logs/:id", handlers.GetDockerImageLogByID)                   // 根据ID获取镜像操作日志
		docker.GET("/image-logs/image/:image_id", handlers.GetDockerImageLogsByImageID) // 根据镜像ID获取操作日志
		docker.DELETE("/image-logs/:id", handlers.DeleteDockerImageLog)                 // 删除镜像操作日志
		docker.GET("/images/db", handlers.GetDockerImages)                              // 获取数据库中的镜像记录
		docker.GET("/images/db/:id", handlers.GetDockerImageByDBID)                     // 根据数据库ID获取镜像记录
		docker.DELETE("/images/db/:id", handlers.DeleteDockerImageRecord)               // 删除镜像数据库记录
	}

	// 暂时禁用 Swagger 文档路由
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 5. 返回路由引擎
	return r
}
