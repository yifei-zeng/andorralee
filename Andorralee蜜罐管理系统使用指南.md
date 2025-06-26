# 🍯 Andorralee 蜜罐管理系统使用指南

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23.4-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-20.10+-blue.svg)](https://docker.com)

## 📋 目录

- [系统概述](#系统概述)
- [快速开始](#快速开始)
- [API接口文档](#api接口文档)
- [容器管理](#容器管理)
- [蜜罐日志分析](#蜜罐日志分析)
- [Docker镜像管理](#docker镜像管理)
- [AI功能](#ai功能)
- [监控和统计](#监控和统计)
- [故障排除](#故障排除)

## 🎯 系统概述

Andorralee是一个现代化的蜜罐管理系统，提供完整的容器化蜜罐部署、管理和分析功能。

### 核心特性

- ✅ **容器化部署** - 基于Docker的蜜罐实例管理
- ✅ **多协议支持** - SSH、HTTP、FTP、Telnet、MySQL等
- ✅ **实时日志分析** - Cowrie和Headling蜜罐日志处理
- ✅ **AI智能分析** - 日志语义分割和攻击行为分析
- ✅ **多数据库支持** - MySQL和达梦数据库
- ✅ **RESTful API** - 完整的REST API接口
- ✅ **银河麒麟兼容** - 专门优化的国产化部署

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面      │    │   API网关       │    │   容器管理      │
│   Web UI        │◄──►│   Gin Router    │◄──►│   Docker API    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   日志分析      │    │   数据存储      │    │   AI分析        │
│   Log Parser    │◄──►│   MySQL/达梦    │◄──►│   Semantic AI   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 快速开始

### 环境要求

- **操作系统**: Linux/Windows/macOS (推荐银河麒麟V10)
- **Go版本**: 1.23.4+
- **Docker**: 20.10+
- **内存**: 4GB+ (推荐8GB)
- **存储**: 20GB+

### 安装部署

#### 方式一：Docker部署 (推荐)
```bash
# 克隆项目
git clone <项目地址>
cd andorralee

# 使用Docker Compose启动
docker-compose up -d

# 银河麒麟系统专用部署
./deploy-kylin.sh
```

#### 方式二：源码编译
```bash
# 编译项目
go mod tidy
go build -o andorralee ./cmd/main.go

# 启动服务
./andorralee
```

### 验证安装

```bash
# 检查服务状态
curl http://localhost:8081/api/v1/health

# 访问调试界面
http://localhost:8081/static/debug-container.html
```

## 📚 API接口文档

### 基础信息

- **基础URL**: `http://localhost:8081/api/v1`
- **内容类型**: `application/json`
- **认证方式**: 暂无 (开发阶段)

### 健康检查接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 简单健康检查 |
| GET | `/api/v1/health` | 详细健康检查 |
| GET | `/api/v1/ready` | 就绪检查 |
| GET | `/api/v1/live` | 存活检查 |

**示例**:
```bash
# 基础健康检查
curl http://localhost:8081/health

# 详细健康检查
curl http://localhost:8081/api/v1/health
```

### 容器实例管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/container-instances` | 创建容器实例 |
| GET | `/container-instances` | 获取所有容器实例 |
| GET | `/container-instances/:id` | 获取指定容器实例 |
| POST | `/container-instances/:id/start` | 启动容器实例 |
| POST | `/container-instances/:id/stop` | 停止容器实例 |
| POST | `/container-instances/:id/restart` | 重启容器实例 |
| DELETE | `/container-instances/:id` | 删除容器实例 |
| GET | `/container-instances/:id/status` | 获取容器状态 |
| POST | `/container-instances/sync` | 同步容器状态 |
| GET | `/container-instances/:id/debug` | 获取调试信息 |

### Docker镜像管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/docker/images` | 获取镜像列表 |
| POST | `/docker/pull` | 拉取镜像 |
| DELETE | `/docker/images/:id` | 删除镜像 |
| GET | `/docker/containers` | 获取容器列表 |
| GET | `/docker/info` | 获取Docker信息 |

### 蜜罐日志接口

#### Headling认证日志
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/headling/pull-logs` | 拉取认证日志 |
| GET | `/headling/logs` | 获取所有日志 |
| GET | `/headling/logs/:id` | 获取指定日志 |
| GET | `/headling/statistics` | 获取统计信息 |
| GET | `/headling/top-attackers` | 获取顶级攻击者 |

#### Cowrie蜜罐日志
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/cowrie/pull-logs` | 拉取Cowrie日志 |
| GET | `/cowrie/logs` | 获取所有日志 |
| GET | `/cowrie/top-commands` | 获取常用命令 |
| GET | `/cowrie/top-attackers` | 获取顶级攻击者 |
| GET | `/cowrie/attacker-behavior` | 获取攻击者行为 |

### AI分析接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/ai/semantic-segment` | 日志语义分割 |
| POST | `/ai/image-segment` | 图像语义分割 |

## 🐳 容器管理

### 创建容器实例

```bash
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
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
    "description": "SSH协议蜜罐实例",
    "auto_start": true
  }'
```

**PowerShell示例**:
```powershell
$body = @{
    name = 'Web蜜罐1'
    honeypot_name = 'web-honeypot-1'
    image_name = 'nginx:latest'
    protocol = 'http'
    interface_type = 'web'
    port_mappings = @{'80' = '8080'}
    environment = @{
        'NGINX_PORT' = '8080'
        'HONEYPOT_TYPE' = 'WEB'
    }
    description = 'Web服务蜜罐实例'
    auto_start = $true
} | ConvertTo-Json -Depth 3

Invoke-WebRequest -Method POST -Uri "http://localhost:8081/api/v1/container-instances" -ContentType "application/json" -Body $body
```

### 容器生命周期管理

```bash
# 启动容器
curl -X POST http://localhost:8081/api/v1/container-instances/1/start

# 停止容器
curl -X POST http://localhost:8081/api/v1/container-instances/1/stop

# 重启容器
curl -X POST http://localhost:8081/api/v1/container-instances/1/restart

# 获取容器状态
curl http://localhost:8081/api/v1/container-instances/1/status

# 同步所有容器状态
curl -X POST http://localhost:8081/api/v1/container-instances/sync
```

### 容器调试

```bash
# 获取容器调试信息
curl http://localhost:8081/api/v1/container-instances/1/debug

# 获取所有容器实例
curl http://localhost:8081/api/v1/container-instances
```

## 📊 蜜罐日志分析

### Headling认证日志

```bash
# 拉取认证日志
curl -X POST http://localhost:8081/api/v1/headling/pull-logs \
  -H "Content-Type: application/json" \
  -d '{"container_id": "your_container_id"}'

# 查看所有认证日志
curl http://localhost:8081/api/v1/headling/logs

# 获取统计信息
curl http://localhost:8081/api/v1/headling/statistics

# 获取顶级攻击者
curl "http://localhost:8081/api/v1/headling/top-attackers?limit=10"

# 根据源IP查询日志
curl "http://localhost:8081/api/v1/headling/logs/source-ip/192.168.1.100"

# 根据时间范围查询
curl "http://localhost:8081/api/v1/headling/logs/time-range?start=2025-01-01&end=2025-01-31"
```

### Cowrie蜜罐日志

```bash
# 拉取Cowrie日志
curl -X POST http://localhost:8081/api/v1/cowrie/pull-logs \
  -H "Content-Type: application/json" \
  -d '{"container_id": "your_cowrie_container_id"}'

# 查看所有日志
curl http://localhost:8081/api/v1/cowrie/logs

# 获取最常用的命令
curl "http://localhost:8081/api/v1/cowrie/top-commands?limit=10"

# 获取最活跃的攻击者
curl "http://localhost:8081/api/v1/cowrie/top-attackers?limit=10"

# 获取攻击者行为统计
curl http://localhost:8081/api/v1/cowrie/attacker-behavior

# 根据命令查询日志
curl "http://localhost:8081/api/v1/cowrie/logs/command/ls"

# 根据用户名查询
curl "http://localhost:8081/api/v1/cowrie/logs/username/root"
```

## 🐋 Docker镜像管理

### 镜像操作

```bash
# 获取镜像列表
curl http://localhost:8081/api/v1/docker/images

# 拉取镜像
curl -X POST http://localhost:8081/api/v1/docker/pull \
  -H "Content-Type: application/json" \
  -d '{"image": "nginx:latest"}'

# 删除镜像
curl -X DELETE http://localhost:8081/api/v1/docker/images/nginx:latest

# 获取Docker信息
curl http://localhost:8081/api/v1/docker/info

# 获取容器列表
curl http://localhost:8081/api/v1/docker/containers
```

**PowerShell示例**:
```powershell
# 拉取常用蜜罐镜像
$images = @("ubuntu:20.04", "nginx:latest", "mysql:8.0", "cowrie/cowrie:latest")

foreach ($image in $images) {
    $body = @{ image = $image } | ConvertTo-Json
    Invoke-WebRequest -Method POST -Uri "http://localhost:8081/api/v1/docker/pull" -ContentType "application/json" -Body $body
    Write-Host "已拉取镜像: $image"
}
```

## 🤖 AI功能

### 日志语义分割

```bash
curl -X POST http://localhost:8081/api/v1/ai/semantic-segment \
  -H "Content-Type: application/json" \
  -d '{
    "container_id": "container_123",
    "log_content": "2025-01-15 10:30:45 [INFO] SSH connection from 192.168.1.100:45678"
  }'
```

### 图像语义分割

```bash
curl -X POST http://localhost:8081/api/v1/ai/image-segment \
  -H "Content-Type: application/json" \
  -d '{
    "image_path": "/path/to/image.jpg",
    "model": "default"
  }'
```

## 📈 监控和统计

### 系统状态监控

```bash
# 同步所有容器状态
curl -X POST http://localhost:8081/api/v1/container-instances/sync

# 获取系统健康状态
curl http://localhost:8081/api/v1/health

# 获取Docker系统信息
curl http://localhost:8081/api/v1/docker/info
```

### 攻击统计分析

```bash
# Headling攻击统计
curl http://localhost:8081/api/v1/headling/statistics
curl http://localhost:8081/api/v1/headling/attacker-statistics

# Cowrie攻击统计
curl http://localhost:8081/api/v1/cowrie/attacker-behavior
curl http://localhost:8081/api/v1/cowrie/top-attackers
```

## 🔧 故障排除

### 常见问题

#### 1. 容器启动失败
```bash
# 检查容器状态
curl http://localhost:8081/api/v1/container-instances/1/debug

# 同步容器状态
curl -X POST http://localhost:8081/api/v1/container-instances/sync

# 查看Docker服务状态
curl http://localhost:8081/api/v1/docker/info
```

#### 2. API无响应
```bash
# 检查服务健康状态
curl http://localhost:8081/health

# 检查详细健康信息
curl http://localhost:8081/api/v1/health
```

#### 3. 数据库连接问题
```bash
# 检查数据库连接状态
curl http://localhost:8081/api/v1/health | jq '.services'
```

### 调试工具

#### Web调试界面
访问: `http://localhost:8081/static/debug-container.html`

功能包括：
- 容器实例列表查看
- 容器启动/停止/重启操作
- 状态同步功能
- 详细调试信息查看
- 新容器创建测试

#### PowerShell测试脚本
```powershell
# 运行API测试脚本
.\test-container-api.ps1 -Verbose

# 自定义服务器测试
.\test-container-api.ps1 -BaseUrl "http://192.168.1.100:8081"
```

#### Linux修复脚本
```bash
# 运行问题诊断和修复
chmod +x fix-container-issues.sh
./fix-container-issues.sh

# 执行所有修复操作
./fix-container-issues.sh all
```

## 📞 技术支持

### 联系方式
- 📧 邮箱: support@andorralee.com
- 🐛 问题反馈: GitHub Issues
- 💬 技术交流: QQ群 123456789

### 相关文档
- [银河麒麟Docker部署指南](银河麒麟Docker部署指南.md)
- [快速开始指南](docs/quick_start_guide.md)
- [PowerShell操作指南](docs/powershell_operations_guide.md)
- [Cowrie实现指南](docs/cowrie_implementation_guide.md)

## 🎯 高级用法

### 批量容器管理

#### 批量创建蜜罐实例
```bash
#!/bin/bash
# 批量创建不同类型的蜜罐

# SSH蜜罐
for i in {1..3}; do
  curl -X POST http://localhost:8081/api/v1/container-instances \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"SSH蜜罐${i}\",
      \"honeypot_name\": \"ssh-honeypot-${i}\",
      \"image_name\": \"ubuntu:20.04\",
      \"protocol\": \"ssh\",
      \"port_mappings\": {\"22\": \"$((2220+i))\"},
      \"auto_start\": true
    }"
done

# Web蜜罐
for i in {1..2}; do
  curl -X POST http://localhost:8081/api/v1/container-instances \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"Web蜜罐${i}\",
      \"honeypot_name\": \"web-honeypot-${i}\",
      \"image_name\": \"nginx:latest\",
      \"protocol\": \"http\",
      \"port_mappings\": {\"80\": \"$((8080+i))\"},
      \"auto_start\": true
    }"
done
```

#### PowerShell批量管理
```powershell
# 批量启动所有停止的容器
function Start-AllStoppedContainers {
    $instances = (Invoke-WebRequest -Uri "http://localhost:8081/api/v1/container-instances").Content | ConvertFrom-Json

    foreach ($instance in $instances.data) {
        if ($instance.status -eq "stopped" -or $instance.status -eq "created") {
            Write-Host "启动容器: $($instance.name)"
            Invoke-WebRequest -Method POST -Uri "http://localhost:8081/api/v1/container-instances/$($instance.id)/start"
        }
    }
}

# 批量停止所有运行的容器
function Stop-AllRunningContainers {
    $instances = (Invoke-WebRequest -Uri "http://localhost:8081/api/v1/container-instances").Content | ConvertFrom-Json

    foreach ($instance in $instances.data) {
        if ($instance.status -eq "running") {
            Write-Host "停止容器: $($instance.name)"
            Invoke-WebRequest -Method POST -Uri "http://localhost:8081/api/v1/container-instances/$($instance.id)/stop"
        }
    }
}

# 获取容器统计信息
function Get-ContainerStatistics {
    $instances = (Invoke-WebRequest -Uri "http://localhost:8081/api/v1/container-instances").Content | ConvertFrom-Json

    $stats = @{
        Total = $instances.data.Count
        Running = ($instances.data | Where-Object {$_.status -eq "running"}).Count
        Stopped = ($instances.data | Where-Object {$_.status -eq "stopped"}).Count
        Created = ($instances.data | Where-Object {$_.status -eq "created"}).Count
    }

    return $stats
}
```

### 日志分析自动化

#### 定时日志收集
```bash
#!/bin/bash
# 定时收集所有容器的日志

# 获取所有容器实例
containers=$(curl -s http://localhost:8081/api/v1/container-instances | jq -r '.data[].id')

for container_id in $containers; do
    echo "收集容器 $container_id 的日志..."

    # 拉取Headling日志
    curl -X POST http://localhost:8081/api/v1/headling/pull-logs \
      -H "Content-Type: application/json" \
      -d "{\"container_id\": \"$container_id\"}"

    # 拉取Cowrie日志
    curl -X POST http://localhost:8081/api/v1/cowrie/pull-logs \
      -H "Content-Type: application/json" \
      -d "{\"container_id\": \"$container_id\"}"
done

echo "日志收集完成"
```

#### 攻击分析报告生成
```bash
#!/bin/bash
# 生成攻击分析报告

REPORT_FILE="attack_report_$(date +%Y%m%d_%H%M%S).json"

echo "生成攻击分析报告..."

# 收集Headling统计信息
echo "收集Headling统计信息..."
curl -s http://localhost:8081/api/v1/headling/statistics > headling_stats.json
curl -s http://localhost:8081/api/v1/headling/top-attackers?limit=20 > headling_attackers.json

# 收集Cowrie统计信息
echo "收集Cowrie统计信息..."
curl -s http://localhost:8081/api/v1/cowrie/attacker-behavior > cowrie_behavior.json
curl -s http://localhost:8081/api/v1/cowrie/top-commands?limit=50 > cowrie_commands.json
curl -s http://localhost:8081/api/v1/cowrie/top-attackers?limit=20 > cowrie_attackers.json

# 合并报告
jq -n \
  --argjson headling_stats "$(cat headling_stats.json)" \
  --argjson headling_attackers "$(cat headling_attackers.json)" \
  --argjson cowrie_behavior "$(cat cowrie_behavior.json)" \
  --argjson cowrie_commands "$(cat cowrie_commands.json)" \
  --argjson cowrie_attackers "$(cat cowrie_attackers.json)" \
  '{
    "report_time": now | strftime("%Y-%m-%d %H:%M:%S"),
    "headling": {
      "statistics": $headling_stats,
      "top_attackers": $headling_attackers
    },
    "cowrie": {
      "behavior": $cowrie_behavior,
      "top_commands": $cowrie_commands,
      "top_attackers": $cowrie_attackers
    }
  }' > $REPORT_FILE

echo "报告已生成: $REPORT_FILE"

# 清理临时文件
rm -f headling_stats.json headling_attackers.json cowrie_behavior.json cowrie_commands.json cowrie_attackers.json
```

### 监控和告警

#### 系统监控脚本
```bash
#!/bin/bash
# 系统监控脚本

ALERT_EMAIL="admin@example.com"
LOG_FILE="/var/log/andorralee_monitor.log"

log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOG_FILE
}

# 检查API服务状态
check_api_health() {
    if ! curl -f http://localhost:8081/api/v1/health > /dev/null 2>&1; then
        log_message "ERROR: API服务异常"
        echo "Andorralee API服务异常，请立即检查" | mail -s "蜜罐系统告警" $ALERT_EMAIL
        return 1
    fi
    return 0
}

# 检查容器状态
check_container_status() {
    local response=$(curl -s http://localhost:8081/api/v1/container-instances)
    local total=$(echo $response | jq '.data | length')
    local running=$(echo $response | jq '.data | map(select(.status == "running")) | length')
    local failed=$(echo $response | jq '.data | map(select(.status == "exited" or .status == "dead")) | length')

    log_message "INFO: 容器状态 - 总数:$total, 运行:$running, 异常:$failed"

    if [ $failed -gt 0 ]; then
        log_message "WARN: 发现 $failed 个异常容器"
        # 尝试重启异常容器
        curl -X POST http://localhost:8081/api/v1/container-instances/sync
    fi
}

# 检查磁盘空间
check_disk_space() {
    local usage=$(df / | awk 'NR==2{print $5}' | sed 's/%//')
    if [ $usage -gt 80 ]; then
        log_message "WARN: 磁盘使用率过高: ${usage}%"
        echo "磁盘使用率过高: ${usage}%" | mail -s "磁盘空间告警" $ALERT_EMAIL
    fi
}

# 主监控循环
main() {
    log_message "INFO: 开始系统监控"

    check_api_health
    check_container_status
    check_disk_space

    log_message "INFO: 监控检查完成"
}

# 执行监控
main
```

## 📝 API响应格式

### 标准响应格式

#### 成功响应
```json
{
  "status": "success",
  "message": "操作成功",
  "data": {
    // 具体数据内容
  },
  "timestamp": "2025-01-15T10:30:45Z"
}
```

#### 错误响应
```json
{
  "status": "error",
  "message": "错误描述",
  "error_code": "ERROR_CODE",
  "timestamp": "2025-01-15T10:30:45Z"
}
```

### 容器实例响应示例

#### 创建容器实例响应
```json
{
  "status": "success",
  "message": "容器实例创建成功",
  "data": {
    "id": 1,
    "name": "SSH蜜罐1",
    "honeypot_name": "ssh-honeypot-1",
    "container_id": "abc123def456",
    "container_name": "ssh-honeypot-1",
    "image_name": "ubuntu:20.04",
    "protocol": "ssh",
    "status": "created",
    "port": 2222,
    "honeypot_ip": "172.17.0.2",
    "create_time": "2025-01-15T10:30:45Z",
    "update_time": "2025-01-15T10:30:45Z"
  }
}
```

#### 容器列表响应
```json
{
  "status": "success",
  "message": "获取容器实例列表成功",
  "data": [
    {
      "id": 1,
      "name": "SSH蜜罐1",
      "status": "running",
      "protocol": "ssh",
      "port": 2222,
      "create_time": "2025-01-15T10:30:45Z"
    },
    {
      "id": 2,
      "name": "Web蜜罐1",
      "status": "stopped",
      "protocol": "http",
      "port": 8080,
      "create_time": "2025-01-15T10:35:20Z"
    }
  ]
}
```

### 日志分析响应示例

#### Headling统计响应
```json
{
  "status": "success",
  "data": {
    "total_logs": 1250,
    "unique_attackers": 45,
    "successful_logins": 23,
    "failed_logins": 1227,
    "top_protocols": [
      {"protocol": "ssh", "count": 800},
      {"protocol": "ftp", "count": 300},
      {"protocol": "telnet", "count": 150}
    ],
    "attack_timeline": [
      {"date": "2025-01-15", "count": 156},
      {"date": "2025-01-14", "count": 203}
    ]
  }
}
```

#### Cowrie行为分析响应
```json
{
  "status": "success",
  "data": {
    "total_sessions": 890,
    "unique_attackers": 67,
    "total_commands": 2340,
    "unique_commands": 156,
    "average_session_duration": 245,
    "top_commands": [
      {"command": "ls", "count": 234},
      {"command": "pwd", "count": 189},
      {"command": "whoami", "count": 167}
    ],
    "attacker_countries": [
      {"country": "CN", "count": 234},
      {"country": "US", "count": 123},
      {"country": "RU", "count": 89}
    ]
  }
}
```

---

**版本**: v1.2.0
**更新时间**: 2025-01-15
**适用系统**: Linux/Windows/macOS/银河麒麟V10
