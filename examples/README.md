# 达梦数据库连接示例

本目录包含两个达梦数据库连接的示例代码，分别使用原生SQL和GORM两种方式实现。

## 文件说明

- `dameng_example.go`: 使用原生SQL方式连接达梦数据库的示例代码
- `dameng_gorm_example.go`: 使用GORM框架连接达梦数据库的示例代码

# 达梦数据库连接示例
## 使用方法
### 1. 安装达梦数据库驱动

```bash
# 安装原生SQL驱动
go get github.com/godoes/dm

# 安装GORM驱动
go get github.com/godoes/gorm-dameng
```

### 2. 代码示例说明

本项目提供了三个达梦数据库连接示例：

1. `dameng_simple_example.go` - 简单连接示例，展示基本的数据库连接和查询操作
2. `dameng_example.go` - 完整示例，包含创建表、插入、更新、查询和删除操作
3. `dameng_gorm_example.go` - 使用GORM框架的示例

> 注意：示例代码已更新，将已弃用的`io/ioutil`包替换为`os`包，以兼容新版Go。

### 2. 配置连接字符串

连接字符串格式：

```
# 原生SQL连接格式
dm://用户名:密码@主机地址:端口号

# GORM连接格式
dm://用户名:密码@主机地址:端口号/数据库名
```

请根据实际环境修改示例代码中的连接字符串。

### 3. 运行示例

```bash
# 运行原生SQL示例
go run dameng_example.go

# 运行GORM示例
go run dameng_gorm_example.go
```

## 注意事项

1. 确保达梦数据库服务已启动并可访问
2. 确保用户名和密码正确
3. 如果使用BLOB/CLOB类型，需要正确处理二进制数据
4. 在生产环境中，请妥善保管数据库连接信息，不要硬编码在源代码中

## 与项目集成

本项目已在`internal/config/config.go`中实现了达梦数据库的配置和初始化，可以通过以下方式使用：

```go
// 初始化达梦数据库连接
if err := config.InitDameng(); err != nil {
    // 处理错误
}

// 使用达梦数据库连接
db := config.DamengDB
```

在`internal/repositories/dameng_repo.go`中已实现了基于GORM的达梦数据库仓库，可以直接使用。