/*
 * 达梦数据库连接示例代码
 * 该例程实现插入数据，修改数据，删除数据，数据查询等基本操作
 */
package examples

// 引入相关包
import (
	"database/sql"
	"fmt"
	"os"
	"time"

	// 导入达梦数据库驱动
	_ "dm"
)

var db *sql.DB
var err error

// RunDamengExample 运行达梦数据库原生驱动示例
func RunDamengExample() {
	// 设置驱动名称和连接字符串
	driverName := "dm"
	// 连接字符串格式：dm://用户名:密码@主机地址:端口号
	dataSourceName := "dm://SYSDBA:SYSDBA@localhost:5236"

	// 连接数据库
	if db, err = connect(driverName, dataSourceName); err != nil {
		fmt.Println("连接数据库失败:", err)
		return
	}

	// 执行数据操作
	if err = createTable(); err != nil {
		fmt.Println("创建表失败:", err)
		return
	}

	if err = insertTable(); err != nil {
		fmt.Println("插入数据失败:", err)
		return
	}

	if err = updateTable(); err != nil {
		fmt.Println("更新数据失败:", err)
		return
	}

	if err = queryTable(); err != nil {
		fmt.Println("查询数据失败:", err)
		return
	}

	if err = deleteTable(); err != nil {
		fmt.Println("删除数据失败:", err)
		return
	}

	// 关闭数据库连接
	if err = disconnect(); err != nil {
		fmt.Println("关闭数据库连接失败:", err)
		return
	}
}

/* 创建数据库连接 */
func connect(driverName string, dataSourceName string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// 打开数据库连接
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试连接是否成功
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("测试数据库连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)           // 最大连接数
	db.SetMaxIdleConns(5)            // 最大空闲连接数
	db.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	fmt.Printf("连接到 \"%s\" 成功\n", dataSourceName)
	return db, nil
}

/* 创建产品信息表 */
func createTable() error {
	// 创建产品信息表的SQL语句
	sql := `CREATE TABLE IF NOT EXISTS product (
		productid INT IDENTITY PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		author VARCHAR(50),
		publisher VARCHAR(100),
		publishtime DATE,
		product_subcategoryid INT,
		productno VARCHAR(50),
		satetystocklevel INT,
		originalprice DECIMAL(10,4),
		nowprice DECIMAL(10,4),
		discount DECIMAL(10,1),
		description CLOB,
		photo BLOB,
		type VARCHAR(20),
		papertotal INT,
		wordtotal INT,
		sellstarttime DATE,
		sellendtime DATE
	)`

	// 执行SQL语句
	_, err := db.Exec(sql)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	fmt.Println("创建表成功")
	return nil
}

/* 往产品信息表插入数据 */
func insertTable() error {
	// 准备图片数据
	var inFileName = "./examples/example.jpg" // 请确保此文件存在或修改为实际路径
	data, err := os.ReadFile(inFileName)
	if err != nil {
		// 如果文件不存在，使用空数据
		fmt.Printf("警告: 图片文件不存在(%s): %v，将使用空数据\n", inFileName, err)
		data = []byte{}
	}

	// 插入数据的SQL语句
	sql := `INSERT INTO product(name, author, publisher, publishtime, 
		product_subcategoryid, productno, satetystocklevel, originalprice, nowprice, discount, 
		description, photo, type, papertotal, wordtotal, sellstarttime, sellendtime) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 准备日期数据
	t1, err := time.Parse("2006-01-02", "2005-04-01")
	if err != nil {
		return fmt.Errorf("解析日期失败: %w", err)
	}
	t2, err := time.Parse("2006-01-02", "2006-03-20")
	if err != nil {
		return fmt.Errorf("解析日期失败: %w", err)
	}
	t3, err := time.Parse("2006-01-02", "1900-01-01")
	if err != nil {
		return fmt.Errorf("解析日期失败: %w", err)
	}

	// 执行SQL语句
	_, err = db.Exec(sql,
		"三国演义", "罗贯中", "中华书局", t1, 4, "9787101046126", 10, 19.0000, 15.2000, 8.0,
		"《三国演义》是中国第一部长篇章回体小说，中国小说由短篇发展至长篇的原因与说书有关。",
		data, "25", 943, 93000, t2, t3)

	if err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	fmt.Println("插入数据成功")
	return nil
}

/* 修改产品信息表数据 */
func updateTable() error {
	// 更新数据的SQL语句
	sql := "UPDATE product SET name = ? WHERE productid = 1"

	// 执行SQL语句
	_, err := db.Exec(sql, "三国演义（上）")
	if err != nil {
		return fmt.Errorf("更新数据失败: %w", err)
	}

	fmt.Println("更新数据成功")
	return nil
}

/* 查询产品信息表 */
func queryTable() error {
	// 查询数据的SQL语句
	sql := "SELECT productid, name, author, description FROM product WHERE productid = 1"

	// 执行查询
	rows, err := db.Query(sql)
	if err != nil {
		return fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	// 打印查询结果
	fmt.Println("查询结果:")
	for rows.Next() {
		var productid int
		var name, author, description string

		// 扫描结果到变量
		if err = rows.Scan(&productid, &name, &author, &description); err != nil {
			return fmt.Errorf("扫描结果失败: %w", err)
		}

		// 打印结果
		fmt.Printf("ID: %d, 名称: %s, 作者: %s, 描述: %s\n",
			productid, name, author, description)
	}

	// 检查遍历过程中是否有错误
	if err = rows.Err(); err != nil {
		return fmt.Errorf("遍历结果失败: %w", err)
	}

	return nil
}

/* 删除产品信息表数据 */
func deleteTable() error {
	// 删除数据的SQL语句
	sql := "DELETE FROM product WHERE productid = 1"

	// 执行SQL语句
	_, err := db.Exec(sql)
	if err != nil {
		return fmt.Errorf("删除数据失败: %w", err)
	}

	fmt.Println("删除数据成功")
	return nil
}

/* 关闭数据库连接 */
func disconnect() error {
	// 关闭数据库连接
	if err := db.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	fmt.Println("数据库连接已关闭")
	return nil
}
