# 蜜罐管理系统 PowerShell 完整操作指南

## 系统概述

本指南基于当前已实现的功能，提供完整的PowerShell操作说明。系统已成功启动并监听在8080端口。

## 基础配置

### 设置环境变量和辅助函数
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
        Write-Host "✅ GET请求成功: $Endpoint" -ForegroundColor Green
        return $Response
    }
    catch {
        Write-Host "❌ GET请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 辅助函数：发送POST请求
function Invoke-ApiPost {
    param([string]$Endpoint, [hashtable]$Body)
    try {
        $JsonBody = $Body | ConvertTo-Json -Depth 10
        $Response = Invoke-RestMethod -Uri "$BaseURL$Endpoint" -Method POST -Headers $Headers -Body $JsonBody
        Write-Host "✅ POST请求成功: $Endpoint" -ForegroundColor Green
        return $Response
    }
    catch {
        Write-Host "❌ POST请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 辅助函数：发送DELETE请求
function Invoke-ApiDelete {
    param([string]$Endpoint)
    try {
        $Response = Invoke-RestMethod -Uri "$BaseURL$Endpoint" -Method DELETE -Headers $Headers
        Write-Host "✅ DELETE请求成功: $Endpoint" -ForegroundColor Green
        return $Response
    }
    catch {
        Write-Host "❌ DELETE请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 辅助函数：发送PUT请求
function Invoke-ApiPut {
    param([string]$Endpoint, [hashtable]$Body)
    try {
        $JsonBody = $Body | ConvertTo-Json -Depth 10
        $Response = Invoke-RestMethod -Uri "$BaseURL$Endpoint" -Method PUT -Headers $Headers -Body $JsonBody
        Write-Host "✅ PUT请求成功: $Endpoint" -ForegroundColor Green
        return $Response
    }
    catch {
        Write-Host "❌ PUT请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 测试连接函数
function Test-ApiConnection {
    Write-Host "🔍 测试API连接..." -ForegroundColor Yellow
    try {
        $Response = Invoke-RestMethod -Uri "$BaseURL/docker/images" -Method GET -Headers $Headers -TimeoutSec 5
        Write-Host "✅ API连接成功！系统正常运行" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "❌ API连接失败: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "请确保系统已启动并监听在8080端口" -ForegroundColor Yellow
        return $false
    }
}

# 首先测试连接
Test-ApiConnection
```

## 1. Docker镜像管理

### 拉取Docker镜像
```powershell
# 拉取常用蜜罐镜像
function Pull-HoneypotImages {
    Write-Host "🐳 开始拉取蜜罐镜像..." -ForegroundColor Cyan
    
    $Images = @(
        "nginx:latest",
        "ubuntu:20.04",
        "alpine:latest",
        "mysql:8.0",
        "redis:latest"
    )
    
    foreach ($Image in $Images) {
        Write-Host "正在拉取镜像: $Image" -ForegroundColor Yellow
        $Body = @{ image = $Image }
        $Result = Invoke-ApiPost -Endpoint "/docker/pull" -Body $Body
        
        if ($Result) {
            Write-Host "✅ 镜像拉取成功: $Image" -ForegroundColor Green
        }
        Start-Sleep -Seconds 2
    }
}

# 拉取指定镜像
function Pull-DockerImage {
    param([string]$ImageName)
    
    Write-Host "🐳 拉取镜像: $ImageName" -ForegroundColor Cyan
    $Body = @{ image = $ImageName }
    $Result = Invoke-ApiPost -Endpoint "/docker/pull" -Body $Body
    
    if ($Result) {
        Write-Host "✅ 镜像拉取成功: $ImageName" -ForegroundColor Green
        return $Result
    }
}

# 使用示例
Pull-DockerImage -ImageName "cowrie/cowrie:latest"
Pull-HoneypotImages
```

### 查看和管理镜像
```powershell
# 获取所有Docker镜像
function Get-DockerImages {
    Write-Host "🐳 获取Docker镜像列表..." -ForegroundColor Cyan
    $Images = Invoke-ApiGet -Endpoint "/docker/images"
    
    if ($Images) {
        Write-Host "📋 找到 $($Images.Count) 个镜像:" -ForegroundColor Green
        $Images | ForEach-Object {
            Write-Host "  🏷️  ID: $($_.Id)" -ForegroundColor Yellow
            Write-Host "     标签: $($_.RepoTags -join ', ')"
            Write-Host "     大小: $([math]::Round($_.Size / 1MB, 2)) MB"
            Write-Host "     创建: $($_.Created)"
            Write-Host ""
        }
        return $Images
    }
}

# 获取数据库中的镜像记录
function Get-DockerImagesFromDB {
    Write-Host "💾 获取数据库中的镜像记录..." -ForegroundColor Cyan
    $Images = Invoke-ApiGet -Endpoint "/docker/images/db"
    
    if ($Images) {
        Write-Host "📋 数据库中有 $($Images.Count) 条镜像记录:" -ForegroundColor Green
        $Images | ForEach-Object {
            Write-Host "  🏷️  镜像名称: $($_.image_name)" -ForegroundColor Yellow
            Write-Host "     镜像ID: $($_.image_id)"
            Write-Host "     状态: $($_.status)"
            Write-Host "     创建时间: $($_.created_at)"
            Write-Host ""
        }
        return $Images
    }
}

# 删除指定镜像
function Remove-DockerImage {
    param([string]$ImageID)
    
    Write-Host "🗑️  删除镜像: $ImageID" -ForegroundColor Red
    $Result = Invoke-ApiDelete -Endpoint "/docker/images/$ImageID"
    
    if ($Result) {
        Write-Host "✅ 镜像删除成功: $ImageID" -ForegroundColor Green
        return $Result
    }
}

# 使用示例
Get-DockerImages
Get-DockerImagesFromDB
```

## 2. 容器实例管理

### 创建容器实例
```powershell
# 创建SSH蜜罐实例
function New-SSHHoneypot {
    param(
        [string]$Name = "SSH蜜罐",
        [string]$HoneypotName = "ssh-honeypot",
        [int]$Port = 2222
    )
    
    Write-Host "🍯 创建SSH蜜罐实例: $Name" -ForegroundColor Cyan
    
    $Body = @{
        name = $Name
        honeypot_name = $HoneypotName
        image_name = "ubuntu:20.04"
        protocol = "ssh"
        interface_type = "terminal"
        port_mappings = @{
            "22" = $Port.ToString()
        }
        environment = @{
            "SSH_PORT" = $Port.ToString()
            "HONEYPOT_TYPE" = "SSH"
        }
        description = "SSH协议蜜罐实例"
    }
    
    $Result = Invoke-ApiPost -Endpoint "/container-instances" -Body $Body
    
    if ($Result) {
        Write-Host "✅ SSH蜜罐创建成功!" -ForegroundColor Green
        Write-Host "  📋 实例ID: $($Result.id)" -ForegroundColor Yellow
        Write-Host "  🐳 容器名称: $($Result.container_name)"
        Write-Host "  🌐 端口: $Port"
        Write-Host "  📊 状态: $($Result.status)"
        return $Result
    }
}

# 创建Web蜜罐实例
function New-WebHoneypot {
    param(
        [string]$Name = "Web蜜罐",
        [string]$HoneypotName = "web-honeypot",
        [int]$Port = 8080
    )
    
    Write-Host "🍯 创建Web蜜罐实例: $Name" -ForegroundColor Cyan
    
    $Body = @{
        name = $Name
        honeypot_name = $HoneypotName
        image_name = "nginx:latest"
        protocol = "http"
        interface_type = "web"
        port_mappings = @{
            "80" = $Port.ToString()
        }
        environment = @{
            "NGINX_PORT" = $Port.ToString()
            "HONEYPOT_TYPE" = "WEB"
        }
        description = "Web服务蜜罐实例"
    }
    
    $Result = Invoke-ApiPost -Endpoint "/container-instances" -Body $Body
    
    if ($Result) {
        Write-Host "✅ Web蜜罐创建成功!" -ForegroundColor Green
        Write-Host "  📋 实例ID: $($Result.id)" -ForegroundColor Yellow
        Write-Host "  🐳 容器名称: $($Result.container_name)"
        Write-Host "  🌐 端口: $Port"
        Write-Host "  📊 状态: $($Result.status)"
        return $Result
    }
}

# 批量创建蜜罐实例
function New-HoneypotCluster {
    param([int]$Count = 3)
    
    Write-Host "🍯 批量创建 $Count 个蜜罐实例..." -ForegroundColor Cyan
    
    $Results = @()
    for ($i = 1; $i -le $Count; $i++) {
        $Name = "蜜罐实例$i"
        $HoneypotName = "honeypot-$i"
        $Port = 2220 + $i
        
        $Result = New-SSHHoneypot -Name $Name -HoneypotName $HoneypotName -Port $Port
        if ($Result) {
            $Results += $Result
        }
        Start-Sleep -Seconds 1
    }
    
    Write-Host "✅ 批量创建完成，成功创建 $($Results.Count) 个实例" -ForegroundColor Green
    return $Results
}

# 使用示例
New-SSHHoneypot -Name "测试SSH蜜罐" -Port 2222
New-WebHoneypot -Name "测试Web蜜罐" -Port 8080
New-HoneypotCluster -Count 3
```

### 管理容器实例
```powershell
# 获取所有容器实例
function Get-ContainerInstances {
    Write-Host "📋 获取所有容器实例..." -ForegroundColor Cyan
    $Instances = Invoke-ApiGet -Endpoint "/container-instances"
    
    if ($Instances) {
        Write-Host "📊 找到 $($Instances.Count) 个容器实例:" -ForegroundColor Green
        $Instances | ForEach-Object {
            $StatusColor = switch ($_.status) {
                "running" { "Green" }
                "stopped" { "Red" }
                "created" { "Yellow" }
                default { "Gray" }
            }
            
            Write-Host "  🍯 ID: $($_.id)" -ForegroundColor Yellow
            Write-Host "     名称: $($_.name)"
            Write-Host "     蜜罐名称: $($_.honeypot_name)"
            Write-Host "     容器名称: $($_.container_name)"
            Write-Host "     状态: $($_.status)" -ForegroundColor $StatusColor
            Write-Host "     协议: $($_.protocol)"
            Write-Host "     端口: $($_.port)"
            Write-Host "     创建时间: $($_.create_time)"
            Write-Host ""
        }
        return $Instances
    }
}

# 根据状态获取容器实例
function Get-ContainerInstancesByStatus {
    param([string]$Status)
    
    Write-Host "📋 获取状态为 '$Status' 的容器实例..." -ForegroundColor Cyan
    $Instances = Invoke-ApiGet -Endpoint "/container-instances/status/$Status"
    
    if ($Instances) {
        Write-Host "📊 找到 $($Instances.Count) 个 '$Status' 状态的实例:" -ForegroundColor Green
        $Instances | ForEach-Object {
            Write-Host "  🍯 $($_.name) (ID: $($_.id))" -ForegroundColor Yellow
        }
        return $Instances
    }
}

# 获取指定容器实例详情
function Get-ContainerInstance {
    param([int]$InstanceID)
    
    Write-Host "🔍 获取容器实例详情: $InstanceID" -ForegroundColor Cyan
    $Instance = Invoke-ApiGet -Endpoint "/container-instances/$InstanceID"
    
    if ($Instance) {
        Write-Host "📋 容器实例详情:" -ForegroundColor Green
        Write-Host "  🍯 ID: $($Instance.id)" -ForegroundColor Yellow
        Write-Host "     名称: $($Instance.name)"
        Write-Host "     蜜罐名称: $($Instance.honeypot_name)"
        Write-Host "     容器名称: $($Instance.container_name)"
        Write-Host "     容器ID: $($Instance.container_id)"
        Write-Host "     状态: $($Instance.status)"
        Write-Host "     协议: $($Instance.protocol)"
        Write-Host "     接口类型: $($Instance.interface_type)"
        Write-Host "     主机IP: $($Instance.ip)"
        Write-Host "     蜜罐IP: $($Instance.honeypot_ip)"
        Write-Host "     端口: $($Instance.port)"
        Write-Host "     镜像名称: $($Instance.image_name)"
        Write-Host "     创建时间: $($Instance.create_time)"
        Write-Host "     更新时间: $($Instance.update_time)"
        Write-Host "     描述: $($Instance.description)"
        return $Instance
    }
}

# 使用示例
Get-ContainerInstances
Get-ContainerInstancesByStatus -Status "running"
Get-ContainerInstance -InstanceID 1

### 控制容器实例
```powershell
# 启动容器实例
function Start-ContainerInstance {
    param([int]$InstanceID)

    Write-Host "▶️  启动容器实例: $InstanceID" -ForegroundColor Green
    $Result = Invoke-ApiPost -Endpoint "/container-instances/$InstanceID/start" -Body @{}

    if ($Result) {
        Write-Host "✅ 容器实例启动成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 停止容器实例
function Stop-ContainerInstance {
    param([int]$InstanceID)

    Write-Host "⏹️  停止容器实例: $InstanceID" -ForegroundColor Red
    $Result = Invoke-ApiPost -Endpoint "/container-instances/$InstanceID/stop" -Body @{}

    if ($Result) {
        Write-Host "✅ 容器实例停止成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 重启容器实例
function Restart-ContainerInstance {
    param([int]$InstanceID)

    Write-Host "🔄 重启容器实例: $InstanceID" -ForegroundColor Yellow
    $Result = Invoke-ApiPost -Endpoint "/container-instances/$InstanceID/restart" -Body @{}

    if ($Result) {
        Write-Host "✅ 容器实例重启成功: $InstanceID" -ForegroundColor Green
        return $Result
    }
}

# 删除容器实例
function Remove-ContainerInstance {
    param([int]$InstanceID)

    Write-Host "🗑️  删除容器实例: $InstanceID" -ForegroundColor Red
    $Confirm = Read-Host "确认删除容器实例 $InstanceID ? (y/N)"

    if ($Confirm -eq 'y' -or $Confirm -eq 'Y') {
        $Result = Invoke-ApiDelete -Endpoint "/container-instances/$InstanceID"

        if ($Result) {
            Write-Host "✅ 容器实例删除成功: $InstanceID" -ForegroundColor Green
            return $Result
        }
    } else {
        Write-Host "❌ 删除操作已取消" -ForegroundColor Yellow
    }
}

# 获取容器实例状态
function Get-ContainerInstanceStatus {
    param([int]$InstanceID)

    $Status = Invoke-ApiGet -Endpoint "/container-instances/$InstanceID/status"

    if ($Status) {
        $StatusColor = switch ($Status.status) {
            "running" { "Green" }
            "stopped" { "Red" }
            "created" { "Yellow" }
            default { "Gray" }
        }
        Write-Host "📊 容器实例 $InstanceID 状态: $($Status.status)" -ForegroundColor $StatusColor
        return $Status
    }
}

# 批量操作容器实例
function Start-AllContainerInstances {
    Write-Host "▶️  启动所有容器实例..." -ForegroundColor Green
    $Instances = Get-ContainerInstances

    if ($Instances) {
        foreach ($Instance in $Instances) {
            if ($Instance.status -ne "running") {
                Start-ContainerInstance -InstanceID $Instance.id
                Start-Sleep -Seconds 1
            }
        }
    }
}

function Stop-AllContainerInstances {
    Write-Host "⏹️  停止所有容器实例..." -ForegroundColor Red
    $Instances = Get-ContainerInstances

    if ($Instances) {
        foreach ($Instance in $Instances) {
            if ($Instance.status -eq "running") {
                Stop-ContainerInstance -InstanceID $Instance.id
                Start-Sleep -Seconds 1
            }
        }
    }
}

# 同步所有容器实例状态
function Sync-AllContainerInstancesStatus {
    Write-Host "🔄 同步所有容器实例状态..." -ForegroundColor Yellow
    $Result = Invoke-ApiPost -Endpoint "/container-instances/sync-status" -Body @{}

    if ($Result) {
        Write-Host "✅ 容器实例状态同步成功" -ForegroundColor Green
        return $Result
    }
}

# 使用示例
Start-ContainerInstance -InstanceID 1
Stop-ContainerInstance -InstanceID 1
Restart-ContainerInstance -InstanceID 1
Get-ContainerInstanceStatus -InstanceID 1
Sync-AllContainerInstancesStatus
```

## 3. Headling认证日志管理

### 拉取和查看Headling日志
```powershell
# 拉取Headling认证日志
function Get-HeadlingLogs {
    param([string]$ContainerID)

    Write-Host "📥 拉取容器 $ContainerID 的Headling认证日志..." -ForegroundColor Cyan
    $Body = @{ container_id = $ContainerID }
    $Result = Invoke-ApiPost -Endpoint "/headling/pull-logs" -Body $Body

    if ($Result) {
        Write-Host "✅ Headling日志拉取成功" -ForegroundColor Green
        return $Result
    }
}

# 获取所有Headling日志
function Get-AllHeadlingLogs {
    Write-Host "📋 获取所有Headling认证日志..." -ForegroundColor Cyan
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs"

    if ($Logs) {
        Write-Host "📊 找到 $($Logs.Count) 条Headling日志:" -ForegroundColor Green
        $Logs | Select-Object -First 10 | ForEach-Object {
            Write-Host "  🔐 ID: $($_.id)" -ForegroundColor Yellow
            Write-Host "     时间: $($_.timestamp)"
            Write-Host "     源IP: $($_.source_ip):$($_.source_port)"
            Write-Host "     目标: $($_.destination_ip):$($_.destination_port)"
            Write-Host "     协议: $($_.protocol)"
            Write-Host "     用户名: $($_.username)"
            Write-Host "     密码: $($_.password)"
            Write-Host ""
        }
        if ($Logs.Count -gt 10) {
            Write-Host "... 还有 $($Logs.Count - 10) 条日志" -ForegroundColor Gray
        }
        return $Logs
    }
}

# 根据源IP获取Headling日志
function Get-HeadlingLogsByIP {
    param([string]$SourceIP)

    Write-Host "🔍 获取源IP $SourceIP 的Headling日志..." -ForegroundColor Cyan
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs/source-ip/$SourceIP"

    if ($Logs) {
        Write-Host "📊 找到 $($Logs.Count) 条来自 $SourceIP 的日志:" -ForegroundColor Green
        return $Logs
    }
}

# 根据协议获取Headling日志
function Get-HeadlingLogsByProtocol {
    param([string]$Protocol)

    Write-Host "🔍 获取协议 $Protocol 的Headling日志..." -ForegroundColor Cyan
    $Logs = Invoke-ApiGet -Endpoint "/headling/logs/protocol/$Protocol"

    if ($Logs) {
        Write-Host "📊 找到 $($Logs.Count) 条 $Protocol 协议的日志:" -ForegroundColor Green
        return $Logs
    }
}

# 使用示例
Get-HeadlingLogs -ContainerID "container_123"
Get-AllHeadlingLogs
Get-HeadlingLogsByIP -SourceIP "192.168.1.100"
Get-HeadlingLogsByProtocol -Protocol "ssh"
```

### Headling统计分析
```powershell
# 获取Headling统计信息
function Get-HeadlingStatistics {
    Write-Host "📊 获取Headling统计信息..." -ForegroundColor Cyan
    $Stats = Invoke-ApiGet -Endpoint "/headling/statistics"

    if ($Stats) {
        Write-Host "📈 Headling统计信息:" -ForegroundColor Green
        $Stats | ForEach-Object {
            Write-Host "  📅 日期: $($_.log_date)" -ForegroundColor Yellow
            Write-Host "     协议: $($_.protocol)"
            Write-Host "     总事件数: $($_.total_events)"
            Write-Host "     唯一IP数: $($_.unique_ips)"
            Write-Host "     成功认证: $($_.successful_auths)"
            Write-Host "     失败认证: $($_.failed_auths)"
            Write-Host ""
        }
        return $Stats
    }
}

# 获取攻击者IP统计
function Get-AttackerIPStatistics {
    Write-Host "🎯 获取攻击者IP统计..." -ForegroundColor Cyan
    $Stats = Invoke-ApiGet -Endpoint "/headling/attacker-statistics"

    if ($Stats) {
        Write-Host "🔍 攻击者IP统计:" -ForegroundColor Green
        $Stats | Select-Object -First 10 | ForEach-Object {
            Write-Host "  🌐 IP: $($_.source_ip)" -ForegroundColor Yellow
            Write-Host "     总尝试次数: $($_.total_attempts)"
            Write-Host "     唯一用户名: $($_.unique_usernames)"
            Write-Host "     唯一密码: $($_.unique_passwords)"
            Write-Host "     首次攻击: $($_.first_attempt)"
            Write-Host "     最后攻击: $($_.last_attempt)"
            Write-Host ""
        }
        return $Stats
    }
}

# 获取顶级攻击者
function Get-TopAttackers {
    param([int]$Limit = 10)

    Write-Host "🏆 获取前 $Limit 个攻击者..." -ForegroundColor Cyan
    $Attackers = Invoke-ApiGet -Endpoint "/headling/top-attackers?limit=$Limit"

    if ($Attackers) {
        Write-Host "🥇 前 $Limit 个攻击者:" -ForegroundColor Green
        for ($i = 0; $i -lt $Attackers.Count; $i++) {
            $Attacker = $Attackers[$i]
            Write-Host "  $($i + 1). 🌐 IP: $($Attacker.source_ip)" -ForegroundColor Yellow
            Write-Host "     🔢 尝试次数: $($Attacker.total_attempts)"
            Write-Host ""
        }
        return $Attackers
    }
}

# 使用示例
Get-HeadlingStatistics
Get-AttackerIPStatistics
Get-TopAttackers -Limit 5
```
```
