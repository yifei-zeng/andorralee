# Headling认证日志管理功能使用指南

## 概述

Headling认证日志管理功能是专门为处理headling蜜罐镜像产生的认证日志而设计的。该功能可以自动从容器中拉取CSV格式的认证日志，解析并存储到数据库中，同时提供丰富的查询和统计分析接口。

## 功能特性

### 1. 日志拉取和解析
- 自动从headling容器中拉取 `/log_auth.csv` 文件
- 智能解析CSV格式的认证日志
- 自动去重，避免重复记录
- 支持批量导入和实时拉取

### 2. 数据存储
- 完整保存认证尝试的所有信息
- 支持微秒级时间戳精度
- 自动关联容器信息
- 提供数据完整性保证

### 3. 查询和分析
- 多维度查询（容器、IP、协议、时间等）
- 攻击者行为分析
- 常用凭据统计
- 实时统计报表

## 数据库表结构

### headling_auth_log 表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | BIGINT UNSIGNED | 主键ID |
| timestamp | DATETIME(6) | 认证行为时间戳（微秒精度） |
| auth_id | VARCHAR(36) | 认证行为唯一ID |
| session_id | VARCHAR(36) | 会话ID |
| source_ip | VARCHAR(45) | 攻击者IP |
| source_port | INT UNSIGNED | 攻击者端口 |
| destination_ip | VARCHAR(45) | 目标IP |
| destination_port | INT UNSIGNED | 目标端口 |
| protocol | VARCHAR(20) | 协议类型 |
| username | VARCHAR(255) | 用户名 |
| password | VARCHAR(255) | 密码 |
| password_hash | VARCHAR(255) | 密码哈希（可选） |
| container_id | VARCHAR(64) | 容器ID |
| container_name | VARCHAR(100) | 容器名称 |
| created_at | DATETIME(3) | 记录创建时间 |

## API接口说明

### 基础URL
```
http://localhost:8080/api/v1/headling
```

### 1. 日志拉取接口

#### 拉取容器认证日志
```http
POST /pull-logs
Content-Type: application/json

{
    "container_id": "container_id_here"
}
```

**响应示例：**
```json
{
    "code": 200,
    "message": "success",
    "data": "headling认证日志拉取成功"
}
```

### 2. 日志查询接口

#### 获取所有认证日志
```http
GET /logs
```

#### 根据容器ID获取日志
```http
GET /logs/container/{container_id}
```

#### 根据源IP获取日志
```http
GET /logs/source-ip/{source_ip}
```

#### 根据协议获取日志
```http
GET /logs/protocol/{protocol}
```

#### 根据时间范围获取日志
```http
GET /logs/time-range?start_time=2025-01-01T00:00:00Z&end_time=2025-01-02T00:00:00Z
```

### 3. 统计分析接口

#### 获取基础统计信息
```http
GET /statistics
```

**响应示例：**
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "log_date": "2025-01-01",
            "protocol": "http",
            "total_attempts": 150,
            "unique_ips": 25,
            "unique_usernames": 45,
            "unique_sessions": 80,
            "first_attempt": "2025-01-01T08:30:00Z",
            "last_attempt": "2025-01-01T23:45:00Z"
        }
    ]
}
```

#### 获取攻击者IP统计
```http
GET /attacker-statistics
```

#### 获取顶级攻击者
```http
GET /top-attackers?limit=10
```

#### 获取常用用户名
```http
GET /top-usernames?limit=10
```

#### 获取常用密码
```http
GET /top-passwords?limit=10
```

## 使用示例

### 1. 部署headling容器并拉取日志

```bash
# 1. 启动headling容器
curl -X POST "http://localhost:8080/api/v1/docker/start" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "headling:latest",
    "container_name": "headling-honeypot-1",
    "ports": {"80": "8080"},
    "environment": {}
  }'

# 2. 等待一段时间让容器产生日志

# 3. 拉取认证日志
curl -X POST "http://localhost:8080/api/v1/headling/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{
    "container_id": "container_id_from_step1"
  }'
```

### 2. 查询特定攻击者的行为

```bash
# 查询特定IP的所有认证尝试
curl "http://localhost:8080/api/v1/headling/logs/source-ip/192.168.1.100"

# 查询HTTP协议的认证尝试
curl "http://localhost:8080/api/v1/headling/logs/protocol/http"
```

### 3. 获取安全分析报告

```bash
# 获取顶级攻击者
curl "http://localhost:8080/api/v1/headling/top-attackers?limit=5"

# 获取最常用的用户名
curl "http://localhost:8080/api/v1/headling/top-usernames?limit=10"

# 获取攻击者详细统计
curl "http://localhost:8080/api/v1/headling/attacker-statistics"
```

## 最佳实践

### 1. 定期日志拉取
建议设置定时任务，定期从容器中拉取日志：

```bash
# 创建定时脚本
#!/bin/bash
CONTAINERS=$(curl -s "http://localhost:8080/api/v1/docker/containers" | jq -r '.data[].ID')

for container in $CONTAINERS; do
    curl -X POST "http://localhost:8080/api/v1/headling/pull-logs" \
      -H "Content-Type: application/json" \
      -d "{\"container_id\": \"$container\"}"
done
```

### 2. 数据清理
定期清理历史数据，避免数据库过大：

```sql
-- 删除30天前的认证日志
DELETE FROM headling_auth_log 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
```

### 3. 监控告警
基于统计数据设置监控告警：

```bash
# 检查是否有异常高频攻击
ATTACK_COUNT=$(curl -s "http://localhost:8080/api/v1/headling/top-attackers?limit=1" | jq '.data[0].total_attempts')

if [ "$ATTACK_COUNT" -gt 1000 ]; then
    echo "警告：检测到高频攻击，攻击次数：$ATTACK_COUNT"
fi
```

## 故障排除

### 1. 日志拉取失败
- 检查容器是否正在运行
- 确认容器中存在 `/log_auth.csv` 文件
- 检查Docker服务是否正常

### 2. 数据解析错误
- 检查CSV文件格式是否正确
- 确认时间戳格式符合预期
- 查看应用程序日志获取详细错误信息

### 3. 查询性能问题
- 为常用查询字段添加索引
- 考虑数据分区策略
- 定期清理历史数据

## 扩展功能

### 1. 自定义日志格式
如果您的headling镜像使用不同的日志格式，可以修改 `parseCSVLogs` 函数来适配。

### 2. 实时监控
可以结合WebSocket技术实现实时日志监控和告警。

### 3. 数据可视化
建议结合前端图表库（如ECharts、D3.js）实现数据可视化展示。
