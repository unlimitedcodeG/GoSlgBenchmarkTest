#!/bin/bash

# GoSlg 负载测试平台启动脚本
# 支持 WebSocket、gRPC、HTTP API 压力测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 protoc
    if ! command -v protoc &> /dev/null; then
        log_warning "protoc 未安装，gRPC 功能可能无法正常工作"
    fi
    
    log_success "依赖检查完成"
}

# 构建项目
build_project() {
    log_info "构建项目..."
    
    # 清理旧的构建文件
    rm -f test-platform-extended
    
    # 构建扩展测试平台
    go build -o test-platform-extended cmd/test-platform/main_extended.go
    
    if [ $? -eq 0 ]; then
        log_success "项目构建成功"
    else
        log_error "项目构建失败"
        exit 1
    fi
}

# 生成 protobuf 代码（如果需要）
generate_proto() {
    if [ "$1" = "--gen-proto" ]; then
        log_info "生成 protobuf 代码..."
        
        # 检查 buf
        if command -v buf &> /dev/null; then
            buf generate
            log_success "使用 buf 生成 protobuf 代码"
        elif command -v protoc &> /dev/null; then
            # 手动生成
            mkdir -p proto/game/v1
            protoc --go_out=. --go_opt=paths=source_relative \
                   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
                   proto/game/v1/game.proto proto/game/v1/game_service.proto
            log_success "使用 protoc 生成 protobuf 代码"
        else
            log_warning "无法生成 protobuf 代码，protoc 和 buf 都未找到"
        fi
    fi
}

# 启动测试平台
start_platform() {
    log_info "启动负载测试平台..."
    
    # 设置环境变量
    export GOMAXPROCS=4
    export TEST_LOG_LEVEL=info
    
    # 显示启动信息
    echo -e "${CYAN}🚀 GoSlg 负载测试平台${NC}"
    echo -e "${CYAN}===========================================${NC}"
    echo -e "📊 Web界面: ${GREEN}http://localhost:8080${NC}"
    echo -e "🧪 负载测试: ${GREEN}http://localhost:8080/loadtest${NC}"
    echo -e "❤️  健康检查: ${GREEN}http://localhost:8080/api/v1/health${NC}"
    echo -e "🌐 HTTP测试服务器: ${GREEN}http://localhost:19000${NC}"
    echo -e "⚡ gRPC测试服务器: ${GREEN}localhost:19001${NC}"
    echo ""
    echo -e "${YELLOW}支持的测试类型:${NC}"
    echo -e "  • WebSocket 长连接压测"
    echo -e "  • gRPC 服务接口压测"
    echo -e "  • HTTP REST API 压测"
    echo ""
    echo -e "${PURPLE}按 Ctrl+C 停止服务${NC}"
    echo ""
    
    # 启动平台
    ./test-platform-extended
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}GoSlg 负载测试平台 - 启动脚本${NC}"
    echo ""
    echo -e "${YELLOW}用法:${NC}"
    echo "  $0 [选项]"
    echo ""
    echo -e "${YELLOW}选项:${NC}"
    echo "  --gen-proto    生成 protobuf 代码"
    echo "  --build-only   仅构建，不启动"
    echo "  --check-deps   仅检查依赖"
    echo "  --help         显示此帮助信息"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo "  $0                    # 正常启动"
    echo "  $0 --gen-proto        # 生成 proto 代码并启动"
    echo "  $0 --build-only       # 仅构建项目"
    echo ""
    echo -e "${YELLOW}功能特性:${NC}"
    echo "  🌐 HTTP API 压力测试"
    echo "  ⚡ gRPC 服务压力测试" 
    echo "  📡 WebSocket 长连接测试"
    echo "  📊 实时性能指标监控"
    echo "  🎯 多协议负载均衡测试"
    echo "  📈 详细的测试报告生成"
}

# 清理函数
cleanup() {
    log_info "正在清理..."
    # 杀死可能运行的测试服务器进程
    pkill -f "test-platform-extended" 2>/dev/null || true
    log_success "清理完成"
}

# 设置信号处理
trap cleanup EXIT

# 主逻辑
main() {
    case "$1" in
        --help|-h)
            show_help
            exit 0
            ;;
        --check-deps)
            check_dependencies
            exit 0
            ;;
        --build-only)
            check_dependencies
            generate_proto "$@"
            build_project
            log_success "构建完成，可执行文件: test-platform-extended"
            exit 0
            ;;
        *)
            check_dependencies
            generate_proto "$@"
            build_project
            start_platform
            ;;
    esac
}

# 运行主函数
main "$@"