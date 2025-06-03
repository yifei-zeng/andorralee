# PowerShell操作命令

## 基础设置

```powershell
# 设置基础URL
$baseUrl = "http://localhost:8080/api/v1"
```

## 1. Docker镜像管理

### 拉取镜像
```powershell
# 拉取nginx镜像
$body = @{image='nginx:latest'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/docker/pull" -ContentType 'application/json' -Body $body

# 拉取达梦数据库镜像
$body = @{image_name='andorralee/dm8'; tag='v0.1'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri 'http://localhost:8080/api/v1/docker/pull' -ContentType 'application/json' -Body $body

# 拉取Cowrie蜜罐镜像
$body = @{image='andorralee/cowrie:v0.1'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/docker/pull" -ContentType 'application/json' -Body $body

$body = @{image_name='andorralee/cowrie'; tag='v0.1'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri 'http://localhost:8080/api/v1/docker/pull' -ContentType 'application/json' -Body $body

# 拉取Heralding蜜罐镜像
$body = @{image='andorralee/heralding:v0.1'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/docker/pull" -ContentType 'application/json' -Body $body

$body = @{image_name='andorralee/heralding'; tag='v0.1'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri 'http://localhost:8080/api/v1/docker/pull' -ContentType 'application/json' -Body $body

```

### 列出镜像

Containers    Created Id                                                                      Labels

```powershell
# 简单列出镜像
Invoke-WebRequest -Method GET -Uri "$baseUrl/docker/images"

# 获取镜像并格式化显示
Containers    Created Id                                                                      Labels
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/docker/images"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize |ForEach-Object { $_.RepoTags } 

# 查看镜像数量和标签
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/docker/images"
$content = $response.Content | ConvertFrom-Json
$content.data.Count  # 查看实际返回了多少个镜像
$content.data | ForEach-Object { $_.RepoTags }  # 查看所有镜像的标签

# 获取数据库中的镜像记录
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/docker/images/db"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize
```

### 获取镜像详情

123456

```powershell
# 获取指定镜像详情（需要替换为实际的镜像ID）
Invoke-WebRequest -Method GET -Uri "$baseUrl/docker/images/sha256:90cd72cf3a5cdd4016ddeefd6fe4fdb3ee45f90fea426a0de97077c5f68225ae"

# 根据数据库ID获取镜像记录
Invoke-WebRequest -Method GET -Uri "$baseUrl/docker/images/db/1"
```

### 删除镜像
```powershell
# 删除指定镜像（需要替换为实际的镜像ID）
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/docker/images/sha256:e7f6e8f6989ab064bdafc72b7be120f471a63f0bf2b256ad5e6eec5f42725bf5"

# 删除镜像数据库记录
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/docker/images/db/1"
```

### 镜像标签操作
```powershell
# 为镜像添加标签
$body = @{repo='my-honeypot'; tag='v1.0'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/docker/images/sha256:abc123.../tag" -ContentType 'application/json' -Body $body
```

## 2. 蜜罐模板管理

### 创建模板
```powershell
$body = @{
    name='SSH蜜罐模板'
    type='SSH'
    description='模拟SSH服务'
    image_name='ubuntu:20.04'
    default_port=22
    config=@{
        hostname='ssh-honeypot'
        log_level='INFO'
    }
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method POST -Uri "$baseUrl/honeypot/templates" -ContentType 'application/json' -Body $body

# 创建Web蜜罐模板
$body = @{
    name='Web蜜罐模板'
    type='HTTP'
    description='模拟Web服务'
    image_name='nginx:latest'
    default_port=80
    config=@{
        server_name='web-honeypot'
        document_root='/var/www/html'
    }
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method POST -Uri "$baseUrl/honeypot/templates" -ContentType 'application/json' -Body $body
```

### 获取模板
```powershell
# 获取所有模板
Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/templates"

# 获取所有模板并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/templates"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 获取指定模板详情
Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/templates/1"
```

### 更新模板
```powershell
$body = @{
    name='SSH蜜罐高级版'
    description='模拟SSH服务(加强版)'
    image_name='ubuntu:20.04'
    default_port=22
    config=@{
        hostname='ssh-honeypot-advanced'
        log_level='DEBUG'
        max_connections=100
    }
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method PUT -Uri "$baseUrl/honeypot/templates/1" -ContentType 'application/json' -Body $body
```

### 删除模板
```powershell
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/honeypot/templates/1"
```

### 导入模板
```powershell
$body = @{
    template_data=@{
        name='FTP蜜罐模板'
        type='FTP'
        description='模拟FTP服务'
        image_name='vsftpd:latest'
        default_port=21
        config=@{
            anonymous_enable='YES'
            local_enable='YES'
        }
    }
} | ConvertTo-Json -Depth 4
Invoke-WebRequest -Method POST -Uri "$baseUrl/honeypot/templates/import" -ContentType 'application/json' -Body $body
```

### 部署模板
```powershell
$body = @{
    instance_name='SSH蜜罐实例1'
    port=2222
} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/honeypot/templates/1/deploy" -ContentType 'application/json' -Body $body
```

## 3. 蜜罐实例管理

### 获取实例
```powershell
# 获取所有蜜罐实例
Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/instances"

# 获取所有实例并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/instances"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 获取指定实例详情
Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/instances/1"
```

### 更新实例
```powershell
$body = @{
    name='更新的SSH蜜罐实例'
    description='更新后的SSH蜜罐描述'
} | ConvertTo-Json
Invoke-WebRequest -Method PUT -Uri "$baseUrl/honeypot/instances/1" -ContentType 'application/json' -Body $body
```

### 部署实例
```powershell
# 部署蜜罐实例
Invoke-WebRequest -Method POST -Uri "$baseUrl/honeypot/instances/1/deploy" -ContentType 'application/json' -Body '{}'
```

### 停止实例
```powershell
# 停止蜜罐实例
Invoke-WebRequest -Method POST -Uri "$baseUrl/honeypot/instances/1/stop" -ContentType 'application/json' -Body '{}'
```

### 删除实例
```powershell
# 删除蜜罐实例
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/honeypot/instances/1"
```

### 获取实例日志
```powershell
# 获取蜜罐实例日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/instances/1/logs"

# 获取日志并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot/instances/1/logs"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize
```

## 4. 容器实例管理

### 创建容器实例
```powershell
# 创建SSH蜜罐容器实例
$body = @{
    name='SSH蜜罐容器1'
    honeypot_name='ssh-honeypot-1'
    image_name='ubuntu:20.04'
    protocol='ssh'
    interface_type='terminal'
    port_mappings=@{
        '22'='2222'
    }
    environment=@{
        'SSH_PORT'='2222'
        'HONEYPOT_TYPE'='SSH'
    }
    description='SSH协议蜜罐容器实例'
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances" -ContentType 'application/json' -Body $body

# 创建Web蜜罐容器实例
$body = @{
    name='Web蜜罐容器1'
    honeypot_name='web-honeypot-1'
    image_name='nginx:latest'
    protocol='http'
    interface_type='web'
    port_mappings=@{
        '80'='8080'
    }
    environment=@{
        'NGINX_PORT'='8080'
        'HONEYPOT_TYPE'='WEB'
    }
    description='Web服务蜜罐容器实例'
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances" -ContentType 'application/json' -Body $body
```

### 获取容器实例
```powershell
# 获取所有容器实例
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances"

# 获取所有容器实例并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 获取指定容器实例详情
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances/1"

# 根据状态获取容器实例
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances/status/running"
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances/status/stopped"
```

### 控制容器实例
```powershell
# 启动容器实例
Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances/1/start" -ContentType 'application/json' -Body '{}'

# 停止容器实例
Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances/1/stop" -ContentType 'application/json' -Body '{}'

# 重启容器实例
Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances/1/restart" -ContentType 'application/json' -Body '{}'

# 获取容器实例状态
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances/1/status"

# 同步所有容器实例状态
Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances/sync-status" -ContentType 'application/json' -Body '{}'
```

### 删除容器实例
```powershell
# 删除容器实例
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/container-instances/1"
```

## 5. Headling认证日志管理

### 拉取日志
```powershell
# 拉取Headling认证日志
$body = @{container_id='container_123'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/headling/pull-logs" -ContentType 'application/json' -Body $body
```

### 获取日志

![image-20250603235539637](C:\Users\ZengYifei\AppData\Roaming\Typora\typora-user-images\image-20250603235539637.png)

id timestamp                        auth_id                              session_id                           source_ip  source_port

```powershell
# 获取所有Headling日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs"

# 获取所有日志并格式化显示


id timestamp                        auth_id                              session_id                           source_ip  source_port destination_ip destination_port protocol username
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 根据ID获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs/1"

# 根据容器ID获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs/container/container_123"

# 根据源IP获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs/source-ip/192.168.1.100"

# 根据协议获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs/protocol/ssh"

# 根据时间范围获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs/time-range?start=2025-01-01T00:00:00Z&end=2025-01-31T23:59:59Z"
```

### 删除日志
```powershell
# 删除容器相关的Headling日志
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/headling/logs/container/container_123"
```

### 统计分析
```powershell
# 获取Headling统计信息
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/statistics"

# 获取统计信息并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/statistics"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 获取攻击者IP统计
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/attacker-statistics"

# 获取顶级攻击者
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/top-attackers?limit=10"

# 获取常用用户名
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/top-usernames?limit=10"

# 获取常用密码
Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/top-passwords?limit=10"
```

## 6. Cowrie蜜罐日志管理

### 拉取日志
```powershell
# 拉取Cowrie蜜罐日志
$body = @{container_id='cowrie_container_123'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/cowrie/pull-logs" -ContentType 'application/json' -Body $body
```

### 获取日志

![image-20250603235620441](C:\Users\ZengYifei\AppData\Roaming\Typora\typora-user-images\image-20250603235620441.png)

id event_time                       auth_id                              session_id                           source_ip     source_port destination_ip destination_port protocol client_info

```powershell
# 获取所有Cowrie日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs"

# 获取所有日志并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 根据ID获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/1"

# 根据容器ID获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/container/cowrie_container_123"

# 根据源IP获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/source-ip/192.168.1.100"

# 根据协议获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/protocol/ssh"

# 根据命令获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/command/ls"

# 根据用户名获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/username/root"

# 根据命令识别状态获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/command-found/true"
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/command-found/false"

# 根据时间范围获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/logs/time-range?start=2025-01-01T00:00:00Z&end=2025-01-31T23:59:59Z"
```

### 删除日志
```powershell
# 删除容器相关的Cowrie日志
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/cowrie/logs/container/cowrie_container_123"
```

### 统计分析
```powershell
# 获取Cowrie统计信息
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/statistics"

# 获取攻击者行为统计
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/attacker-behavior"

# 获取顶级攻击者
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/top-attackers?limit=10"

# 获取常用命令
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/top-commands?limit=10"

# 获取常用用户名
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/top-usernames?limit=10"

# 获取常用密码
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/top-passwords?limit=10"

# 获取常用客户端指纹
Invoke-WebRequest -Method GET -Uri "$baseUrl/cowrie/top-fingerprints?limit=10"
```

## 7. 诱饵(蜜签)管理

### 创建蜜签
```powershell
$body = @{
    name='敏感文件蜜签'
    type='file'
    content='这是一个诱饵文件'
    path='/tmp/sensitive.txt'
    description='用于检测文件访问的蜜签'
} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/baits" -ContentType 'application/json' -Body $body
```

### 获取蜜签
```powershell
# 获取所有蜜签
Invoke-WebRequest -Method GET -Uri "$baseUrl/baits"

# 获取所有蜜签并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/baits"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 根据ID获取蜜签
Invoke-WebRequest -Method GET -Uri "$baseUrl/baits/1"
```

### 更新蜜签
```powershell
$body = @{
    name='更新的蜜签'
    description='更新后的描述'
} | ConvertTo-Json
Invoke-WebRequest -Method PUT -Uri "$baseUrl/baits/1" -ContentType 'application/json' -Body $body
```

### 部署蜜签
```powershell
$body = @{target_path='/var/log/sensitive.log'} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/baits/1/deploy" -ContentType 'application/json' -Body $body
```

### 删除蜜签
```powershell
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/baits/1"
```

## 8. 安全规则管理

### 创建规则
```powershell
$body = @{
    name='SSH暴力破解检测'
    type='detection'
    condition='failed_login_attempts > 5'
    action='block_ip'
    description='检测SSH暴力破解攻击'
} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/rules" -ContentType 'application/json' -Body $body
```

### 获取规则
```powershell
# 获取所有安全规则
Invoke-WebRequest -Method GET -Uri "$baseUrl/rules"

# 获取所有规则并格式化显示
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/rules"
$content = $response.Content | ConvertFrom-Json
$content.data | Format-Table -AutoSize

# 根据ID获取规则
Invoke-WebRequest -Method GET -Uri "$baseUrl/rules/1"
```

### 更新规则
```powershell
$body = @{
    name='更新的安全规则'
    condition='failed_login_attempts > 3'
} | ConvertTo-Json
Invoke-WebRequest -Method PUT -Uri "$baseUrl/rules/1" -ContentType 'application/json' -Body $body
```

### 启用/禁用规则
```powershell
# 启用安全规则
Invoke-WebRequest -Method PUT -Uri "$baseUrl/rules/1/enable" -ContentType 'application/json' -Body '{}'

# 禁用安全规则
Invoke-WebRequest -Method PUT -Uri "$baseUrl/rules/1/disable" -ContentType 'application/json' -Body '{}'
```

### 删除规则
```powershell
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/rules/1"
```

### 规则日志
```powershell
# 获取所有规则日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/rules/logs"

# 根据ID获取规则日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/rules/logs/1"

# 根据规则ID获取日志
Invoke-WebRequest -Method GET -Uri "$baseUrl/rules/logs/rule/1"
```

## 9. AI功能

### 语义分割
```powershell
# 日志语义分割
$body = @{
    container_id='container_123'
    log_content='2025-01-15 10:30:45 [INFO] SSH connection from 192.168.1.100:45678'
} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/ai/semantic-segment" -ContentType 'application/json' -Body $body

# 图像语义分割
$body = @{
    image_path='/path/to/image.jpg'
    model='default'
} | ConvertTo-Json
Invoke-WebRequest -Method POST -Uri "$baseUrl/ai/image-segment" -ContentType 'application/json' -Body $body
```

## 10. 容器日志分析

### 获取分析结果
```powershell
# 获取所有日志分析结果
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-logs/segments"

# 根据ID获取分析结果
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-logs/segments/1"

# 根据容器ID获取分析结果
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-logs/segments/container/container_123"

# 根据类型获取分析结果
Invoke-WebRequest -Method GET -Uri "$baseUrl/container-logs/segments/type/error"
```

### 删除分析结果
```powershell
# 删除分析结果
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/container-logs/segments/1"

# 删除容器相关分析结果
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/container-logs/segments/container/container_123"
```

## 11. 数据库操作

### 通用数据操作
```powershell
# 查询数据
Invoke-WebRequest -Method GET -Uri "$baseUrl/data?table=honeypot_instance&limit=10"

# 创建数据
$body = @{
    table='honeypot_instance'
    data=@{
        name='新蜜罐实例'
        status='created'
    }
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method POST -Uri "$baseUrl/data" -ContentType 'application/json' -Body $body

# 更新数据
$body = @{
    table='honeypot_instance'
    id=1
    data=@{
        status='running'
    }
} | ConvertTo-Json -Depth 3
Invoke-WebRequest -Method PUT -Uri "$baseUrl/data" -ContentType 'application/json' -Body $body

# 删除数据
Invoke-WebRequest -Method DELETE -Uri "$baseUrl/data?table=honeypot_instance&id=1"

# 根据ID获取数据
Invoke-WebRequest -Method GET -Uri "$baseUrl/data/id?table=honeypot_instance&id=1"

# 根据名称获取数据
Invoke-WebRequest -Method GET -Uri "$baseUrl/data/name?table=honeypot_instance&name=测试蜜罐"
```

## 12. 批量操作示例

### 批量创建容器实例
```powershell
# 批量创建5个SSH蜜罐实例
for ($i = 1; $i -le 5; $i++) {
    $body = @{
        name="SSH蜜罐$i"
        honeypot_name="ssh-honeypot-$i"
        image_name='ubuntu:20.04'
        protocol='ssh'
        interface_type='terminal'
        port_mappings=@{
            '22'="$((2220 + $i))"
        }
        environment=@{
            'SSH_PORT'="$((2220 + $i))"
            'HONEYPOT_TYPE'='SSH'
            'INSTANCE_ID'="$i"
        }
        description="SSH蜜罐实例$i"
    } | ConvertTo-Json -Depth 3

    Write-Host "创建SSH蜜罐$i..."
    Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances" -ContentType 'application/json' -Body $body
    Start-Sleep -Seconds 1
}
```

### 批量启动容器实例
```powershell
# 获取所有stopped状态的实例并启动
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances/status/stopped"
$content = $response.Content | ConvertFrom-Json

foreach ($instance in $content.data) {
    Write-Host "启动容器实例$($instance.id)..."
    Invoke-WebRequest -Method POST -Uri "$baseUrl/container-instances/$($instance.id)/start" -ContentType 'application/json' -Body '{}'
    Start-Sleep -Seconds 1
}
```

### 批量拉取日志
```powershell
# 获取所有running状态的实例并拉取日志
$response = Invoke-WebRequest -Method GET -Uri "$baseUrl/container-instances/status/running"
$content = $response.Content | ConvertFrom-Json

foreach ($instance in $content.data) {
    if ($instance.container_id) {
        Write-Host "拉取容器$($instance.container_id)的日志..."

        # 拉取Headling日志
        $body = @{container_id=$instance.container_id} | ConvertTo-Json
        Invoke-WebRequest -Method POST -Uri "$baseUrl/headling/pull-logs" -ContentType 'application/json' -Body $body

        # 拉取Cowrie日志
        Invoke-WebRequest -Method POST -Uri "$baseUrl/cowrie/pull-logs" -ContentType 'application/json' -Body $body

        Start-Sleep -Seconds 2
    }
}
```
```
