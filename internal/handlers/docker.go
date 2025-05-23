package handlers

import (
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"github.com/gin-gonic/gin"
)

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
	var req struct {
		Image string `json:"image" binding:"required"`
		Tag   string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	imageName := req.Image
	if req.Tag != "" {
		imageName += ":" + req.Tag
	}

	if err := services.PullDockerImage(imageName); err != nil {
		utils.ResponseError(c, 500, "拉取失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "镜像拉取成功")
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
