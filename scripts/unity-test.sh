#!/bin/bash

# Unity游戏测试录制脚本 (Linux/macOS版本)
# 用于配合Unity客户端进行自动化测试录制

set -e

# 默认参数
ENVIRONMENT="development"
PLAYER=""
DURATION="30m"
CONFIG="configs/test-environments.yaml"
OUTPUT=""
VERBOSE=false
DRY_RUN=false
HELP=false

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    echo -e "${GREEN}🎮 Unity游戏测试录制脚本${NC}"
    echo -e "${GREEN}=================================${NC}"
    echo ""
    echo -e "${YELLOW}用法:${NC}"
    echo -e "  ${WHITE}./scripts/unity-test.sh [选项]${NC}"
    echo ""
    echo -e "${YELLOW}选项:${NC}"
    echo -e "  ${WHITE}-e, --environment ENV    环境类型 (development|testing|staging|local)${NC}"
    echo -e "  ${WHITE}-p, --player PLAYER      测试账号用户名${NC}"
    echo -e "  ${WHITE}-d, --duration DURATION  录制时长 (例如: 30m, 1h)${NC}"
    echo -e "  ${WHITE}-c, --config CONFIG      配置文件路径${NC}"
    echo -e "  ${WHITE}-o, --output OUTPUT      输出目录${NC}"
    echo -e "  ${WHITE}-v, --verbose            启用详细日志${NC}"
    echo -e "  ${WHITE}    --dry-run            干运行模式，只检查配置${NC}"
    echo -e "  ${WHITE}-h, --help               显示此帮助信息${NC}"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo -e "  ${GREEN}# 使用开发环境录制30分钟${NC}"
    echo -e "  ${WHITE}./scripts/unity-test.sh -e development -d 30m${NC}"
    echo ""
    echo -e "  ${GREEN}# 指定测试账号和输出目录${NC}"
    echo -e "  ${WHITE}./scripts/unity-test.sh -p qa_tester_001 -o ./recordings/unity${NC}"
    echo ""
    echo -e "  ${GREEN}# 干运行模式检查配置${NC}"
    echo -e "  ${WHITE}./scripts/unity-test.sh --dry-run${NC}"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -p|--player)
            PLAYER="$2"
            shift 2
            ;;
        -d|--duration)
            DURATION="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            HELP=true
            shift
            ;;
        *)
            echo -e "${RED}❌ 未知参数: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# 显示帮助信息
if [[ "$HELP" == true ]]; then
    show_help
    exit 0
fi

echo -e "${GREEN}🎮 Unity游戏测试录制工具${NC}"
echo -e "${GREEN}=========================${NC}"
echo ""

# 验证环境参数
case $ENVIRONMENT in
    development|testing|staging|local)
        ;;
    *)
        echo -e "${RED}❌ 无效的环境类型: $ENVIRONMENT${NC}"
        echo -e "${YELLOW}可用环境: development, testing, staging, local${NC}"
        exit 1
        ;;
esac

# 检查Go环境
echo -n "🔧 检查Go环境... "
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go未安装或不在PATH中${NC}"
    echo -e "${YELLOW}请确保已安装Go 1.25+并添加到PATH中${NC}"
    exit 1
fi

GO_VERSION=$(go version)
echo -e "${GREEN}✅ $GO_VERSION${NC}"

# 检查配置文件
echo -n "📄 检查配置文件... "
if [[ ! -f "$CONFIG" ]]; then
    echo -e "${RED}❌ 配置文件不存在: $CONFIG${NC}"
    echo -e "${YELLOW}请确保配置文件存在或指定正确的路径${NC}"
    exit 1
fi
echo -e "${GREEN}✅ $CONFIG${NC}"

# 构建录制工具
echo -e "${CYAN}🔧 构建Unity录制工具...${NC}"
mkdir -p ./bin

if ! go build -o "./bin/unity-recorder" "./cmd/unity-recorder"; then
    echo -e "${RED}❌ 构建失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 构建完成${NC}"

# 准备命令行参数
RECORDER_ARGS=(
    "-env" "$ENVIRONMENT"
    "-config" "$CONFIG"
    "-duration" "$DURATION"
)

if [[ -n "$PLAYER" ]]; then
    RECORDER_ARGS+=("-player" "$PLAYER")
fi

if [[ -n "$OUTPUT" ]]; then
    RECORDER_ARGS+=("-output" "$OUTPUT")
fi

if [[ "$VERBOSE" == true ]]; then
    RECORDER_ARGS+=("-verbose")
fi

if [[ "$DRY_RUN" == true ]]; then
    RECORDER_ARGS+=("-dry-run")
    echo -e "${CYAN}🔍 干运行模式 - 只检查配置${NC}"
fi

# 显示即将执行的命令
echo ""
echo -e "${CYAN}🚀 即将执行的命令:${NC}"
echo -e "  ${WHITE}./bin/unity-recorder ${RECORDER_ARGS[*]}${NC}"
echo ""

# 如果不是干运行模式，询问是否继续
if [[ "$DRY_RUN" != true ]]; then
    echo -e "${YELLOW}⚠️  录制即将开始，请确保:${NC}"
    echo -e "   ${WHITE}1. Unity客户端已准备就绪${NC}"
    echo -e "   ${WHITE}2. 游戏服务器正常运行${NC}"
    echo -e "   ${WHITE}3. 网络连接稳定${NC}"
    echo ""
    
    read -p "是否开始录制? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}❌ 用户取消操作${NC}"
        exit 0
    fi
fi

# 执行录制工具
echo -e "${GREEN}🎬 启动录制工具...${NC}"
echo ""

# 运行录制工具
if ./bin/unity-recorder "${RECORDER_ARGS[@]}"; then
    echo ""
    echo -e "${GREEN}🎉 录制完成!${NC}"
    
    # 显示输出目录中的文件
    OUTPUT_DIR=${OUTPUT:-"recordings"}
    if [[ -d "$OUTPUT_DIR" ]]; then
        echo ""
        echo -e "${CYAN}📁 生成的文件:${NC}"
        find "$OUTPUT_DIR" -name "session_*" -type f -printf "%T@ %p\n" 2>/dev/null | \
        sort -rn | head -5 | while IFS= read -r line; do
            FILE=$(echo "$line" | cut -d' ' -f2-)
            if [[ -f "$FILE" ]]; then
                SIZE=$(du -h "$FILE" | cut -f1)
                BASENAME=$(basename "$FILE")
                echo -e "   ${WHITE}$BASENAME ($SIZE)${NC}"
            fi
        done
    fi
    
    echo ""
    echo -e "${YELLOW}💡 提示:${NC}"
    echo -e "   ${WHITE}- 可以使用生成的JSON文件进行回放分析${NC}"
    echo -e "   ${WHITE}- 断言测试结果已包含在输出中${NC}"
    echo -e "   ${WHITE}- 如需详细分析，可运行: go run cmd/session-analyzer/main.go${NC}"
    
else
    echo -e "${RED}❌ 录制失败${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}🚀 所有操作完成!${NC}"