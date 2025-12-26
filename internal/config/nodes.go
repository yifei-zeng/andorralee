package config

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

// NodeConfig 表示一个远程节点的基本信息
type NodeConfig struct {
	ID      string `json:"id" gorm:"-"`
	Name    string `json:"name" gorm:"-"`
	BaseURL string `json:"base_url" gorm:"-"`
	Token   string `json:"token" gorm:"-"`
}

// NodeInfo 返回给前端的精简信息（不包含 token）
type NodeInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

var (
	nodes = []NodeConfig{
		{ID: "node-1", Name: "edge-1", BaseURL: "http://127.0.0.1:8081", Token: ""},
		{ID: "node-2", Name: "edge-2", BaseURL: "http://127.0.0.1:8082", Token: ""},
	}
	nodesMu sync.RWMutex
)

func init() {
	loadNodesFromEnv()
}

func loadNodesFromEnv() {
	raw := strings.TrimSpace(os.Getenv("NODES_JSON"))
	if raw == "" {
		return
	}
	var parsed []NodeConfig
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return
	}
	if len(parsed) == 0 {
		return
	}
	nodesMu.Lock()
	nodes = parsed
	nodesMu.Unlock()
}

// GetNodes 返回当前静态节点列表（不含 token）
func GetNodes() []NodeInfo {
	nodesMu.RLock()
	defer nodesMu.RUnlock()
	out := make([]NodeInfo, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, NodeInfo{ID: n.ID, Name: n.Name, BaseURL: n.BaseURL})
	}
	return out
}

// FindNodeByID 查找节点
func FindNodeByID(id string) *NodeConfig {
	nodesMu.RLock()
	defer nodesMu.RUnlock()
	for _, n := range nodes {
		if n.ID == id {
			nn := n
			return &nn
		}
	}
	return nil
}
