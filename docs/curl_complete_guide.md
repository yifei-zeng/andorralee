# 蜜罐管理系统 cURL 完整操作指南

## 系统概述

本指南基于当前已实现的功能，提供完整的cURL命令操作说明。系统已成功启动并监听在8080端口。

## 基础配置

### 环境变量设置
```bash
# 设置API基础URL
export BASE_URL="http://localhost:8080/api/v1"

# 设置通用请求头
export HEADERS='-H "Content-Type: application/json" -H "Accept: application/json"'

# 测试API连接
echo "🔍 测试API连接..."
curl -s -X GET "$BASE_URL/docker/images" $HEADERS > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ API连接成功！系统正常运行"
else
    echo "❌ API连接失败，请确保系统已启动并监听在8080端口"
fi
```

## 1. Docker镜像管理

### 拉取Docker镜像
```bash
# 拉取单个镜像
echo "🐳 拉取Docker镜像..."
curl -X POST "$BASE_URL/docker/pull" \
  $HEADERS \
  -d '{
    "image": "nginx:latest"
  }'

# 拉取常用蜜罐镜像
echo "🍯 拉取蜜罐相关镜像..."

# 拉取Ubuntu镜像
curl -X POST "$BASE_URL/docker/pull" \
  $HEADERS \
  -d '{
    "image": "ubuntu:20.04"
  }'

# 拉取Alpine镜像
curl -X POST "$BASE_URL/docker/pull" \
  $HEADERS \
  -d '{
    "image": "alpine:latest"
  }'

# 拉取MySQL镜像
curl -X POST "$BASE_URL/docker/pull" \
  $HEADERS \
  -d '{
    "image": "mysql:8.0"
  }'

# 拉取Redis镜像
curl -X POST "$BASE_URL/docker/pull" \
  $HEADERS \
  -d '{
    "image": "redis:latest"
  }'

# 拉取Cowrie蜜罐镜像
curl -X POST "$BASE_URL/docker/pull" \
  $HEADERS \
  -d '{
    "image": "cowrie/cowrie:latest"
  }'
```

### 查看和管理镜像
```bash
# 获取所有Docker镜像
echo "📋 获取Docker镜像列表..."
curl -X GET "$BASE_URL/docker/images" $HEADERS

# 获取数据库中的镜像记录
echo "💾 获取数据库中的镜像记录..."
curl -X GET "$BASE_URL/docker/images/db" $HEADERS

# 根据ID获取镜像详情
echo "🔍 获取镜像详情..."
curl -X GET "$BASE_URL/docker/images/sha256:abc123..." $HEADERS

# 删除指定镜像
echo "🗑️ 删除镜像..."
curl -X DELETE "$BASE_URL/docker/images/sha256:abc123..." $HEADERS

# 为镜像添加标签
echo "🏷️ 为镜像添加标签..."
curl -X POST "$BASE_URL/docker/images/sha256:abc123.../tag" \
  $HEADERS \
  -d '{
    "repo": "my-honeypot",
    "tag": "v1.0"
  }'
```

## 2. 容器实例管理

### 创建容器实例
```bash
# 创建SSH蜜罐实例
echo "🍯 创建SSH蜜罐实例..."
curl -X POST "$BASE_URL/container-instances" \
  $HEADERS \
  -d '{
    "name": "SSH蜜罐实例1",
    "honeypot_name": "ssh-honeypot-1",
    "image_name": "ubuntu:20.04",
    "protocol": "ssh",
    "interface_type": "terminal",
    "port_mappings": {
      "22": "2222"
    },
    "environment": {
      "SSH_PORT": "2222",
      "HONEYPOT_TYPE": "SSH"
    },
    "description": "SSH协议蜜罐实例"
  }'

# 创建Web蜜罐实例
echo "🌐 创建Web蜜罐实例..."
curl -X POST "$BASE_URL/container-instances" \
  $HEADERS \
  -d '{
    "name": "Web蜜罐实例1",
    "honeypot_name": "web-honeypot-1",
    "image_name": "nginx:latest",
    "protocol": "http",
    "interface_type": "web",
    "port_mappings": {
      "80": "8080"
    },
    "environment": {
      "NGINX_PORT": "8080",
      "HONEYPOT_TYPE": "WEB"
    },
    "description": "Web服务蜜罐实例"
  }'

# 创建数据库蜜罐实例
echo "🗄️ 创建数据库蜜罐实例..."
curl -X POST "$BASE_URL/container-instances" \
  $HEADERS \
  -d '{
    "name": "MySQL蜜罐实例1",
    "honeypot_name": "mysql-honeypot-1",
    "image_name": "mysql:8.0",
    "protocol": "mysql",
    "interface_type": "database",
    "port_mappings": {
      "3306": "3307"
    },
    "environment": {
      "MYSQL_ROOT_PASSWORD": "honeypot123",
      "MYSQL_DATABASE": "honeypot",
      "HONEYPOT_TYPE": "MYSQL"
    },
    "description": "MySQL数据库蜜罐实例"
  }'

# 批量创建多个SSH蜜罐实例
echo "🔄 批量创建SSH蜜罐实例..."
for i in {1..5}; do
  echo "创建SSH蜜罐实例 $i..."
  curl -X POST "$BASE_URL/container-instances" \
    $HEADERS \
    -d "{
      \"name\": \"SSH蜜罐$i\",
      \"honeypot_name\": \"ssh-honeypot-$i\",
      \"image_name\": \"ubuntu:20.04\",
      \"protocol\": \"ssh\",
      \"interface_type\": \"terminal\",
      \"port_mappings\": {
        \"22\": \"$((2220 + i))\"
      },
      \"environment\": {
        \"SSH_PORT\": \"$((2220 + i))\",
        \"HONEYPOT_TYPE\": \"SSH\",
        \"INSTANCE_ID\": \"$i\"
      },
      \"description\": \"SSH蜜罐实例$i\"
    }"
  echo ""
  sleep 1
done
```

### 查看容器实例
```bash
# 获取所有容器实例
echo "📋 获取所有容器实例..."
curl -X GET "$BASE_URL/container-instances" $HEADERS

# 根据ID获取容器实例详情
echo "🔍 获取容器实例详情..."
curl -X GET "$BASE_URL/container-instances/1" $HEADERS

# 根据状态获取容器实例
echo "📊 获取运行中的容器实例..."
curl -X GET "$BASE_URL/container-instances/status/running" $HEADERS

echo "📊 获取已停止的容器实例..."
curl -X GET "$BASE_URL/container-instances/status/stopped" $HEADERS

echo "📊 获取已创建的容器实例..."
curl -X GET "$BASE_URL/container-instances/status/created" $HEADERS

# 获取容器实例状态
echo "📈 获取容器实例状态..."
curl -X GET "$BASE_URL/container-instances/1/status" $HEADERS
```

### 控制容器实例
```bash
# 启动容器实例
echo "▶️ 启动容器实例..."
curl -X POST "$BASE_URL/container-instances/1/start" \
  $HEADERS \
  -d '{}'

# 停止容器实例
echo "⏹️ 停止容器实例..."
curl -X POST "$BASE_URL/container-instances/1/stop" \
  $HEADERS \
  -d '{}'

# 重启容器实例
echo "🔄 重启容器实例..."
curl -X POST "$BASE_URL/container-instances/1/restart" \
  $HEADERS \
  -d '{}'

# 删除容器实例
echo "🗑️ 删除容器实例..."
curl -X DELETE "$BASE_URL/container-instances/1" $HEADERS

# 同步所有容器实例状态
echo "🔄 同步所有容器实例状态..."
curl -X POST "$BASE_URL/container-instances/sync-status" \
  $HEADERS \
  -d '{}'

# 批量启动所有stopped状态的实例
echo "▶️ 批量启动所有停止的实例..."
stopped_instances=$(curl -s -X GET "$BASE_URL/container-instances/status/stopped" $HEADERS)
echo "$stopped_instances" | jq -r '.[].id' | while read id; do
  echo "启动容器实例 $id..."
  curl -X POST "$BASE_URL/container-instances/$id/start" \
    $HEADERS \
    -d '{}'
  sleep 1
done

# 批量停止所有running状态的实例
echo "⏹️ 批量停止所有运行的实例..."
running_instances=$(curl -s -X GET "$BASE_URL/container-instances/status/running" $HEADERS)
echo "$running_instances" | jq -r '.[].id' | while read id; do
  echo "停止容器实例 $id..."
  curl -X POST "$BASE_URL/container-instances/$id/stop" \
    $HEADERS \
    -d '{}'
  sleep 1
done
```

## 3. Headling认证日志管理

### 拉取和查看Headling日志
```bash
# 拉取Headling认证日志
echo "📥 拉取Headling认证日志..."
curl -X POST "$BASE_URL/headling/pull-logs" \
  $HEADERS \
  -d '{
    "container_id": "container_123"
  }'

# 获取所有Headling日志
echo "📋 获取所有Headling日志..."
curl -X GET "$BASE_URL/headling/logs" $HEADERS

# 根据ID获取Headling日志
echo "🔍 根据ID获取Headling日志..."
curl -X GET "$BASE_URL/headling/logs/1" $HEADERS

# 根据容器ID获取Headling日志
echo "🐳 根据容器ID获取Headling日志..."
curl -X GET "$BASE_URL/headling/logs/container/container_123" $HEADERS

# 根据源IP获取Headling日志
echo "🌐 根据源IP获取Headling日志..."
curl -X GET "$BASE_URL/headling/logs/source-ip/192.168.1.100" $HEADERS

# 根据协议获取Headling日志
echo "🔌 根据协议获取Headling日志..."
curl -X GET "$BASE_URL/headling/logs/protocol/ssh" $HEADERS

# 根据时间范围获取Headling日志
echo "📅 根据时间范围获取Headling日志..."
curl -X GET "$BASE_URL/headling/logs/time-range?start=2025-01-01T00:00:00Z&end=2025-01-31T23:59:59Z" $HEADERS

# 删除容器相关的Headling日志
echo "🗑️ 删除容器相关的Headling日志..."
curl -X DELETE "$BASE_URL/headling/logs/container/container_123" $HEADERS
```

### Headling统计分析
```bash
# 获取Headling统计信息
echo "📊 获取Headling统计信息..."
curl -X GET "$BASE_URL/headling/statistics" $HEADERS

# 获取攻击者IP统计
echo "🎯 获取攻击者IP统计..."
curl -X GET "$BASE_URL/headling/attacker-statistics" $HEADERS

# 获取顶级攻击者
echo "🏆 获取前10个攻击者..."
curl -X GET "$BASE_URL/headling/top-attackers?limit=10" $HEADERS

# 获取常用用户名
echo "👤 获取常用用户名..."
curl -X GET "$BASE_URL/headling/top-usernames?limit=10" $HEADERS

# 获取常用密码
echo "🔑 获取常用密码..."
curl -X GET "$BASE_URL/headling/top-passwords?limit=10" $HEADERS
```

## 4. Cowrie蜜罐日志管理

### 拉取和查看Cowrie日志
```bash
# 拉取Cowrie蜜罐日志
echo "📥 拉取Cowrie蜜罐日志..."
curl -X POST "$BASE_URL/cowrie/pull-logs" \
  $HEADERS \
  -d '{
    "container_id": "cowrie_container_123"
  }'

# 获取所有Cowrie日志
echo "📋 获取所有Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs" $HEADERS

# 根据ID获取Cowrie日志
echo "🔍 根据ID获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/1" $HEADERS

# 根据容器ID获取Cowrie日志
echo "🐳 根据容器ID获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/container/cowrie_container_123" $HEADERS

# 根据源IP获取Cowrie日志
echo "🌐 根据源IP获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/source-ip/192.168.1.100" $HEADERS

# 根据协议获取Cowrie日志
echo "🔌 根据协议获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/protocol/ssh" $HEADERS

# 根据命令获取Cowrie日志
echo "💻 根据命令获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/command/ls" $HEADERS

# 根据用户名获取Cowrie日志
echo "👤 根据用户名获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/username/root" $HEADERS

# 根据命令识别状态获取Cowrie日志
echo "✅ 获取命令识别成功的日志..."
curl -X GET "$BASE_URL/cowrie/logs/command-found/true" $HEADERS

echo "❌ 获取命令识别失败的日志..."
curl -X GET "$BASE_URL/cowrie/logs/command-found/false" $HEADERS

# 根据时间范围获取Cowrie日志
echo "📅 根据时间范围获取Cowrie日志..."
curl -X GET "$BASE_URL/cowrie/logs/time-range?start=2025-01-01T00:00:00Z&end=2025-01-31T23:59:59Z" $HEADERS

# 删除容器相关的Cowrie日志
echo "🗑️ 删除容器相关的Cowrie日志..."
curl -X DELETE "$BASE_URL/cowrie/logs/container/cowrie_container_123" $HEADERS
```

### Cowrie统计分析
```bash
# 获取Cowrie统计信息
echo "📊 获取Cowrie统计信息..."
curl -X GET "$BASE_URL/cowrie/statistics" $HEADERS

# 获取Cowrie攻击者行为统计
echo "🎯 获取Cowrie攻击者行为统计..."
curl -X GET "$BASE_URL/cowrie/attacker-behavior" $HEADERS

# 获取Cowrie顶级攻击者
echo "🏆 获取Cowrie前10个攻击者..."
curl -X GET "$BASE_URL/cowrie/top-attackers?limit=10" $HEADERS

# 获取Cowrie常用命令
echo "💻 获取Cowrie常用命令..."
curl -X GET "$BASE_URL/cowrie/top-commands?limit=10" $HEADERS

# 获取Cowrie常用用户名
echo "👤 获取Cowrie常用用户名..."
curl -X GET "$BASE_URL/cowrie/top-usernames?limit=10" $HEADERS

# 获取Cowrie常用密码
echo "🔑 获取Cowrie常用密码..."
curl -X GET "$BASE_URL/cowrie/top-passwords?limit=10" $HEADERS

# 获取Cowrie常用客户端指纹
echo "🔍 获取Cowrie常用客户端指纹..."
curl -X GET "$BASE_URL/cowrie/top-fingerprints?limit=10" $HEADERS
```

## 5. 蜜罐模板和实例管理

### 蜜罐模板管理
```bash
# 获取所有蜜罐模板
echo "📋 获取所有蜜罐模板..."
curl -X GET "$BASE_URL/honeypot/templates" $HEADERS

# 根据ID获取蜜罐模板
echo "🔍 获取蜜罐模板详情..."
curl -X GET "$BASE_URL/honeypot/templates/1" $HEADERS

# 创建蜜罐模板
echo "➕ 创建蜜罐模板..."
curl -X POST "$BASE_URL/honeypot/templates" \
  $HEADERS \
  -d '{
    "name": "SSH蜜罐模板",
    "type": "SSH",
    "description": "标准SSH蜜罐模板",
    "image_name": "ubuntu:20.04",
    "default_port": 22,
    "config": {
      "hostname": "server01",
      "log_level": "INFO",
      "max_connections": 100
    }
  }'

# 更新蜜罐模板
echo "✏️ 更新蜜罐模板..."
curl -X PUT "$BASE_URL/honeypot/templates/1" \
  $HEADERS \
  -d '{
    "name": "更新的SSH蜜罐模板",
    "description": "更新后的SSH蜜罐模板",
    "config": {
      "hostname": "updated-server",
      "log_level": "DEBUG",
      "max_connections": 200
    }
  }'

# 删除蜜罐模板
echo "🗑️ 删除蜜罐模板..."
curl -X DELETE "$BASE_URL/honeypot/templates/1" $HEADERS

# 部署蜜罐模板
echo "🚀 部署蜜罐模板..."
curl -X POST "$BASE_URL/honeypot/templates/1/deploy" \
  $HEADERS \
  -d '{
    "instance_name": "SSH蜜罐实例1",
    "port": 2222
  }'

# 导入蜜罐模板
echo "📥 导入蜜罐模板..."
curl -X POST "$BASE_URL/honeypot/templates/import" \
  $HEADERS \
  -d '{
    "template_data": {
      "name": "导入的Web蜜罐模板",
      "type": "WEB",
      "description": "从外部导入的Web蜜罐模板",
      "image_name": "nginx:latest",
      "default_port": 80,
      "config": {
        "server_name": "honeypot.local",
        "document_root": "/var/www/html"
      }
    }
  }'
```

### 蜜罐实例管理
```bash
# 获取所有蜜罐实例
echo "📋 获取所有蜜罐实例..."
curl -X GET "$BASE_URL/honeypot/instances" $HEADERS

# 根据ID获取蜜罐实例
echo "🔍 获取蜜罐实例详情..."
curl -X GET "$BASE_URL/honeypot/instances/1" $HEADERS

# 更新蜜罐实例
echo "✏️ 更新蜜罐实例..."
curl -X PUT "$BASE_URL/honeypot/instances/1" \
  $HEADERS \
  -d '{
    "name": "更新的蜜罐实例",
    "description": "更新后的描述"
  }'

# 部署蜜罐实例
echo "🚀 部署蜜罐实例..."
curl -X POST "$BASE_URL/honeypot/instances/1/deploy" \
  $HEADERS \
  -d '{}'

# 停止蜜罐实例
echo "⏹️ 停止蜜罐实例..."
curl -X POST "$BASE_URL/honeypot/instances/1/stop" \
  $HEADERS \
  -d '{}'

# 获取蜜罐实例日志
echo "📄 获取蜜罐实例日志..."
curl -X GET "$BASE_URL/honeypot/instances/1/logs" $HEADERS

# 删除蜜罐实例
echo "🗑️ 删除蜜罐实例..."
curl -X DELETE "$BASE_URL/honeypot/instances/1" $HEADERS
```

### 蜜罐日志管理
```bash
# 获取所有蜜罐日志
echo "📋 获取所有蜜罐日志..."
curl -X GET "$BASE_URL/honeypot/logs" $HEADERS

# 根据ID获取蜜罐日志
echo "🔍 根据ID获取蜜罐日志..."
curl -X GET "$BASE_URL/honeypot/logs/1" $HEADERS

# 根据实例ID获取蜜罐日志
echo "🍯 根据实例ID获取蜜罐日志..."
curl -X GET "$BASE_URL/honeypot/logs/instance/1" $HEADERS
```
```
