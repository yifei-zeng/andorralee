package handlers

import (
	"andorralee/internal/config"
	"andorralee/pkg/utils"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	proxyTimeout     = 10 * time.Second
	maxResponseBytes = 10 * 1024 * 1024 // 10MB
)

// NodesListHandler 返回静态配置的节点列表
func NodesListHandler(c *gin.Context) {
	nodes := config.GetNodes()
	utils.ResponseSuccess(c, nodes)
}

// ProxyToNodeHandler 将请求转发到指定节点
func ProxyToNodeHandler(c *gin.Context) {
	nodeID := c.Param("node")
	if nodeID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "节点ID不能为空")
		return
	}
	path := c.Param("path")
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	node := config.FindNodeByID(nodeID)
	if node == nil {
		utils.ResponseError(c, http.StatusNotFound, "未找到指定节点")
		return
	}

	// 构造目标 URL
	u, err := url.Parse(node.BaseURL)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "节点 base_url 配置错误")
		return
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/" + path
	q := u.Query()
	for k, vals := range c.Request.URL.Query() {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(c.Request.Context(), proxyTimeout)
	defer cancel()

	// 准备转发请求
	req, err := http.NewRequestWithContext(ctx, c.Request.Method, u.String(), c.Request.Body)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "构建转发请求失败: "+err.Error())
		return
	}

	// 复制 headers
	for k, vals := range c.Request.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	// 自动注入节点 token（优先使用配置的 token，不覆盖已有 Authorization）
	if token := strings.TrimSpace(node.Token); token != "" {
		req.Header.Set("X-Node-Token", token)
		if req.Header.Get("Authorization") == "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	client := &http.Client{Timeout: proxyTimeout}
	resp, err := client.Do(req)
	if err != nil {
		utils.ResponseError(c, http.StatusBadGateway, "转发请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 复制状态码与 headers
	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Status(resp.StatusCode)

	// 限制响应体大小，避免过大响应拖垮 central
	_, copyErr := io.Copy(c.Writer, io.LimitReader(resp.Body, maxResponseBytes))
	if copyErr != nil {
		return
	}
}
