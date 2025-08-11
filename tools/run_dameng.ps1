Param(
    [string]$DamengHost = '127.0.0.1',
    [string]$DamengPort = '5236',
    [string]$DamengUser = 'SYSDBA',
    [string]$DamengPassword = 'Dm123456',
    [string]$DamengDatabase = 'DAMENG'
)

Write-Host "[run_dameng] 设置环境变量并构建服务" -ForegroundColor Cyan
$env:DB_MODE = 'dameng'
$env:DAMENG_HOST = $DamengHost
$env:DAMENG_PORT = $DamengPort
$env:DAMENG_USER = $DamengUser
$env:DAMENG_PASSWORD = $DamengPassword
$env:DAMENG_DATABASE = $DamengDatabase

# 可选：减少 MySQL 连接等待
$env:MYSQL_HOST = '127.0.0.1'
$env:MYSQL_PORT = '13306'

Write-Host "[run_dameng] Go build ..." -ForegroundColor Cyan
go build -o andorralee_app.exe ./cmd
if ($LASTEXITCODE -ne 0) {
    Write-Error "[run_dameng] 构建失败"
    exit 1
}

Write-Host "[run_dameng] 启动服务 (后台进程)" -ForegroundColor Cyan
Start-Process -FilePath (Join-Path (Get-Location) 'andorralee_app.exe') -WindowStyle Hidden

Start-Sleep -Seconds 3

Write-Host "[run_dameng] 健康检查请求 http://localhost:9090/api/v1/health" -ForegroundColor Cyan
try {
    $resp = Invoke-RestMethod -Uri 'http://localhost:9090/api/v1/health' -TimeoutSec 5
    $resp | ConvertTo-Json -Depth 6
} catch {
    Write-Warning "[run_dameng] 健康检查失败: $($_.Exception.Message)"
    exit 2
}

Write-Host "[run_dameng] 完成" -ForegroundColor Green
