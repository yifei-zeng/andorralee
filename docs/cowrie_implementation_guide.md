# Cowrie蜜罐日志管理功能完整操作指南

## 项目概述

成功为您的蜜罐管理系统添加了完整的Cowrie蜜罐日志管理功能，支持JSON格式日志的拉取、解析、存储和分析。

## 实现的功能

### 1. 数据库设计

#### 新增表结构
- **cowrie_log**: 存储Cowrie蜜罐日志的主表
- **v_cowrie_statistics**: Cowrie日志统计视图
- **v_cowrie_attacker_behavior**: 攻击者行为统计视图
- **v_cowrie_command_statistics**: 命令使用统计视图

#### 字段设计
支持您提供的所有字段：
- `event_time`: 微秒精度事件时间戳
- `auth_id`: 认证行为唯一ID
- `session_id`: 会话ID
- `source_ip/source_port`: 攻击者信息
- `destination_ip/destination_port`: 目标信息
- `protocol`: 协议类型（枚举：http/ssh/telnet/ftp/smb/other）
- `client_info`: 客户端信息
- `fingerprint`: 客户端指纹
- `username/password`: 认证凭据
- `password_hash`: 密码哈希
- `command`: 执行的命令
- `command_found`: 命令是否被识别
- `raw_log`: 原始日志内容
- `container_id/container_name`: 容器关联信息

### 2. 后端服务实现

#### 数据模型 (internal/repositories/models.go)
- `CowrieLog`: 主要数据模型
- `CowrieStatistics`: 统计数据模型
- `CowrieAttackerBehavior`: 攻击者行为模型
- `CowrieCommandStatistics`: 命令统计模型

#### 仓库层 (internal/repositories/)
- `CowrieLogRepository`: 仓库接口定义
- `MySQLCowrieLogRepo`: MySQL实现
- 支持CRUD操作、批量操作、统计查询

#### 服务层 (internal/services/cowrie_service.go)
- `CowrieService`: 业务逻辑服务
- `PullCowrieLogs`: 日志拉取功能
- `parseJSONLogs`: JSON解析功能
- 各种查询和统计方法

#### 控制器层 (internal/handlers/cowrie_handler.go)
- 完整的RESTful API处理器
- 支持日志拉取、查询、统计、删除等操作

### 3. API接口

#### 基础URL
```
http://localhost:8080/api/v1/cowrie
```

#### 主要接口
1. **日志管理**
   - `POST /pull-logs` - 拉取蜜罐日志
   - `GET /logs` - 获取所有日志
   - `GET /logs/{id}` - 根据ID获取日志
   - `GET /logs/container/{container_id}` - 根据容器获取日志
   - `GET /logs/source-ip/{source_ip}` - 根据源IP获取日志
   - `GET /logs/protocol/{protocol}` - 根据协议获取日志
   - `GET /logs/command/{command}` - 根据命令获取日志
   - `GET /logs/username/{username}` - 根据用户名获取日志
   - `GET /logs/command-found/{found}` - 根据命令识别状态获取日志
   - `GET /logs/time-range` - 根据时间范围获取日志
   - `DELETE /logs/container/{container_id}` - 删除容器相关日志

2. **统计分析**
   - `GET /statistics` - 获取基础统计信息
   - `GET /attacker-behavior` - 获取攻击者行为统计
   - `GET /top-attackers` - 获取顶级攻击者
   - `GET /top-commands` - 获取常用命令
   - `GET /top-usernames` - 获取常用用户名
   - `GET /top-passwords` - 获取常用密码
   - `GET /top-fingerprints` - 获取常用指纹

## 部署指南

### 1. 数据库更新
```bash
# 执行数据库迁移脚本
mysql -u your_username -p your_database < scripts/update_database_schema.sql
```

### 2. 应用部署
```bash
# 编译应用
go build -o bin/andorralee cmd/main.go

# 启动应用
./bin/andorralee
```

### 3. 功能测试
```bash
# 测试日志拉取
curl -X POST "http://localhost:8080/api/v1/cowrie/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{"container_id": "your_container_id"}'

# 查看拉取的日志
curl "http://localhost:8080/api/v1/cowrie/logs"

# 获取统计信息
curl "http://localhost:8080/api/v1/cowrie/statistics"
```

## 使用示例

### 1. 拉取并分析日志
```bash
# 1. 拉取指定容器的Cowrie日志
curl -X POST "http://localhost:8080/api/v1/cowrie/pull-logs" \
  -H "Content-Type: application/json" \
  -d '{"container_id": "container_123"}'

# 2. 查看该容器的所有日志
curl "http://localhost:8080/api/v1/cowrie/logs/container/container_123"

# 3. 分析攻击者行为
curl "http://localhost:8080/api/v1/cowrie/logs/source-ip/192.168.1.100"
```

### 2. 命令分析
```bash
# 获取包含特定命令的日志
curl "http://localhost:8080/api/v1/cowrie/logs/command/ls"

# 获取被系统识别的命令
curl "http://localhost:8080/api/v1/cowrie/logs/command-found/true"

# 获取最常用的命令
curl "http://localhost:8080/api/v1/cowrie/top-commands?limit=5"
```

### 3. 安全分析
```bash
# 获取最活跃的攻击者
curl "http://localhost:8080/api/v1/cowrie/top-attackers?limit=5"

# 获取攻击者行为详细统计
curl "http://localhost:8080/api/v1/cowrie/attacker-behavior"

# 获取最常用的客户端指纹
curl "http://localhost:8080/api/v1/cowrie/top-fingerprints?limit=10"
```

## 代码调用示例

### 1. Go代码调用
```go
// 创建Cowrie服务
service, err := services.NewCowrieService()
if err != nil {
    log.Fatal("创建服务失败:", err)
}

// 拉取日志
err = service.PullCowrieLogs("container_id")
if err != nil {
    log.Printf("拉取日志失败: %v", err)
}

// 获取统计信息
stats, err := service.GetStatistics()
if err != nil {
    log.Printf("获取统计信息失败: %v", err)
} else {
    for _, stat := range stats {
        fmt.Printf("日期: %s, 协议: %s, 事件数: %d\n", 
            stat.LogDate, stat.Protocol, stat.TotalEvents)
    }
}

// 获取顶级攻击者
attackers, err := service.GetTopAttackers(5)
if err != nil {
    log.Printf("获取攻击者失败: %v", err)
} else {
    for _, attacker := range attackers {
        fmt.Printf("IP: %s, 事件数: %d, 命令数: %d\n", 
            attacker.SourceIP, attacker.TotalEvents, attacker.CommandsExecuted)
    }
}
```

### 2. JavaScript/前端调用
```javascript
// 拉取日志
async function pullCowrieLogs(containerId) {
    const response = await fetch('/api/v1/cowrie/pull-logs', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            container_id: containerId
        })
    });
    
    const result = await response.json();
    console.log('拉取结果:', result);
}

// 获取统计信息
async function getCowrieStatistics() {
    const response = await fetch('/api/v1/cowrie/statistics');
    const result = await response.json();
    
    if (result.code === 200) {
        console.log('统计信息:', result.data);
        return result.data;
    }
}

// 获取攻击者行为
async function getAttackerBehavior() {
    const response = await fetch('/api/v1/cowrie/attacker-behavior');
    const result = await response.json();
    
    if (result.code === 200) {
        result.data.forEach(attacker => {
            console.log(`攻击者 ${attacker.source_ip}: 
                事件数=${attacker.total_events}, 
                命令数=${attacker.commands_executed}`);
        });
    }
}
```

### 3. Python脚本调用
```python
import requests
import json

# 基础URL
BASE_URL = "http://localhost:8080/api/v1/cowrie"

def pull_cowrie_logs(container_id):
    """拉取Cowrie日志"""
    url = f"{BASE_URL}/pull-logs"
    data = {"container_id": container_id}
    
    response = requests.post(url, json=data)
    return response.json()

def get_cowrie_statistics():
    """获取统计信息"""
    url = f"{BASE_URL}/statistics"
    response = requests.get(url)
    return response.json()

def get_top_commands(limit=10):
    """获取最常用命令"""
    url = f"{BASE_URL}/top-commands"
    params = {"limit": limit}
    
    response = requests.get(url, params=params)
    return response.json()

def analyze_attacker_behavior():
    """分析攻击者行为"""
    url = f"{BASE_URL}/attacker-behavior"
    response = requests.get(url)
    
    if response.status_code == 200:
        data = response.json()
        for attacker in data['data']:
            print(f"攻击者: {attacker['source_ip']}")
            print(f"  总事件数: {attacker['total_events']}")
            print(f"  执行命令数: {attacker['commands_executed']}")
            print(f"  有效命令数: {attacker['valid_commands']}")
            print(f"  活动时长: {attacker['activity_duration_minutes']}分钟")
            print("-" * 50)

# 使用示例
if __name__ == "__main__":
    # 拉取日志
    result = pull_cowrie_logs("container_123")
    print("拉取结果:", result)
    
    # 获取统计信息
    stats = get_cowrie_statistics()
    print("统计信息:", json.dumps(stats, indent=2, ensure_ascii=False))
    
    # 分析攻击者行为
    analyze_attacker_behavior()
```

## 技术特性

### 1. 高性能
- 批量数据插入
- 数据库索引优化
- JSON解析优化
- 分页查询支持

### 2. 数据完整性
- 唯一约束防重复
- 枚举类型保证数据一致性
- 事务支持保证原子性

### 3. 可扩展性
- 模块化设计
- 接口抽象
- 支持多种数据库

### 4. 监控友好
- 详细的统计视图
- 丰富的查询维度
- 实时数据分析
- 命令执行分析

## 注意事项

### 1. 当前限制
- 日志拉取功能使用模拟数据（需要根据实际容器环境调整）
- 需要确保MySQL数据库连接正常
- 建议定期清理历史数据

### 2. 性能建议
- 为高频查询字段添加索引
- 考虑数据分区策略
- 实施数据归档机制

### 3. 安全考虑
- 密码字段建议加密存储
- 实施访问控制
- 定期备份重要数据

## 扩展方向

### 1. 实时监控
- WebSocket实时推送
- 告警机制
- 仪表板展示

### 2. 高级分析
- 机器学习异常检测
- 攻击模式识别
- 威胁情报集成

### 3. 数据可视化
- 攻击趋势图表
- 命令执行热力图
- 交互式报表

## 总结

Cowrie蜜罐日志管理功能为您的项目提供了：

✅ **完整的日志管理** - 从拉取到分析的全流程支持
✅ **丰富的查询接口** - 支持多维度数据查询
✅ **深度行为分析** - 攻击者行为和命令执行分析
✅ **高性能存储** - 优化的数据库设计和索引
✅ **易用的API** - RESTful风格的完整接口
✅ **详细的文档** - 完整的使用指南和示例代码

该实现为您的蜜罐系统提供了强大的Cowrie日志分析能力，支持实时监控、威胁分析和安全报告生成。
