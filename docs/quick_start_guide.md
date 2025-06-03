# 蜜罐管理系统快速开始指南

## 🚀 系统启动

### 1. 启动系统
```bash
# 在项目根目录下启动系统
go run cmd/main.go
```

系统启动成功后会显示：
```
✅ Docker 客户端初始化成功
✅ MySQL 数据库连接成功
✅ MySQL数据库表初始化成功
✅ 达梦数据库连接成功！
🚀 服务启动中，监听端口: 8080...
```

### 2. 验证系统状态
```bash
# 使用cURL测试连接
curl -X GET "http://localhost:8080/api/v1/docker/images" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json"
```

## 🐳 Docker镜像管理

### 拉取常用蜜罐镜像
```bash
# 设置环境变量
export BASE_URL="http://localhost:8080/api/v1"
export HEADERS='-H "Content-Type: application/json" -H "Accept: application/json"'

# 拉取Ubuntu镜像（用于SSH蜜罐）
curl -X POST "$BASE_URL/docker/pull" $HEADERS -d '{"image": "ubuntu:20.04"}'

# 拉取Nginx镜像（用于Web蜜罐）
curl -X POST "$BASE_URL/docker/pull" $HEADERS -d '{"image": "nginx:latest"}'

# 拉取MySQL镜像（用于数据库蜜罐）
curl -X POST "$BASE_URL/docker/pull" $HEADERS -d '{"image": "mysql:8.0"}'

# 查看已拉取的镜像
curl -X GET "$BASE_URL/docker/images" $HEADERS
```

### PowerShell版本
```powershell
# 设置环境变量
$BaseURL = "http://localhost:8080/api/v1"
$Headers = @{"Content-Type" = "application/json"; "Accept" = "application/json"}

# 拉取镜像函数
function Pull-DockerImage {
    param([string]$ImageName)
    $Body = @{ image = $ImageName } | ConvertTo-Json
    Invoke-RestMethod -Uri "$BaseURL/docker/pull" -Method POST -Headers $Headers -Body $Body
}

# 拉取常用镜像
Pull-DockerImage -ImageName "ubuntu:20.04"
Pull-DockerImage -ImageName "nginx:latest"
Pull-DockerImage -ImageName "mysql:8.0"

# 查看镜像列表
Invoke-RestMethod -Uri "$BaseURL/docker/images" -Method GET -Headers $Headers
```

## 🍯 创建蜜罐实例

### 创建SSH蜜罐
```bash
# 创建SSH蜜罐实例
curl -X POST "$BASE_URL/container-instances" $HEADERS -d '{
  "name": "SSH蜜罐1",
  "honeypot_name": "ssh-honeypot-1",
  "image_name": "ubuntu:20.04",
  "protocol": "ssh",
  "interface_type": "terminal",
  "port_mappings": {"22": "2222"},
  "environment": {
    "SSH_PORT": "2222",
    "HONEYPOT_TYPE": "SSH"
  },
  "description": "SSH协议蜜罐实例"
}'
```

### 创建Web蜜罐
```bash
# 创建Web蜜罐实例
curl -X POST "$BASE_URL/container-instances" $HEADERS -d '{
  "name": "Web蜜罐1",
  "honeypot_name": "web-honeypot-1",
  "image_name": "nginx:latest",
  "protocol": "http",
  "interface_type": "web",
  "port_mappings": {"80": "8080"},
  "environment": {
    "NGINX_PORT": "8080",
    "HONEYPOT_TYPE": "WEB"
  },
  "description": "Web服务蜜罐实例"
}'
```

### PowerShell版本
```powershell
# 创建SSH蜜罐
function New-SSHHoneypot {
    param([string]$Name, [int]$Port = 2222)
    
    $Body = @{
        name = $Name
        honeypot_name = "ssh-honeypot-$(Get-Random)"
        image_name = "ubuntu:20.04"
        protocol = "ssh"
        interface_type = "terminal"
        port_mappings = @{"22" = $Port.ToString()}
        environment = @{
            "SSH_PORT" = $Port.ToString()
            "HONEYPOT_TYPE" = "SSH"
        }
        description = "SSH协议蜜罐实例"
    } | ConvertTo-Json -Depth 10
    
    Invoke-RestMethod -Uri "$BaseURL/container-instances" -Method POST -Headers $Headers -Body $Body
}

# 创建Web蜜罐
function New-WebHoneypot {
    param([string]$Name, [int]$Port = 8080)
    
    $Body = @{
        name = $Name
        honeypot_name = "web-honeypot-$(Get-Random)"
        image_name = "nginx:latest"
        protocol = "http"
        interface_type = "web"
        port_mappings = @{"80" = $Port.ToString()}
        environment = @{
            "NGINX_PORT" = $Port.ToString()
            "HONEYPOT_TYPE" = "WEB"
        }
        description = "Web服务蜜罐实例"
    } | ConvertTo-Json -Depth 10
    
    Invoke-RestMethod -Uri "$BaseURL/container-instances" -Method POST -Headers $Headers -Body $Body
}

# 使用示例
New-SSHHoneypot -Name "测试SSH蜜罐" -Port 2222
New-WebHoneypot -Name "测试Web蜜罐" -Port 8080
```

## 📋 管理蜜罐实例

### 查看所有实例
```bash
# 获取所有容器实例
curl -X GET "$BASE_URL/container-instances" $HEADERS

# 获取运行中的实例
curl -X GET "$BASE_URL/container-instances/status/running" $HEADERS

# 获取已停止的实例
curl -X GET "$BASE_URL/container-instances/status/stopped" $HEADERS
```

### 控制实例状态
```bash
# 启动实例（假设实例ID为1）
curl -X POST "$BASE_URL/container-instances/1/start" $HEADERS -d '{}'

# 停止实例
curl -X POST "$BASE_URL/container-instances/1/stop" $HEADERS -d '{}'

# 重启实例
curl -X POST "$BASE_URL/container-instances/1/restart" $HEADERS -d '{}'

# 获取实例状态
curl -X GET "$BASE_URL/container-instances/1/status" $HEADERS

# 删除实例
curl -X DELETE "$BASE_URL/container-instances/1" $HEADERS
```

### PowerShell版本
```powershell
# 获取所有实例
function Get-AllInstances {
    Invoke-RestMethod -Uri "$BaseURL/container-instances" -Method GET -Headers $Headers
}

# 启动实例
function Start-Instance {
    param([int]$ID)
    Invoke-RestMethod -Uri "$BaseURL/container-instances/$ID/start" -Method POST -Headers $Headers -Body '{}'
}

# 停止实例
function Stop-Instance {
    param([int]$ID)
    Invoke-RestMethod -Uri "$BaseURL/container-instances/$ID/stop" -Method POST -Headers $Headers -Body '{}'
}

# 获取实例状态
function Get-InstanceStatus {
    param([int]$ID)
    Invoke-RestMethod -Uri "$BaseURL/container-instances/$ID/status" -Method GET -Headers $Headers
}

# 使用示例
$Instances = Get-AllInstances
Start-Instance -ID 1
Get-InstanceStatus -ID 1
```

## 📊 日志管理

### Headling认证日志
```bash
# 拉取认证日志
curl -X POST "$BASE_URL/headling/pull-logs" $HEADERS -d '{
  "container_id": "your_container_id"
}'

# 查看所有认证日志
curl -X GET "$BASE_URL/headling/logs" $HEADERS

# 获取统计信息
curl -X GET "$BASE_URL/headling/statistics" $HEADERS

# 获取顶级攻击者
curl -X GET "$BASE_URL/headling/top-attackers?limit=10" $HEADERS
```

### Cowrie蜜罐日志
```bash
# 拉取Cowrie日志
curl -X POST "$BASE_URL/cowrie/pull-logs" $HEADERS -d '{
  "container_id": "your_cowrie_container_id"
}'

# 查看所有Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs" $HEADERS

# 获取常用命令
curl -X GET "$BASE_URL/cowrie/top-commands?limit=10" $HEADERS

# 获取攻击者行为统计
curl -X GET "$BASE_URL/cowrie/attacker-behavior" $HEADERS
```

## 🔧 批量操作

### 批量创建蜜罐
```bash
# 批量创建5个SSH蜜罐
for i in {1..5}; do
  echo "创建SSH蜜罐 $i..."
  curl -X POST "$BASE_URL/container-instances" $HEADERS -d "{
    \"name\": \"SSH蜜罐$i\",
    \"honeypot_name\": \"ssh-honeypot-$i\",
    \"image_name\": \"ubuntu:20.04\",
    \"protocol\": \"ssh\",
    \"interface_type\": \"terminal\",
    \"port_mappings\": {\"22\": \"$((2220 + i))\"},
    \"environment\": {
      \"SSH_PORT\": \"$((2220 + i))\",
      \"HONEYPOT_TYPE\": \"SSH\"
    },
    \"description\": \"SSH蜜罐实例$i\"
  }"
  sleep 1
done
```

### 批量启动实例
```bash
# 获取所有停止的实例并启动
stopped_instances=$(curl -s -X GET "$BASE_URL/container-instances/status/stopped" $HEADERS)
echo "$stopped_instances" | jq -r '.[].id' | while read id; do
  echo "启动实例 $id..."
  curl -X POST "$BASE_URL/container-instances/$id/start" $HEADERS -d '{}'
  sleep 1
done
```

### PowerShell批量操作
```powershell
# 批量创建蜜罐
function New-HoneypotCluster {
    param([int]$Count = 3)
    
    for ($i = 1; $i -le $Count; $i++) {
        $Name = "蜜罐实例$i"
        $Port = 2220 + $i
        New-SSHHoneypot -Name $Name -Port $Port
        Start-Sleep -Seconds 1
    }
}

# 批量启动所有停止的实例
function Start-AllStoppedInstances {
    $StoppedInstances = Invoke-RestMethod -Uri "$BaseURL/container-instances/status/stopped" -Method GET -Headers $Headers
    
    foreach ($Instance in $StoppedInstances) {
        Write-Host "启动实例: $($Instance.name)"
        Start-Instance -ID $Instance.id
        Start-Sleep -Seconds 1
    }
}

# 使用示例
New-HoneypotCluster -Count 3
Start-AllStoppedInstances
```

## 📈 监控和统计

### 系统状态监控
```bash
# 同步所有容器状态
curl -X POST "$BASE_URL/container-instances/sync-status" $HEADERS -d '{}'

# 获取Docker镜像日志
curl -X GET "$BASE_URL/docker/image-logs" $HEADERS

# 获取容器日志分析结果
curl -X GET "$BASE_URL/container-logs/segments" $HEADERS
```

### AI功能
```bash
# 日志语义分割
curl -X POST "$BASE_URL/ai/semantic-segment" $HEADERS -d '{
  "container_id": "container_123",
  "log_content": "2025-01-15 10:30:45 [INFO] SSH connection from 192.168.1.100:45678"
}'

# 图像语义分割
curl -X POST "$BASE_URL/ai/image-segment" $HEADERS -d '{
  "image_path": "/path/to/image.jpg",
  "model": "default"
}'
```

## 🎯 常用操作组合

### 完整的蜜罐部署流程
```bash
#!/bin/bash
echo "🚀 开始部署蜜罐集群..."

# 1. 拉取必要镜像
echo "📥 拉取Docker镜像..."
curl -X POST "$BASE_URL/docker/pull" $HEADERS -d '{"image": "ubuntu:20.04"}'
curl -X POST "$BASE_URL/docker/pull" $HEADERS -d '{"image": "nginx:latest"}'

# 2. 创建SSH蜜罐
echo "🍯 创建SSH蜜罐..."
ssh_result=$(curl -s -X POST "$BASE_URL/container-instances" $HEADERS -d '{
  "name": "生产SSH蜜罐",
  "honeypot_name": "prod-ssh-honeypot",
  "image_name": "ubuntu:20.04",
  "protocol": "ssh",
  "interface_type": "terminal",
  "port_mappings": {"22": "2222"},
  "environment": {"SSH_PORT": "2222", "HONEYPOT_TYPE": "SSH"},
  "description": "生产环境SSH蜜罐"
}')

ssh_id=$(echo "$ssh_result" | jq -r '.id')
echo "✅ SSH蜜罐创建成功，ID: $ssh_id"

# 3. 创建Web蜜罐
echo "🌐 创建Web蜜罐..."
web_result=$(curl -s -X POST "$BASE_URL/container-instances" $HEADERS -d '{
  "name": "生产Web蜜罐",
  "honeypot_name": "prod-web-honeypot",
  "image_name": "nginx:latest",
  "protocol": "http",
  "interface_type": "web",
  "port_mappings": {"80": "8080"},
  "environment": {"NGINX_PORT": "8080", "HONEYPOT_TYPE": "WEB"},
  "description": "生产环境Web蜜罐"
}')

web_id=$(echo "$web_result" | jq -r '.id')
echo "✅ Web蜜罐创建成功，ID: $web_id"

# 4. 启动蜜罐
echo "▶️ 启动蜜罐实例..."
curl -X POST "$BASE_URL/container-instances/$ssh_id/start" $HEADERS -d '{}'
curl -X POST "$BASE_URL/container-instances/$web_id/start" $HEADERS -d '{}'

# 5. 验证状态
echo "📊 验证蜜罐状态..."
curl -X GET "$BASE_URL/container-instances" $HEADERS

echo "🎉 蜜罐集群部署完成！"
echo "SSH蜜罐端口: 2222"
echo "Web蜜罐端口: 8080"
```

## 📚 更多资源

- **完整PowerShell指南**: `docs/powershell_complete_guide.md`
- **完整cURL指南**: `docs/curl_complete_guide.md`
- **API文档**: `docs/api_documentation.md`
- **系统功能清单**: `docs/complete_system_features.md`

## 🆘 故障排除

### 常见问题
1. **API连接失败**: 确保系统已启动并监听8080端口
2. **Docker镜像拉取失败**: 检查网络连接和Docker服务状态
3. **容器启动失败**: 检查端口是否被占用
4. **数据库连接失败**: 确认MySQL和达梦数据库配置正确

### 检查系统状态
```bash
# 检查系统健康状态
curl -X GET "$BASE_URL/docker/images" $HEADERS
curl -X GET "$BASE_URL/container-instances" $HEADERS
```

🎊 **恭喜！您已经掌握了蜜罐管理系统的基本操作！**
