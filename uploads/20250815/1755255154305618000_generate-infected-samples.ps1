Param(
    [int[]]$SizesMB = @(5,6,7,8,10)
)

# 生成包含 EICAR 测试串的“已感染”样本文件（安全，无真实恶意代码）
$ErrorActionPreference = 'Stop'
$eicar = 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'

function New-InfectedFile {
    Param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)][int]$SizeMB,
        [ValidateSet('start','middle','end')][string]$PlaceAt = 'start'
    )

    $total = [long]$SizeMB * 1MB
    $eicarBytes = [System.Text.Encoding]::ASCII.GetBytes($eicar)

    $dir = [System.IO.Path]::GetDirectoryName($Path)
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }

    $fs = [System.IO.File]::Open($Path, [System.IO.FileMode]::Create, [System.IO.FileAccess]::Write, [System.IO.FileShare]::None)
    try {
        $blockSize = 1024 * 1024
        $block = New-Object byte[] $blockSize
        for ($i=0; $i -lt $block.Length; $i++) { $block[$i] = 0x41 } # 'A'

        $writePad = {
            param($bytesToWrite)
            $remaining = [long]$bytesToWrite
            while ($remaining -gt 0) {
                $toWrite = [int]([System.Math]::Min($blockSize, $remaining))
                $fs.Write($block, 0, $toWrite)
                $remaining -= $toWrite
            }
        }

        switch ($PlaceAt) {
            'start' {
                # 开头写入 EICAR
                $fs.Write($eicarBytes, 0, $eicarBytes.Length)
                $pad = $total - [long]$eicarBytes.Length
                & $writePad $pad
            }
            'middle' {
                # 先写前半填充，再写 EICAR，然后补足剩余
                $front = [long]([System.Math]::Max(0, ($total - $eicarBytes.Length) / 2))
                & $writePad $front
                $fs.Write($eicarBytes, 0, $eicarBytes.Length)
                $tail = $total - $front - [long]$eicarBytes.Length
                & $writePad $tail
            }
            'end' {
                # 先写填充，最后写 EICAR
                $pad = $total - [long]$eicarBytes.Length
                & $writePad $pad
                $fs.Write($eicarBytes, 0, $eicarBytes.Length)
            }
        }
    }
    finally {
        $fs.Close()
    }

    $md5 = (Get-FileHash -Algorithm MD5 -Path $Path).Hash
    $sha = (Get-FileHash -Algorithm SHA256 -Path $Path).Hash
    Write-Host ("Created: {0}  Size: {1} MB  MD5: {2}  SHA256: {3}" -f $Path, $SizeMB, $md5, $sha)
}

# 生成多种位置与扩展名的样本
$samples = @(
    @{ Name = 'infected-eicar-start-5mb.txt';   Size = $SizesMB[0]; Pos = 'start' },
    @{ Name = 'infected-eicar-middle-6mb.bin';  Size = $SizesMB[1]; Pos = 'middle' },
    @{ Name = 'infected-eicar-end-7mb.bin';     Size = $SizesMB[2]; Pos = 'end' },
    @{ Name = 'infected-eicar-middle-8mb.jpg';  Size = $SizesMB[3]; Pos = 'middle' },
    @{ Name = 'infected-eicar-start-10mb.exe';  Size = $SizesMB[4]; Pos = 'start' }
)

foreach ($s in $samples) {
    $path = Join-Path $PSScriptRoot $s.Name
    New-InfectedFile -Path $path -SizeMB $s.Size -PlaceAt $s.Pos
}

Write-Host "Done."
