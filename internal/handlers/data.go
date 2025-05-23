package handlers

import (
	"andorralee/internal/repositories"
	"andorralee/internal/services"
	"andorralee/pkg/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

// QueryData 查询数据
// @Summary 查询数据
// @Description 从 MySQL 或达梦数据库查询数据
// @Tags 数据库
// @Produce json
// @Param   db_type  query  string  true  "数据库类型 (mysql 或 dameng)"
// @Success 200 {object} utils.Response
// @Router /data [get]
func QueryData(c *gin.Context) {
	dbType := c.Query("db_type")
	dbService, err := services.NewDatabaseService(dbType)
	if err != nil {
		utils.ResponseError(c, 400, "数据库类型错误: "+err.Error())
		return
	}

	data, err := dbService.QueryData()
	if err != nil {
		utils.ResponseError(c, 500, "查询失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, data)
}

// CreateData 创建数据
// @Summary 创建数据
// @Description 在指定数据库中创建新数据
// @Tags 数据库
// @Accept json
// @Produce json
// @Param   db_type  query  string  true  "数据库类型 (mysql 或 dameng)"
// @Param   data  body  repositories.DataModel  true  "数据信息"
// @Success 200 {object} utils.Response
// @Router /data [post]
func CreateData(c *gin.Context) {
	dbType := c.Query("db_type")
	dbService, err := services.NewDatabaseService(dbType)
	if err != nil {
		utils.ResponseError(c, 400, "数据库类型错误: "+err.Error())
		return
	}

	var data repositories.DataModel
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := dbService.CreateData(&data); err != nil {
		utils.ResponseError(c, 500, "创建失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, data)
}

// UpdateData 更新数据
// @Summary 更新数据
// @Description 更新指定数据库中的数据
// @Tags 数据库
// @Accept json
// @Produce json
// @Param   db_type  query  string  true  "数据库类型 (mysql 或 dameng)"
// @Param   data  body  repositories.DataModel  true  "数据信息"
// @Success 200 {object} utils.Response
// @Router /data [put]
func UpdateData(c *gin.Context) {
	dbType := c.Query("db_type")
	dbService, err := services.NewDatabaseService(dbType)
	if err != nil {
		utils.ResponseError(c, 400, "数据库类型错误: "+err.Error())
		return
	}

	var data repositories.DataModel
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ResponseError(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := dbService.UpdateData(&data); err != nil {
		utils.ResponseError(c, 500, "更新失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, data)
}

// DeleteData 删除数据
// @Summary 删除数据
// @Description 从指定数据库中删除数据
// @Tags 数据库
// @Produce json
// @Param   db_type  query  string  true  "数据库类型 (mysql 或 dameng)"
// @Param   id  query  string  true  "数据ID"
// @Success 200 {object} utils.Response
// @Router /data [delete]
func DeleteData(c *gin.Context) {
	dbType := c.Query("db_type")
	dbService, err := services.NewDatabaseService(dbType)
	if err != nil {
		utils.ResponseError(c, 400, "数据库类型错误: "+err.Error())
		return
	}

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, 400, "ID格式错误: "+err.Error())
		return
	}

	if err := dbService.DeleteData(uint(id)); err != nil {
		utils.ResponseError(c, 500, "删除失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, "删除成功")
}

// GetDataByID 根据ID获取数据
// @Summary 获取数据
// @Description 根据ID从指定数据库获取数据
// @Tags 数据库
// @Produce json
// @Param   db_type  query  string  true  "数据库类型 (mysql 或 dameng)"
// @Param   id  query  string  true  "数据ID"
// @Success 200 {object} utils.Response
// @Router /data/id [get]
func GetDataByID(c *gin.Context) {
	dbType := c.Query("db_type")
	dbService, err := services.NewDatabaseService(dbType)
	if err != nil {
		utils.ResponseError(c, 400, "数据库类型错误: "+err.Error())
		return
	}

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, 400, "ID格式错误: "+err.Error())
		return
	}

	data, err := dbService.GetDataByID(uint(id))
	if err != nil {
		utils.ResponseError(c, 500, "获取数据失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, data)
}

// GetDataByName 根据名称获取数据
// @Summary 根据名称获取数据
// @Description 根据名称从指定数据库获取数据
// @Tags 数据库
// @Produce json
// @Param   db_type  query  string  true  "数据库类型 (mysql 或 dameng)"
// @Param   name  query  string  true  "数据名称"
// @Success 200 {object} utils.Response
// @Router /data/name [get]
func GetDataByName(c *gin.Context) {
	dbType := c.Query("db_type")
	dbService, err := services.NewDatabaseService(dbType)
	if err != nil {
		utils.ResponseError(c, 400, "数据库类型错误: "+err.Error())
		return
	}

	name := c.Query("name")
	if name == "" {
		utils.ResponseError(c, 400, "名称不能为空")
		return
	}

	data, err := dbService.GetDataByName(name)
	if err != nil {
		utils.ResponseError(c, 500, "获取数据失败: "+err.Error())
		return
	}
	utils.ResponseSuccess(c, data)
}
