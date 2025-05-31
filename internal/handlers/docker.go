package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"

	"github.com/gin-gonic/gin"
)

// PullImageRequest 拉取镜像请求结构
type PullImageRequest struct {
	ImageName string `json:"image_name" binding:"required"` // 镜像名称，例如 andorralee/cowrie
	Tag       string `json:"tag"`                           // 可选的标签，默认为latest
}

// PullImage 拉取镜像
// @Summary 拉取 Docker 镜像
// @Description 根据镜像名称和标签拉取镜像
// @Tags Docker
// @Accept json
// @Produce json
// @Param   payload  body   PullImageRequest  true  "镜像信息"
// @Success 200 {object} utils.Response
// @Router /docker/pull [post]
func PullImage(c *gin.Context) {
	var req PullImageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	imageName := req.ImageName
	if req.Tag != "" {
		imageName += ":" + req.Tag
	}

	// 检查Docker客户端是否初始化
	if services.IsDockerAvailable() {
		if err := services.PullDockerImage(imageName); err != nil {
			utils.ResponseError(c, 500, "拉取镜像失败: "+err.Error())
			return
		}
		utils.ResponseSuccess(c, "镜像 "+imageName+" 拉取成功")
	} else {
		utils.ResponseError(c, 500, "Docker服务不可用")
	}
}

// ListImages 获取本地镜像列表
// @Summary 获取本地镜像
// @Description 返回所有本地 Docker 镜像
// @Tags Docker
// @Produce json
// @Success 200 {object} utils.Response
// @Router /docker/images [get]
func ListImages(c *gin.Context) {
	images, err := services.ListDockerImages()
	if err != nil {
		utils.ResponseError(c, 500, "获取镜像失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, images)
}
