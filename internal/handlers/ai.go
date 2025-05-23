package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SemanticSegmentRequest 语义分割请求
type SemanticSegmentRequest struct {
	ContainerID string `json:"container_id" binding:"required"` // 容器ID
}

// SemanticSegment 语义分割接口
// @Summary 日志语义分割
// @Description 对Docker容器日志进行语义分割处理并存储到数据库
// @Tags AI
// @Accept json
// @Produce json
// @Param   payload  body   SemanticSegmentRequest  true  "容器信息"
// @Success 200 {object} utils.Response
// @Router /ai/semantic-segment [post]
func SemanticSegment(c *gin.Context) {
	var req SemanticSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	result, err := services.SemanticSegment(req.ContainerID)
	if err != nil {
		utils.ResponseError(c, 500, "处理失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, result)
}
