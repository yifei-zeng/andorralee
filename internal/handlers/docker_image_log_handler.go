package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllDockerImageLogs 获取所有Docker镜像操作日志
// @Summary 获取所有Docker镜像操作日志
// @Description 获取所有Docker镜像的操作日志记录
// @Tags Docker镜像日志
// @Produce json
// @Success 200 {object} utils.Response
// @Router /docker/image-logs [get]
func GetAllDockerImageLogs(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageLogRepo(config.MySQLDB)
	logs, err := repo.List()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取镜像操作日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// GetDockerImageLogByID 根据ID获取Docker镜像操作日志
// @Summary 根据ID获取Docker镜像操作日志
// @Description 根据ID获取指定的Docker镜像操作日志记录
// @Tags Docker镜像日志
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /docker/image-logs/{id} [get]
func GetDockerImageLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageLogRepo(config.MySQLDB)
	log, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "镜像操作日志不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, log)
}

// GetDockerImageLogsByImageID 根据镜像ID获取操作日志
// @Summary 根据镜像ID获取操作日志
// @Description 获取指定镜像的所有操作日志记录
// @Tags Docker镜像日志
// @Produce json
// @Param image_id path string true "镜像ID"
// @Success 200 {object} utils.Response
// @Router /docker/image-logs/image/{image_id} [get]
func GetDockerImageLogsByImageID(c *gin.Context) {
	imageID := c.Param("image_id")
	if imageID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "镜像ID不能为空")
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageLogRepo(config.MySQLDB)
	logs, err := repo.GetByImageID(imageID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取镜像操作日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, logs)
}

// DeleteDockerImageLog 删除Docker镜像操作日志
// @Summary 删除Docker镜像操作日志
// @Description 根据ID删除指定的Docker镜像操作日志记录
// @Tags Docker镜像日志
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /docker/image-logs/{id} [delete]
func DeleteDockerImageLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageLogRepo(config.MySQLDB)
	if err := repo.Delete(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除镜像操作日志失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "镜像操作日志删除成功")
}

// GetDockerImages 获取所有Docker镜像记录
// @Summary 获取所有Docker镜像记录
// @Description 从数据库获取所有Docker镜像记录
// @Tags Docker镜像管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /docker/images/db [get]
func GetDockerImages(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageRepo(config.MySQLDB)
	images, err := repo.List()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取镜像记录失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, images)
}

// GetDockerImageByDBID 根据数据库ID获取Docker镜像记录
// @Summary 根据数据库ID获取Docker镜像记录
// @Description 根据数据库ID获取指定的Docker镜像记录
// @Tags Docker镜像管理
// @Produce json
// @Param id path int true "数据库记录ID"
// @Success 200 {object} utils.Response
// @Router /docker/images/db/{id} [get]
func GetDockerImageByDBID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageRepo(config.MySQLDB)
	image, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "镜像记录不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, image)
}

// DeleteDockerImageRecord 删除Docker镜像数据库记录
// @Summary 删除Docker镜像数据库记录
// @Description 根据数据库ID删除指定的Docker镜像记录（不删除实际镜像）
// @Tags Docker镜像管理
// @Produce json
// @Param id path int true "数据库记录ID"
// @Success 200 {object} utils.Response
// @Router /docker/images/db/{id} [delete]
func DeleteDockerImageRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLDockerImageRepo(config.MySQLDB)
	if err := repo.Delete(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除镜像记录失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "镜像记录删除成功")
}
