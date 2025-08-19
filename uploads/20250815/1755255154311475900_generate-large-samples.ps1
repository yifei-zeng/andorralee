Param(
    [int[]]$SizesMB = @(5,6,7,8,10)
)

# 安全测试用：生成包含 EICAR 测试串且体积>=5MB 的样本文件
# 这些文件仅用于触发检测流程，不包含真实恶意代码

$ErrorActionPreference = 'Stop'
$eicar = 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'

function New-LargeEicarFile {
    Param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)][int]$SizeMB
    )
    $total = [long]$SizeMB * 1MB
    $eicarBytes = [System.Text.Encoding]::ASCII.GetBytes($eicar)

    $dir = [System.IO.Path]::GetDirectoryName($Path)
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }

    $fs = [System.IO.File]::Open($Path, [System.IO.FileMode]::Create, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
    try {
        # 开头写入 EICAR 测试串
        $fs.Write($eicarBytes, 0, $eicarBytes.Length)

        # 用 'A' 填充直至达到目标大小
        $blockSize = 1024 * 1024
        $block = New-Object byte[] $blockSize
        for ($i=0; $i -lt $block.Length; $i++) { $block[$i] = 0x41 }

        $remaining = $total - [long]$eicarBytes.Length
        while ($remaining -gt 0) {
            $toWrite = [int]([System.Math]::Min($blockSize, $remaining))
            $fs.Write($block, 0, $toWrite)
            $remaining -= $toWrite
        }
    }
    finally {
        $fs.Close()
    }

    # 输出校验信息
    $md5 = (Get-FileHash -Algorithm MD5 -Path $Path).Hash
    $sha = (Get-FileHash -Algorithm SHA256 -Path $Path).Hash
    Write-Host ("Created: {0}  Size: {1} MB  MD5: {2}  SHA256: {3}" -f $Path, $SizeMB, $md5, $sha)
}

# 生成 5 个样本
$names = @(
    'eicar-large-5mb.txt',
    'eicar-large-6mb.bin',
    'eicar-large-7mb.txt',
    'eicar-large-8mb.bin',
    'eicar-large-10mb.bin'
)

for ($i=0; $i -lt $names.Count; $i++) {
    $name = $names[$i]
    $size = $SizesMB[[System.Math]::Min($i, $SizesMB.Length-1)]
    New-LargeEicarFile -Path (Join-Path $PSScriptRoot $name) -SizeMB $size
}

Write-Host "Done."
