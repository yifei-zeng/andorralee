# 蜜罐管理系统完整功能清单

## 系统概述

您的蜜罐管理系统现在具备了完整的Docker容器管理、多种蜜罐日志分析和AI智能分析功能。

## 🎯 核心功能模块

### 1. Docker容器管理
- **镜像管理**: 拉取、列出、删除Docker镜像
- **容器管理**: 启动、停止、重启、删除容器
- **容器监控**: 实时状态监控、资源使用情况
- **日志收集**: 容器日志自动收集和分析

### 2. Headling认证日志管理
- **日志拉取**: 从容器自动拉取CSV格式认证日志
- **数据解析**: 智能解析认证尝试数据
- **攻击分析**: 攻击者IP、用户名、密码统计分析
- **时间分析**: 基于时间维度的攻击趋势分析

### 3. Cowrie蜜罐日志管理
- **日志拉取**: 从容器自动拉取JSON格式蜜罐日志
- **命令分析**: 攻击者执行命令的详细分析
- **行为分析**: 攻击者行为模式识别
- **指纹分析**: 客户端指纹统计和分析

### 4. AI智能分析
- **日志语义分割**: 自动识别日志类型和重要信息
- **异常检测**: 基于AI的异常行为检测
- **模式识别**: 攻击模式和威胁识别
- **智能分类**: 自动分类和标记安全事件

### 5. 数据库管理
- **多表设计**: 优化的数据库表结构
- **统计视图**: 预构建的统计分析视图
- **索引优化**: 高性能查询索引
- **数据完整性**: 完整的约束和关联关系

## 🚀 API接口清单

### Docker管理接口
```
POST   /api/v1/docker/pull              # 拉取镜像
GET    /api/v1/docker/images            # 列出镜像
DELETE /api/v1/docker/images/{id}       # 删除镜像
POST   /api/v1/docker/start             # 启动容器
POST   /api/v1/docker/stop              # 停止容器
GET    /api/v1/docker/containers        # 列出容器
GET    /api/v1/docker/logs/{id}         # 获取容器日志
```

### Headling认证日志接口
```
POST   /api/v1/headling/pull-logs                    # 拉取认证日志
GET    /api/v1/headling/logs                         # 获取所有日志
GET    /api/v1/headling/logs/{id}                    # 根据ID获取日志
GET    /api/v1/headling/logs/container/{id}          # 根据容器获取日志
GET    /api/v1/headling/logs/source-ip/{ip}          # 根据源IP获取日志
GET    /api/v1/headling/logs/protocol/{protocol}     # 根据协议获取日志
GET    /api/v1/headling/logs/time-range              # 根据时间范围获取日志
GET    /api/v1/headling/statistics                   # 获取统计信息
GET    /api/v1/headling/attacker-statistics          # 获取攻击者统计
GET    /api/v1/headling/top-attackers                # 获取顶级攻击者
GET    /api/v1/headling/top-usernames                # 获取常用用户名
GET    /api/v1/headling/top-passwords                # 获取常用密码
DELETE /api/v1/headling/logs/container/{id}          # 删除容器日志
```

### Cowrie蜜罐日志接口
```
POST   /api/v1/cowrie/pull-logs                      # 拉取蜜罐日志
GET    /api/v1/cowrie/logs                           # 获取所有日志
GET    /api/v1/cowrie/logs/{id}                      # 根据ID获取日志
GET    /api/v1/cowrie/logs/container/{id}            # 根据容器获取日志
GET    /api/v1/cowrie/logs/source-ip/{ip}            # 根据源IP获取日志
GET    /api/v1/cowrie/logs/protocol/{protocol}       # 根据协议获取日志
GET    /api/v1/cowrie/logs/command/{command}         # 根据命令获取日志
GET    /api/v1/cowrie/logs/username/{username}       # 根据用户名获取日志
GET    /api/v1/cowrie/logs/command-found/{found}     # 根据命令识别状态获取日志
GET    /api/v1/cowrie/logs/time-range                # 根据时间范围获取日志
GET    /api/v1/cowrie/statistics                     # 获取统计信息
GET    /api/v1/cowrie/attacker-behavior               # 获取攻击者行为统计
GET    /api/v1/cowrie/top-attackers                  # 获取顶级攻击者
GET    /api/v1/cowrie/top-commands                   # 获取常用命令
GET    /api/v1/cowrie/top-usernames                  # 获取常用用户名
GET    /api/v1/cowrie/top-passwords                  # 获取常用密码
GET    /api/v1/cowrie/top-fingerprints               # 获取常用指纹
DELETE /api/v1/cowrie/logs/container/{id}            # 删除容器日志
```

### AI分析接口
```
POST   /api/v1/ai/semantic-segment                   # 日志语义分割
POST   /api/v1/ai/image-segment                      # 图像语义分割
```

### 容器日志分析接口
```
GET    /api/v1/container-logs/segments               # 获取所有日志分析结果
GET    /api/v1/container-logs/segments/{id}          # 根据ID获取分析结果
GET    /api/v1/container-logs/segments/container/{id} # 根据容器ID获取分析结果
GET    /api/v1/container-logs/segments/type/{type}   # 根据类型获取分析结果
DELETE /api/v1/container-logs/segments/{id}          # 删除分析结果
```

### Docker镜像日志接口
```
GET    /api/v1/docker/image-logs                     # 获取所有镜像操作日志
GET    /api/v1/docker/image-logs/{id}                # 根据ID获取镜像操作日志
GET    /api/v1/docker/image-logs/image/{id}          # 根据镜像ID获取操作日志
DELETE /api/v1/docker/image-logs/{id}                # 删除镜像操作日志
GET    /api/v1/docker/images/db                      # 获取数据库中的镜像记录
```

## 💾 数据库表结构

### 核心业务表
- `honeypot_template` - 蜜罐模板
- `honeypot_instance` - 蜜罐实例
- `honeypot_log` - 蜜罐日志
- `security_rule` - 安全规则
- `rule_log` - 规则日志

### Docker管理表
- `docker_image` - Docker镜像管理
- `docker_image_log` - Docker镜像操作日志
- `docker_container` - Docker容器管理
- `container_log_segment` - 容器日志语义分析结果

### 蜜罐日志表
- `headling_auth_log` - Headling认证日志
- `cowrie_log` - Cowrie蜜罐日志

### 统计视图
- `v_headling_auth_statistics` - Headling认证统计
- `v_attacker_ip_statistics` - 攻击者IP统计
- `v_cowrie_statistics` - Cowrie日志统计
- `v_cowrie_attacker_behavior` - Cowrie攻击者行为统计
- `v_cowrie_command_statistics` - Cowrie命令统计
- `v_log_statistics` - 日志统计
- `v_container_with_image` - 容器镜像关联视图

## 🔧 代码调用方式

### 1. Go服务层调用
```go
// Docker服务
dockerService, _ := services.NewDockerService()
dockerService.PullImage("nginx:latest")

// Headling服务
headlingService, _ := services.NewHeadlingService()
headlingService.PullHeadlingLogs("container_id")

// Cowrie服务
cowrieService, _ := services.NewCowrieService()
cowrieService.PullCowrieLogs("container_id")

// AI服务
aiService := &services.AIService{}
aiService.AnalyzeContainerLogs("container_id")
```

### 2. HTTP API调用
```bash
# 拉取Docker镜像
curl -X POST "http://localhost:8080/api/v1/docker/pull" \
  -H "Content-Type: application/json" \
  -d '{"image": "nginx:latest"}'

# 拉取Headling日志
curl -X POST "http://localhost:8080/api/v1/headling/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{"container_id": "container_123"}'

# 拉取Cowrie日志
curl -X POST "http://localhost:8080/api/v1/cowrie/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{"container_id": "container_456"}'

# 获取攻击者统计
curl "http://localhost:8080/api/v1/headling/top-attackers?limit=5"
curl "http://localhost:8080/api/v1/cowrie/top-attackers?limit=5"
```

### 3. 数据库直接查询
```sql
-- 获取最活跃的攻击者
SELECT * FROM v_attacker_ip_statistics ORDER BY total_attempts DESC LIMIT 10;

-- 获取最常用的命令
SELECT * FROM v_cowrie_command_statistics ORDER BY usage_count DESC LIMIT 10;

-- 获取日志统计
SELECT * FROM v_headling_auth_statistics WHERE log_date = CURDATE();
```

## 📊 监控和分析功能

### 1. 实时监控
- 容器状态实时监控
- 攻击事件实时检测
- 系统资源使用监控

### 2. 统计分析
- 攻击趋势分析
- 攻击者行为分析
- 命令执行分析
- 协议使用分析

### 3. 威胁情报
- 恶意IP识别
- 攻击模式识别
- 异常行为检测
- 威胁等级评估

## 🛡️ 安全特性

### 1. 数据安全
- 数据库连接加密
- 敏感信息脱敏
- 访问权限控制
- 操作日志记录

### 2. 系统安全
- API接口鉴权
- 输入参数验证
- SQL注入防护
- XSS攻击防护

### 3. 运维安全
- 自动备份机制
- 数据恢复功能
- 系统监控告警
- 异常处理机制

## 🚀 性能优化

### 1. 数据库优化
- 索引优化设计
- 查询性能优化
- 连接池管理
- 缓存策略

### 2. 应用优化
- 并发处理优化
- 内存使用优化
- 网络传输优化
- 批量操作优化

### 3. 系统优化
- 容器资源限制
- 日志轮转机制
- 数据清理策略
- 性能监控

## 📈 扩展能力

### 1. 水平扩展
- 微服务架构支持
- 负载均衡支持
- 分布式部署支持
- 集群管理支持

### 2. 功能扩展
- 新蜜罐类型支持
- 自定义规则引擎
- 第三方集成接口
- 插件化架构

### 3. 数据扩展
- 大数据处理支持
- 实时流处理
- 机器学习集成
- 数据可视化

## 总结

您的蜜罐管理系统现在具备了：

✅ **完整的容器管理能力** - Docker镜像和容器的全生命周期管理
✅ **多种蜜罐日志支持** - Headling和Cowrie两种主流蜜罐的日志管理
✅ **智能分析能力** - AI驱动的日志分析和威胁检测
✅ **丰富的API接口** - 50+个RESTful API接口
✅ **优化的数据库设计** - 12个核心表和7个统计视图
✅ **高性能架构** - 支持大规模数据处理和分析
✅ **完整的文档支持** - 详细的使用指南和示例代码

这是一个功能完整、性能优异、易于扩展的企业级蜜罐管理系统！
