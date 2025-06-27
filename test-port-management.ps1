# 端口管理功能测试脚本
# 测试新的端口管理API和自动端口分配功能

param(
    [string]$BaseUrl = "http://localhost:8081/api/v1",
    [switch]$Verbose
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

# API调用函数
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null
    )
    
    $uri = "$BaseUrl$Endpoint"
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            if ($Verbose) {
                Write-Host "请求: $Method $uri" -ForegroundColor Gray
                Write-Host "Body: $jsonBody" -ForegroundColor Gray
            }
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        } else {
            if ($Verbose) {
                Write-Host "请求: $Method $uri" -ForegroundColor Gray
            }
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        
        if ($Verbose) {
            Write-Host "响应: $($response | ConvertTo-Json -Depth 5)" -ForegroundColor Gray
        }
        
        return $response
    }
    catch {
        Write-Error "API请求失败: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $responseBody = $reader.ReadToEnd()
            Write-Error "响应内容: $responseBody"
        }
        throw
    }
}

# 测试端口统计信息
function Test-PortStatistics {
    Write-Info "=== 测试端口统计信息 ==="
    
    try {
        $stats = Invoke-ApiRequest -Method "GET" -Endpoint "/ports/statistics"
        Write-Success "获取端口统计信息成功"
        Write-Host "总分配端口数: $($stats.data.total_allocated)" -ForegroundColor Yellow
        
        if ($stats.data.by_service_type) {
            Write-Host "按服务类型统计:" -ForegroundColor Yellow
            $stats.data.by_service_type.PSObject.Properties | ForEach-Object {
                Write-Host "  $($_.Name): $($_.Value)" -ForegroundColor White
            }
        }
        
        return $true
    }
    catch {
        Write-Error "获取端口统计信息失败: $($_.Exception.Message)"
        return $false
    }
}

# 测试自动端口分配
function Test-AutoPortAllocation {
    Write-Info "=== 测试自动端口分配 ==="
    
    try {
        # 测试MySQL服务端口分配
        $allocateRequest = @{
            container_id = "test-mysql-container"
            service_type = "mysql"
            description = "测试MySQL端口分配"
        }
        
        $result = Invoke-ApiRequest -Method "POST" -Endpoint "/ports/allocate" -Body $allocateRequest
        Write-Success "MySQL端口分配成功: $($result.data.port)"
        
        # 测试SSH服务端口分配
        $allocateRequest = @{
            container_id = "test-ssh-container"
            service_type = "ssh"
            description = "测试SSH端口分配"
        }
        
        $result2 = Invoke-ApiRequest -Method "POST" -Endpoint "/ports/allocate" -Body $allocateRequest
        Write-Success "SSH端口分配成功: $($result2.data.port)"
        
        return @($result.data.port, $result2.data.port)
    }
    catch {
        Write-Error "自动端口分配失败: $($_.Exception.Message)"
        return $null
    }
}

# 测试指定端口分配
function Test-SpecificPortAllocation {
    Write-Info "=== 测试指定端口分配 ==="
    
    try {
        # 尝试分配一个高端口号
        $testPort = 12345
        $allocateRequest = @{
            port = $testPort
            container_id = "test-specific-container"
            service_type = "http"
            description = "测试指定端口分配"
        }
        
        $result = Invoke-ApiRequest -Method "POST" -Endpoint "/ports/allocate-specific" -Body $allocateRequest
        Write-Success "指定端口 $testPort 分配成功"
        
        return $testPort
    }
    catch {
        Write-Error "指定端口分配失败: $($_.Exception.Message)"
        return $null
    }
}

# 测试端口映射自动分配
function Test-AutoPortMapping {
    Write-Info "=== 测试端口映射自动分配 ==="
    
    try {
        $mappingRequest = @{
            container_id = "test-mapping-container"
            port_mappings = @{
                "22" = "auto"      # SSH端口自动分配
                "80" = "auto"      # HTTP端口自动分配
                "3306" = "auto"    # MySQL端口自动分配
                "8080" = "18080"   # 指定端口映射
            }
        }
        
        $result = Invoke-ApiRequest -Method "POST" -Endpoint "/ports/auto-allocate-mapping" -Body $mappingRequest
        Write-Success "端口映射自动分配成功"
        Write-Host "分配的端口映射:" -ForegroundColor Yellow
        $result.data.port_mappings.PSObject.Properties | ForEach-Object {
            Write-Host "  容器端口 $($_.Name) -> 主机端口 $($_.Value)" -ForegroundColor White
        }
        
        return $result.data.port_mappings
    }
    catch {
        Write-Error "端口映射自动分配失败: $($_.Exception.Message)"
        return $null
    }
}

# 测试容器创建（使用自动端口分配）
function Test-ContainerCreationWithAutoPort {
    Write-Info "=== 测试容器创建（自动端口分配） ==="
    
    try {
        $containerRequest = @{
            name = "测试容器-自动端口"
            honeypot_name = "test-auto-port"
            image_name = "nginx:latest"
            protocol = "http"
            interface_type = "web"
            port_mappings = @{
                "80" = "auto"    # 自动分配HTTP端口
                "443" = "auto"   # 自动分配HTTPS端口
            }
            environment = @{
                "TEST_MODE" = "true"
            }
            description = "测试自动端口分配的容器"
            auto_start = $false
        }
        
        $result = Invoke-ApiRequest -Method "POST" -Endpoint "/container-instances" -Body $containerRequest
        Write-Success "容器创建成功，ID: $($result.data.id)"
        Write-Host "分配的端口映射:" -ForegroundColor Yellow
        $result.data.port_mappings.PSObject.Properties | ForEach-Object {
            Write-Host "  容器端口 $($_.Name) -> 主机端口 $($_.Value)" -ForegroundColor White
        }
        
        return $result.data.id
    }
    catch {
        Write-Error "容器创建失败: $($_.Exception.Message)"
        return $null
    }
}

# 测试端口查询
function Test-PortQuery {
    param([array]$AllocatedPorts)
    
    Write-Info "=== 测试端口查询 ==="
    
    try {
        # 查询所有已分配端口
        $allPorts = Invoke-ApiRequest -Method "GET" -Endpoint "/ports/allocated"
        Write-Success "查询所有已分配端口成功，总数: $($allPorts.data.total)"
        
        # 查询特定端口信息
        if ($AllocatedPorts -and $AllocatedPorts.Count -gt 0) {
            $testPort = $AllocatedPorts[0]
            $portInfo = Invoke-ApiRequest -Method "GET" -Endpoint "/ports/$testPort"
            Write-Success "查询端口 $testPort 信息成功"
            Write-Host "端口信息: 容器ID=$($portInfo.data.container_id), 服务类型=$($portInfo.data.service_type)" -ForegroundColor White
        }
        
        return $true
    }
    catch {
        Write-Error "端口查询失败: $($_.Exception.Message)"
        return $false
    }
}

# 清理测试数据
function Cleanup-TestData {
    param([array]$AllocatedPorts, [string]$ContainerID)
    
    Write-Info "=== 清理测试数据 ==="
    
    try {
        # 删除测试容器
        if ($ContainerID) {
            try {
                Invoke-ApiRequest -Method "DELETE" -Endpoint "/container-instances/$ContainerID"
                Write-Success "删除测试容器成功"
            }
            catch {
                Write-Warning "删除测试容器失败: $($_.Exception.Message)"
            }
        }
        
        # 释放测试端口
        if ($AllocatedPorts) {
            foreach ($port in $AllocatedPorts) {
                try {
                    Invoke-ApiRequest -Method "DELETE" -Endpoint "/ports/$port/release"
                    Write-Success "释放端口 $port 成功"
                }
                catch {
                    Write-Warning "释放端口 $port 失败: $($_.Exception.Message)"
                }
            }
        }
        
        # 释放容器相关端口
        $testContainers = @("test-mysql-container", "test-ssh-container", "test-specific-container", "test-mapping-container")
        foreach ($containerName in $testContainers) {
            try {
                Invoke-ApiRequest -Method "DELETE" -Endpoint "/ports/container/$containerName/release"
                Write-Success "释放容器 $containerName 的端口成功"
            }
            catch {
                Write-Warning "释放容器 $containerName 的端口失败: $($_.Exception.Message)"
            }
        }
    }
    catch {
        Write-Warning "清理过程中出现错误: $($_.Exception.Message)"
    }
}

# 主测试流程
function Main {
    Write-Info "开始端口管理功能测试..."
    Write-Info "测试服务器: $BaseUrl"
    
    $allocatedPorts = @()
    $containerID = $null
    
    try {
        # 1. 测试端口统计
        if (-not (Test-PortStatistics)) {
            throw "端口统计测试失败"
        }
        
        # 2. 测试自动端口分配
        $autoPorts = Test-AutoPortAllocation
        if ($autoPorts) {
            $allocatedPorts += $autoPorts
        }
        
        # 3. 测试指定端口分配
        $specificPort = Test-SpecificPortAllocation
        if ($specificPort) {
            $allocatedPorts += $specificPort
        }
        
        # 4. 测试端口映射自动分配
        $mappingResult = Test-AutoPortMapping
        
        # 5. 测试容器创建
        $containerID = Test-ContainerCreationWithAutoPort
        
        # 6. 测试端口查询
        Test-PortQuery -AllocatedPorts $allocatedPorts
        
        Write-Success "所有测试完成！"
        
    }
    catch {
        Write-Error "测试过程中出现错误: $($_.Exception.Message)"
    }
    finally {
        # 清理测试数据
        Cleanup-TestData -AllocatedPorts $allocatedPorts -ContainerID $containerID
    }
}

# 运行测试
Main
