# 蜜罐系统日志记录功能验证报告

## 📋 项目需求验证

根据项目需求书要求：**系统需对交互过程中的关键事件进行详细日志记录，包括时间戳、IP信息、协议类型、用户行为等，日志格式规范，支持导出与长期存储。**

### 考核指标验证

#### (1) ✅ 启动蜜罐系统并进行基础部署
- 系统已成功启动，监听端口8081
- 数据库连接正常（MySQL + 达梦数据库）
- 所有API接口正常工作

#### (2) ✅ 从攻击端执行若干正常/异常交互
- 已模拟认证尝试（成功/失败）
- 已模拟命令输入和响应
- 已模拟连接建立和断开

#### (3) ✅ 查看生成的日志文件，验证记录信息

**✅ 连接时间与IP**
- Headling日志: `timestamp` 字段记录精确时间戳
- Cowrie日志: `event_time` 字段记录微秒精度时间戳
- 会话管理: `start_time`, `end_time` 记录会话生命周期
- IP信息: `source_ip`, `destination_ip` 字段

**✅ 所使用的协议与端口**
- 协议类型: `protocol` 字段（支持http/ssh/telnet/ftp/smb/other）
- 端口信息: `source_port`, `destination_port` 字段
- 网络层信息完整记录

**✅ 用户名与密码**
- 用户名: `username` 字段
- 密码: `password` 字段（明文）
- 密码哈希: `password_hash` 字段（加密存储）
- 认证结果: 会话事件中的 `success` 字段

**✅ 输入命令与响应**
- 命令内容: `command` 字段（支持完整命令行）
- 命令识别: `command_found` 字段（布尔值）
- 命令响应: 会话事件中的 `response` 字段
- 执行结果: 会话事件中的 `success` 字段

**✅ 会话关闭时间与持续时长**
- 会话开始: `start_time` 字段
- 会话结束: `end_time` 字段
- 持续时长: `duration_seconds` 字段（秒）
- 格式化显示: `duration_formatted`（如"29秒"）

## 🗄️ 数据库表结构

### 核心日志表

#### 1. `headling_auth_log` - Headling认证日志
```sql
- id: 主键
- timestamp: 认证时间戳（微秒精度）
- auth_id: 认证行为唯一ID
- session_id: 会话ID
- source_ip: 攻击者IP
- source_port: 攻击者端口
- destination_ip: 目标IP
- destination_port: 目标端口
- protocol: 协议类型
- username: 用户名
- password: 密码
- password_hash: 密码哈希
- container_id: 容器ID
- container_name: 容器名称
```

#### 2. `cowrie_log` - Cowrie蜜罐日志
```sql
- id: 主键
- event_time: 事件时间戳（微秒精度）
- auth_id: 认证行为唯一ID
- session_id: 会话ID
- source_ip: 攻击者IP
- source_port: 攻击者端口
- destination_ip: 目标IP
- destination_port: 目标端口
- protocol: 协议类型
- client_info: 客户端信息
- fingerprint: 客户端指纹
- username: 用户名
- password: 密码
- password_hash: 密码哈希
- command: 执行的命令
- command_found: 命令是否被识别
- raw_log: 原始日志内容
- container_id: 容器ID
- container_name: 容器名称
```

#### 3. `honeypot_session` - 蜜罐会话管理（新增）
```sql
- id: 主键
- session_id: 会话唯一ID
- source_ip: 攻击者IP
- source_port: 攻击者端口
- destination_ip: 目标IP
- destination_port: 目标端口
- protocol: 协议类型
- container_id: 容器ID
- container_name: 容器名称
- start_time: 会话开始时间
- end_time: 会话结束时间
- duration_seconds: 会话持续时长（秒）
- status: 会话状态（active/closed/timeout）
- event_count: 事件数量
- auth_attempts: 认证尝试次数
- command_count: 命令执行次数
- last_activity: 最后活动时间
- user_agent: 用户代理
- client_info: 客户端信息
- fingerprint: 客户端指纹
```

#### 4. `session_event` - 会话事件记录（新增）
```sql
- id: 主键
- session_id: 会话ID
- event_type: 事件类型（connect/auth/command/disconnect/error）
- event_time: 事件时间戳（微秒精度）
- username: 用户名
- password: 密码
- command: 命令内容
- response: 响应内容
- success: 操作是否成功
- error_msg: 错误信息
- details: 详细信息
```

## 🔧 API接口

### 日志查询接口

#### Headling日志
- `GET /api/v1/headling/logs` - 获取所有认证日志
- `GET /api/v1/headling/logs/source-ip/{ip}` - 根据源IP获取日志
- `GET /api/v1/headling/logs/protocol/{protocol}` - 根据协议获取日志
- `GET /api/v1/headling/logs/time-range` - 根据时间范围获取日志

#### Cowrie日志
- `GET /api/v1/cowrie/logs` - 获取所有蜜罐日志
- `GET /api/v1/cowrie/logs/source-ip/{ip}` - 根据源IP获取日志
- `GET /api/v1/cowrie/logs/command/{command}` - 根据命令获取日志
- `GET /api/v1/cowrie/logs/username/{username}` - 根据用户名获取日志

#### 会话管理接口（新增）
- `GET /api/v1/sessions/{id}` - 获取会话基本信息
- `GET /api/v1/sessions/{id}/details` - 获取会话详细信息
- `GET /api/v1/sessions/{id}/events` - 获取会话事件
- `POST /api/v1/sessions/{id}/close` - 关闭会话
- `GET /api/v1/sessions/statistics` - 获取会话统计
- `POST /api/v1/sessions/auth` - 记录认证尝试
- `POST /api/v1/sessions/command` - 记录命令执行

### 日志导出接口
- `POST /api/v1/logs/export` - 导出日志（支持CSV/JSON格式）

## 📊 实际测试结果

### 测试场景
1. **认证尝试记录**
   - 用户名: admin
   - 密码: 123456
   - 结果: 失败
   - ✅ 成功记录到数据库

2. **命令执行记录**
   - 命令: `ls -la`
   - 响应: `total 8\ndrwxr-xr-x 2 root root 4096 Jan 1 12:00 .`
   - 结果: 成功
   - ✅ 成功记录到数据库

3. **会话生命周期**
   - 会话ID: d8ad1350-e8a1-4e71-b674-e33b26930605
   - 开始时间: 2025-06-29T21:23:09
   - 结束时间: 2025-06-29T21:23:38
   - 持续时长: 29秒
   - ✅ 完整记录会话生命周期

### 统计信息
- 总会话数: 1
- 活跃会话数: 0
- 已关闭会话数: 1
- 平均持续时长: 29秒
- 事件总数: 3（1个认证 + 1个命令 + 1个断开）

## 🎯 功能完整性验证

### ✅ 已实现的核心功能
1. **完整的时间戳记录** - 微秒精度，支持时区
2. **网络信息记录** - IP地址、端口、协议完整记录
3. **认证信息记录** - 用户名、密码、哈希值、认证结果
4. **命令交互记录** - 命令内容、响应、执行结果
5. **会话生命周期管理** - 开始、结束、持续时长
6. **事件详细记录** - 连接、认证、命令、断开等事件
7. **日志导出功能** - 支持CSV/JSON格式导出
8. **统计分析功能** - 会话统计、攻击者行为分析

### 🔍 日志格式规范
- **时间格式**: RFC3339标准（2025-06-29T21:23:09.0666076+08:00）
- **数据类型**: 严格的数据类型定义（字符串、整数、布尔值、时间戳）
- **字段完整性**: 所有必要字段都有明确定义和约束
- **索引优化**: 关键查询字段都建立了数据库索引

### 📈 长期存储支持
- **数据库持久化**: MySQL + 达梦数据库双重保障
- **批量操作**: 支持批量插入和查询优化
- **数据清理**: 支持按容器、时间范围删除历史数据
- **备份恢复**: 标准SQL格式，支持数据库备份恢复

## ✅ 结论

**蜜罐系统的日志记录功能已完全满足项目需求书的所有要求：**

1. ✅ **连接时间与IP** - 完整记录
2. ✅ **协议与端口** - 完整记录  
3. ✅ **用户名与密码** - 完整记录
4. ✅ **命令与响应** - 完整记录
5. ✅ **会话持续时长** - 完整记录

系统不仅满足了基本的日志记录需求，还提供了：
- 实时会话管理
- 详细的事件追踪
- 灵活的查询接口
- 标准化的导出功能
- 完整的统计分析

**项目考核指标全部通过！** 🎉
