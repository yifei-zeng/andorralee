# 🐳 Docker镜像管理测试脚本
# 测试Andorralee蜜罐系统的镜像管理功能

param(
    [string]$BaseUrl = "http://localhost:8081/api/v1",
    [switch]$Verbose
)

# 设置请求头
$Headers = @{
    "Content-Type" = "application/json"
    "Accept" = "application/json"
}

# 辅助函数：发送API请求
function Invoke-ApiRequest {
    param(
        [string]$Endpoint,
        [string]$Method = "GET",
        [hashtable]$Body = $null
    )
    
    $Uri = "$BaseUrl$Endpoint"
    $RequestParams = @{
        Uri = $Uri
        Method = $Method
        Headers = $Headers
    }
    
    if ($Body) {
        $RequestParams.Body = $Body | ConvertTo-Json -Depth 3
    }
    
    try {
        if ($Verbose) {
            Write-Host "🔄 $Method $Endpoint" -ForegroundColor Cyan
        }
        
        $Response = Invoke-WebRequest @RequestParams
        $Content = $Response.Content | ConvertFrom-Json
        
        if ($Verbose) {
            Write-Host "✅ 成功: $($Content.message)" -ForegroundColor Green
        }
        
        return $Content
    }
    catch {
        Write-Host "❌ 失败: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# 测试函数
function Test-HealthCheck {
    Write-Host "`n🏥 测试健康检查..." -ForegroundColor Yellow
    $result = Invoke-ApiRequest -Endpoint "/health"
    if ($result) {
        Write-Host "✅ 系统健康状态正常" -ForegroundColor Green
        return $true
    }
    return $false
}

function Test-DockerStatus {
    Write-Host "`n🐳 测试Docker状态..." -ForegroundColor Yellow
    $result = Invoke-ApiRequest -Endpoint "/docker/info"
    if ($result) {
        Write-Host "✅ Docker服务正常" -ForegroundColor Green
        return $true
    }
    return $false
}

function Test-PullImage {
    param([string]$ImageName, [string]$Tag = "latest")
    
    Write-Host "`n📥 测试拉取镜像: $ImageName:$Tag..." -ForegroundColor Yellow
    
    $body = @{
        image_name = $ImageName
        tag = $Tag
    }
    
    $result = Invoke-ApiRequest -Endpoint "/docker/pull" -Method "POST" -Body $body
    if ($result -and $result.status -eq "success") {
        Write-Host "✅ 镜像拉取成功" -ForegroundColor Green
        return $true
    }
    return $false
}

function Test-ListImages {
    Write-Host "`n📋 测试获取镜像列表..." -ForegroundColor Yellow
    
    $result = Invoke-ApiRequest -Endpoint "/docker/images"
    if ($result -and $result.status -eq "success") {
        $images = $result.data
        Write-Host "✅ 获取到 $($images.Count) 个镜像" -ForegroundColor Green
        
        if ($Verbose -and $images.Count -gt 0) {
            Write-Host "`n📋 镜像列表:" -ForegroundColor Cyan
            foreach ($image in $images) {
                $repoTags = $image.RepoTags -join ", "
                $size = [math]::Round($image.Size / 1MB, 2)
                $created = [DateTime]::new(1970,1,1,0,0,0,0,[DateTimeKind]::Utc).AddSeconds($image.Created).ToLocalTime()
                
                Write-Host "  🖼️  $repoTags" -ForegroundColor White
                Write-Host "      ID: $($image.Id.Substring(7, 12))..." -ForegroundColor Gray
                Write-Host "      大小: $size MB" -ForegroundColor Gray
                Write-Host "      创建: $($created.ToString('yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Gray
                Write-Host ""
            }
        }
        return $images
    }
    return $null
}

function Test-CreateContainer {
    param([string]$ImageName)
    
    Write-Host "`n🏗️  测试创建容器实例..." -ForegroundColor Yellow
    
    $containerName = "test-container-$(Get-Date -Format 'HHmmss')"
    
    $body = @{
        name = $containerName
        honeypot_name = $containerName
        image_name = $ImageName
        protocol = "http"
        interface_type = "network"
        port_mappings = @{"80" = "8080"}
        environment = @{"TEST" = "true"}
        description = "测试容器实例"
        auto_start = $false
    }
    
    $result = Invoke-ApiRequest -Endpoint "/container-instances" -Method "POST" -Body $body
    if ($result -and $result.status -eq "success") {
        Write-Host "✅ 容器创建成功: $containerName" -ForegroundColor Green
        return $result.data
    }
    return $null
}

function Test-ContainerOperations {
    param([int]$ContainerId)
    
    Write-Host "`n🔧 测试容器操作..." -ForegroundColor Yellow
    
    # 启动容器
    Write-Host "  🚀 启动容器..." -ForegroundColor Cyan
    $result = Invoke-ApiRequest -Endpoint "/container-instances/$ContainerId/start" -Method "POST"
    if ($result -and $result.status -eq "success") {
        Write-Host "  ✅ 容器启动成功" -ForegroundColor Green
    }
    
    Start-Sleep -Seconds 2
    
    # 获取容器状态
    Write-Host "  📊 获取容器状态..." -ForegroundColor Cyan
    $result = Invoke-ApiRequest -Endpoint "/container-instances/$ContainerId/status"
    if ($result) {
        Write-Host "  ✅ 容器状态: $($result.data.status)" -ForegroundColor Green
    }
    
    Start-Sleep -Seconds 2
    
    # 停止容器
    Write-Host "  🛑 停止容器..." -ForegroundColor Cyan
    $result = Invoke-ApiRequest -Endpoint "/container-instances/$ContainerId/stop" -Method "POST"
    if ($result -and $result.status -eq "success") {
        Write-Host "  ✅ 容器停止成功" -ForegroundColor Green
    }
}

# 主测试流程
function Start-ImageManagementTest {
    Write-Host "🎯 开始Docker镜像管理功能测试" -ForegroundColor Magenta
    Write-Host "目标服务器: $BaseUrl" -ForegroundColor Magenta
    Write-Host "=" * 50
    
    # 1. 健康检查
    if (-not (Test-HealthCheck)) {
        Write-Host "❌ 系统健康检查失败，停止测试" -ForegroundColor Red
        return
    }
    
    # 2. Docker状态检查
    if (-not (Test-DockerStatus)) {
        Write-Host "⚠️  Docker状态检查失败，但继续测试" -ForegroundColor Yellow
    }
    
    # 3. 获取当前镜像列表
    $currentImages = Test-ListImages
    
    # 4. 测试拉取镜像
    $testImages = @(
        @{name="hello-world"; tag="latest"},
        @{name="nginx"; tag="latest"}
    )
    
    $pulledImages = @()
    foreach ($img in $testImages) {
        if (Test-PullImage -ImageName $img.name -Tag $img.tag) {
            $pulledImages += "$($img.name):$($img.tag)"
        }
    }
    
    # 5. 再次获取镜像列表，验证拉取结果
    Write-Host "`n🔄 验证镜像拉取结果..." -ForegroundColor Yellow
    $newImages = Test-ListImages
    
    # 6. 测试创建容器（使用第一个成功拉取的镜像）
    if ($pulledImages.Count -gt 0) {
        $testImage = $pulledImages[0]
        $container = Test-CreateContainer -ImageName $testImage
        
        if ($container -and $container.id) {
            Test-ContainerOperations -ContainerId $container.id
        }
    }
    
    # 7. 测试总结
    Write-Host "`n📊 测试总结" -ForegroundColor Magenta
    Write-Host "=" * 30
    Write-Host "✅ 成功拉取镜像: $($pulledImages.Count) 个" -ForegroundColor Green
    if ($pulledImages.Count -gt 0) {
        $pulledImages | ForEach-Object { Write-Host "   - $_" -ForegroundColor White }
    }
    
    if ($newImages) {
        Write-Host "📋 当前镜像总数: $($newImages.Count) 个" -ForegroundColor Cyan
    }
    
    Write-Host "`n🎉 测试完成！" -ForegroundColor Green
    Write-Host "💡 现在可以访问调试界面测试镜像管理功能:" -ForegroundColor Yellow
    Write-Host "   $($BaseUrl.Replace('/api/v1', ''))/static/debug-container.html" -ForegroundColor Cyan
}

# 执行测试
Start-ImageManagementTest
