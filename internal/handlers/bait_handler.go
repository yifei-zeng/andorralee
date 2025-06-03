package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BaitHandler 诱饵处理器
type BaitHandler struct {
	baitDir string
}

// NewBaitHandler 创建诱饵处理器
func NewBaitHandler(baitDir string) *BaitHandler {
	return &BaitHandler{
		baitDir: baitDir,
	}
}

// CreateBait 创建诱饵
func (h *BaitHandler) CreateBait(c *gin.Context) {
	var bait repositories.Bait
	if err := c.ShouldBindJSON(&bait); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	if err := repo.Create(&bait); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, bait)
}

// GetBait 获取诱饵
func (h *BaitHandler) GetBait(c *gin.Context) {
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

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	bait, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, bait)
}

// ListBaits 列出所有诱饵
func (h *BaitHandler) ListBaits(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	baits, err := repo.List()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, baits)
}

// DeleteBait 删除诱饵
func (h *BaitHandler) DeleteBait(c *gin.Context) {
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

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	if err := repo.Delete(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "删除成功")
}

// MonitorBait 监控诱饵
func (h *BaitHandler) MonitorBait(c *gin.Context) {
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

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	bait, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取诱饵失败: "+err.Error())
		return
	}

	// 这里应该实现监控诱饵的逻辑
	// 例如，记录访问日志，触发告警等

	utils.ResponseSuccess(c, map[string]interface{}{
		"message": "开始监控诱饵",
		"bait":    bait,
	})
}

// GetAllBaits 获取所有诱饵
// @Summary 获取所有诱饵
// @Description 获取所有诱饵信息
// @Tags 诱饵管理
// @Produce json
// @Success 200 {object} utils.Response
// @Router /baits [get]
func GetAllBaits(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	baits, err := repo.List()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, baits)
}

// GetBaitByID 根据ID获取诱饵
// @Summary 根据ID获取诱饵
// @Description 根据ID获取诱饵详情
// @Tags 诱饵管理
// @Produce json
// @Param id path int true "诱饵ID"
// @Success 200 {object} utils.Response
// @Router /baits/{id} [get]
func GetBaitByID(c *gin.Context) {
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

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	bait, err := repo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, bait)
}

// CreateBait 创建诱饵
// @Summary 创建诱饵
// @Description 创建新的诱饵
// @Tags 诱饵管理
// @Accept json
// @Produce json
// @Param bait body repositories.Bait true "诱饵信息"
// @Success 200 {object} utils.Response
// @Router /baits [post]
func CreateBait(c *gin.Context) {
	var bait repositories.Bait
	if err := c.ShouldBindJSON(&bait); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	if err := repo.Create(&bait); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "创建诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, bait)
}

// UpdateBait 更新诱饵
// @Summary 更新诱饵
// @Description 更新现有的诱饵
// @Tags 诱饵管理
// @Accept json
// @Produce json
// @Param id path int true "诱饵ID"
// @Param bait body repositories.Bait true "诱饵信息"
// @Success 200 {object} utils.Response
// @Router /baits/{id} [put]
func UpdateBait(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的ID: "+err.Error())
		return
	}

	var bait repositories.Bait
	if err := c.ShouldBindJSON(&bait); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的请求参数: "+err.Error())
		return
	}

	bait.ID = uint(id)

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	if err := repo.Update(&bait); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "更新诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, bait)
}

// DeleteBait 删除诱饵
// @Summary 删除诱饵
// @Description 删除指定的诱饵
// @Tags 诱饵管理
// @Produce json
// @Param id path int true "诱饵ID"
// @Success 200 {object} utils.Response
// @Router /baits/{id} [delete]
func DeleteBait(c *gin.Context) {
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

	repo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	if err := repo.Delete(uint(id)); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "删除诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, "删除成功")
}

// DeployBait 部署诱饵
// @Summary 部署诱饵
// @Description 部署指定的诱饵到蜜罐实例
// @Tags 诱饵管理
// @Accept json
// @Produce json
// @Param id path int true "诱饵ID"
// @Param instance_id query int true "蜜罐实例ID"
// @Success 200 {object} utils.Response
// @Router /baits/{id}/deploy [post]
func DeployBait(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的诱饵ID: "+err.Error())
		return
	}

	instanceIDStr := c.Query("instance_id")
	instanceID, err := strconv.ParseUint(instanceIDStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "无效的实例ID: "+err.Error())
		return
	}

	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "MySQL数据库未初始化")
		return
	}

	// 获取诱饵信息
	baitRepo := repositories.NewMySQLBaitRepo(config.MySQLDB)
	bait, err := baitRepo.GetByID(uint(id))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取诱饵失败: "+err.Error())
		return
	}

	// 获取实例信息
	instanceRepo := repositories.NewMySQLHoneypotInstanceRepo(config.MySQLDB)
	instance, err := instanceRepo.GetByID(uint(instanceID))
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "获取实例失败: "+err.Error())
		return
	}

	// 更新诱饵信息
	bait.IsDeployed = true
	bait.InstanceID = instance.ID
	if err := baitRepo.Update(bait); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, "部署诱饵失败: "+err.Error())
		return
	}

	utils.ResponseSuccess(c, bait)
}
