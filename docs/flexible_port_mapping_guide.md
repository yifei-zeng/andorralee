# 灵活端口映射功能使用指南

## 📋 概述

灵活端口映射功能解决了原有端口映射固定到8080端口的问题，现在支持：
- 人为指定具体的主机端口
- 自动分配可用端口
- 混合模式（部分指定，部分自动）
- 智能端口冲突检测

## 🚀 新增API

### POST /api/v1/ports/flexible-mapping

创建灵活的端口映射，支持多种配置方式。

#### 请求参数

```json
{
  "container_name": "string (必需)",
  "service_type": "string (可选)",
  "allow_auto_allocation": "boolean (可选)",
  
  // 方式1: 使用容器端口数组
  "container_ports": [22, 80, 3306],
  "preferred_host_ports": [12345, 0, 13306],
  
  // 方式2: 使用端口映射对象
  "port_mappings": {
    "22": "12345",
    "80": "auto",
    "3306": "13306"
  }
}
```

#### 参数说明

- `container_name`: 容器名称（必需）
- `service_type`: 服务类型，如ssh、mysql、web等（可选）
- `allow_auto_allocation`: 是否允许自动分配端口（默认false）
- `container_ports`: 容器端口数组
- `preferred_host_ports`: 首选主机端口数组，0表示自动分配
- `port_mappings`: 端口映射对象，值为"auto"表示自动分配

## 📝 使用示例

### 1. 指定具体端口映射

```powershell
# 创建SSH蜜罐，映射到指定端口12345
$body = @{
    container_name = "ssh-honeypot-1"
    service_type = "ssh"
    port_mappings = @{
        "22" = "12345"
    }
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/flexible-mapping" -Method POST -ContentType "application/json" -Body $body
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "port_mapping": {"22": "12345"},
    "container_name": "ssh-honeypot-1",
    "allocated_ports": [12345],
    "service_type": "ssh",
    "message": "灵活端口映射创建成功"
  }
}
```

### 2. 自动分配端口

```powershell
# 创建MySQL蜜罐，自动分配端口
$body = @{
    container_name = "mysql-honeypot-1"
    service_type = "mysql"
    port_mappings = @{
        "3306" = "auto"
    }
    allow_auto_allocation = $true
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/flexible-mapping" -Method POST -ContentType "application/json" -Body $body
```

### 3. 混合模式

```powershell
# 部分指定，部分自动分配
$body = @{
    container_name = "web-honeypot-1"
    service_type = "web"
    port_mappings = @{
        "80" = "18080"      # 指定端口
        "443" = "auto"      # 自动分配
        "22" = "auto"       # 自动分配
    }
    allow_auto_allocation = $true
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/flexible-mapping" -Method POST -ContentType "application/json" -Body $body
```

### 4. 使用数组方式

```powershell
# 使用容器端口数组和首选主机端口数组
$body = @{
    container_name = "multi-service-honeypot"
    service_type = "multi"
    container_ports = @(22, 80, 3306)
    preferred_host_ports = @(12345, 0, 13306)  # 0表示自动分配
    allow_auto_allocation = $true
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/flexible-mapping" -Method POST -ContentType "application/json" -Body $body
```

## 🔧 与容器创建集成

### 更新的容器创建API

现在创建容器时会根据协议自动设置合适的端口映射：

```powershell
# 创建SSH蜜罐，自动分配SSH端口
$body = @{
    name = "ssh-honeypot-1"
    honeypot_name = "ssh-honeypot-1"
    image_name = "andorralee/cowrie:v0.1"
    protocol = "ssh"
    interface_type = "network"
    # port_mappings 会根据协议自动设置为 {"22": "auto"}
    environment = @{"TEST" = "true"}
    description = "SSH蜜罐实例"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/container-instances" -Method POST -ContentType "application/json" -Body $body
```

### 手动指定端口映射

```powershell
# 创建MySQL蜜罐，指定端口映射
$body = @{
    name = "mysql-honeypot-1"
    honeypot_name = "mysql-honeypot-1"
    image_name = "mysql:8.0"
    protocol = "mysql"
    interface_type = "network"
    port_mappings = @{"3306" = "13306"}  # 指定映射到13306端口
    environment = @{
        "MYSQL_ROOT_PASSWORD" = "honeypot123"
    }
    description = "MySQL蜜罐实例"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/container-instances" -Method POST -ContentType "application/json" -Body $body
```

## 🌐 Nginx流量转发配置

创建端口映射后，可以配置Nginx进行流量转发：

### 1. 手动配置Nginx

```nginx
# SSH蜜罐代理配置
upstream ssh_honeypot_backend {
    server 127.0.0.1:12345;
}

server {
    listen 22;
    proxy_pass ssh_honeypot_backend;
    proxy_timeout 1s;
    proxy_responses 1;
}

# MySQL蜜罐代理配置
upstream mysql_honeypot_backend {
    server 127.0.0.1:13306;
}

server {
    listen 3306;
    proxy_pass mysql_honeypot_backend;
    proxy_timeout 1s;
    proxy_responses 1;
}
```

### 2. 使用Nginx管理API（如果已实现）

```powershell
# 为SSH蜜罐创建Nginx代理
$body = @{
    name = "ssh-honeypot-proxy"
    listen_port = 22
    target_port = 12345
    service_type = "ssh"
    description = "SSH蜜罐代理"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8081/api/v1/nginx/proxy" -Method POST -ContentType "application/json" -Body $body
```

## 🎯 最佳实践

### 1. 端口范围规划

- **SSH服务**: 12000-12999
- **Web服务**: 18000-18999  
- **数据库服务**: 13000-13999
- **其他服务**: 19000-19999

### 2. 命名规范

```powershell
# 推荐的容器命名规范
$containerName = "{service}-honeypot-{instance_number}"
# 例如: ssh-honeypot-1, mysql-honeypot-2, web-honeypot-1
```

### 3. 端口冲突处理

```powershell
# 检查端口可用性
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/12345/check" -Method GET

# 获取下一个可用端口
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/next-available" -Method GET
```

## ⚠️ 注意事项

1. **端口冲突**: 系统会自动检测端口冲突，如果指定端口已被占用会返回错误
2. **权限要求**: 监听1024以下的端口需要管理员权限
3. **防火墙**: 确保防火墙允许相应端口的流量
4. **资源限制**: 避免分配过多端口导致系统资源不足

## 🔍 故障排除

### 端口分配失败

```bash
# 检查端口状态
curl http://localhost:8081/api/v1/ports/statistics

# 查看已分配端口
curl http://localhost:8081/api/v1/ports/allocated
```

### 容器无法访问

1. 检查端口映射是否正确
2. 验证防火墙设置
3. 确认容器服务正在运行
4. 检查Nginx代理配置

## 📊 监控和管理

```powershell
# 查看端口统计
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/statistics" -Method GET

# 查看容器端口分配
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/container/ssh-honeypot-1" -Method GET

# 释放端口
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/ports/12345/release" -Method DELETE
```
