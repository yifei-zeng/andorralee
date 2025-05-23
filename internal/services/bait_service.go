package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BaitType 蜜签类型
type BaitType string

const (
	BaitTypeFile    BaitType = "file"
	BaitTypeAPI     BaitType = "api"
	BaitTypeService BaitType = "service"
)

// BaitConfig 蜜签配置
type BaitConfig struct {
	Type        BaitType `json:"type"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	CreatedAt   string   `json:"created_at"`
}

// BaitService 蜜签服务
type BaitService struct {
	basePath string
}

// NewBaitService 创建蜜签服务实例
func NewBaitService(basePath string) *BaitService {
	return &BaitService{
		basePath: basePath,
	}
}

// CreateBait 创建蜜签
func (s *BaitService) CreateBait(config BaitConfig) error {
	// 生成唯一ID
	id := generateUniqueID()

	// 创建蜜签目录
	baitPath := filepath.Join(s.basePath, id)
	if err := os.MkdirAll(baitPath, 0755); err != nil {
		return fmt.Errorf("failed to create bait directory: %v", err)
	}

	// 创建蜜签文件
	filePath := filepath.Join(baitPath, config.Name)
	if err := os.WriteFile(filePath, []byte(config.Content), 0644); err != nil {
		return fmt.Errorf("failed to create bait file: %v", err)
	}

	// 创建元数据文件
	metadata := map[string]interface{}{
		"id":          id,
		"type":        config.Type,
		"name":        config.Name,
		"path":        config.Path,
		"description": config.Description,
		"created_at":  time.Now().Format(time.RFC3339),
	}

	metadataPath := filepath.Join(baitPath, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(fmt.Sprintf("%+v", metadata)), 0644); err != nil {
		return fmt.Errorf("failed to create metadata file: %v", err)
	}

	return nil
}

// GetBait 获取蜜签信息
func (s *BaitService) GetBait(id string) (*BaitConfig, error) {
	baitPath := filepath.Join(s.basePath, id)
	metadataPath := filepath.Join(baitPath, "metadata.json")

	// 读取元数据
	metadata, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %v", err)
	}

	// 解析元数据
	var config BaitConfig
	if err := json.Unmarshal(metadata, &config); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	return &config, nil
}

// ListBaits 列出所有蜜签
func (s *BaitService) ListBaits() ([]BaitConfig, error) {
	var baits []BaitConfig

	// 遍历蜜签目录
	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否是蜜签目录
		if info.IsDir() && path != s.basePath {
			metadataPath := filepath.Join(path, "metadata.json")
			if _, err := os.Stat(metadataPath); err == nil {
				// 读取元数据
				metadata, err := os.ReadFile(metadataPath)
				if err != nil {
					return err
				}

				// 解析元数据
				var config BaitConfig
				if err := json.Unmarshal(metadata, &config); err != nil {
					return err
				}

				baits = append(baits, config)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list baits: %v", err)
	}

	return baits, nil
}

// DeleteBait 删除蜜签
func (s *BaitService) DeleteBait(id string) error {
	baitPath := filepath.Join(s.basePath, id)
	return os.RemoveAll(baitPath)
}

// MonitorBait 监控蜜签访问
func (s *BaitService) MonitorBait(id string) error {
	baitPath := filepath.Join(s.basePath, id)
	filePath := filepath.Join(baitPath, "access.log")

	// 创建访问日志文件
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create access log: %v", err)
	}
	defer file.Close()

	// 记录访问信息
	accessInfo := fmt.Sprintf("[%s] Bait accessed\n", time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(accessInfo); err != nil {
		return fmt.Errorf("failed to write access log: %v", err)
	}

	return nil
}

// 辅助函数

// generateUniqueID 生成唯一ID
func generateUniqueID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
