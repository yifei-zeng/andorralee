# 蜜罐管理系统 cURL 操作指南

## 系统概述

本文档提供了使用cURL命令行工具操作蜜罐管理系统的完整指南，包括Docker镜像管理、容器实例管理、蜜罐日志分析等所有功能。

## 基础配置

### 设置环境变量
```bash
# 设置API基础URL
export BASE_URL="http://localhost:8080/api/v1"

# 设置通用请求头
export HEADERS="-H 'Content-Type: application/json' -H 'Accept: application/json'"
```

## 1. Docker镜像管理

### 拉取Docker镜像
```bash
# 拉取指定镜像
curl -X POST "$BASE_URL/docker/pull" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "nginx:latest"
  }'

# 拉取Cowrie蜜罐镜像
curl -X POST "$BASE_URL/docker/pull" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "cowrie/cowrie:latest"
  }'

# 拉取Headling蜜罐镜像
curl -X POST "$BASE_URL/docker/pull" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "headling/headling:latest"
  }'
```

### 查看镜像列表
```bash
# 获取所有Docker镜像
curl -X GET "$BASE_URL/docker/images" \
  -H "Accept: application/json"

# 获取数据库中的镜像记录
curl -X GET "$BASE_URL/docker/images/db" \
  -H "Accept: application/json"

# 根据ID获取镜像详情
curl -X GET "$BASE_URL/docker/images/sha256:abc123..." \
  -H "Accept: application/json"

# 根据数据库ID获取镜像记录
curl -X GET "$BASE_URL/docker/images/db/1" \
  -H "Accept: application/json"
```

### 删除Docker镜像
```bash
# 删除指定镜像
curl -X DELETE "$BASE_URL/docker/images/sha256:abc123..." \
  -H "Accept: application/json"

# 删除镜像数据库记录
curl -X DELETE "$BASE_URL/docker/images/db/1" \
  -H "Accept: application/json"
```

### 镜像操作日志
```bash
# 获取所有镜像操作日志
curl -X GET "$BASE_URL/docker/image-logs" \
  -H "Accept: application/json"

# 根据ID获取镜像操作日志
curl -X GET "$BASE_URL/docker/image-logs/1" \
  -H "Accept: application/json"

# 根据镜像ID获取操作日志
curl -X GET "$BASE_URL/docker/image-logs/image/sha256:abc123..." \
  -H "Accept: application/json"

# 删除镜像操作日志
curl -X DELETE "$BASE_URL/docker/image-logs/1" \
  -H "Accept: application/json"
```

## 2. 容器实例管理

### 创建容器实例
```bash
# 创建基础容器实例
curl -X POST "$BASE_URL/container-instances" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试蜜罐1",
    "honeypot_name": "cowrie-test",
    "image_name": "cowrie/cowrie:latest",
    "protocol": "ssh",
    "interface_type": "terminal",
    "port_mappings": {
      "22": "2222",
      "80": "8080"
    },
    "environment": {
      "COWRIE_HOSTNAME": "server01",
      "COWRIE_LOG_LEVEL": "INFO"
    },
    "description": "测试用Cowrie蜜罐"
  }'

# 创建Web蜜罐实例
curl -X POST "$BASE_URL/container-instances" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web蜜罐",
    "honeypot_name": "nginx-honeypot",
    "image_name": "nginx:latest",
    "protocol": "http",
    "interface_type": "web",
    "port_mappings": {
      "80": "8080",
      "443": "8443"
    },
    "environment": {
      "NGINX_HOST": "localhost",
      "NGINX_PORT": "80"
    },
    "description": "Web服务蜜罐"
  }'

# 创建SSH蜜罐实例
curl -X POST "$BASE_URL/container-instances" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SSH蜜罐",
    "honeypot_name": "ssh-honeypot",
    "image_name": "headling/headling:latest",
    "protocol": "ssh",
    "interface_type": "terminal",
    "port_mappings": {
      "22": "2222"
    },
    "environment": {
      "SSH_HOST_KEY": "/etc/ssh/ssh_host_rsa_key"
    },
    "description": "SSH认证蜜罐"
  }'
```

### 查看容器实例
```bash
# 获取所有容器实例
curl -X GET "$BASE_URL/container-instances" \
  -H "Accept: application/json"

# 根据ID获取容器实例详情
curl -X GET "$BASE_URL/container-instances/1" \
  -H "Accept: application/json"

# 根据状态获取容器实例
curl -X GET "$BASE_URL/container-instances/status/running" \
  -H "Accept: application/json"

curl -X GET "$BASE_URL/container-instances/status/stopped" \
  -H "Accept: application/json"

curl -X GET "$BASE_URL/container-instances/status/created" \
  -H "Accept: application/json"
```

### 控制容器实例
```bash
# 启动容器实例
curl -X POST "$BASE_URL/container-instances/1/start" \
  -H "Content-Type: application/json" \
  -d '{}'

# 停止容器实例
curl -X POST "$BASE_URL/container-instances/1/stop" \
  -H "Content-Type: application/json" \
  -d '{}'

# 重启容器实例
curl -X POST "$BASE_URL/container-instances/1/restart" \
  -H "Content-Type: application/json" \
  -d '{}'

# 获取容器实例状态
curl -X GET "$BASE_URL/container-instances/1/status" \
  -H "Accept: application/json"

# 同步所有容器实例状态
curl -X POST "$BASE_URL/container-instances/sync-status" \
  -H "Content-Type: application/json" \
  -d '{}'

# 删除容器实例
curl -X DELETE "$BASE_URL/container-instances/1" \
  -H "Accept: application/json"
```

## 3. Headling认证日志管理

### 拉取和查看Headling日志
```bash
# 拉取Headling认证日志
curl -X POST "$BASE_URL/headling/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{
    "container_id": "container_123"
  }'

# 获取所有Headling日志
curl -X GET "$BASE_URL/headling/logs" \
  -H "Accept: application/json"

# 根据ID获取Headling日志
curl -X GET "$BASE_URL/headling/logs/1" \
  -H "Accept: application/json"

# 根据容器ID获取Headling日志
curl -X GET "$BASE_URL/headling/logs/container/container_123" \
  -H "Accept: application/json"

# 根据源IP获取Headling日志
curl -X GET "$BASE_URL/headling/logs/source-ip/192.168.1.100" \
  -H "Accept: application/json"

# 根据协议获取Headling日志
curl -X GET "$BASE_URL/headling/logs/protocol/ssh" \
  -H "Accept: application/json"

# 根据时间范围获取Headling日志
curl -X GET "$BASE_URL/headling/logs/time-range?start_time=2025-01-01T00:00:00Z&end_time=2025-01-31T23:59:59Z" \
  -H "Accept: application/json"

# 删除容器相关的Headling日志
curl -X DELETE "$BASE_URL/headling/logs/container/container_123" \
  -H "Accept: application/json"
```

### Headling统计分析
```bash
# 获取Headling统计信息
curl -X GET "$BASE_URL/headling/statistics" \
  -H "Accept: application/json"

# 获取攻击者IP统计
curl -X GET "$BASE_URL/headling/attacker-statistics" \
  -H "Accept: application/json"

# 获取前10个攻击者
curl -X GET "$BASE_URL/headling/top-attackers?limit=10" \
  -H "Accept: application/json"

# 获取前5个攻击者
curl -X GET "$BASE_URL/headling/top-attackers?limit=5" \
  -H "Accept: application/json"

# 获取常用用户名
curl -X GET "$BASE_URL/headling/top-usernames?limit=10" \
  -H "Accept: application/json"

# 获取常用密码
curl -X GET "$BASE_URL/headling/top-passwords?limit=10" \
  -H "Accept: application/json"
```

## 4. Cowrie蜜罐日志管理

### 拉取和查看Cowrie日志
```bash
# 拉取Cowrie蜜罐日志
curl -X POST "$BASE_URL/cowrie/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{
    "container_id": "container_456"
  }'

# 获取所有Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs" \
  -H "Accept: application/json"

# 根据ID获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/1" \
  -H "Accept: application/json"

# 根据容器ID获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/container/container_456" \
  -H "Accept: application/json"

# 根据源IP获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/source-ip/192.168.1.100" \
  -H "Accept: application/json"

# 根据协议获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/protocol/ssh" \
  -H "Accept: application/json"

# 根据命令获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/command/ls" \
  -H "Accept: application/json"

# 根据用户名获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/username/admin" \
  -H "Accept: application/json"

# 根据命令识别状态获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/command-found/true" \
  -H "Accept: application/json"

curl -X GET "$BASE_URL/cowrie/logs/command-found/false" \
  -H "Accept: application/json"

# 根据时间范围获取Cowrie日志
curl -X GET "$BASE_URL/cowrie/logs/time-range?start_time=2025-01-01T00:00:00Z&end_time=2025-01-31T23:59:59Z" \
  -H "Accept: application/json"

# 删除容器相关的Cowrie日志
curl -X DELETE "$BASE_URL/cowrie/logs/container/container_456" \
  -H "Accept: application/json"
```

### Cowrie统计分析
```bash
# 获取Cowrie统计信息
curl -X GET "$BASE_URL/cowrie/statistics" \
  -H "Accept: application/json"

# 获取攻击者行为统计
curl -X GET "$BASE_URL/cowrie/attacker-behavior" \
  -H "Accept: application/json"

# 获取前10个攻击者
curl -X GET "$BASE_URL/cowrie/top-attackers?limit=10" \
  -H "Accept: application/json"

# 获取最常用的命令
curl -X GET "$BASE_URL/cowrie/top-commands?limit=10" \
  -H "Accept: application/json"

# 获取常用用户名
curl -X GET "$BASE_URL/cowrie/top-usernames?limit=10" \
  -H "Accept: application/json"

# 获取常用密码
curl -X GET "$BASE_URL/cowrie/top-passwords?limit=10" \
  -H "Accept: application/json"

# 获取常用客户端指纹
curl -X GET "$BASE_URL/cowrie/top-fingerprints?limit=10" \
  -H "Accept: application/json"
```

## 5. 蜜罐模板和实例管理

### 蜜罐模板管理
```bash
# 获取所有蜜罐模板
curl -X GET "$BASE_URL/honeypot/templates" \
  -H "Accept: application/json"

# 根据ID获取蜜罐模板
curl -X GET "$BASE_URL/honeypot/templates/1" \
  -H "Accept: application/json"

# 创建蜜罐模板
curl -X POST "$BASE_URL/honeypot/templates" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SSH蜜罐模板",
    "type": "SSH",
    "description": "标准SSH蜜罐模板",
    "image_name": "cowrie/cowrie:latest",
    "default_port": 22,
    "config": {
      "hostname": "server01",
      "log_level": "INFO"
    }
  }'

# 更新蜜罐模板
curl -X PUT "$BASE_URL/honeypot/templates/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新的SSH蜜罐模板",
    "description": "更新后的SSH蜜罐模板",
    "config": {
      "hostname": "updated-server",
      "log_level": "DEBUG"
    }
  }'

# 删除蜜罐模板
curl -X DELETE "$BASE_URL/honeypot/templates/1" \
  -H "Accept: application/json"

# 部署蜜罐模板
curl -X POST "$BASE_URL/honeypot/templates/1/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "instance_name": "SSH蜜罐实例1",
    "port": 2222
  }'
```

### 蜜罐实例管理
```bash
# 获取所有蜜罐实例
curl -X GET "$BASE_URL/honeypot/instances" \
  -H "Accept: application/json"

# 根据ID获取蜜罐实例
curl -X GET "$BASE_URL/honeypot/instances/1" \
  -H "Accept: application/json"

# 更新蜜罐实例
curl -X PUT "$BASE_URL/honeypot/instances/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新的蜜罐实例",
    "description": "更新后的描述"
  }'

# 部署蜜罐实例
curl -X POST "$BASE_URL/honeypot/instances/1/deploy" \
  -H "Content-Type: application/json" \
  -d '{}'

# 停止蜜罐实例
curl -X POST "$BASE_URL/honeypot/instances/1/stop" \
  -H "Content-Type: application/json" \
  -d '{}'

# 获取蜜罐实例日志
curl -X GET "$BASE_URL/honeypot/instances/1/logs" \
  -H "Accept: application/json"

# 删除蜜罐实例
curl -X DELETE "$BASE_URL/honeypot/instances/1" \
  -H "Accept: application/json"
```

### 蜜罐日志管理
```bash
# 获取所有蜜罐日志
curl -X GET "$BASE_URL/honeypot/logs" \
  -H "Accept: application/json"

# 根据ID获取蜜罐日志
curl -X GET "$BASE_URL/honeypot/logs/1" \
  -H "Accept: application/json"

# 根据实例ID获取蜜罐日志
curl -X GET "$BASE_URL/honeypot/logs/instance/1" \
  -H "Accept: application/json"
```

## 6. 诱饵(蜜签)管理

### 蜜签管理
```bash
# 获取所有蜜签
curl -X GET "$BASE_URL/baits" \
  -H "Accept: application/json"

# 根据ID获取蜜签
curl -X GET "$BASE_URL/baits/1" \
  -H "Accept: application/json"

# 创建蜜签
curl -X POST "$BASE_URL/baits" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "敏感文件蜜签",
    "type": "file",
    "content": "这是一个诱饵文件",
    "path": "/tmp/sensitive.txt",
    "description": "用于检测文件访问的蜜签"
  }'

# 更新蜜签
curl -X PUT "$BASE_URL/baits/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新的蜜签",
    "description": "更新后的描述"
  }'

# 部署蜜签
curl -X POST "$BASE_URL/baits/1/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "target_path": "/var/log/sensitive.log"
  }'

# 删除蜜签
curl -X DELETE "$BASE_URL/baits/1" \
  -H "Accept: application/json"
```

## 7. 安全规则管理

### 规则管理
```bash
# 获取所有安全规则
curl -X GET "$BASE_URL/rules" \
  -H "Accept: application/json"

# 根据ID获取安全规则
curl -X GET "$BASE_URL/rules/1" \
  -H "Accept: application/json"

# 创建安全规则
curl -X POST "$BASE_URL/rules" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SSH暴力破解检测",
    "type": "detection",
    "condition": "failed_login_attempts > 5",
    "action": "block_ip",
    "description": "检测SSH暴力破解攻击"
  }'

# 更新安全规则
curl -X PUT "$BASE_URL/rules/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新的安全规则",
    "condition": "failed_login_attempts > 3"
  }'

# 启用安全规则
curl -X PUT "$BASE_URL/rules/1/enable" \
  -H "Content-Type: application/json" \
  -d '{}'

# 禁用安全规则
curl -X PUT "$BASE_URL/rules/1/disable" \
  -H "Content-Type: application/json" \
  -d '{}'

# 删除安全规则
curl -X DELETE "$BASE_URL/rules/1" \
  -H "Accept: application/json"
```

### 规则日志管理
```bash
# 获取所有规则日志
curl -X GET "$BASE_URL/rules/logs" \
  -H "Accept: application/json"

# 根据ID获取规则日志
curl -X GET "$BASE_URL/rules/logs/1" \
  -H "Accept: application/json"

# 根据规则ID获取日志
curl -X GET "$BASE_URL/rules/logs/rule/1" \
  -H "Accept: application/json"
```

## 8. AI功能

### 语义分割
```bash
# 日志语义分割
curl -X POST "$BASE_URL/ai/semantic-segment" \
  -H "Content-Type: application/json" \
  -d '{
    "container_id": "container_123",
    "log_content": "2025-01-15 10:30:45 [INFO] SSH connection from 192.168.1.100:45678"
  }'

# 图像语义分割
curl -X POST "$BASE_URL/ai/image-segment" \
  -H "Content-Type: application/json" \
  -d '{
    "image_path": "/path/to/image.jpg",
    "model": "default"
  }'
```

## 9. 容器日志分析

### 日志分析结果管理
```bash
# 获取所有日志分析结果
curl -X GET "$BASE_URL/container-logs/segments" \
  -H "Accept: application/json"

# 根据ID获取分析结果
curl -X GET "$BASE_URL/container-logs/segments/1" \
  -H "Accept: application/json"

# 根据容器ID获取分析结果
curl -X GET "$BASE_URL/container-logs/segments/container/container_123" \
  -H "Accept: application/json"

# 根据类型获取分析结果
curl -X GET "$BASE_URL/container-logs/segments/type/error" \
  -H "Accept: application/json"

# 删除分析结果
curl -X DELETE "$BASE_URL/container-logs/segments/1" \
  -H "Accept: application/json"

# 删除容器相关分析结果
curl -X DELETE "$BASE_URL/container-logs/segments/container/container_123" \
  -H "Accept: application/json"
```

## 10. 数据库操作

### 通用数据操作
```bash
# 查询数据
curl -X GET "$BASE_URL/data?table=honeypot_instance&limit=10" \
  -H "Accept: application/json"

# 创建数据
curl -X POST "$BASE_URL/data" \
  -H "Content-Type: application/json" \
  -d '{
    "table": "honeypot_instance",
    "data": {
      "name": "新蜜罐实例",
      "status": "created"
    }
  }'

# 更新数据
curl -X PUT "$BASE_URL/data" \
  -H "Content-Type: application/json" \
  -d '{
    "table": "honeypot_instance",
    "id": 1,
    "data": {
      "status": "running"
    }
  }'

# 删除数据
curl -X DELETE "$BASE_URL/data?table=honeypot_instance&id=1" \
  -H "Accept: application/json"

# 根据ID获取数据
curl -X GET "$BASE_URL/data/id?table=honeypot_instance&id=1" \
  -H "Accept: application/json"

# 根据名称获取数据
curl -X GET "$BASE_URL/data/name?table=honeypot_instance&name=测试蜜罐" \
  -H "Accept: application/json"
```

## 批量操作示例

### 批量创建容器实例
```bash
#!/bin/bash

# 批量创建多个SSH蜜罐实例
for i in {1..5}; do
  curl -X POST "$BASE_URL/container-instances" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"SSH蜜罐$i\",
      \"honeypot_name\": \"ssh-honeypot-$i\",
      \"image_name\": \"cowrie/cowrie:latest\",
      \"protocol\": \"ssh\",
      \"interface_type\": \"terminal\",
      \"port_mappings\": {
        \"22\": \"$((2220 + i))\"
      },
      \"environment\": {
        \"COWRIE_HOSTNAME\": \"server$i\"
      },
      \"description\": \"SSH蜜罐实例$i\"
    }"
  echo "创建SSH蜜罐$i完成"
  sleep 2
done
```

### 批量启动容器实例
```bash
#!/bin/bash

# 获取所有stopped状态的实例并启动
instances=$(curl -s -X GET "$BASE_URL/container-instances/status/stopped" -H "Accept: application/json")

echo "$instances" | jq -r '.[].id' | while read id; do
  curl -X POST "$BASE_URL/container-instances/$id/start" \
    -H "Content-Type: application/json" \
    -d '{}'
  echo "启动容器实例$id完成"
  sleep 1
done
```

### 批量拉取日志
```bash
#!/bin/bash

# 获取所有running状态的实例并拉取日志
instances=$(curl -s -X GET "$BASE_URL/container-instances/status/running" -H "Accept: application/json")

echo "$instances" | jq -r '.[].container_id' | while read container_id; do
  # 拉取Headling日志
  curl -X POST "$BASE_URL/headling/pull-logs" \
    -H "Content-Type: application/json" \
    -d "{\"container_id\": \"$container_id\"}"

  # 拉取Cowrie日志
  curl -X POST "$BASE_URL/cowrie/pull-logs" \
    -H "Content-Type: application/json" \
    -d "{\"container_id\": \"$container_id\"}"

  echo "拉取容器$container_id日志完成"
  sleep 2
done
```
```
