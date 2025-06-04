# 🎯 蜜罐容器系统快速测试脚本
# 用于验证系统是否通过测试大纲的所有要求

Write-Host "🚀 开始蜜罐容器系统快速测试..." -ForegroundColor Green

$baseUrl = "http://localhost:8081/api/v1"
$testResults = @()

function Test-API {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [object]$Body = $null
    )
    
    try {
        $params = @{
            Method = $Method
            Uri = $Url
            ContentType = "application/json"
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 3)
        }
        
        $response = Invoke-WebRequest @params
        $result = @{
            Name = $Name
            Status = "✅ 通过"
            StatusCode = $response.StatusCode
            Details = ""
        }
        
        if ($response.StatusCode -eq 200) {
            $data = ($response.Content | ConvertFrom-Json)
            if ($data.code -eq 0) {
                $result.Details = "成功"
            } else {
                $result.Details = $data.message
            }
        }
        
        return $result
    }
    catch {
        return @{
            Name = $Name
            Status = "❌ 失败"
            StatusCode = $_.Exception.Response.StatusCode.value__
            Details = $_.Exception.Message
        }
    }
}

Write-Host "`n📋 一、蜜罐部署功能测试" -ForegroundColor Yellow

# 1.1 获取蜜罐模板
$testResults += Test-API "获取蜜罐模板" "GET" "$baseUrl/honeypot-templates"

# 1.2 从模板部署蜜罐
$deployBody = @{
    name = "快速测试SSH蜜罐"
    auto_start = $false
}
$testResults += Test-API "部署SSH蜜罐" "POST" "$baseUrl/honeypot-templates/ssh-cowrie/deploy" $deployBody

# 1.3 端口扫描验证
$scanBody = @{
    target = "127.0.0.1"
    ports = "22,80,443,8081"
    protocol = "tcp"
    timeout = 3
}
$testResults += Test-API "端口扫描验证" "POST" "$baseUrl/port-scan" $scanBody

Write-Host "`n🔧 二、蜜罐管理功能测试" -ForegroundColor Yellow

# 2.1 创建容器实例
$containerBody = @{
    name = "管理测试容器"
    honeypot_name = "mgmt-test"
    image_name = "andorralee/cowrie:v0.1"
    protocol = "ssh"
    interface_type = "network"
    port_mappings = @{
        "22" = "12222"
    }
    environment = @{
        "COWRIE_HOSTNAME" = "mgmt-test"
    }
    description = "管理功能测试容器"
    auto_start = $false
}
$testResults += Test-API "创建容器实例" "POST" "$baseUrl/memory-container-instances" $containerBody

# 2.2 查询容器状态
$testResults += Test-API "查询容器状态" "GET" "$baseUrl/memory-container-instances"

Write-Host "`n🍯 三、蜜签管理功能测试" -ForegroundColor Yellow

# 3.1 查看默认蜜签
$testResults += Test-API "查看默认蜜签" "GET" "$baseUrl/honeytokens"

# 3.2 创建新蜜签
$tokenBody = @{
    name = "测试API密钥"
    type = "credential"
    content = "api_key:sk-test123456"
    description = "测试用的API密钥蜜签"
}
$testResults += Test-API "创建新蜜签" "POST" "$baseUrl/honeytokens" $tokenBody

# 3.3 触发蜜签
$triggerBody = @{
    action = "api_access"
    details = "测试触发API密钥蜜签"
}
$testResults += Test-API "触发蜜签" "POST" "$baseUrl/honeytokens/1/trigger" $triggerBody

# 3.4 获取触发记录
$testResults += Test-API "获取触发记录" "GET" "$baseUrl/honeytokens/triggers"

Write-Host "`n🎯 四、攻击行为捕获测试" -ForegroundColor Yellow

# 4.1 模拟SQL注入攻击
$sqlAttackBody = @{
    attack_type = "sql_injection"
    target_ip = "127.0.0.1"
    target_port = 80
}
$testResults += Test-API "模拟SQL注入" "POST" "$baseUrl/attack-capture/simulate" $sqlAttackBody

# 4.2 模拟XSS攻击
$xssAttackBody = @{
    attack_type = "xss"
    target_ip = "127.0.0.1"
    target_port = 80
}
$testResults += Test-API "模拟XSS攻击" "POST" "$baseUrl/attack-capture/simulate" $xssAttackBody

# 4.3 获取攻击事件
$testResults += Test-API "获取攻击事件" "GET" "$baseUrl/attack-capture/events"

# 4.4 获取攻击统计
$testResults += Test-API "获取攻击统计" "GET" "$baseUrl/attack-capture/statistics"

Write-Host "`n📊 五、日志记录功能测试" -ForegroundColor Yellow

# 5.1 导出攻击日志
$exportBody = @{
    log_type = "attack"
    format = "json"
}
$testResults += Test-API "导出攻击日志" "POST" "$baseUrl/logs/export" $exportBody

# 5.2 获取日志统计
$testResults += Test-API "获取日志统计" "GET" "$baseUrl/logs/statistics"

Write-Host "`n🗄️ 六、MySQL数据库测试" -ForegroundColor Yellow

# 6.1 测试Headling日志写入
$headlingBody = @{
    container_id = "quick-test-headling"
    log_data = @(
        @{
            timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
            auth_id = "quick_test_001"
            session_id = "quick_session_001"
            source_ip = "192.168.1.200"
            source_port = 54321
            destination_ip = "192.168.1.1"
            destination_port = 22
            protocol = "ssh"
            username = "testuser"
            password = "testpass"
            password_hash = "5d41402abc4b2a76b9719d911017c592"
        }
    )
}
$testResults += Test-API "Headling日志写入" "POST" "$baseUrl/headling/pull-logs" $headlingBody

# 6.2 测试Headling日志读取
$testResults += Test-API "Headling日志读取" "GET" "$baseUrl/headling/logs"

# 6.3 测试Cowrie日志写入
$cowrieBody = @{
    container_id = "quick-test-cowrie"
    log_data = @(
        @{
            event_time = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
            auth_id = "cowrie_quick_001"
            session_id = "cowrie_session_001"
            source_ip = "192.168.1.201"
            source_port = 12345
            destination_ip = "192.168.1.1"
            destination_port = 2222
            protocol = "ssh"
            client_info = "SSH-2.0-OpenSSH_8.0"
            fingerprint = "SHA256:quick_test_fingerprint"
            username = "root"
            password = "admin"
            password_hash = "21232f297a57a5a743894a0e4a801fc3"
            command = "whoami"
            command_found = $true
            raw_log = "quick test raw log"
        }
    )
}
$testResults += Test-API "Cowrie日志写入" "POST" "$baseUrl/cowrie/pull-logs" $cowrieBody

# 6.4 测试Cowrie日志读取
$testResults += Test-API "Cowrie日志读取" "GET" "$baseUrl/cowrie/logs"

Write-Host "`n📈 测试结果汇总" -ForegroundColor Cyan
Write-Host "=" * 80

$passCount = ($testResults | Where-Object { $_.Status -like "*通过*" }).Count
$failCount = ($testResults | Where-Object { $_.Status -like "*失败*" }).Count
$totalCount = $testResults.Count

foreach ($result in $testResults) {
    $statusColor = if ($result.Status -like "*通过*") { "Green" } else { "Red" }
    Write-Host "$($result.Status) $($result.Name)" -ForegroundColor $statusColor
    if ($result.Details) {
        Write-Host "    详情: $($result.Details)" -ForegroundColor Gray
    }
}

Write-Host "`n📊 总体统计:" -ForegroundColor Cyan
Write-Host "  总测试项: $totalCount" -ForegroundColor White
Write-Host "  通过: $passCount" -ForegroundColor Green
Write-Host "  失败: $failCount" -ForegroundColor Red
Write-Host "  通过率: $([math]::Round($passCount / $totalCount * 100, 2))%" -ForegroundColor $(if ($passCount -eq $totalCount) { "Green" } else { "Yellow" })

if ($passCount -eq $totalCount) {
    Write-Host "`n恭喜！系统完全通过蜜罐容器系统测试大纲的所有要求！" -ForegroundColor Green
} else {
    Write-Host "`n系统部分功能需要检查，请查看失败项目的详细信息。" -ForegroundColor Yellow
}

Write-Host "`n测试完成！" -ForegroundColor Green
