package main

// 暂时禁用swagger文档
// _ "andorralee/docs"
import (
	"andorralee/internal/config"
	"andorralee/routers" // 导入路由包
	"fmt"
)

// @title           Andorralee Docker API
// @version         1.0
// @description     管理 Docker 镜像和数据库的接口
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// 初始化配置
	// 尝试初始化Docker客户端，但允许失败
	if err := config.InitDockerClient(); err != nil {
		fmt.Println("警告: Docker服务未启动或不可用，部分功能将不可用")
	}

	// 尝试初始化数据库，但允许失败
	if err := config.InitDatabase(); err != nil {
		fmt.Println("警告: 数据库连接失败，相关功能将不可用")
	} else {
		// 初始化数据库表
		if err := config.InitTables(); err != nil {
			fmt.Println("警告: 数据库表初始化失败，相关功能可能不可用:", err)
		}
	}

	// 初始化路由
	r := routers.SetupRouter() // 通过路由包获取已配置的 Gin 引擎

	fmt.Println("服务启动中，监听端口: 8848...")
	// 启动服务
	err := r.Run(":8848")
	if err != nil {
		fmt.Println("服务启动失败:", err)
		return
	}
}
