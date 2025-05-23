package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"errors"
)

// DatabaseService 数据库服务
type DatabaseService struct {
	repo repositories.DataRepository
}

// NewDatabaseService 创建数据库服务实例
func NewDatabaseService(dbType string) (*DatabaseService, error) {
	var repo repositories.DataRepository

	switch dbType {
	case "mysql":
		repo = repositories.NewMySQLRepository(config.MySQLDB)
	case "dameng":
		repo = repositories.NewDamengRepository(config.DamengDB)
	default:
		return nil, errors.New("不支持的数据库类型")
	}

	return &DatabaseService{repo: repo}, nil
}

// QueryData 查询数据
func (s *DatabaseService) QueryData() ([]repositories.DataModel, error) {
	return s.repo.QueryData()
}

// CreateData 创建数据
func (s *DatabaseService) CreateData(data *repositories.DataModel) error {
	return s.repo.CreateData(data)
}

// UpdateData 更新数据
func (s *DatabaseService) UpdateData(data *repositories.DataModel) error {
	return s.repo.UpdateData(data)
}

// DeleteData 删除数据
func (s *DatabaseService) DeleteData(id uint) error {
	return s.repo.DeleteData(id)
}

// GetDataByID 根据ID获取数据
func (s *DatabaseService) GetDataByID(id uint) (*repositories.DataModel, error) {
	return s.repo.GetDataByID(id)
}

// GetDataByName 根据名称获取数据
func (s *DatabaseService) GetDataByName(name string) (*repositories.DataModel, error) {
	return s.repo.GetDataByName(name)
}
