#!/bin/bash

# WebSocket测试运行脚本
# 用于本地开发环境和CI环境

set -e

echo "🚀 WebSocket集成测试启动脚本"
echo "================================="

# 设置默认值
SERVER_TYPE="${SERVER_TYPE:-websocket}"
WS_PORT="${WS_PORT:-18090}"
TEST_TIMEOUT="${TEST_TIMEOUT:-10m}"
CI="${CI:-false}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help          显示此帮助信息"
    echo "  -p, --port PORT     设置WebSocket服务器端口 (默认: 18090)"
    echo "  -t, --timeout TIME  设置测试超时时间 (默认: 10m)"
    echo "  -c, --ci            CI环境模式"
    echo "  -l, --local         本地开发模式"
    echo ""
    echo "环境变量:"
    echo "  WS_PORT             WebSocket服务器端口"
    echo "  TEST_TIMEOUT        测试超时时间"
    echo "  CI                  CI环境标识 (true/false)"
    echo ""
    echo "示例:"
    echo "  $0 -p 18090 -t 15m    # 本地测试，端口18090，超时15分钟"
    echo "  $0 -c                 # CI环境模式"
    exit 0
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            print_help
            ;;
        -p|--port)
            WS_PORT="$2"
            shift 2
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        -c|--ci)
            CI=true
            shift
            ;;
        -l|--local)
            CI=false
            shift
            ;;
        *)
            echo -e "${RED}错误: 未知参数 $1${NC}"
            print_help
            ;;
    esac
done

echo -e "${BLUE}配置信息:${NC}"
echo "  服务器类型: $SERVER_TYPE"
echo "  WebSocket端口: $WS_PORT"
echo "  测试超时: $TEST_TIMEOUT"
echo "  CI环境: $CI"
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到Go命令${NC}"
    exit 1
fi

# 检查端口是否被占用
if lsof -Pi :$WS_PORT -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${YELLOW}警告: 端口 $WS_PORT 已被占用，尝试停止现有进程...${NC}"
    if [ "$CI" = true ]; then
        # 在CI环境中强制停止
        sudo lsof -ti :$WS_PORT | xargs -r sudo kill -9 2>/dev/null || true
    else
        # 在本地环境中询问用户
        read -p "是否要停止占用端口的进程？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            lsof -ti :$WS_PORT | xargs kill -9 2>/dev/null || true
            sleep 2
        else
            echo -e "${RED}错误: 端口 $WS_PORT 被占用，无法启动服务器${NC}"
            exit 1
        fi
    fi
fi

# 启动WebSocket服务器
echo -e "${GREEN}启动WebSocket服务器...${NC}"
if [ "$CI" = true ]; then
    # CI环境：后台运行，设置环境变量
    export SERVER_TYPE=$SERVER_TYPE
    export WS_PORT=$WS_PORT
    export CI=$CI

    go run tools/grpc-server.go &
    SERVER_PID=$!
    echo "SERVER_PID=$SERVER_PID"

    # 等待服务器启动
    echo "等待服务器启动..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if nc -z localhost $WS_PORT 2>/dev/null; then
            echo -e "${GREEN}✅ WebSocket服务器启动成功 (PID: $SERVER_PID)${NC}"
            break
        fi
        sleep 2
        timeout=$((timeout - 2))
        echo "等待服务器启动... 剩余 ${timeout}s"
    done

    if [ $timeout -le 0 ]; then
        echo -e "${RED}❌ WebSocket服务器启动失败${NC}"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
else
    # 本地环境：前台运行，用户手动Ctrl+C停止
    echo -e "${YELLOW}本地开发模式 - 服务器将在前台运行${NC}"
    echo -e "${YELLOW}请在另一个终端运行测试，或按Ctrl+C停止服务器${NC}"
    echo ""

    export SERVER_TYPE=$SERVER_TYPE
    export WS_PORT=$WS_PORT

    # 启动服务器
    go run tools/grpc-server.go
    exit 0
fi

# 运行WebSocket测试
echo -e "${GREEN}开始WebSocket集成测试...${NC}"

# 设置测试环境变量
export WS_PORT=$WS_PORT

# 运行主要集成测试
echo "运行5分钟墙钟时间测试..."
if go test ./test/session -v -run TestTimelineAnalysisRealWallClock \
    -timeout=$TEST_TIMEOUT \
    -parallel=1 \
    -count=1; then
    echo -e "${GREEN}✅ 5分钟集成测试通过${NC}"
else
    TEST_EXIT_CODE=$?
    echo -e "${RED}❌ 5分钟集成测试失败 (退出码: $TEST_EXIT_CODE)${NC}"
fi

# 运行其他WebSocket测试
echo "运行其他WebSocket测试..."
if go test ./test/session -v -run "TestSessionRecordingAndReplay|TestSessionAssertions" \
    -timeout=2m; then
    echo -e "${GREEN}✅ 其他WebSocket测试通过${NC}"
else
    OTHER_TEST_EXIT_CODE=$?
    echo -e "${RED}❌ 其他WebSocket测试失败 (退出码: $OTHER_TEST_EXIT_CODE)${NC}"
fi

# 清理服务器进程
echo "清理服务器进程..."
if [ ! -z "$SERVER_PID" ]; then
    echo "停止WebSocket服务器 (PID: $SERVER_PID)"
    kill $SERVER_PID 2>/dev/null || true
    sleep 2
    # 强制终止如果还在运行
    kill -9 $SERVER_PID 2>/dev/null || true
fi

# 检查测试结果
if [ -z "$TEST_EXIT_CODE" ] && [ -z "$OTHER_TEST_EXIT_CODE" ]; then
    echo -e "${GREEN}🎉 所有WebSocket测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 部分WebSocket测试失败${NC}"
    if [ ! -z "$TEST_EXIT_CODE" ]; then
        echo -e "${RED}  - 5分钟集成测试失败 (退出码: $TEST_EXIT_CODE)${NC}"
    fi
    if [ ! -z "$OTHER_TEST_EXIT_CODE" ]; then
        echo -e "${RED}  - 其他测试失败 (退出码: $OTHER_TEST_EXIT_CODE)${NC}"
    fi
    exit 1
fi