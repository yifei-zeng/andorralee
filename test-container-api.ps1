# 容器API测试脚本
# 用于测试和调试容器实例相关的API接口

param(
    [string]$BaseUrl = "http://localhost:8081",
    [string]$ApiVersion = "v1",
    [switch]$Verbose = $false,
    [switch]$Help = $false
)

# 显示帮助信息
if ($Help) {
    Write-Host @"
容器API测试脚本

用法:
    .\test-container-api.ps1 [参数]

参数:
    -BaseUrl <URL>        API基础URL (默认: http://localhost:8081)
    -ApiVersion <版本>    API版本 (默认: v1)
    -Verbose             显示详细输出
    -Help                显示此帮助信息

示例:
    # 基本测试
    .\test-container-api.ps1
    
    # 使用自定义URL
    .\test-container-api.ps1 -BaseUrl "http://192.168.1.100:8081"
    
    # 详细输出
    .\test-container-api.ps1 -Verbose

"@
    exit 0
}

# 全局变量
$ApiBase = "$BaseUrl/api/$ApiVersion"
$TestResults = @()

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] " -NoNewline -ForegroundColor Gray
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput $Message "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput $Message "Red" }
function Write-Warning { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Write-Info { param([string]$Message) Write-ColorOutput $Message "Cyan" }

# API调用函数
function Invoke-ApiCall {
    param(
        [string]$Endpoint,
        [string]$Method = "GET",
        [hashtable]$Body = $null,
        [string]$Description = ""
    )
    
    $url = "$ApiBase$Endpoint"
    $testResult = @{
        Endpoint = $Endpoint
        Method = $Method
        Description = $Description
        Success = $false
        StatusCode = 0
        Response = $null
        Error = $null
        Duration = 0
    }
    
    try {
        Write-Info "测试: $Description"
        Write-Info "请求: $Method $url"
        
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        
        $requestParams = @{
            Uri = $url
            Method = $Method
            ContentType = "application/json"
            TimeoutSec = 30
        }
        
        if ($Body) {
            $requestParams.Body = $Body | ConvertTo-Json -Depth 10
            if ($Verbose) {
                Write-Info "请求体: $($requestParams.Body)"
            }
        }
        
        $response = Invoke-WebRequest @requestParams
        $stopwatch.Stop()
        
        $testResult.Success = $true
        $testResult.StatusCode = $response.StatusCode
        $testResult.Duration = $stopwatch.ElapsedMilliseconds
        
        try {
            $testResult.Response = $response.Content | ConvertFrom-Json
        } catch {
            $testResult.Response = $response.Content
        }
        
        Write-Success "✅ 成功 (${$testResult.Duration}ms) - 状态码: $($response.StatusCode)"
        
        if ($Verbose -and $testResult.Response) {
            Write-Host "响应数据:" -ForegroundColor Yellow
            Write-Host ($testResult.Response | ConvertTo-Json -Depth 3) -ForegroundColor White
        }
        
    } catch {
        $stopwatch.Stop()
        $testResult.Error = $_.Exception.Message
        $testResult.Duration = $stopwatch.ElapsedMilliseconds
        
        if ($_.Exception.Response) {
            $testResult.StatusCode = $_.Exception.Response.StatusCode.value__
        }
        
        Write-Error "❌ 失败 (${$testResult.Duration}ms) - $($testResult.Error)"
        
        if ($Verbose) {
            Write-Host "详细错误:" -ForegroundColor Red
            Write-Host $_.Exception.ToString() -ForegroundColor White
        }
    }
    
    $TestResults += $testResult
    Write-Host ""
    return $testResult
}

# 测试健康检查
function Test-HealthCheck {
    Write-Info "=== 健康检查测试 ==="
    
    Invoke-ApiCall -Endpoint "/health" -Description "简单健康检查"
    Invoke-ApiCall -Endpoint "/health" -Description "详细健康检查"
    Invoke-ApiCall -Endpoint "/ready" -Description "就绪检查"
    Invoke-ApiCall -Endpoint "/live" -Description "存活检查"
}

# 测试Docker相关API
function Test-DockerAPI {
    Write-Info "=== Docker API测试 ==="
    
    Invoke-ApiCall -Endpoint "/docker/info" -Description "Docker信息"
    Invoke-ApiCall -Endpoint "/docker/images" -Description "Docker镜像列表"
    Invoke-ApiCall -Endpoint "/docker/containers" -Description "Docker容器列表"
}

# 测试容器实例API
function Test-ContainerInstanceAPI {
    Write-Info "=== 容器实例API测试 ==="
    
    # 获取所有容器实例
    $containersResult = Invoke-ApiCall -Endpoint "/container-instances" -Description "获取所有容器实例"
    
    # 同步容器状态
    Invoke-ApiCall -Endpoint "/container-instances/sync" -Method "POST" -Description "同步容器状态"
    
    # 如果有容器实例，测试单个容器操作
    if ($containersResult.Success -and $containersResult.Response.data -and $containersResult.Response.data.Count -gt 0) {
        $firstContainer = $containersResult.Response.data[0]
        $containerId = $firstContainer.id
        
        Write-Info "找到容器实例，ID: $containerId，名称: $($firstContainer.name)"
        
        # 获取单个容器信息
        Invoke-ApiCall -Endpoint "/container-instances/$containerId" -Description "获取容器实例详情"
        
        # 获取调试信息
        Invoke-ApiCall -Endpoint "/container-instances/$containerId/debug" -Description "获取容器调试信息"
        
        # 获取容器状态
        Invoke-ApiCall -Endpoint "/container-instances/$containerId/status" -Description "获取容器状态"
        
        # 测试容器操作（谨慎操作）
        $userChoice = Read-Host "是否要测试容器启动/停止操作？这可能会影响正在运行的容器 (y/N)"
        if ($userChoice -eq 'y' -or $userChoice -eq 'Y') {
            Write-Warning "开始测试容器操作..."
            
            # 如果容器是停止状态，尝试启动
            if ($firstContainer.status -eq "stopped" -or $firstContainer.status -eq "created") {
                Invoke-ApiCall -Endpoint "/container-instances/$containerId/start" -Method "POST" -Description "启动容器"
                Start-Sleep -Seconds 2
            }
            
            # 如果容器是运行状态，尝试停止
            if ($firstContainer.status -eq "running") {
                Invoke-ApiCall -Endpoint "/container-instances/$containerId/stop" -Method "POST" -Description "停止容器"
                Start-Sleep -Seconds 2
            }
        }
    } else {
        Write-Warning "没有找到容器实例，跳过单个容器测试"
        
        # 询问是否创建测试容器
        $createTest = Read-Host "是否要创建一个测试容器实例？(y/N)"
        if ($createTest -eq 'y' -or $createTest -eq 'Y') {
            Test-CreateContainer
        }
    }
}

# 测试创建容器
function Test-CreateContainer {
    Write-Info "=== 创建测试容器 ==="
    
    $testContainer = @{
        name = "api-test-container-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
        honeypot_name = "test-honeypot"
        image_name = "hello-world:latest"
        protocol = "http"
        interface_type = "network"
        port_mappings = @{
            "80" = "8080"
        }
        environment = @{
            "TEST" = "true"
            "CREATED_BY" = "api-test-script"
        }
        description = "通过API测试脚本创建的测试容器"
    }
    
    $createResult = Invoke-ApiCall -Endpoint "/container-instances" -Method "POST" -Body $testContainer -Description "创建测试容器实例"
    
    if ($createResult.Success -and $createResult.Response.data) {
        $newContainerId = $createResult.Response.data.id
        Write-Success "测试容器创建成功，ID: $newContainerId"
        
        # 等待一下然后获取容器信息
        Start-Sleep -Seconds 2
        Invoke-ApiCall -Endpoint "/container-instances/$newContainerId" -Description "获取新创建的容器信息"
        Invoke-ApiCall -Endpoint "/container-instances/$newContainerId/debug" -Description "获取新容器调试信息"
        
        # 询问是否删除测试容器
        $deleteTest = Read-Host "是否要删除刚创建的测试容器？(y/N)"
        if ($deleteTest -eq 'y' -or $deleteTest -eq 'Y') {
            Invoke-ApiCall -Endpoint "/container-instances/$newContainerId" -Method "DELETE" -Description "删除测试容器"
        }
    }
}

# 生成测试报告
function Generate-TestReport {
    Write-Info "=== 测试报告 ==="
    
    $totalTests = $TestResults.Count
    $successfulTests = ($TestResults | Where-Object { $_.Success }).Count
    $failedTests = $totalTests - $successfulTests
    $averageDuration = ($TestResults | Measure-Object -Property Duration -Average).Average
    
    Write-Host ""
    Write-Host "📊 测试统计:" -ForegroundColor Cyan
    Write-Host "  总测试数: $totalTests" -ForegroundColor White
    Write-Host "  成功: $successfulTests" -ForegroundColor Green
    Write-Host "  失败: $failedTests" -ForegroundColor Red
    Write-Host "  成功率: $([math]::Round($successfulTests / $totalTests * 100, 2))%" -ForegroundColor Yellow
    Write-Host "  平均响应时间: $([math]::Round($averageDuration, 2))ms" -ForegroundColor White
    
    if ($failedTests -gt 0) {
        Write-Host ""
        Write-Host "❌ 失败的测试:" -ForegroundColor Red
        $TestResults | Where-Object { -not $_.Success } | ForEach-Object {
            Write-Host "  - $($_.Description): $($_.Error)" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    Write-Host "📋 详细结果:" -ForegroundColor Cyan
    $TestResults | ForEach-Object {
        $status = if ($_.Success) { "✅" } else { "❌" }
        $statusCode = if ($_.StatusCode -gt 0) { " ($($_.StatusCode))" } else { "" }
        Write-Host "  $status $($_.Description) - $($_.Duration)ms$statusCode" -ForegroundColor White
    }
    
    # 保存报告到文件
    $reportFile = "api-test-report-$(Get-Date -Format 'yyyyMMdd-HHmmss').json"
    $TestResults | ConvertTo-Json -Depth 10 | Out-File -FilePath $reportFile -Encoding UTF8
    Write-Info "详细报告已保存到: $reportFile"
}

# 主函数
function Main {
    Write-Host "🚀 开始API测试" -ForegroundColor Green
    Write-Host "目标服务器: $BaseUrl" -ForegroundColor Yellow
    Write-Host "API版本: $ApiVersion" -ForegroundColor Yellow
    Write-Host ""
    
    try {
        # 基础连接测试
        Write-Info "测试基础连接..."
        $pingResult = Test-Connection -ComputerName ([System.Uri]$BaseUrl).Host -Count 1 -Quiet
        if (-not $pingResult) {
            Write-Warning "无法ping通目标主机，但继续测试..."
        }
        
        # 执行各项测试
        Test-HealthCheck
        Test-DockerAPI
        Test-ContainerInstanceAPI
        
        # 生成报告
        Generate-TestReport
        
        Write-Success "🎉 API测试完成！"
        
    } catch {
        Write-Error "测试过程中发生错误: $($_.Exception.Message)"
        if ($Verbose) {
            Write-Host $_.Exception.ToString() -ForegroundColor Red
        }
    }
}

# 执行主函数
Main
