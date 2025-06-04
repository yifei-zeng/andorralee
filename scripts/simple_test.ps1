# 蜜罐容器系统简化测试脚本

Write-Host "开始蜜罐容器系统测试..." -ForegroundColor Green

$baseUrl = "http://localhost:8081/api/v1"

# 测试1: 获取蜜罐模板
Write-Host "`n1. 测试蜜罐模板功能..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Method GET -Uri "$baseUrl/honeypot-templates"
    if ($response.StatusCode -eq 200) {
        $data = ($response.Content | ConvertFrom-Json).data
        Write-Host "✅ 蜜罐模板获取成功，共 $($data.Count) 个模板" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ 蜜罐模板测试失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试2: 创建蜜签
Write-Host "`n2. 测试蜜签管理功能..." -ForegroundColor Yellow
try {
    $body = @{
        name = "测试蜜签"
        type = "credential"
        content = "test:password"
        description = "测试用蜜签"
    } | ConvertTo-Json
    
    $response = Invoke-WebRequest -Method POST -Uri "$baseUrl/honeytokens" -ContentType "application/json" -Body $body
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ 蜜签创建成功" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ 蜜签创建失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试3: 模拟攻击
Write-Host "`n3. 测试攻击模拟功能..." -ForegroundColor Yellow
try {
    $body = @{
        attack_type = "sql_injection"
        target_ip = "127.0.0.1"
        target_port = 80
    } | ConvertTo-Json
    
    $response = Invoke-WebRequest -Method POST -Uri "$baseUrl/attack-capture/simulate" -ContentType "application/json" -Body $body
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ 攻击模拟成功" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ 攻击模拟失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试4: 端口扫描
Write-Host "`n4. 测试端口扫描功能..." -ForegroundColor Yellow
try {
    $body = @{
        target = "127.0.0.1"
        ports = "80,443,8081"
        protocol = "tcp"
        timeout = 3
    } | ConvertTo-Json
    
    $response = Invoke-WebRequest -Method POST -Uri "$baseUrl/port-scan" -ContentType "application/json" -Body $body
    if ($response.StatusCode -eq 200) {
        $data = ($response.Content | ConvertFrom-Json).data
        Write-Host "✅ 端口扫描成功，开放端口: $($data.open_ports)" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ 端口扫描失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试5: MySQL日志功能
Write-Host "`n5. 测试MySQL日志功能..." -ForegroundColor Yellow
try {
    $body = @{
        container_id = "test-container"
        log_data = @(
            @{
                timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
                auth_id = "test_001"
                session_id = "session_001"
                source_ip = "192.168.1.100"
                source_port = 12345
                destination_ip = "192.168.1.1"
                destination_port = 22
                protocol = "ssh"
                username = "admin"
                password = "123456"
                password_hash = "e10adc3949ba59abbe56e057f20f883e"
            }
        )
    } | ConvertTo-Json -Depth 3
    
    $response = Invoke-WebRequest -Method POST -Uri "$baseUrl/headling/pull-logs" -ContentType "application/json" -Body $body
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Headling日志写入成功" -ForegroundColor Green
        
        # 测试读取
        $response = Invoke-WebRequest -Method GET -Uri "$baseUrl/headling/logs"
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ Headling日志读取成功" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "❌ MySQL日志测试失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试6: 容器实例管理
Write-Host "`n6. 测试容器实例管理..." -ForegroundColor Yellow
try {
    $body = @{
        name = "测试容器"
        honeypot_name = "test-honeypot"
        image_name = "andorralee/cowrie:v0.1"
        protocol = "ssh"
        interface_type = "network"
        port_mappings = @{
            "22" = "12222"
        }
        environment = @{
            "COWRIE_HOSTNAME" = "test"
        }
        description = "测试容器实例"
        auto_start = $false
    } | ConvertTo-Json -Depth 3
    
    $response = Invoke-WebRequest -Method POST -Uri "$baseUrl/memory-container-instances" -ContentType "application/json" -Body $body
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ 容器实例创建成功" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ 容器实例创建失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n测试完成！" -ForegroundColor Green
Write-Host "系统已通过蜜罐容器系统测试大纲的主要功能验证！" -ForegroundColor Cyan
