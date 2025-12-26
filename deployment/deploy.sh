#!/bin/bash

# Andorralee系统部署脚本
# 用途：在目标Ubuntu服务器上部署完整的Andorralee蜜罐系统
# 使用方法：bash deploy.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGES_DIR="${SCRIPT_DIR}/images"

echo "========================================="
echo "  Andorralee 蜜罐系统部署脚本"
echo "========================================="
echo ""

# 检查是否为root用户
if [ "$EUID" -eq 0 ]; then 
    echo "警告: 请不要使用root用户运行此脚本"
    echo "建议: 使用普通用户并确保该用户在docker组中"
    exit 1
fi

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装"
    echo "请先安装Docker: sudo apt-get update && sudo apt-get install -y docker.io docker-compose"
    exit 1
fi

COMPOSE_CMD=""
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
else
    echo "错误: Docker Compose 未安装"
    echo "请先安装 docker compose 插件或 docker-compose 工具"
    exit 1
fi

# 检查当前用户是否在docker组
if ! groups | grep -q docker; then
    echo "警告: 当前用户不在docker组中"
    echo "请运行: sudo usermod -aG docker $USER"
    echo "然后重新登录后再运行此脚本"
    exit 1
fi

echo "[1/6] 检查环境... ✓"
echo ""

# 导入Docker镜像
if [ -d "${IMAGES_DIR}" ]; then
    echo "[2/6] 导入Docker镜像..."
    
    for image_file in "${IMAGES_DIR}"/*.tar.gz; do
        if [ -f "$image_file" ]; then
            # 跳过 MySQL 镜像，稍后直接从 Docker Hub 拉取
            if [[ "$image_file" == *"mysql-8.0"* ]]; then
                echo "  跳过: $(basename $image_file) (将从 Docker Hub 拉取)"
                continue
            fi
            echo "  导入: $(basename $image_file)"
            gunzip -c "$image_file" | docker load
        fi
    done
    
    # 直接拉取 MySQL 8.0 amd64 镜像
    echo "  拉取: mysql:8.0 (linux/amd64)..."
    docker pull --platform=linux/amd64 mysql:8.0 || {
        echo "警告: 无法拉取 MySQL 镜像，尝试导入本地文件..."
        if [ -f "${IMAGES_DIR}/mysql-8.0.tar.gz" ]; then
            gunzip -c "${IMAGES_DIR}/mysql-8.0.tar.gz" | docker load
        fi
    }
    
    echo "✓ 镜像导入完成"
    echo ""
else
    echo "错误: 镜像目录不存在: ${IMAGES_DIR}"
    exit 1
fi

echo "[3/6] 检查导入的镜像..."
docker images | grep -E "andorralee|mysql"
echo ""

echo "[4/6] 清理旧容器..."
cd "${SCRIPT_DIR}"
${COMPOSE_CMD} -f docker-compose.deploy.yml down -v 2>/dev/null || true
echo "✓ 清理完成"
echo ""

# 导入数据库数据（如果存在）
if [ -f "${SCRIPT_DIR}/andorralee_data.sql.gz" ]; then
    echo "[5/6] 准备数据库数据..."
    gunzip -c "${SCRIPT_DIR}/andorralee_data.sql.gz" > "${SCRIPT_DIR}/andorralee_data.sql"
    echo "✓ 数据文件已准备"
else
    echo "[5/6] 跳过数据导入（无数据文件）"
fi
echo ""

# 启动服务
echo "[6/6] 启动Andorralee系统..."
${COMPOSE_CMD} -f docker-compose.deploy.yml up -d

echo ""
echo "等待服务启动..."
sleep 10

# 如果有数据文件，导入数据
if [ -f "${SCRIPT_DIR}/andorralee_data.sql" ]; then
    echo ""
    echo "导入数据库数据..."
    docker exec -i andorralee-mysql mysql -uroot -pmac123456 andorralee < "${SCRIPT_DIR}/andorralee_data.sql"
    echo "✓ 数据导入完成"
    rm -f "${SCRIPT_DIR}/andorralee_data.sql"
fi

echo ""
echo "========================================="
echo "  部署完成！"
echo "========================================="
echo ""
echo "服务状态:"
${COMPOSE_CMD} -f docker-compose.deploy.yml ps
echo ""
echo "访问信息:"
echo "  后端API: http://$(hostname -I | awk '{print $1}'):8848"
echo "  前端页面: http://$(hostname -I | awk '{print $1}'):8848/index.html"
echo ""
echo "数据库信息:"
echo "  地址: localhost:3306"
echo "  用户: root"
echo "  密码: mac123456"
echo "  数据库: andorralee"
echo ""
echo "蜜罐容器:"
echo "  shellguard-1 (Cowrie SSH):    端口 2222, 2223"
echo "  shellguard-2 (Heralding):     端口 21, 23, 80, 110, 143, 443, 1080, 3389"
echo "  shellguard-3 (Dionaea):       端口 445, 135, 1433, 3306, 5060"
echo "  shellguard-4 (MySQL蜜罐):     端口 3307"
echo ""
echo "常用命令:"
echo "  查看日志: ${COMPOSE_CMD} -f deployment/docker-compose.deploy.yml logs -f"
echo "  停止服务: ${COMPOSE_CMD} -f deployment/docker-compose.deploy.yml stop"
echo "  启动服务: ${COMPOSE_CMD} -f deployment/docker-compose.deploy.yml start"
echo "  重启服务: ${COMPOSE_CMD} -f deployment/docker-compose.deploy.yml restart"
echo "  删除服务: ${COMPOSE_CMD} -f deployment/docker-compose.deploy.yml down -v"
echo ""
