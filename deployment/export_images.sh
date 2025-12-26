#!/bin/bash

# 导出Docker镜像脚本
# 用途：在本地Mac上导出所有需要的Docker镜像（包含后端镜像）

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGES_DIR="${SCRIPT_DIR}/images"

echo "=== 开始导出Docker镜像 ==="
echo "导出目录: ${IMAGES_DIR}"

# 创建镜像目录
mkdir -p "${IMAGES_DIR}"

# 导出后端镜像
echo ""
echo "[1/6] 导出 andorralee/backend:latest ..."
docker save andorralee/backend:latest -o "${IMAGES_DIR}/backend-latest.tar"
gzip -f "${IMAGES_DIR}/backend-latest.tar"
echo "✓ 已导出: backend-latest.tar.gz"

# 导出蜜罐镜像
echo ""
echo "[2/6] 导出 andorralee/cowrie:v0.1 ..."
docker save andorralee/cowrie:v0.1 -o "${IMAGES_DIR}/cowrie-v0.1.tar"
gzip -f "${IMAGES_DIR}/cowrie-v0.1.tar"
echo "✓ 已导出: cowrie-v0.1.tar.gz"

echo ""
echo "[3/6] 导出 andorralee/heralding:v0.1 ..."
docker save andorralee/heralding:v0.1 -o "${IMAGES_DIR}/heralding-v0.1.tar"
gzip -f "${IMAGES_DIR}/heralding-v0.1.tar"
echo "✓ 已导出: heralding-v0.1.tar.gz"

echo ""
echo "[4/6] 导出 andorralee/dionaea:v1 ..."
docker save andorralee/dionaea:v1 -o "${IMAGES_DIR}/dionaea-v1.tar"
gzip -f "${IMAGES_DIR}/dionaea-v1.tar"
echo "✓ 已导出: dionaea-v1.tar.gz"

echo ""
echo "[5/6] 导出 andorralee/dbdefende-r2:v1.0 ..."
docker save andorralee/dbdefende-r2:v1.0 -o "${IMAGES_DIR}/dbdefende-r2-v1.0.tar"
gzip -f "${IMAGES_DIR}/dbdefende-r2-v1.0.tar"
echo "✓ 已导出: dbdefende-r2-v1.0.tar.gz"

echo ""
echo "[6/6] 导出 mysql:8.0 (amd64) ..."
docker save mysql:8.0 -o "${IMAGES_DIR}/mysql-8.0.tar"
gzip -f "${IMAGES_DIR}/mysql-8.0.tar"
echo "✓ 已导出: mysql-8.0.tar.gz"

echo ""
echo "=== 镜像导出完成 ==="
echo "所有镜像已保存到: ${IMAGES_DIR}"
echo ""
ls -lh "${IMAGES_DIR}"/*.tar.gz
echo ""
echo "提示: 请将整个 deployment 目录打包传输到目标服务器"
