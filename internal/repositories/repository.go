package repositories

import "time"

// DataModel 数据模型
type DataModel struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Behavior  string    `json:"behavior"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DataRepository 定义数据库操作的统一接口
type DataRepository interface {
	// QueryData 查询数据
	QueryData() ([]DataModel, error)
	
	// CreateData 创建数据
	CreateData(data *DataModel) error
	
	// UpdateData 更新数据
	UpdateData(data *DataModel) error
	
	// DeleteData 删除数据
	DeleteData(id uint) error
	
	// GetDataByID 根据ID获取数据
	GetDataByID(id uint) (*DataModel, error)
	
	// GetDataByName 根据名称获取数据
	GetDataByName(name string) (*DataModel, error)
}
