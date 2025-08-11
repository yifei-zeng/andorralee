package integration_test

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"os"
	"testing"
)

// TestDBConnection 基础连接 + CRUD 测试，依据 DB_MODE 运行
func TestDBConnection(t *testing.T) {
	mode := os.Getenv("DB_MODE")
	if mode == "" { // 默认 mysql
		mode = "mysql"
	}

	// 初始化对应数据库
	if mode == "dameng" {
		if err := config.InitDameng(); err != nil {
			t.Fatalf("InitDameng failed: %v", err)
		}
	} else {
		if err := config.InitMySQL(); err != nil {
			t.Fatalf("InitMySQL failed: %v", err)
		}
	}

	// 取得活动 DB
	db := config.GetActiveDB()
	if db == nil {
		t.Fatalf("active DB is nil for mode %s", mode)
	}

	// 迁移一个最小表 (使用 MalwareSignature)
	if err := db.AutoMigrate(&repositories.MalwareSignature{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// 插入
	rec := &repositories.MalwareSignature{Name: "TestSig", Pattern: "abc123", Type: "hash", Severity: "low", Description: "test", IsActive: true}
	if err := db.Create(rec).Error; err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if rec.ID == 0 {
		t.Fatalf("expected auto ID > 0, got 0")
	}

	// 查询
	var out repositories.MalwareSignature
	if err := db.First(&out, rec.ID).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if out.Name != rec.Name {
		t.Fatalf("expected name %s got %s", rec.Name, out.Name)
	}
}
