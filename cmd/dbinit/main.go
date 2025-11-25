package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		user   = flag.String("user", getenv("DB_USER", "root"), "MySQL 用户名")
		pass   = flag.String("password", getenv("DB_PASSWORD", "123456"), "MySQL 密码")
		host   = flag.String("host", getenv("DB_HOST", "127.0.0.1"), "MySQL 主机")
		port   = flag.String("port", getenv("DB_PORT", "3306"), "MySQL 端口")
		script = flag.String("script", filepath.Join("deployment", "bootstrap_schema.sql"), "SQL 脚本路径")
	)
	flag.Parse()

	payload, err := os.ReadFile(*script)
	if err != nil {
		log.Fatalf("读取脚本失败: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true", *user, *pass, *host, *port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库无法访问: %v", err)
	}

	scriptText := normalizeNewlines(string(payload))
	if _, err := db.Exec(scriptText); err != nil {
		log.Fatalf("执行脚本失败: %v", err)
	}

	log.Println("✅ 数据库初始化完成")
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func getenv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
