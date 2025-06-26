#!/bin/bash

# 银河麒麟系统Docker部署测试脚本
# 用于验证Andorralee蜜罐管理系统在银河麒麟系统上的部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "运行测试: $test_name"
    
    if eval "$test_command"; then
        log_info "✅ 测试通过: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        log_error "❌ 测试失败: $test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 检查系统环境
test_system_environment() {
    log_info "检查系统环境..."
    
    run_test "检查银河麒麟系统" "[ -f /etc/kylin-release ]"
    run_test "检查系统架构" "uname -m | grep -E '(x86_64|aarch64)'"
    run_test "检查内核版本" "[ $(uname -r | cut -d. -f1) -ge 4 ]"
    run_test "检查内存大小" "[ $(free -g | awk '/^Mem:/{print $2}') -ge 3 ]"
    run_test "检查磁盘空间" "[ $(df -BG . | awk 'NR==2{print $4}' | sed 's/G//') -ge 10 ]"
}

# 检查Docker环境
test_docker_environment() {
    log_info "检查Docker环境..."
    
    run_test "检查Docker安装" "command -v docker"
    run_test "检查Docker服务状态" "systemctl is-active docker"
    run_test "检查Docker版本" "docker --version | grep -E '(19|20|21|22|23|24)'"
    run_test "检查Docker Compose安装" "command -v docker-compose"
    run_test "检查Docker Compose版本" "docker-compose --version"
    run_test "检查Docker权限" "docker ps"
}

# 检查网络连通性
test_network_connectivity() {
    log_info "检查网络连通性..."
    
    run_test "检查端口8081可用性" "! netstat -tuln | grep ':8081 '"
    run_test "检查端口3306可用性" "! netstat -tuln | grep ':3306 '"
    run_test "检查端口5236可用性" "! netstat -tuln | grep ':5236 '"
    
    # 可选：检查外网连通性
    if ping -c 1 8.8.8.8 &> /dev/null; then
        run_test "检查外网连通性" "true"
        run_test "检查Docker Hub连通性" "docker pull hello-world:latest"
    else
        log_warn "外网不可达，跳过相关测试"
    fi
}

# 测试项目文件
test_project_files() {
    log_info "检查项目文件..."
    
    run_test "检查Dockerfile.kylin" "[ -f Dockerfile.kylin ]"
    run_test "检查docker-compose.kylin.yml" "[ -f docker-compose.kylin.yml ]"
    run_test "检查部署脚本" "[ -f deploy-kylin.sh ]"
    run_test "检查go.mod文件" "[ -f go.mod ]"
    run_test "检查源代码目录" "[ -d cmd ] && [ -d internal ] && [ -d pkg ]"
    run_test "检查静态资源" "[ -d static ]"
    run_test "检查脚本目录" "[ -d scripts ]"
}

# 测试Docker镜像构建
test_docker_build() {
    log_info "测试Docker镜像构建..."
    
    # 设置构建参数
    ARCH=$(uname -m)
    if [ "$ARCH" = "aarch64" ]; then
        PLATFORM="linux/arm64"
    else
        PLATFORM="linux/amd64"
    fi
    
    run_test "构建应用镜像" "docker build --platform $PLATFORM -f Dockerfile.kylin -t andorralee:test-kylin ."
    run_test "检查镜像创建" "docker images | grep 'andorralee.*test-kylin'"
}

# 测试Docker Compose配置
test_docker_compose_config() {
    log_info "测试Docker Compose配置..."
    
    run_test "验证docker-compose配置" "docker-compose -f docker-compose.kylin.yml config"
    run_test "检查服务定义" "docker-compose -f docker-compose.kylin.yml config | grep -E '(andorralee|mysql|dameng)'"
}

# 测试服务启动
test_service_startup() {
    log_info "测试服务启动..."
    
    # 创建必要的目录
    mkdir -p data/{mysql,dameng,redis} logs config mysql/conf.d dameng/scripts redis
    
    # 启动服务
    run_test "启动Docker Compose服务" "docker-compose -f docker-compose.kylin.yml up -d"
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 30
    
    run_test "检查容器状态" "docker-compose -f docker-compose.kylin.yml ps | grep 'Up'"
    run_test "检查应用容器" "docker-compose -f docker-compose.kylin.yml ps andorralee | grep 'Up'"
    run_test "检查MySQL容器" "docker-compose -f docker-compose.kylin.yml ps mysql | grep 'Up'"
}

# 测试服务健康状态
test_service_health() {
    log_info "测试服务健康状态..."
    
    # 等待服务完全启动
    log_info "等待服务完全启动..."
    sleep 60
    
    run_test "检查应用健康状态" "curl -f http://localhost:8081/health"
    run_test "检查API健康状态" "curl -f http://localhost:8081/api/v1/health"
    run_test "检查就绪状态" "curl -f http://localhost:8081/api/v1/ready"
    run_test "检查存活状态" "curl -f http://localhost:8081/api/v1/live"
}

# 测试API功能
test_api_functionality() {
    log_info "测试API功能..."
    
    run_test "测试Docker镜像列表API" "curl -s http://localhost:8081/api/v1/docker/images | grep -E '(\\[|\\]|{|})"
    run_test "测试容器实例API" "curl -s http://localhost:8081/api/v1/container-instances | grep -E '(\\[|\\]|{|})"
    
    # 测试创建容器实例
    local test_payload='{
        "name": "测试容器",
        "honeypot_name": "test-honeypot",
        "image_name": "hello-world:latest",
        "protocol": "http",
        "interface_type": "network",
        "port_mappings": {"80": "8080"},
        "environment": {"TEST": "true"},
        "description": "测试容器实例"
    }'
    
    run_test "测试创建容器实例API" "curl -s -X POST -H 'Content-Type: application/json' -d '$test_payload' http://localhost:8081/api/v1/container-instances | grep -E '(success|id)'"
}

# 测试数据库连接
test_database_connectivity() {
    log_info "测试数据库连接..."
    
    run_test "测试MySQL连接" "docker-compose -f docker-compose.kylin.yml exec -T mysql mysql -u root -pKylin123456! -e 'SELECT 1'"
    
    # 检查达梦数据库（如果可用）
    if docker-compose -f docker-compose.kylin.yml ps dameng | grep -q 'Up'; then
        log_info "达梦数据库容器运行中，测试连接..."
        # 达梦数据库连接测试（根据实际情况调整）
        run_test "检查达梦数据库容器" "docker-compose -f docker-compose.kylin.yml ps dameng | grep 'Up'"
    else
        log_warn "达梦数据库容器未运行，跳过连接测试"
    fi
}

# 清理测试环境
cleanup_test_environment() {
    log_info "清理测试环境..."
    
    # 停止服务
    docker-compose -f docker-compose.kylin.yml down -v || true
    
    # 删除测试镜像
    docker rmi andorralee:test-kylin || true
    
    # 清理测试数据
    rm -rf data logs config mysql dameng redis || true
    
    log_info "测试环境清理完成"
}

# 生成测试报告
generate_test_report() {
    log_info "生成测试报告..."
    
    echo
    echo "=========================================="
    echo "  银河麒麟Docker部署测试报告"
    echo "=========================================="
    echo "测试时间: $(date)"
    echo "系统信息: $(cat /etc/kylin-release 2>/dev/null || echo '未知系统')"
    echo "系统架构: $(uname -m)"
    echo "内核版本: $(uname -r)"
    echo "Docker版本: $(docker --version 2>/dev/null || echo '未安装')"
    echo "Docker Compose版本: $(docker-compose --version 2>/dev/null || echo '未安装')"
    echo
    echo "测试结果:"
    echo "  总测试数: $TOTAL_TESTS"
    echo "  通过测试: $PASSED_TESTS"
    echo "  失败测试: $FAILED_TESTS"
    echo "  成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✅ 所有测试通过！系统可以正常部署。${NC}"
        echo "=========================================="
        return 0
    else
        echo -e "${RED}❌ 有 $FAILED_TESTS 个测试失败，请检查相关问题。${NC}"
        echo "=========================================="
        return 1
    fi
}

# 主函数
main() {
    log_info "开始银河麒麟Docker部署测试"
    
    # 运行测试套件
    test_system_environment
    test_docker_environment
    test_network_connectivity
    test_project_files
    test_docker_build
    test_docker_compose_config
    test_service_startup
    test_service_health
    test_api_functionality
    test_database_connectivity
    
    # 清理环境
    cleanup_test_environment
    
    # 生成报告
    generate_test_report
}

# 捕获退出信号，确保清理
trap cleanup_test_environment EXIT

# 执行主函数
main "$@"
