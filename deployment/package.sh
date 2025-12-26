#!/bin/bash

# 打包脚本 - 在Mac上执行
# 用途：打包整个部署目录，准备传输到目标服务器

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "${SCRIPT_DIR}")"
PACKAGE_NAME="andorralee-deploy-$(date +%Y%m%d-%H%M%S).tar.gz"
TEMP_DIR="/tmp/andorralee-package"

echo "========================================="
echo "  Andorralee 打包脚本"
echo "========================================="
echo ""

# 清理临时目录
rm -rf "${TEMP_DIR}"
mkdir -p "${TEMP_DIR}/deployment"

echo "[1/9] 构建后端二进制 (linux/amd64) ..."
cd "${PROJECT_DIR}"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o build/andorralee ./cmd/main.go
echo "✓ 后端二进制构建完成"
echo ""

echo "[2/9] 构建后端镜像 andorralee/backend:latest (离线, amd64) ..."
docker buildx build --platform linux/amd64 -f deployment/Dockerfile.backend -t andorralee/backend:latest --load .
echo "✓ 后端镜像构建完成 (linux/amd64)"
echo ""

echo "[3/9] 导出Docker镜像..."
cd "${SCRIPT_DIR}"
bash export_images.sh
echo ""

echo "[4/9] 导出数据库..."
bash export_database.sh
echo ""

echo "[5/9] 复制部署文件（包含导出的镜像与数据）..."
cp -r "${SCRIPT_DIR}"/* "${TEMP_DIR}/deployment/"
echo "✓ 部署文件已复制"
echo ""

echo "[6/9] 复制前端文件..."
mkdir -p "${TEMP_DIR}/frontend"
cp -r "${PROJECT_DIR}/frontend"/* "${TEMP_DIR}/frontend/"
echo "✓ 前端文件已复制"
echo ""

echo "[7/9] 复制Go项目文件..."
cp "${PROJECT_DIR}/go.mod" "${TEMP_DIR}/"
cp "${PROJECT_DIR}/go.sum" "${TEMP_DIR}/"
mkdir -p "${TEMP_DIR}/cmd"
cp -r "${PROJECT_DIR}/cmd"/* "${TEMP_DIR}/cmd/"
mkdir -p "${TEMP_DIR}/internal"
cp -r "${PROJECT_DIR}/internal"/* "${TEMP_DIR}/internal/"
mkdir -p "${TEMP_DIR}/pkg"
cp -r "${PROJECT_DIR}/pkg"/* "${TEMP_DIR}/pkg/"
mkdir -p "${TEMP_DIR}/routers"
cp -r "${PROJECT_DIR}/routers"/* "${TEMP_DIR}/routers/"
echo "✓ Go项目文件已复制"
echo ""

echo "[8/9] 创建部署包..."
cd /tmp
tar -czf "${PROJECT_DIR}/${PACKAGE_NAME}" andorralee-package/
echo "✓ 部署包已创建"
echo ""

echo "[9/9] 清理临时文件..."
rm -rf "${TEMP_DIR}"
echo "✓ 清理完成"
echo ""

echo "========================================="
echo "  打包完成！"
echo "========================================="
echo ""
echo "部署包: ${PROJECT_DIR}/${PACKAGE_NAME}"
ls -lh "${PROJECT_DIR}/${PACKAGE_NAME}"
echo ""
echo "传输到目标服务器:"
echo "  scp ${PACKAGE_NAME} httc@192.168.1.21:~/"
echo ""
echo "在目标服务器上执行:"
echo "  tar -xzf ${PACKAGE_NAME}"
echo "  cd andorralee-package/deployment"
echo "  bash deploy.sh"
echo ""
