package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllContainerLogSegments 获取所有容器日志分析结果
// @Summary 获取所有容器日志分析结果
// @Description 获取所有容器日志语义分析结果
// @Tags 容器日志分析
// @Produce json
// @Success 200 {object} utils.Response
// @Router /container-logs/segments [get]
func GetAllContainerLogSegments(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	segments, err := repo.List()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志分析结果失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, segments)
}

// GetContainerLogSegmentByID 根据ID获取容器日志分析结果
// @Summary 根据ID获取容器日志分析结果
// @Description 根据ID获取指定的容器日志语义分析结果
// @Tags 容器日志分析
// @Produce json
// @Param id path int true "分析结果ID"
// @Success 200 {object} utils.Response
// @Router /container-logs/segments/{id} [get]
func GetContainerLogSegmentByID(c *gin.Context) {
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

	repo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	segment, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, "日志分析结果不存在: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, segment)
}

// GetLogSegmentsByContainerID 根据容器ID获取日志分析结果
// @Summary 根据容器ID获取日志分析结果
// @Description 获取指定容器的所有日志语义分析结果
// @Tags 容器日志分析
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /container-logs/segments/container/{container_id} [get]
func GetLogSegmentsByContainerID(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	segments, err := repo.GetByContainerID(containerID)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志分析结果失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, segments)
}

// GetLogSegmentsByType 根据日志类型获取分析结果
// @Summary 根据日志类型获取分析结果
// @Description 获取指定类型的所有日志语义分析结果
// @Tags 容器日志分析
// @Produce json
// @Param type path string true "日志类型(error/warning/info/debug)"
// @Success 200 {object} utils.Response
// @Router /container-logs/segments/type/{type} [get]
func GetLogSegmentsByType(c *gin.Context) {
	segmentType := c.Param("type")
	if segmentType == "" {
		utils.ResponseError(c, http.StatusBadRequest, "日志类型不能为空")
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	segments, err := repo.GetBySegmentType(segmentType)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取日志分析结果失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, segments)
}

// DeleteContainerLogSegment 删除容器日志分析结果
// @Summary 删除容器日志分析结果
// @Description 根据ID删除指定的容器日志语义分析结果
// @Tags 容器日志分析
// @Produce json
// @Param id path int true "分析结果ID"
// @Success 200 {object} utils.Response
// @Router /container-logs/segments/{id} [delete]
func DeleteContainerLogSegment(c *gin.Context) {
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

	repo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	if err := repo.Delete(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除日志分析结果失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "日志分析结果删除成功")
}

// DeleteLogSegmentsByContainerID 根据容器ID删除所有相关日志分析结果
// @Summary 根据容器ID删除日志分析结果
// @Description 删除指定容器的所有日志语义分析结果
// @Tags 容器日志分析
// @Produce json
// @Param container_id path string true "容器ID"
// @Success 200 {object} utils.Response
// @Router /container-logs/segments/container/{container_id} [delete]
func DeleteLogSegmentsByContainerID(c *gin.Context) {
	containerID := c.Param("container_id")
	if containerID == "" {
		utils.ResponseError(c, http.StatusBadRequest, "容器ID不能为空")
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLContainerLogSegmentRepo(config.MySQLDB)
	if err := repo.DeleteByContainerID(containerID); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除日志分析结果失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "容器相关日志分析结果删除成功")
}
