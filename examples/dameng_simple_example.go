/*
 * 达梦数据库简单连接示例
 * 该示例展示如何正确导入和使用达梦数据库驱动
 */
package examples

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	// 导入达梦数据库驱动
	_ "dm"
)

// RunDamengSimpleExample 运行达梦数据库简单连接示例
func RunDamengSimpleExample() {
	// 设置驱动名称和连接字符串
	driverName := "dm"
	// 连接字符串格式：dm://用户名:密码@主机地址:端口号
	dataSourceName := "dm://SYSDBA:SYSDBA@localhost:5236"

	// 打开数据库连接
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatalf("打开数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试连接是否成功
	if err = db.Ping(); err != nil {
		log.Fatalf("测试数据库连接失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)           // 最大连接数
	db.SetMaxIdleConns(5)            // 最大空闲连接数
	db.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	fmt.Printf("连接到 \"%s\" 成功\n", dataSourceName)

	// 执行简单查询
	rows, err := db.Query("SELECT 1 AS test")
	if err != nil {
		log.Fatalf("执行查询失败: %v", err)
	}
	defer rows.Close()

	// 打印查询结果
	fmt.Println("查询结果:")
	for rows.Next() {
		var test int
		if err = rows.Scan(&test); err != nil {
			log.Fatalf("扫描结果失败: %v", err)
		}
		fmt.Printf("测试值: %d\n", test)
	}

	// 检查遍历过程中是否有错误
	if err = rows.Err(); err != nil {
		log.Fatalf("遍历结果失败: %v", err)
	}

	fmt.Println("数据库连接测试完成")
}
