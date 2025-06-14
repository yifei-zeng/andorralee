package main

// 暂时禁用swagger文档
// _ "andorralee/docs"
import (
	"andorralee/internal/config"
	"andorralee/internal/handlers"
	"andorralee/routers" // 导入路由包
	"fmt"
	"os"
)

// @title           Andorralee Docker API
// @version         1.0
// @description     管理 Docker 镜像和数据库的接口
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// 设置达梦数据库环境变量，避免重复定义标志
	os.Setenv("DM_HOME", "./dm_home")

	// 创建dm_home目录
	if err := os.MkdirAll("dm_home", 0755); err != nil {
		fmt.Println("创建dm_home目录失败:", err)
	}

	// 初始化配置
	// 尝试初始化Docker客户端，但允许失败
	if err := config.InitDockerClient(); err != nil {
		fmt.Println("警告: Docker服务未启动或不可用，部分功能将不可用")
	}

	// 尝试初始化MySQL，但允许失败
	if err := config.InitMySQL(); err != nil {
		fmt.Println("警告: MySQL数据库连接失败，相关功能将不可用")
	} else {
		// 初始化数据库表
		if err := config.InitTables(); err != nil {
			fmt.Println("警告: MySQL数据库表初始化失败，相关功能可能不可用:", err)
		}
	}

	// 尝试初始化达梦数据库，但允许失败
	if err := config.InitDameng(); err != nil {
		fmt.Println("警告: 达梦数据库连接失败，相关功能将不可用")
	} else {
		fmt.Println("达梦数据库连接成功！")
	}

	// 初始化路由
	r := routers.SetupRouter() // 通过路由包获取已配置的 Gin 引擎

	// 初始化默认蜜签
	handlers.CreateDefaultHoneyTokens()
	fmt.Println("✅ 默认蜜签初始化完成")

	fmt.Println("服务启动中，监听端口: 8081...")
	// 启动服务
	err := r.Run(":8081")
	if err != nil {
		fmt.Println("服务启动失败:", err)
		return
	}
}
