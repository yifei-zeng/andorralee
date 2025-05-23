package repositories

import (
	"gorm.io/gorm"
)

// MySQLRepository 实现 DataRepository 接口
type MySQLRepository struct {
	db *gorm.DB
}

// NewMySQLRepository 创建MySQL仓库实例
func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	// 自动迁移表结构
	err := db.AutoMigrate(&DataModel{})
	if err != nil {
		panic("MySQL表结构迁移失败: " + err.Error())
	}
	return &MySQLRepository{db: db}
}

// QueryData 查询所有数据
func (r *MySQLRepository) QueryData() ([]DataModel, error) {
	var results []DataModel
	err := r.db.Find(&results).Error
	return results, err
}

// CreateData 创建数据
func (r *MySQLRepository) CreateData(data *DataModel) error {
	return r.db.Create(data).Error
}

// UpdateData 更新数据
func (r *MySQLRepository) UpdateData(data *DataModel) error {
	return r.db.Save(data).Error
}

// DeleteData 删除数据
func (r *MySQLRepository) DeleteData(id uint) error {
	return r.db.Delete(&DataModel{}, id).Error
}

// GetDataByID 根据ID获取数据
func (r *MySQLRepository) GetDataByID(id uint) (*DataModel, error) {
	var data DataModel
	err := r.db.First(&data, id).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetDataByName 根据名称获取数据
func (r *MySQLRepository) GetDataByName(name string) (*DataModel, error) {
	var data DataModel
	err := r.db.Where("name = ?", name).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}
