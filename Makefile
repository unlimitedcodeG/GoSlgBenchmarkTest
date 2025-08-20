# Go Unity长连接+Protobuf测试项目 Makefile
# 项目: GoSlgBenchmarkTest

.PHONY: help proto test test-race test-short fuzz bench clean install-deps lint format run-server run-client deps-check tools-install ci-prepare ci-test ci-lint

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
GO_VERSION := $(shell go version | awk '{print $$3}')
PROTO_DIR := proto
BUILD_DIR := build
COVERAGE_DIR := coverage
TOOLS_DIR := tools

# 颜色定义
COLOR_RED := \033[31m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_RESET := \033[0m

# 帮助信息
help: ## 显示帮助信息
	@echo "$(COLOR_BLUE)GoSlgBenchmarkTest - Unity长连接+Protobuf测试框架$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Go版本: $(GO_VERSION)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)基础命令:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(COLOR_YELLOW)%-20s$(COLOR_RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(COLOR_GREEN)SLG协议管理:$(COLOR_RESET)"
	@echo "  $(COLOR_YELLOW)slg-help$(COLOR_RESET)             SLG协议管理帮助"
	@echo "  $(COLOR_YELLOW)integrate-dev-proto$(COLOR_RESET)  集成研发协议(交互式)"
	@echo "  $(COLOR_YELLOW)generate-slg-proto$(COLOR_RESET)   生成SLG协议代码 VERSION=v1.0.0"
	@echo "  $(COLOR_YELLOW)validate-slg-proto$(COLOR_RESET)   验证SLG协议 VERSION=v1.0.0"
	@echo "  $(COLOR_YELLOW)list-slg-versions$(COLOR_RESET)    列出所有SLG协议版本"
	@echo ""
	@echo "$(COLOR_GREEN)CI 相关命令:$(COLOR_RESET)"
	@echo "  $(COLOR_YELLOW)ci-prepare$(COLOR_RESET)          准备CI环境"
	@echo "  $(COLOR_YELLOW)ci-test$(COLOR_RESET)             运行CI测试"
	@echo "  $(COLOR_YELLOW)ci-lint$(COLOR_RESET)             运行CI代码检查"

# 安装工具依赖
tools-install: ## 安装必要的工具
	@echo "$(COLOR_BLUE)安装开发工具...$(COLOR_RESET)"
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "$(COLOR_YELLOW)安装 buf...$(COLOR_RESET)"; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(COLOR_YELLOW)安装 golangci-lint...$(COLOR_RESET)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi

# 检查依赖
deps-check: ## 检查Go模块依赖
	@echo "$(COLOR_BLUE)检查依赖...$(COLOR_RESET)"
	go mod verify
	go mod tidy
	@echo "$(COLOR_GREEN)依赖检查完成$(COLOR_RESET)"

# 安装项目依赖
install-deps: deps-check ## 安装项目依赖
	@echo "$(COLOR_BLUE)下载依赖...$(COLOR_RESET)"
	go mod download
	@echo "$(COLOR_GREEN)依赖安装完成$(COLOR_RESET)"

# 生成Protobuf代码
proto: tools-install ## 生成Protobuf代码
	@echo "$(COLOR_BLUE)生成Protobuf代码...$(COLOR_RESET)"
	@if [ ! -f buf.yaml ]; then \
		echo "$(COLOR_RED)错误: buf.yaml文件不存在$(COLOR_RESET)"; \
		exit 1; \
	fi
	buf lint
	buf generate
	@echo "$(COLOR_GREEN)Protobuf代码生成完成$(COLOR_RESET)"

# 代码格式化
format: ## 格式化Go代码
	@echo "$(COLOR_BLUE)格式化代码...$(COLOR_RESET)"
	go fmt ./...
	goimports -w .
	@echo "$(COLOR_GREEN)代码格式化完成$(COLOR_RESET)"

# 代码检查
lint: tools-install ## 运行代码检查
	@echo "$(COLOR_BLUE)运行代码检查...$(COLOR_RESET)"
	golangci-lint run ./...
	@echo "$(COLOR_GREEN)代码检查完成$(COLOR_RESET)"

# 运行测试
test: proto ## 运行所有测试
	@echo "$(COLOR_BLUE)运行测试...$(COLOR_RESET)"
	go test ./... -v -count=1
	@echo "$(COLOR_GREEN)测试完成$(COLOR_RESET)"

# 运行竞态检测测试
test-race: proto ## 运行竞态检测测试
	@echo "$(COLOR_BLUE)运行竞态检测测试...$(COLOR_RESET)"
	go test ./... -v -race -count=1
	@echo "$(COLOR_GREEN)竞态检测测试完成$(COLOR_RESET)"

# 运行短测试
test-short: proto ## 运行短测试
	@echo "$(COLOR_BLUE)运行短测试...$(COLOR_RESET)"
	go test ./... -v -short -count=1
	@echo "$(COLOR_GREEN)短测试完成$(COLOR_RESET)"

# 运行基准测试
bench: proto ## 运行基准测试
	@echo "$(COLOR_BLUE)运行基准测试...$(COLOR_RESET)"
	go test ./test -run=^$ -bench=. -benchmem -count=3
	@echo "$(COLOR_GREEN)基准测试完成$(COLOR_RESET)"

# 运行模糊测试
fuzz: proto ## 运行模糊测试
	@echo "$(COLOR_BLUE)运行模糊测试...$(COLOR_RESET)"
	@targets="FuzzBattlePushUnmarshal FuzzLoginReqUnmarshal FuzzPlayerActionUnmarshal FuzzFrameDecode FuzzFrameDecoder FuzzErrorResp"; \
	for target in $$targets; do \
		echo "$(COLOR_YELLOW)运行 $$target...$(COLOR_RESET)"; \
		go test ./test -run=^$$ -fuzz=^$$target$$ -fuzztime=30s || exit 1; \
	done
	@echo "$(COLOR_GREEN)模糊测试完成$(COLOR_RESET)"

# 清理构建文件
clean: ## 清理构建文件
	@echo "$(COLOR_BLUE)清理构建文件...$(COLOR_RESET)"
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -rf generated/slg/v1_0_0/GoSlgBenchmarkTest
	rm -rf generated/slg/v1_1_0/GoSlgBenchmarkTest
	go clean -cache -testcache
	@echo "$(COLOR_GREEN)清理完成$(COLOR_RESET)"

# CI 准备
ci-prepare: ## 准备CI环境
	@echo "$(COLOR_BLUE)准备CI环境...$(COLOR_RESET)"
	go mod download
	@echo "$(COLOR_GREEN)CI环境准备完成$(COLOR_RESET)"

# CI 测试
ci-test: ci-prepare ## 运行CI测试
	@echo "$(COLOR_BLUE)运行CI测试...$(COLOR_RESET)"
	buf generate
	go test ./... -v -race -count=1 -timeout=10m
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic -timeout=10m
	@echo "$(COLOR_GREEN)CI测试完成$(COLOR_RESET)"

# CI 代码检查
ci-lint: ci-prepare ## 运行CI代码检查
	@echo "$(COLOR_BLUE)运行CI代码检查...$(COLOR_RESET)"
	golangci-lint run ./...
	@echo "$(COLOR_GREEN)CI代码检查完成$(COLOR_RESET)"

# 运行测试服务器
run-server: build-server ## 运行测试服务器
	@echo "$(COLOR_BLUE)启动测试服务器...$(COLOR_RESET)"
	$(BUILD_DIR)/test-server

# 运行客户端压力测试
run-loadtest: build-client ## 运行客户端压力测试
	@echo "$(COLOR_BLUE)运行压力测试...$(COLOR_RESET)"
	$(BUILD_DIR)/test-client -mode=loadtest -clients=10 -duration=30s

# 网络模拟（Linux）
simulate-network: ## 模拟网络延迟和丢包（需要sudo权限）
	@echo "$(COLOR_BLUE)应用网络模拟...$(COLOR_RESET)"
	@if [ "$(shell uname)" = "Linux" ]; then \
		sudo tc qdisc replace dev lo root netem delay 100ms 20ms distribution normal loss 1% reorder 1%; \
		echo "$(COLOR_GREEN)网络模拟已应用: 延迟100±20ms, 丢包率1%, 重排序1%$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)使用 'make reset-network' 恢复网络$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)网络模拟仅支持Linux系统$(COLOR_RESET)"; \
	fi

# 重置网络（Linux）
reset-network: ## 重置网络模拟（需要sudo权限）
	@echo "$(COLOR_BLUE)重置网络设置...$(COLOR_RESET)"
	@if [ "$(shell uname)" = "Linux" ]; then \
		sudo tc qdisc del dev lo root 2>/dev/null || true; \
		echo "$(COLOR_GREEN)网络设置已重置$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)网络重置仅支持Linux系统$(COLOR_RESET)"; \
	fi

# 生成测试数据
generate-testdata: proto ## 生成测试数据文件
	@echo "$(COLOR_BLUE)生成测试数据...$(COLOR_RESET)"
	@mkdir -p testdata
	go run tools/generate_testdata.go
	@echo "$(COLOR_GREEN)测试数据生成完成$(COLOR_RESET)"

# 验证项目完整性
verify: proto test-race lint ## 完整验证项目（proto + 测试 + 检查）
	@echo "$(COLOR_GREEN)项目验证完成!$(COLOR_RESET)"

# 快速检查
quick: proto test-short lint ## 快速检查（短测试 + 代码检查）
	@echo "$(COLOR_GREEN)快速检查完成!$(COLOR_RESET)"

# 发布准备
release-check: verify test-coverage bench ## 发布前检查
	@echo "$(COLOR_GREEN)发布检查完成!$(COLOR_RESET)"

# 显示项目状态
status: ## 显示项目状态
	@echo "$(COLOR_BLUE)项目状态:$(COLOR_RESET)"
	@echo "  Go版本: $(GO_VERSION)"
	@echo "  项目路径: $(shell pwd)"
	@echo "  模块: $(shell go list -m)"
	@echo ""
	@echo "$(COLOR_BLUE)依赖状态:$(COLOR_RESET)"
	@go list -m all | head -10
	@echo ""
	@echo "$(COLOR_BLUE)构建文件:$(COLOR_RESET)"
	@ls -la $(BUILD_DIR) 2>/dev/null || echo "  无构建文件"

# 开发环境设置
dev-setup: tools-install install-deps proto ## 设置开发环境
	@echo "$(COLOR_GREEN)开发环境设置完成!$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)建议的开发流程:$(COLOR_RESET)"
	@echo "  1. make quick    # 快速检查"
	@echo "  2. make test     # 运行测试"
	@echo "  3. make bench    # 性能测试"
	@echo "  4. make verify   # 完整验证"

# 监控文件变化并自动测试（需要 inotify-tools）
watch: ## 监控文件变化并自动运行测试
	@echo "$(COLOR_BLUE)监控文件变化...$(COLOR_RESET)"
	@if command -v inotifywait >/dev/null 2>&1; then \
		while inotifywait -r -e modify --exclude '\.git|build|coverage' . >/dev/null 2>&1; do \
			echo "$(COLOR_YELLOW)文件变化检测到，运行测试...$(COLOR_RESET)"; \
			make test-short; \
		done; \
	else \
		echo "$(COLOR_RED)错误: 需要安装 inotify-tools$(COLOR_RESET)"; \
		echo "Ubuntu/Debian: sudo apt-get install inotify-tools"; \
		echo "CentOS/RHEL: sudo yum install inotify-tools"; \
	fi# SLG协议管理扩展 - 追加到主Makefile中

# SLG协议相关变量
SLG_PROTO_DIR := slg-proto
SLG_GENERATED_DIR := generated/slg
SLG_CONFIG_DIR := configs

# SLG协议帮助
.PHONY: slg-help
slg-help: ## SLG协议管理完整帮助
	@echo "$(COLOR_BLUE)🎮 SLG协议管理工具$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)========================================$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)📋 命令列表:$(COLOR_RESET)"
	@echo "  $(COLOR_YELLOW)make integrate-dev-proto$(COLOR_RESET)"
	@echo "    交互式集成研发团队提供的协议文件"
	@echo ""
	@echo "  $(COLOR_YELLOW)make generate-slg-proto VERSION=v1.0.0$(COLOR_RESET)"
	@echo "    生成指定版本的SLG协议Go代码"
	@echo ""
	@echo "  $(COLOR_YELLOW)make validate-slg-proto VERSION=v1.0.0$(COLOR_RESET)"
	@echo "    验证指定版本的协议格式和兼容性"
	@echo ""
	@echo "  $(COLOR_YELLOW)make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0$(COLOR_RESET)"
	@echo "    测试两个版本间的协议兼容性"
	@echo ""
	@echo "  $(COLOR_YELLOW)make list-slg-versions$(COLOR_RESET)"
	@echo "    列出所有可用的SLG协议版本"
	@echo ""
	@echo "$(COLOR_GREEN)📁 目录结构:$(COLOR_RESET)"
	@echo "  slg-proto/v1.0.0/    - 协议定义文件"
	@echo "  generated/slg/       - 生成的Go代码"
	@echo "  test/slg/           - SLG专用测试"
	@echo "  configs/            - 配置文件"
	@echo ""
	@echo "$(COLOR_GREEN)🔄 典型工作流:$(COLOR_RESET)"
	@echo "  1. make integrate-dev-proto  # 集成新协议"
	@echo "  2. make validate-slg-proto VERSION=v1.1.0  # 验证格式"
	@echo "  3. make generate-slg-proto VERSION=v1.1.0  # 生成代码"
	@echo "  4. make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0  # 兼容性测试"
	@echo "  5. go test ./test/slg -v  # 运行SLG测试"

# 交互式集成研发协议
.PHONY: integrate-dev-proto
integrate-dev-proto: ## 交互式集成研发协议
	@echo "$(COLOR_BLUE)🔄 集成研发协议...$(COLOR_RESET)"
	@go run tools/slg-proto-manager.go integrate

# 生成SLG协议代码
.PHONY: generate-slg-proto
generate-slg-proto: ## 生成SLG协议Go代码 (需要VERSION参数)
ifndef VERSION
	@echo "$(COLOR_RED)错误: 请指定VERSION参数$(COLOR_RESET)"
	@echo "示例: make generate-slg-proto VERSION=v1.0.0"
	@exit 1
endif
	@echo "$(COLOR_BLUE)🔧 生成SLG协议代码 $(VERSION)...$(COLOR_RESET)"
	@if [ ! -d "$(SLG_PROTO_DIR)/$(VERSION)" ]; then \
		echo "$(COLOR_RED)错误: 协议版本不存在: $(VERSION)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@go run tools/slg-proto-manager.go generate $(VERSION)
	@echo "$(COLOR_GREEN)✅ 代码生成完成: $(SLG_GENERATED_DIR)/$(subst .,_,$(VERSION))$(COLOR_RESET)"

# 验证SLG协议
.PHONY: validate-slg-proto
validate-slg-proto: ## 验证SLG协议格式 (需要VERSION参数)
ifndef VERSION
	@echo "$(COLOR_RED)错误: 请指定VERSION参数$(COLOR_RESET)"
	@echo "示例: make validate-slg-proto VERSION=v1.0.0"
	@exit 1
endif
	@echo "$(COLOR_BLUE)🔍 验证SLG协议 $(VERSION)...$(COLOR_RESET)"
	@go run tools/slg-proto-manager.go validate $(VERSION)
	@echo "$(COLOR_GREEN)✅ 协议验证通过$(COLOR_RESET)"

# 兼容性测试
.PHONY: test-slg-compatibility
test-slg-compatibility: ## 测试SLG协议兼容性 (需要FROM和TO参数)
ifndef FROM
	@echo "$(COLOR_RED)错误: 请指定FROM参数$(COLOR_RESET)"
	@echo "示例: make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0"
	@exit 1
endif
ifndef TO
	@echo "$(COLOR_RED)错误: 请指定TO参数$(COLOR_RESET)"
	@echo "示例: make test-slg-compatibility FROM=v1.0.0 TO=v1.1.0"
	@exit 1
endif
	@echo "$(COLOR_BLUE)🔍 兼容性测试: $(FROM) -> $(TO)$(COLOR_RESET)"
	@go run tools/slg-proto-manager.go compatibility-check $(FROM) $(TO)
	@echo "$(COLOR_GREEN)✅ 兼容性检查通过$(COLOR_RESET)"

# 列出SLG协议版本
.PHONY: list-slg-versions
list-slg-versions: ## 列出所有SLG协议版本
	@echo "$(COLOR_BLUE)📋 SLG协议版本列表$(COLOR_RESET)"
	@go run tools/slg-proto-manager.go list-versions

# 运行SLG测试
.PHONY: test-slg
test-slg: ## 运行SLG专用测试
	@echo "$(COLOR_BLUE)🧪 运行SLG测试...$(COLOR_RESET)"
	@go test ./test/slg -v
	@echo "$(COLOR_GREEN)✅ SLG测试完成$(COLOR_RESET)"

# 生成SLG测试数据
.PHONY: generate-slg-testdata
generate-slg-testdata: ## 生成SLG测试数据
	@echo "$(COLOR_BLUE)📊 生成SLG测试数据...$(COLOR_RESET)"
	@mkdir -p testdata/slg
	@go run tools/generate-slg-testdata.go
	@echo "$(COLOR_GREEN)✅ SLG测试数据生成完成$(COLOR_RESET)"

# 清理SLG生成文件
.PHONY: clean-slg
clean-slg: ## 清理SLG生成的文件
	@echo "$(COLOR_BLUE)🧹 清理SLG生成文件...$(COLOR_RESET)"
	@rm -rf $(SLG_GENERATED_DIR)
	@rm -rf testdata/slg
	@echo "$(COLOR_GREEN)✅ SLG清理完成$(COLOR_RESET)"

# 完整SLG验证
.PHONY: verify-slg
verify-slg: ## 完整SLG项目验证
	@echo "$(COLOR_BLUE)🔍 完整SLG项目验证...$(COLOR_RESET)"
	@make list-slg-versions
	@make validate-slg-proto VERSION=$$(go run tools/slg-proto-manager.go list-versions | grep "当前版本" | awk '{print $$2}')
	@make test-slg
	@echo "$(COLOR_GREEN)✅ SLG项目验证完成$(COLOR_RESET)"
