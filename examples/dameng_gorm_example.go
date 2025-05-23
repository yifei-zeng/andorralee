/*
 * 达梦数据库GORM连接示例代码
 * 该例程使用GORM实现插入数据，修改数据，删除数据，数据查询等基本操作
 * 与项目中现有的达梦数据库连接方式保持一致
 */
package examples

import (
	"fmt"
	"time"

	// 导入达梦数据库GORM驱动
	_ "github.com/godoes/gorm-dameng"
	dm "github.com/godoes/gorm-dameng"
	"gorm.io/gorm"
)

// Product 产品信息模型
type Product struct {
	ID                   uint      `gorm:"column:productid;primaryKey;autoIncrement"`
	Name                 string    `gorm:"column:name;type:varchar(100);not null"`
	Author               string    `gorm:"column:author;type:varchar(50)"`
	Publisher            string    `gorm:"column:publisher;type:varchar(100)"`
	PublishTime          time.Time `gorm:"column:publishtime"`
	ProductSubcategoryID int       `gorm:"column:product_subcategoryid"`
	ProductNo            string    `gorm:"column:productno;type:varchar(50)"`
	SafetyStockLevel     int       `gorm:"column:satetystocklevel"`
	OriginalPrice        float64   `gorm:"column:originalprice;type:decimal(10,4)"`
	NowPrice             float64   `gorm:"column:nowprice;type:decimal(10,4)"`
	Discount             float64   `gorm:"column:discount;type:decimal(10,1)"`
	Description          string    `gorm:"column:description;type:text"`
	Photo                []byte    `gorm:"column:photo;type:blob"`
	Type                 string    `gorm:"column:type;type:varchar(20)"`
	PaperTotal           int       `gorm:"column:papertotal"`
	WordTotal            int       `gorm:"column:wordtotal"`
	SellStartTime        time.Time `gorm:"column:sellstarttime"`
	SellEndTime          time.Time `gorm:"column:sellendtime"`
}

// TableName 设置表名
func (Product) TableName() string {
	return "product"
}

func main() {
	// 连接达梦数据库
	db, err := connectDB()
	if err != nil {
		fmt.Println("连接数据库失败:", err)
		return
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&Product{}); err != nil {
		fmt.Println("表结构迁移失败:", err)
		return
	}
	fmt.Println("表结构迁移成功")

	// 插入数据
	if err := insertData(db); err != nil {
		fmt.Println("插入数据失败:", err)
		return
	}

	// 查询数据
	if err := queryData(db); err != nil {
		fmt.Println("查询数据失败:", err)
		return
	}

	// 更新数据
	if err := updateData(db); err != nil {
		fmt.Println("更新数据失败:", err)
		return
	}

	// 删除数据
	if err := deleteData(db); err != nil {
		fmt.Println("删除数据失败:", err)
		return
	}

	fmt.Println("所有操作完成")
}

// connectDB 连接达梦数据库
func connectDB() (*gorm.DB, error) {
	// 连接字符串格式：dm://用户名:密码@主机地址:端口号/数据库名
	dsn := "dm://SYSDBA:SYSDBA@localhost:5236/DAMENG"

	// 打开数据库连接
	db, err := gorm.Open(dm.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 获取底层SQL DB对象并设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取SQL DB对象失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(10)           // 最大连接数
	sqlDB.SetMaxIdleConns(5)            // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	fmt.Println("连接到达梦数据库成功")
	return db, nil
}

// insertData 插入数据
func insertData(db *gorm.DB) error {
	// 准备日期数据
	t1, _ := time.Parse("2006-01-02", "2005-04-01")
	t2, _ := time.Parse("2006-01-02", "2006-03-20")
	t3, _ := time.Parse("2006-01-02", "1900-01-01")

	// 创建产品对象
	product := Product{
		Name:                 "三国演义",
		Author:               "罗贯中",
		Publisher:            "中华书局",
		PublishTime:          t1,
		ProductSubcategoryID: 4,
		ProductNo:            "9787101046126",
		SafetyStockLevel:     10,
		OriginalPrice:        19.0000,
		NowPrice:             15.2000,
		Discount:             8.0,
		Description:          "《三国演义》是中国第一部长篇章回体小说，中国小说由短篇发展至长篇的原因与说书有关。",
		Photo:                []byte{}, // 这里可以读取实际图片文件
		Type:                 "25",
		PaperTotal:           943,
		WordTotal:            93000,
		SellStartTime:        t2,
		SellEndTime:          t3,
	}

	// 插入数据
	result := db.Create(&product)
	if result.Error != nil {
		return fmt.Errorf("插入数据失败: %w", result.Error)
	}

	fmt.Printf("插入数据成功，ID: %d\n", product.ID)
	return nil
}

// queryData 查询数据
func queryData(db *gorm.DB) error {
	// 查询所有产品
	var products []Product
	result := db.Find(&products)
	if result.Error != nil {
		return fmt.Errorf("查询所有数据失败: %w", result.Error)
	}

	fmt.Printf("共查询到 %d 条记录\n", len(products))

	// 查询单个产品
	var product Product
	result = db.First(&product)
	if result.Error != nil {
		return fmt.Errorf("查询单个数据失败: %w", result.Error)
	}

	fmt.Printf("查询到产品: ID=%d, 名称=%s, 作者=%s\n",
		product.ID, product.Name, product.Author)

	// 条件查询
	var conditionalProducts []Product
	result = db.Where("author = ?", "罗贯中").Find(&conditionalProducts)
	if result.Error != nil {
		return fmt.Errorf("条件查询失败: %w", result.Error)
	}

	fmt.Printf("条件查询结果: %d 条记录\n", len(conditionalProducts))

	return nil
}

// updateData 更新数据
func updateData(db *gorm.DB) error {
	// 先查询要更新的记录
	var product Product
	result := db.First(&product)
	if result.Error != nil {
		return fmt.Errorf("查询要更新的记录失败: %w", result.Error)
	}

	// 更新字段
	product.Name = "三国演义（上）"
	product.NowPrice = 16.8000

	// 保存更新
	result = db.Save(&product)
	if result.Error != nil {
		return fmt.Errorf("更新数据失败: %w", result.Error)
	}

	fmt.Printf("更新数据成功，ID: %d, 新名称: %s\n", product.ID, product.Name)

	// 另一种更新方式：直接更新特定字段
	result = db.Model(&Product{}).Where("productid = ?", product.ID).Update("discount", 7.5)
	if result.Error != nil {
		return fmt.Errorf("更新折扣失败: %w", result.Error)
	}

	fmt.Printf("更新折扣成功，影响行数: %d\n", result.RowsAffected)

	return nil
}

// deleteData 删除数据
func deleteData(db *gorm.DB) error {
	// 删除指定ID的记录
	result := db.Delete(&Product{}, 1) // 删除ID为1的记录
	if result.Error != nil {
		return fmt.Errorf("删除数据失败: %w", result.Error)
	}

	fmt.Printf("删除数据成功，影响行数: %d\n", result.RowsAffected)

	// 条件删除
	result = db.Where("originalprice < ?", 10.0).Delete(&Product{})
	if result.Error != nil {
		return fmt.Errorf("条件删除失败: %w", result.Error)
	}

	fmt.Printf("条件删除成功，影响行数: %d\n", result.RowsAffected)

	return nil
}
