# 蜜罐管理系统 PowerShell 操作指南

## 系统概述

本文档提供了使用PowerShell操作蜜罐管理系统的完整指南，包括Docker镜像管理、容器实例管理、蜜罐日志分析等所有功能。

## 基础配置

### 设置基础变量
```powershell
# 设置API基础URL
$BaseURL = "http://localhost:8080/api/v1"

# 设置请求头
$Headers = @{
    "Content-Type" = "application/json"
    "Accept" = "application/json"
}

# 辅助函数：发送GET请求
function Invoke-ApiGet {
    param([string]$Endpoint)
    try {
        $Response = Invoke-RestMethod -Uri "$BaseURL$Endpoint" -Method GET -Headers $Headers
        return $Response
    }
    catch {
        Write-Error "GET请求失败: $($_.Exception.Message)"
        return $null
    }
}

# 辅助函数：发送POST请求
function Invoke-ApiPost {
    param([string]$Endpoint, [hashtable]$Body)
    try {
        $JsonBody = $Body | ConvertTo-Json -Depth 10
        $Response = Invoke-RestMethod -Uri "$BaseURL$Endpoint" -Method POST -Headers $Headers -Body $JsonBody
        return $Response
    }
    catch {
        Write-Error "POST请求失败: $($_.Exception.Message)"
        return $null
    }
}

# 辅助函数：发送DELETE请求
function Invoke-ApiDelete {
    param([string]$Endpoint)
    try {
        $Response = Invoke-RestMethod -Uri "$BaseURL$Endpoint" -Method DELETE -Headers $Headers
        return $Response
    }
    catch {
        Write-Error "DELETE请求失败: $($_.Exception.Message)"
        return $null
    }
}
```

## 1. Docker镜像管理

### 拉取Docker镜像
```powershell
# 拉取指定镜像
function Pull-DockerImage {
    param([string]$ImageName)
    
    $Body = @{
        image = $ImageName
    }
    
    Write-Host "正在拉取镜像: $ImageName"
    $Result = Invoke-ApiPost -Endpoint "/docker/pull" -Body $Body
    
    if ($Result) {
        Write-Host "镜像拉取成功: $ImageName" -ForegroundColor Green
        return $Result
    }
}

# 使用示例
Pull-DockerImage -ImageName "nginx:latest"
Pull-DockerImage -ImageName "cowrie/cowrie:latest"
Pull-DockerImage -ImageName "headling/headling:latest"
```

### 查看镜像列表
```powershell
# 获取所有镜像
function Get-DockerImages {
    Write-Host "获取Docker镜像列表..."
    $Images = Invoke-ApiGet -Endpoint "/docker/images"
    
    if ($Images) {
        Write-Host "找到 $($Images.Count) 个镜像:" -ForegroundColor Green
        $Images | ForEach-Object {
            Write-Host "  - ID: $($_.Id)" -ForegroundColor Yellow
            Write-Host "    标签: $($_.RepoTags -join ', ')"
            Write-Host "    大小: $([math]::Round($_.Size / 1MB, 2)) MB"
            Write-Host "    创建时间: $($_.Created)"
            Write-Host ""
        }
        return $Images
    }
}

# 获取数据库中的镜像记录
function Get-DockerImagesFromDB {
    Write-Host "获取数据库中的镜像记录..."
    $Images = Invoke-ApiGet -Endpoint "/docker/images/db"
    
    if ($Images) {
        Write-Host "数据库中有 $($Images.Count) 条镜像记录:" -ForegroundColor Green
        $Images | ForEach-Object {
            Write-Host "  - 镜像名称: $($_.image_name)" -ForegroundColor Yellow
            Write-Host "    镜像ID: $($_.image_id)"
            Write-Host "    状态: $($_.status)"
            Write-Host "    创建时间: $($_.created_at)"
            Write-Host ""
        }
        return $Images
    }
}

# 使用示例
Get-DockerImages
Get-DockerImagesFromDB
```

### 删除Docker镜像
```powershell
# 删除指定镜像
function Remove-DockerImage {
    param([string]$ImageID)
    
    Write-Host "正在删除镜像: $ImageID"
    $Result = Invoke-ApiDelete -Endpoint "/docker/images/$ImageID"
    
    if ($Result) {
        Write-Host "镜像删除成功: $ImageID" -ForegroundColor Green
        return $Result
    }
}

# 使用示例
Remove-DockerImage -ImageID "sha256:abc123..."
```

## 2. 容器实例管理

### 创建容器实例
```powershell
# 创建新的容器实例
function New-ContainerInstance {
    param(
        [string]$Name,
        [string]$HoneypotName,
        [string]$ImageName,
        [string]$Protocol,
        [string]$InterfaceType = "web",
        [hashtable]$PortMappings = @{},
        [hashtable]$Environment = @{},
        [string]$Description = "",
        [int]$TemplateID = 0
    )
    
    $Body = @{
        name = $Name
        honeypot_name = $HoneypotName
        image_name = $ImageName
        protocol = $Protocol
        interface_type = $InterfaceType
        port_mappings = $PortMappings
        environment = $Environment
        description = $Description
    }
    
    if ($TemplateID -gt 0) {
        $Body.template_id = $TemplateID
    }
    
    Write-Host "正在创建容器实例: $Name"
    $Result = Invoke-ApiPost -Endpoint "/container-instances" -Body $Body
    
    if ($Result) {
        Write-Host "容器实例创建成功:" -ForegroundColor Green
        Write-Host "  - 实例ID: $($Result.id)" -ForegroundColor Yellow
        Write-Host "  - 容器名称: $($Result.container_name)"
        Write-Host "  - 容器ID: $($Result.container_id)"
        Write-Host "  - 蜜罐IP: $($Result.honeypot_ip)"
        Write-Host "  - 状态: $($Result.status)"
        return $Result
    }
}

# 使用示例
$PortMappings = @{
    "22" = "2222"
    "80" = "8080"
}

$Environment = @{
    "COWRIE_HOSTNAME" = "server01"
    "COWRIE_LOG_LEVEL" = "INFO"
}

New-ContainerInstance -Name "测试蜜罐1" -HoneypotName "cowrie-test" -ImageName "cowrie/cowrie:latest" -Protocol "ssh" -InterfaceType "terminal" -PortMappings $PortMappings -Environment $Environment -Description "测试用Cowrie蜜罐"
```

### 管理容器实例
```powershell
# 获取所有容器实例
function Get-ContainerInstances {
    Write-Host "获取所有容器实例..."
    $Instances = Invoke-ApiGet -Endpoint "/container-instances"
    
    if ($Instances) {
        Write-Host "找到 $($Instances.Count) 个容器实例:" -ForegroundColor Green
        $Instances | ForEach-Object {
            Write-Host "  - ID: $($_.id)" -ForegroundColor Yellow
            Write-Host "    名称: $($_.name)"
            Write-Host "    蜜罐名称: $($_.honeypot_name)"
            Write-Host "    容器名称: $($_.container_name)"
            Write-Host "    状态: $($_.status)"
            Write-Host "    协议: $($_.protocol)"
            Write-Host "    蜜罐IP: $($_.honeypot_ip)"
            Write-Host "    创建时间: $($_.create_time)"
            Write-Host ""
        }
        return $Instances
    }
}

# 根据状态获取容器实例
function Get-ContainerInstancesByStatus {
    param([string]$Status)
    
    Write-Host "获取状态为 '$Status' 的容器实例..."
    $Instances = Invoke-ApiGet -Endpoint "/container-instances/status/$Status"
    
    if ($Instances) {
        Write-Host "找到 $($Instances.Count) 个 '$Status' 状态的实例:" -ForegroundColor Green
        $Instances | ForEach-Object {
            Write-Host "  - $($_.name) (ID: $($_.id))" -ForegroundColor Yellow
        }
        return $Instances
    }
}

# 获取指定容器实例详情
function Get-ContainerInstance {
    param([int]$InstanceID)
    
    Write-Host "获取容器实例详情: $InstanceID"
    $Instance = Invoke-ApiGet -Endpoint "/container-instances/$InstanceID"
    
    if ($Instance) {
        Write-Host "容器实例详情:" -ForegroundColor Green
        Write-Host "  - ID: $($Instance.id)" -ForegroundColor Yellow
        Write-Host "  - 名称: $($Instance.name)"
        Write-Host "  - 蜜罐名称: $($Instance.honeypot_name)"
        Write-Host "  - 容器名称: $($Instance.container_name)"
        Write-Host "  - 容器ID: $($Instance.container_id)"
        Write-Host "  - 状态: $($Instance.status)"
        Write-Host "  - 协议: $($Instance.protocol)"
        Write-Host "  - 接口类型: $($Instance.interface_type)"
        Write-Host "  - 主机IP: $($Instance.ip)"
        Write-Host "  - 蜜罐IP: $($Instance.honeypot_ip)"
        Write-Host "  - 端口: $($Instance.port)"
        Write-Host "  - 镜像名称: $($Instance.image_name)"
        Write-Host "  - 创建时间: $($Instance.create_time)"
        Write-Host "  - 更新时间: $($Instance.update_time)"
        Write-Host "  - 描述: $($Instance.description)"
        return $Instance
    }
}

# 使用示例
Get-ContainerInstances
Get-ContainerInstancesByStatus -Status "running"
Get-ContainerInstance -InstanceID 1
```

### 控制容器实例
```powershell
# 启动容器实例
function Start-ContainerInstance {
    param([int]$InstanceID)
    
    Write-Host "正在启动容器实例: $InstanceID"
    $Result = Invoke-ApiPost -Endpoint "/container-instances/$InstanceID/start" -Body @{}
    
    if ($Result) {
        Write-Host "容器实例启动成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 停止容器实例
function Stop-ContainerInstance {
    param([int]$InstanceID)
    
    Write-Host "正在停止容器实例: $InstanceID"
    $Result = Invoke-ApiPost -Endpoint "/container-instances/$InstanceID/stop" -Body @{}
    
    if ($Result) {
        Write-Host "容器实例停止成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 重启容器实例
function Restart-ContainerInstance {
    param([int]$InstanceID)
    
    Write-Host "正在重启容器实例: $InstanceID"
    $Result = Invoke-ApiPost -Endpoint "/container-instances/$InstanceID/restart" -Body @{}
    
    if ($Result) {
        Write-Host "容器实例重启成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 删除容器实例
function Remove-ContainerInstance {
    param([int]$InstanceID)
    
    Write-Host "正在删除容器实例: $InstanceID"
    $Result = Invoke-ApiDelete -Endpoint "/container-instances/$InstanceID"
    
    if ($Result) {
        Write-Host "容器实例删除成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 获取容器实例状态
function Get-ContainerInstanceStatus {
    param([int]$InstanceID)
    
    $Status = Invoke-ApiGet -Endpoint "/container-instances/$InstanceID/status"
    
    if ($Status) {
        Write-Host "容器实例 $InstanceID 状态: $($Status.status)" -ForegroundColor Green
        return $Status
    }
}

# 同步所有容器实例状态
function Sync-AllContainerInstancesStatus {
    Write-Host "正在同步所有容器实例状态..."
    $Result = Invoke-ApiPost -Endpoint "/container-instances/sync-status" -Body @{}
    
    if ($Result) {
        Write-Host "容器实例状态同步成功" -ForegroundColor Green
        return $Result
    }
}

# 使用示例
Start-ContainerInstance -InstanceID 1
Stop-ContainerInstance -InstanceID 1
Restart-ContainerInstance -InstanceID 1
Get-ContainerInstanceStatus -InstanceID 1
Sync-AllContainerInstancesStatus
Remove-ContainerInstance -InstanceID 1

## 3. Headling认证日志管理

### 拉取和查看Headling日志
```powershell
# 拉取Headling认证日志
function Get-HeadlingLogs {
    param([string]$ContainerID)

    $Body = @{
        container_id = $ContainerID
    }

    Write-Host "正在拉取容器 $ContainerID 的Headling认证日志..."
    $Result = Invoke-ApiPost -Endpoint "/headling/pull-logs" -Body $Body

    if ($Result) {
        Write-Host "Headling日志拉取成功" -ForegroundColor Green
        return $Result
    }
}

# 获取所有Headling日志
function Get-AllHeadlingLogs {
    Write-Host "获取所有Headling认证日志..."
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs"

    if ($Logs) {
        Write-Host "找到 $($Logs.Count) 条Headling日志:" -ForegroundColor Green
        $Logs | ForEach-Object {
            Write-Host "  - ID: $($_.id)" -ForegroundColor Yellow
            Write-Host "    时间: $($_.timestamp)"
            Write-Host "    源IP: $($_.source_ip):$($_.source_port)"
            Write-Host "    目标: $($_.destination_ip):$($_.destination_port)"
            Write-Host "    协议: $($_.protocol)"
            Write-Host "    用户名: $($_.username)"
            Write-Host "    密码: $($_.password)"
            Write-Host ""
        }
        return $Logs
    }
}

# 根据源IP获取Headling日志
function Get-HeadlingLogsBySourceIP {
    param([string]$SourceIP)

    Write-Host "获取源IP $SourceIP 的Headling日志..."
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs/source-ip/$SourceIP"

    if ($Logs) {
        Write-Host "找到 $($Logs.Count) 条来自 $SourceIP 的日志:" -ForegroundColor Green
        return $Logs
    }
}

# 根据协议获取Headling日志
function Get-HeadlingLogsByProtocol {
    param([string]$Protocol)

    Write-Host "获取协议 $Protocol 的Headling日志..."
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs/protocol/$Protocol"

    if ($Logs) {
        Write-Host "找到 $($Logs.Count) 条 $Protocol 协议的日志:" -ForegroundColor Green
        return $Logs
    }
}

# 根据时间范围获取Headling日志
function Get-HeadlingLogsByTimeRange {
    param(
        [datetime]$StartTime,
        [datetime]$EndTime
    )

    $StartTimeStr = $StartTime.ToString("yyyy-MM-ddTHH:mm:ssZ")
    $EndTimeStr = $EndTime.ToString("yyyy-MM-ddTHH:mm:ssZ")

    Write-Host "获取时间范围 $StartTimeStr 到 $EndTimeStr 的Headling日志..."
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs/time-range?start_time=$StartTimeStr&end_time=$EndTimeStr"

    if ($Logs) {
        Write-Host "找到 $($Logs.Count) 条指定时间范围的日志:" -ForegroundColor Green
        return $Logs
    }
}

# 使用示例
Get-HeadlingLogs -ContainerID "container_123"
Get-AllHeadlingLogs
Get-HeadlingLogsBySourceIP -SourceIP "192.168.1.100"
Get-HeadlingLogsByProtocol -Protocol "ssh"
Get-HeadlingLogsByTimeRange -StartTime (Get-Date).AddDays(-1) -EndTime (Get-Date)
```

### Headling统计分析
```powershell
# 获取Headling统计信息
function Get-HeadlingStatistics {
    Write-Host "获取Headling统计信息..."
    $Stats = Invoke-ApiGet -Endpoint "/headling/statistics"

    if ($Stats) {
        Write-Host "Headling统计信息:" -ForegroundColor Green
        $Stats | ForEach-Object {
            Write-Host "  - 日期: $($_.log_date)" -ForegroundColor Yellow
            Write-Host "    协议: $($_.protocol)"
            Write-Host "    总事件数: $($_.total_events)"
            Write-Host "    唯一IP数: $($_.unique_ips)"
            Write-Host "    成功认证: $($_.successful_auths)"
            Write-Host "    失败认证: $($_.failed_auths)"
            Write-Host ""
        }
        return $Stats
    }
}

# 获取攻击者IP统计
function Get-AttackerIPStatistics {
    Write-Host "获取攻击者IP统计..."
    $Stats = Invoke-ApiGet -Endpoint "/headling/attacker-statistics"

    if ($Stats) {
        Write-Host "攻击者IP统计:" -ForegroundColor Green
        $Stats | ForEach-Object {
            Write-Host "  - IP: $($_.source_ip)" -ForegroundColor Yellow
            Write-Host "    总尝试次数: $($_.total_attempts)"
            Write-Host "    唯一用户名: $($_.unique_usernames)"
            Write-Host "    唯一密码: $($_.unique_passwords)"
            Write-Host "    首次攻击: $($_.first_attempt)"
            Write-Host "    最后攻击: $($_.last_attempt)"
            Write-Host ""
        }
        return $Stats
    }
}

# 获取顶级攻击者
function Get-TopAttackers {
    param([int]$Limit = 10)

    Write-Host "获取前 $Limit 个攻击者..."
    $Attackers = Invoke-ApiGet -Endpoint "/headling/top-attackers?limit=$Limit"

    if ($Attackers) {
        Write-Host "前 $Limit 个攻击者:" -ForegroundColor Green
        for ($i = 0; $i -lt $Attackers.Count; $i++) {
            $Attacker = $Attackers[$i]
            Write-Host "  $($i + 1). IP: $($Attacker.source_ip)" -ForegroundColor Yellow
            Write-Host "     尝试次数: $($Attacker.total_attempts)"
            Write-Host ""
        }
        return $Attackers
    }
}

# 获取常用用户名
function Get-TopUsernames {
    param([int]$Limit = 10)

    Write-Host "获取前 $Limit 个常用用户名..."
    $Usernames = Invoke-ApiGet -Endpoint "/headling/top-usernames?limit=$Limit"

    if ($Usernames) {
        Write-Host "前 $Limit 个常用用户名:" -ForegroundColor Green
        for ($i = 0; $i -lt $Usernames.Count; $i++) {
            $Username = $Usernames[$i]
            Write-Host "  $($i + 1). 用户名: $($Username.username)" -ForegroundColor Yellow
            Write-Host "     使用次数: $($Username.count)"
            Write-Host "     唯一IP数: $($Username.unique_ips)"
            Write-Host ""
        }
        return $Usernames
    }
}

# 获取常用密码
function Get-TopPasswords {
    param([int]$Limit = 10)

    Write-Host "获取前 $Limit 个常用密码..."
    $Passwords = Invoke-ApiGet -Endpoint "/headling/top-passwords?limit=$Limit"

    if ($Passwords) {
        Write-Host "前 $Limit 个常用密码:" -ForegroundColor Green
        for ($i = 0; $i -lt $Passwords.Count; $i++) {
            $Password = $Passwords[$i]
            Write-Host "  $($i + 1). 密码: $($Password.password)" -ForegroundColor Yellow
            Write-Host "     使用次数: $($Password.count)"
            Write-Host "     唯一IP数: $($Password.unique_ips)"
            Write-Host ""
        }
        return $Passwords
    }
}

# 使用示例
Get-HeadlingStatistics
Get-AttackerIPStatistics
Get-TopAttackers -Limit 5
Get-TopUsernames -Limit 5
Get-TopPasswords -Limit 5
```
```
