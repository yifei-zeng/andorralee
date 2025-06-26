# 银河麒麟系统Docker部署脚本 (PowerShell版本)
# Andorralee蜜罐管理系统 - 银河麒麟V10专用部署脚本
# 适用于Windows环境下远程部署到银河麒麟系统

param(
    [string]$KylinHost = "localhost",
    [string]$KylinUser = "root",
    [string]$KylinPassword = "",
    [switch]$UseSSHKey = $false,
    [string]$SSHKeyPath = "",
    [switch]$LocalDeploy = $false,
    [switch]$Help = $false
)

# 显示帮助信息
if ($Help) {
    Write-Host @"
银河麒麟系统Docker部署脚本 (PowerShell版本)

用法:
    .\deploy-kylin.ps1 [参数]

参数:
    -KylinHost <主机地址>     银河麒麟系统的IP地址或主机名 (默认: localhost)
    -KylinUser <用户名>       SSH登录用户名 (默认: root)
    -KylinPassword <密码>     SSH登录密码
    -UseSSHKey               使用SSH密钥认证
    -SSHKeyPath <密钥路径>    SSH私钥文件路径
    -LocalDeploy             本地部署模式 (在Windows Docker Desktop上运行)
    -Help                    显示此帮助信息

示例:
    # 使用密码认证部署到远程银河麒麟系统
    .\deploy-kylin.ps1 -KylinHost "192.168.1.100" -KylinUser "admin" -KylinPassword "password"
    
    # 使用SSH密钥认证
    .\deploy-kylin.ps1 -KylinHost "192.168.1.100" -KylinUser "admin" -UseSSHKey -SSHKeyPath "C:\Users\user\.ssh\id_rsa"
    
    # 本地部署 (Windows Docker Desktop)
    .\deploy-kylin.ps1 -LocalDeploy

"@
    exit 0
}

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] " -NoNewline -ForegroundColor Gray
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info { param([string]$Message) Write-ColorOutput $Message "Green" }
function Write-Warn { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput $Message "Red" }
function Write-Debug { param([string]$Message) Write-ColorOutput $Message "Cyan" }

# 检查PowerShell版本
function Test-PowerShellVersion {
    Write-Info "检查PowerShell版本..."
    
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "需要PowerShell 5.0或更高版本"
        exit 1
    }
    
    Write-Info "PowerShell版本: $($PSVersionTable.PSVersion)"
}

# 检查必要的模块和工具
function Test-Prerequisites {
    Write-Info "检查必要的工具..."
    
    # 检查Docker (本地部署模式)
    if ($LocalDeploy) {
        try {
            $dockerVersion = docker --version
            Write-Info "Docker已安装: $dockerVersion"
        }
        catch {
            Write-Error "Docker未安装或不可用，请安装Docker Desktop"
            exit 1
        }
        
        try {
            $composeVersion = docker-compose --version
            Write-Info "Docker Compose已安装: $composeVersion"
        }
        catch {
            Write-Error "Docker Compose未安装或不可用"
            exit 1
        }
    }
    
    # 检查SSH客户端 (远程部署模式)
    if (-not $LocalDeploy) {
        try {
            $sshVersion = ssh -V 2>&1
            Write-Info "SSH客户端可用: $sshVersion"
        }
        catch {
            Write-Error "SSH客户端不可用，请安装OpenSSH客户端"
            exit 1
        }
    }
}

# 创建SSH连接字符串
function Get-SSHConnection {
    if ($UseSSHKey -and $SSHKeyPath) {
        return "ssh -i `"$SSHKeyPath`" $KylinUser@$KylinHost"
    }
    else {
        return "ssh $KylinUser@$KylinHost"
    }
}

# 执行远程命令
function Invoke-RemoteCommand {
    param(
        [string]$Command,
        [switch]$IgnoreError = $false
    )
    
    if ($LocalDeploy) {
        Write-Debug "本地执行: $Command"
        Invoke-Expression $Command
    }
    else {
        $sshCmd = Get-SSHConnection
        $fullCommand = "$sshCmd `"$Command`""
        Write-Debug "远程执行: $Command"
        
        if ($IgnoreError) {
            Invoke-Expression $fullCommand 2>$null
        }
        else {
            Invoke-Expression $fullCommand
        }
    }
}

# 上传文件到远程系统
function Copy-FileToRemote {
    param(
        [string]$LocalPath,
        [string]$RemotePath
    )
    
    if ($LocalDeploy) {
        Copy-Item $LocalPath $RemotePath -Force
    }
    else {
        if ($UseSSHKey -and $SSHKeyPath) {
            $scpCmd = "scp -i `"$SSHKeyPath`" `"$LocalPath`" $KylinUser@$KylinHost:`"$RemotePath`""
        }
        else {
            $scpCmd = "scp `"$LocalPath`" $KylinUser@$KylinHost:`"$RemotePath`""
        }
        
        Write-Debug "上传文件: $LocalPath -> $RemotePath"
        Invoke-Expression $scpCmd
    }
}

# 检测远程系统信息
function Get-RemoteSystemInfo {
    Write-Info "检测远程系统信息..."
    
    # 检测操作系统
    $osInfo = Invoke-RemoteCommand "cat /etc/kylin-release 2>/dev/null || cat /etc/os-release 2>/dev/null || echo 'Unknown'"
    Write-Info "操作系统信息: $osInfo"
    
    # 检测架构
    $arch = Invoke-RemoteCommand "uname -m"
    Write-Info "系统架构: $arch"
    
    # 检测内核版本
    $kernel = Invoke-RemoteCommand "uname -r"
    Write-Info "内核版本: $kernel"
    
    # 检测Docker
    $dockerStatus = Invoke-RemoteCommand "docker --version 2>/dev/null || echo 'Not installed'" -IgnoreError
    Write-Info "Docker状态: $dockerStatus"
    
    # 检测Docker Compose
    $composeStatus = Invoke-RemoteCommand "docker-compose --version 2>/dev/null || echo 'Not installed'" -IgnoreError
    Write-Info "Docker Compose状态: $composeStatus"
}

# 创建远程目录结构
function New-RemoteDirectories {
    Write-Info "创建远程目录结构..."
    
    $directories = @(
        "andorralee-kylin",
        "andorralee-kylin/data/mysql",
        "andorralee-kylin/data/dameng", 
        "andorralee-kylin/data/redis",
        "andorralee-kylin/logs",
        "andorralee-kylin/config",
        "andorralee-kylin/mysql/conf.d",
        "andorralee-kylin/dameng/scripts",
        "andorralee-kylin/redis",
        "andorralee-kylin/scripts"
    )
    
    foreach ($dir in $directories) {
        Invoke-RemoteCommand "mkdir -p $dir"
    }
    
    Write-Info "目录结构创建完成"
}

# 上传项目文件
function Copy-ProjectFiles {
    Write-Info "上传项目文件..."
    
    $filesToCopy = @(
        @{ Local = "Dockerfile.kylin"; Remote = "andorralee-kylin/Dockerfile.kylin" },
        @{ Local = "docker-compose.kylin.yml"; Remote = "andorralee-kylin/docker-compose.yml" },
        @{ Local = "deploy-kylin.sh"; Remote = "andorralee-kylin/deploy.sh" },
        @{ Local = "go.mod"; Remote = "andorralee-kylin/go.mod" },
        @{ Local = "go.sum"; Remote = "andorralee-kylin/go.sum" }
    )
    
    foreach ($file in $filesToCopy) {
        if (Test-Path $file.Local) {
            Copy-FileToRemote $file.Local $file.Remote
            Write-Debug "已上传: $($file.Local)"
        }
        else {
            Write-Warn "文件不存在: $($file.Local)"
        }
    }
    
    # 上传源代码目录
    $sourceDirs = @("cmd", "internal", "pkg", "routers", "static", "scripts")
    foreach ($dir in $sourceDirs) {
        if (Test-Path $dir) {
            if ($LocalDeploy) {
                Copy-Item $dir "andorralee-kylin/" -Recurse -Force
            }
            else {
                $tarCmd = "tar -czf $dir.tar.gz $dir"
                Invoke-Expression $tarCmd
                Copy-FileToRemote "$dir.tar.gz" "andorralee-kylin/$dir.tar.gz"
                Invoke-RemoteCommand "cd andorralee-kylin && tar -xzf $dir.tar.gz && rm $dir.tar.gz"
                Remove-Item "$dir.tar.gz" -Force
            }
            Write-Debug "已上传目录: $dir"
        }
    }
    
    Write-Info "项目文件上传完成"
}

# 执行远程部署
function Start-RemoteDeployment {
    Write-Info "开始远程部署..."
    
    # 设置执行权限
    Invoke-RemoteCommand "chmod +x andorralee-kylin/deploy.sh"
    
    # 执行部署脚本
    Invoke-RemoteCommand "cd andorralee-kylin && ./deploy.sh"
    
    Write-Info "远程部署完成"
}

# 本地部署
function Start-LocalDeployment {
    Write-Info "开始本地部署..."
    
    # 检查文件是否存在
    if (-not (Test-Path "docker-compose.kylin.yml")) {
        Write-Error "docker-compose.kylin.yml文件不存在"
        exit 1
    }
    
    # 创建本地目录
    New-Item -ItemType Directory -Path "data/mysql", "data/dameng", "data/redis", "logs", "config" -Force | Out-Null
    
    # 构建镜像
    Write-Info "构建Docker镜像..."
    docker build -f Dockerfile.kylin -t andorralee:kylin-latest .
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "镜像构建失败"
        exit 1
    }
    
    # 启动服务
    Write-Info "启动服务..."
    docker-compose -f docker-compose.kylin.yml up -d
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "服务启动失败"
        exit 1
    }
    
    Write-Info "本地部署完成"
}

# 验证部署结果
function Test-Deployment {
    Write-Info "验证部署结果..."
    
    if ($LocalDeploy) {
        $healthUrl = "http://localhost:8081/api/v1/health"
    }
    else {
        $healthUrl = "http://$KylinHost:8081/api/v1/health"
    }
    
    # 等待服务启动
    $maxAttempts = 30
    $attempt = 0
    
    do {
        $attempt++
        Write-Debug "健康检查尝试 $attempt/$maxAttempts"
        
        try {
            if ($LocalDeploy) {
                $response = Invoke-WebRequest -Uri $healthUrl -TimeoutSec 5 -UseBasicParsing
            }
            else {
                $response = Invoke-RemoteCommand "curl -f $healthUrl" -IgnoreError
            }
            
            if ($response) {
                Write-Info "服务健康检查通过"
                return $true
            }
        }
        catch {
            Write-Debug "健康检查失败: $_"
        }
        
        Start-Sleep -Seconds 10
    } while ($attempt -lt $maxAttempts)
    
    Write-Warn "服务健康检查超时，请手动验证"
    return $false
}

# 显示部署信息
function Show-DeploymentInfo {
    Write-Info "部署完成！"
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host "  Andorralee蜜罐管理系统 - 银河麒麟版" -ForegroundColor Cyan
    Write-Host "==========================================" -ForegroundColor Cyan
    
    if ($LocalDeploy) {
        $host = "localhost"
    }
    else {
        $host = $KylinHost
    }
    
    Write-Host "服务地址:" -ForegroundColor Yellow
    Write-Host "  应用服务: http://$host:8081" -ForegroundColor White
    Write-Host "  MySQL: $host:3306" -ForegroundColor White
    Write-Host "  达梦数据库: $host:5236" -ForegroundColor White
    Write-Host "  Redis: $host:6379" -ForegroundColor White
    Write-Host ""
    Write-Host "管理命令:" -ForegroundColor Yellow
    if ($LocalDeploy) {
        Write-Host "  查看日志: docker-compose -f docker-compose.kylin.yml logs -f" -ForegroundColor White
        Write-Host "  停止服务: docker-compose -f docker-compose.kylin.yml down" -ForegroundColor White
        Write-Host "  重启服务: docker-compose -f docker-compose.kylin.yml restart" -ForegroundColor White
    }
    else {
        $sshCmd = Get-SSHConnection
        Write-Host "  查看日志: $sshCmd 'cd andorralee-kylin && docker-compose logs -f'" -ForegroundColor White
        Write-Host "  停止服务: $sshCmd 'cd andorralee-kylin && docker-compose down'" -ForegroundColor White
        Write-Host "  重启服务: $sshCmd 'cd andorralee-kylin && docker-compose restart'" -ForegroundColor White
    }
    Write-Host "==========================================" -ForegroundColor Cyan
}

# 主函数
function Main {
    Write-Info "开始部署Andorralee蜜罐管理系统 - 银河麒麟版"
    
    try {
        Test-PowerShellVersion
        Test-Prerequisites
        
        if ($LocalDeploy) {
            Start-LocalDeployment
        }
        else {
            Get-RemoteSystemInfo
            New-RemoteDirectories
            Copy-ProjectFiles
            Start-RemoteDeployment
        }
        
        Test-Deployment
        Show-DeploymentInfo
        
        Write-Info "部署流程完成！"
    }
    catch {
        Write-Error "部署过程中发生错误: $_"
        exit 1
    }
}

# 执行主函数
Main
