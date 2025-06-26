#!/bin/bash

# 容器问题修复脚本
# 自动诊断和修复常见的容器实例问题

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
API_BASE="http://localhost:8081/api/v1"
DOCKER_SOCKET="/var/run/docker.sock"

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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖工具..."
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        log_error "curl未安装，请先安装curl"
        exit 1
    fi
    
    # 检查jq
    if ! command -v jq &> /dev/null; then
        log_warn "jq未安装，JSON解析功能受限"
        JQ_AVAILABLE=false
    else
        JQ_AVAILABLE=true
    fi
    
    # 检查docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装"
        exit 1
    fi
    
    log_info "依赖检查完成"
}

# API调用函数
api_call() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="${3:-}"
    
    local url="${API_BASE}${endpoint}"
    local curl_opts="-s -w %{http_code}"
    
    if [ "$method" = "POST" ]; then
        curl_opts="$curl_opts -X POST"
        if [ -n "$data" ]; then
            curl_opts="$curl_opts -H 'Content-Type: application/json' -d '$data'"
        fi
    elif [ "$method" = "DELETE" ]; then
        curl_opts="$curl_opts -X DELETE"
    fi
    
    log_debug "API调用: $method $url"
    
    local response=$(eval "curl $curl_opts '$url'")
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo "$body"
        return 0
    else
        log_error "API调用失败: HTTP $http_code"
        echo "$body" >&2
        return 1
    fi
}

# 检查服务状态
check_service_status() {
    log_info "检查服务状态..."
    
    # 检查API服务
    if api_call "/health" > /dev/null 2>&1; then
        log_info "✅ API服务正常"
    else
        log_error "❌ API服务异常"
        return 1
    fi
    
    # 检查Docker服务
    if docker info > /dev/null 2>&1; then
        log_info "✅ Docker服务正常"
    else
        log_error "❌ Docker服务异常"
        return 1
    fi
    
    # 检查Docker套接字权限
    if [ -S "$DOCKER_SOCKET" ]; then
        if [ -r "$DOCKER_SOCKET" ] && [ -w "$DOCKER_SOCKET" ]; then
            log_info "✅ Docker套接字权限正常"
        else
            log_warn "⚠️ Docker套接字权限可能有问题"
        fi
    else
        log_error "❌ Docker套接字不存在"
        return 1
    fi
    
    return 0
}

# 同步容器状态
sync_container_status() {
    log_info "同步容器状态..."
    
    local response=$(api_call "/container-instances/sync" "POST")
    if [ $? -eq 0 ]; then
        log_info "✅ 容器状态同步成功"
        if [ "$JQ_AVAILABLE" = true ]; then
            echo "$response" | jq '.data.results[] | select(.updated == true) | "更新: \(.name) \(.old_status) -> \(.new_status)"' -r
        fi
    else
        log_error "❌ 容器状态同步失败"
        return 1
    fi
}

# 清理无效容器记录
cleanup_invalid_containers() {
    log_info "清理无效容器记录..."
    
    local containers_response=$(api_call "/container-instances")
    if [ $? -ne 0 ]; then
        log_error "无法获取容器列表"
        return 1
    fi
    
    if [ "$JQ_AVAILABLE" = true ]; then
        # 使用jq解析JSON
        local invalid_containers=$(echo "$containers_response" | jq '.data[] | select(.status == "deleted" or .container_id == "" or .container_id == null) | .id' -r)
        
        if [ -n "$invalid_containers" ]; then
            log_info "发现无效容器记录，准备清理..."
            echo "$invalid_containers" | while read -r container_id; do
                log_info "删除无效容器记录: ID $container_id"
                if api_call "/container-instances/$container_id" "DELETE" > /dev/null 2>&1; then
                    log_info "✅ 删除成功: ID $container_id"
                else
                    log_warn "⚠️ 删除失败: ID $container_id"
                fi
            done
        else
            log_info "没有发现无效容器记录"
        fi
    else
        log_warn "jq不可用，跳过无效容器清理"
    fi
}

# 修复容器ID问题
fix_container_id_issues() {
    log_info "修复容器ID问题..."
    
    local containers_response=$(api_call "/container-instances")
    if [ $? -ne 0 ]; then
        log_error "无法获取容器列表"
        return 1
    fi
    
    if [ "$JQ_AVAILABLE" = true ]; then
        # 查找没有容器ID但有容器名称的记录
        local containers_without_id=$(echo "$containers_response" | jq '.data[] | select(.container_id == "" or .container_id == null) | select(.name != null and .name != "") | {id: .id, name: .name}' -c)
        
        if [ -n "$containers_without_id" ]; then
            log_info "发现缺少容器ID的记录，尝试修复..."
            echo "$containers_without_id" | while read -r container_info; do
                local db_id=$(echo "$container_info" | jq '.id' -r)
                local container_name=$(echo "$container_info" | jq '.name' -r)
                
                log_info "检查容器: $container_name (DB ID: $db_id)"
                
                # 在Docker中查找匹配的容器
                local docker_containers=$(docker ps -a --format "{{.ID}}\t{{.Names}}" | grep "$container_name" || true)
                
                if [ -n "$docker_containers" ]; then
                    local docker_id=$(echo "$docker_containers" | head -1 | cut -f1)
                    log_info "找到匹配的Docker容器: $docker_id"
                    
                    # 这里需要调用更新API（如果存在）
                    log_warn "需要手动更新数据库中的容器ID: $db_id -> $docker_id"
                else
                    log_warn "未找到匹配的Docker容器: $container_name"
                fi
            done
        else
            log_info "所有容器记录都有有效的容器ID"
        fi
    else
        log_warn "jq不可用，跳过容器ID修复"
    fi
}

# 重启异常容器
restart_failed_containers() {
    log_info "重启异常容器..."
    
    local containers_response=$(api_call "/container-instances")
    if [ $? -ne 0 ]; then
        log_error "无法获取容器列表"
        return 1
    fi
    
    if [ "$JQ_AVAILABLE" = true ]; then
        # 查找状态异常的容器
        local failed_containers=$(echo "$containers_response" | jq '.data[] | select(.status == "exited" or .status == "dead") | select(.container_id != "" and .container_id != null) | {id: .id, name: .name, container_id: .container_id}' -c)
        
        if [ -n "$failed_containers" ]; then
            log_info "发现异常容器，尝试重启..."
            echo "$failed_containers" | while read -r container_info; do
                local db_id=$(echo "$container_info" | jq '.id' -r)
                local container_name=$(echo "$container_info" | jq '.name' -r)
                local container_id=$(echo "$container_info" | jq '.container_id' -r)
                
                log_info "重启容器: $container_name (ID: $container_id)"
                
                if api_call "/container-instances/$db_id/start" "POST" > /dev/null 2>&1; then
                    log_info "✅ 容器重启成功: $container_name"
                else
                    log_warn "⚠️ 容器重启失败: $container_name"
                fi
            done
        else
            log_info "没有发现需要重启的异常容器"
        fi
    else
        log_warn "jq不可用，跳过异常容器重启"
    fi
}

# 生成诊断报告
generate_diagnostic_report() {
    log_info "生成诊断报告..."
    
    local report_file="container-diagnostic-$(date +%Y%m%d-%H%M%S).txt"
    
    {
        echo "容器实例诊断报告"
        echo "生成时间: $(date)"
        echo "========================================"
        echo ""
        
        echo "1. 系统信息"
        echo "操作系统: $(uname -a)"
        echo "Docker版本: $(docker --version)"
        echo ""
        
        echo "2. 服务状态"
        if api_call "/health" > /dev/null 2>&1; then
            echo "API服务: 正常"
        else
            echo "API服务: 异常"
        fi
        
        if docker info > /dev/null 2>&1; then
            echo "Docker服务: 正常"
        else
            echo "Docker服务: 异常"
        fi
        echo ""
        
        echo "3. 容器实例列表"
        local containers_response=$(api_call "/container-instances" 2>/dev/null)
        if [ $? -eq 0 ]; then
            if [ "$JQ_AVAILABLE" = true ]; then
                echo "$containers_response" | jq '.data[] | "ID: \(.id), 名称: \(.name), 状态: \(.status), 容器ID: \(.container_id // "无")"' -r
            else
                echo "$containers_response"
            fi
        else
            echo "无法获取容器列表"
        fi
        echo ""
        
        echo "4. Docker容器列表"
        docker ps -a --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}"
        echo ""
        
        echo "5. 建议操作"
        echo "- 如果API服务异常，请检查应用日志"
        echo "- 如果Docker服务异常，请重启Docker服务"
        echo "- 如果容器状态不一致，请运行状态同步"
        echo "- 如果有无效记录，请运行清理操作"
        
    } > "$report_file"
    
    log_info "诊断报告已保存到: $report_file"
}

# 显示菜单
show_menu() {
    echo ""
    echo "🔧 容器问题修复工具"
    echo "===================="
    echo "1. 检查服务状态"
    echo "2. 同步容器状态"
    echo "3. 清理无效容器记录"
    echo "4. 修复容器ID问题"
    echo "5. 重启异常容器"
    echo "6. 生成诊断报告"
    echo "7. 执行全部修复操作"
    echo "0. 退出"
    echo ""
}

# 主函数
main() {
    log_info "容器问题修复工具启动"
    
    check_dependencies
    
    if [ $# -eq 0 ]; then
        # 交互模式
        while true; do
            show_menu
            read -p "请选择操作 (0-7): " choice
            
            case $choice in
                1)
                    check_service_status
                    ;;
                2)
                    sync_container_status
                    ;;
                3)
                    cleanup_invalid_containers
                    ;;
                4)
                    fix_container_id_issues
                    ;;
                5)
                    restart_failed_containers
                    ;;
                6)
                    generate_diagnostic_report
                    ;;
                7)
                    log_info "执行全部修复操作..."
                    check_service_status
                    sync_container_status
                    cleanup_invalid_containers
                    fix_container_id_issues
                    restart_failed_containers
                    generate_diagnostic_report
                    log_info "全部修复操作完成"
                    ;;
                0)
                    log_info "退出修复工具"
                    exit 0
                    ;;
                *)
                    log_error "无效选择，请重新输入"
                    ;;
            esac
            
            echo ""
            read -p "按回车键继续..."
        done
    else
        # 命令行模式
        case "$1" in
            "check")
                check_service_status
                ;;
            "sync")
                sync_container_status
                ;;
            "cleanup")
                cleanup_invalid_containers
                ;;
            "fix-id")
                fix_container_id_issues
                ;;
            "restart")
                restart_failed_containers
                ;;
            "report")
                generate_diagnostic_report
                ;;
            "all")
                check_service_status
                sync_container_status
                cleanup_invalid_containers
                fix_container_id_issues
                restart_failed_containers
                generate_diagnostic_report
                ;;
            *)
                echo "用法: $0 [check|sync|cleanup|fix-id|restart|report|all]"
                echo "或者不带参数运行进入交互模式"
                exit 1
                ;;
        esac
    fi
}

# 执行主函数
main "$@"
