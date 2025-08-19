# 生成 EICAR 标准测试文件（非真正病毒）。
# 某些杀软会拦截生成过程，如遇拦截请临时允许。

$eicar = 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'
Set-Content -Path "$PSScriptRoot/eicar.com.txt" -Value $eicar -NoNewline -Encoding ASCII
Write-Host "已生成: $PSScriptRoot/eicar.com.txt"
