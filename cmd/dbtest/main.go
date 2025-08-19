package main

import (
	"andorralee/internal/config"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Minimal MySQL connectivity tester using environment variables.
// Env: MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE
func main() {
	cfg := config.LoadConfig()

	host := cfg.MySQL.Host
	port := cfg.MySQL.Port
	user := cfg.MySQL.User
	pass := cfg.MySQL.Password
	dbname := cfg.MySQL.Database

	// Allow override via single DSN env if provided
	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		tryNative(dsn)
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s",
		user, pass, host, port, dbname,
	)
	tryNative(dsn)
}

func tryNative(dsn string) {
	fmt.Println("[dbtest] Trying DSN:")
	fmt.Println("  ", sanitizeDSN(dsn))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("[dbtest] sql.Open error:", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	// Ping with timeout
	pingCh := make(chan error, 1)
	go func() { pingCh <- db.Ping() }()

	select {
	case err := <-pingCh:
		if err != nil {
			fmt.Println("[dbtest] Ping failed:", err)
			os.Exit(2)
		}
		fmt.Println("[dbtest] ✅ MySQL connection OK")
	case <-time.After(8 * time.Second):
		fmt.Println("[dbtest] ❌ Ping timeout (8s)")
		os.Exit(3)
	}
}

// sanitizeDSN masks password for display
func sanitizeDSN(dsn string) string {
	// Very simple masking: user:*****@tcp(host:port)/db?...
	// Find '://'? Not present in mysql DSN; mask between first ':' after user and '@'
	at := -1
	colon := -1
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == ':' && colon == -1 {
			colon = i
		}
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if colon != -1 && at != -1 && at > colon+1 {
		return dsn[:colon+1] + "*****" + dsn[at:]
	}
	return dsn
}
