# 端口管理功能使用指南

## 概述

新的端口管理功能解决了甲方提出的端口映射需求，支持：

1. **自动端口分配** - 系统自动分配空闲端口
2. **指定端口分配** - 支持指定特定端口
3. **智能端口范围** - 根据服务类型选择合适的端口范围
4. **端口冲突检测** - 避免端口冲突
5. **端口生命周期管理** - 自动释放容器删除时的端口

## 端口分配策略

### 服务类型端口范围

- **MySQL**: 13306-13399
- **SSH**: 12222-12299  
- **HTTP/Web**: 18000-18999
- **默认动态范围**: 10000-19999, 30000-39999

### 使用场景示例

#### 场景1: MySQL容器自动端口分配

```bash
# 创建MySQL容器，自动分配端口
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MySQL蜜罐",
    "honeypot_name": "mysql-honeypot",
    "image_name": "mysql:8.0",
    "protocol": "mysql",
    "port_mappings": {
      "3306": "auto"
    },
    "environment": {
      "MYSQL_ROOT_PASSWORD": "honeypot123"
    },
    "description": "MySQL蜜罐实例"
  }'
```

**响应示例:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "port_mappings": {
      "3306": "13306"
    },
    "requested_port_mappings": {
      "3306": "auto"
    }
  }
}
```

#### 场景2: 指定端口映射

```bash
# 创建容器，指定端口12345
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web蜜罐",
    "honeypot_name": "web-honeypot",
    "image_name": "nginx:latest",
    "protocol": "http",
    "port_mappings": {
      "80": "12345"
    },
    "description": "指定端口的Web蜜罐"
  }'
```

#### 场景3: 混合端口分配

```bash
# 混合使用自动分配和指定端口
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "多服务蜜罐",
    "honeypot_name": "multi-service",
    "image_name": "custom-honeypot:latest",
    "protocol": "multi",
    "port_mappings": {
      "22": "auto",      // SSH自动分配
      "80": "auto",      // HTTP自动分配
      "3306": "auto",    // MySQL自动分配
      "8080": "18080"    // 指定端口
    },
    "description": "多服务蜜罐实例"
  }'
```

## 端口管理API

### 1. 自动分配端口

```bash
POST /api/v1/ports/allocate
{
  "container_id": "container-name",
  "service_type": "mysql",
  "description": "MySQL服务端口"
}
```

### 2. 分配指定端口

```bash
POST /api/v1/ports/allocate-specific
{
  "port": 12345,
  "container_id": "container-name", 
  "service_type": "http",
  "description": "Web服务端口"
}
```

### 3. 自动分配端口映射

```bash
POST /api/v1/ports/auto-allocate-mapping
{
  "container_id": "container-name",
  "port_mappings": {
    "22": "auto",
    "80": "auto", 
    "3306": "13306"
  }
}
```

### 4. 查询端口信息

```bash
# 查询特定端口
GET /api/v1/ports/12345

# 查询所有已分配端口
GET /api/v1/ports/allocated

# 查询容器的端口
GET /api/v1/ports/container/container-name

# 获取端口统计
GET /api/v1/ports/statistics
```

### 5. 释放端口

```bash
# 释放特定端口
DELETE /api/v1/ports/12345/release

# 释放容器的所有端口
DELETE /api/v1/ports/container/container-name/release
```

### 6. 检查端口可用性

```bash
GET /api/v1/ports/12345/check
```

### 7. 获取可用端口

```bash
POST /api/v1/ports/available
{
  "start": 10000,
  "end": 20000,
  "limit": 10
}
```

## PowerShell使用示例

```powershell
# 设置基础URL
$BaseUrl = "http://localhost:8081/api/v1"

# 创建自动端口分配的容器
$containerRequest = @{
    name = "MySQL蜜罐"
    honeypot_name = "mysql-test"
    image_name = "mysql:8.0"
    protocol = "mysql"
    port_mappings = @{
        "3306" = "auto"
    }
    environment = @{
        "MYSQL_ROOT_PASSWORD" = "honeypot123"
    }
    description = "自动端口分配的MySQL蜜罐"
} | ConvertTo-Json -Depth 3

$result = Invoke-RestMethod -Uri "$BaseUrl/container-instances" -Method POST -Body $containerRequest -ContentType "application/json"

Write-Host "容器创建成功，分配的端口: $($result.data.port_mappings.'3306')"
```

## 与Nginx流量转发集成

创建容器后，可以配置Nginx进行流量转发：

```nginx
# /etc/nginx/sites-available/honeypot-mysql
upstream mysql_honeypot {
    server 127.0.0.1:13306;  # 自动分配的端口
}

server {
    listen 3306;
    proxy_pass mysql_honeypot;
    proxy_timeout 1s;
    proxy_responses 1;
}
```

## 最佳实践

1. **使用自动分配** - 对于大多数场景，推荐使用"auto"自动分配端口
2. **服务类型标识** - 正确设置protocol字段，帮助系统选择合适的端口范围
3. **端口监控** - 定期查询端口统计信息，监控端口使用情况
4. **及时清理** - 删除容器时会自动释放端口，无需手动操作
5. **冲突处理** - 如果指定端口被占用，系统会返回错误，可以选择其他端口

## 故障排除

### 端口分配失败
- 检查端口范围是否有可用端口
- 确认指定端口未被占用
- 查看系统端口统计信息

### 容器创建失败
- 检查端口映射格式是否正确
- 确认Docker服务正常运行
- 查看系统日志获取详细错误信息

### 端口冲突
- 使用端口检查API确认端口状态
- 选择其他端口或使用自动分配
- 检查系统防火墙设置

## 测试脚本

使用提供的测试脚本验证功能：

```powershell
# 运行完整测试
.\test-port-management.ps1

# 详细输出模式
.\test-port-management.ps1 -Verbose

# 指定服务器地址
.\test-port-management.ps1 -BaseUrl "http://your-server:8081/api/v1"
```

## 技术实现

- **端口管理器**: 单例模式，线程安全
- **端口范围**: 可配置的端口分配策略
- **冲突检测**: 实时检查端口占用状态
- **生命周期**: 与容器生命周期绑定
- **持久化**: 内存管理，重启后重新扫描
