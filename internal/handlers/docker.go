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

// TagImageRequest 标记镜像请求结构
type TagImageRequest struct {
	NewRepo string `json:"new_repo" binding:"required"` // 新的仓库名称
	NewTag  string `json:"new_tag" binding:"required"`  // 新的标签
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

// GetImageByID 获取指定ID的Docker镜像详情
// @Summary 获取镜像详情
// @Description 根据ID获取Docker镜像详细信息
// @Tags Docker
// @Produce json
// @Param id path string true "镜像ID"
// @Success 200 {object} utils.Response
// @Router /docker/images/{id} [get]
func GetImageByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ResponseError(c, 400, "无效的镜像ID")
		return
	}

	image, err := services.GetDockerImageByID(id)
	if err != nil {
		utils.ResponseError(c, 404, "镜像不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, image)
}

// DeleteImage 删除指定ID的Docker镜像
// @Summary 删除镜像
// @Description 根据ID删除Docker镜像
// @Tags Docker
// @Produce json
// @Param id path string true "镜像ID"
// @Success 200 {object} utils.Response
// @Router /docker/images/{id} [delete]
func DeleteImage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ResponseError(c, 400, "无效的镜像ID")
		return
	}

	if err := services.DeleteDockerImage(id); err != nil {
		utils.ResponseError(c, 500, "删除镜像失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "镜像删除成功")
}

// TagImage 为指定ID的Docker镜像添加新标签
// @Summary 标记镜像
// @Description 为指定ID的Docker镜像添加新标签
// @Tags Docker
// @Accept json
// @Produce json
// @Param id path string true "镜像ID"
// @Param payload body TagImageRequest true "新标签信息"
// @Success 200 {object} utils.Response
// @Router /docker/images/{id}/tag [post]
func TagImage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ResponseError(c, 400, "无效的镜像ID")
		return
	}

	var req TagImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := services.TagDockerImage(id, req.NewRepo, req.NewTag); err != nil {
		utils.ResponseError(c, 500, "标记镜像失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "镜像标记成功")
}
