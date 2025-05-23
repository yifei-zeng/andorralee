package repositories

import (
	"gorm.io/gorm"
)

// DamengRepository 实现 DataRepository 接口
type DamengRepository struct {
	db *gorm.DB
}

// NewDamengRepository 创建达梦数据库仓库实例
func NewDamengRepository(db *gorm.DB) *DamengRepository {
	// 自动迁移表结构
	err := db.AutoMigrate(&DataModel{})
	if err != nil {
		panic("达梦数据库表结构迁移失败: " + err.Error())
	}
	return &DamengRepository{db: db}
}

// QueryData 查询所有数据
func (r *DamengRepository) QueryData() ([]DataModel, error) {
	var results []DataModel
	err := r.db.Find(&results).Error
	return results, err
}

// CreateData 创建数据
func (r *DamengRepository) CreateData(data *DataModel) error {
	return r.db.Create(data).Error
}

// UpdateData 更新数据
func (r *DamengRepository) UpdateData(data *DataModel) error {
	return r.db.Save(data).Error
}

// DeleteData 删除数据
func (r *DamengRepository) DeleteData(id uint) error {
	return r.db.Delete(&DataModel{}, id).Error
}

// GetDataByID 根据ID获取数据
func (r *DamengRepository) GetDataByID(id uint) (*DataModel, error) {
	var data DataModel
	err := r.db.First(&data, id).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetDataByName 根据名称获取数据
func (r *DamengRepository) GetDataByName(name string) (*DataModel, error) {
	var data DataModel
	err := r.db.Where("name = ?", name).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}
