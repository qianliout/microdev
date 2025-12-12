# MicroDev Makefile

.PHONY: build test clean install help

# 默认目标
all: build install help

# 构建项目
build:
	@echo "🔨 构建 micro 命令..."
	@go build -o micro ./cmd/micro
	@echo "✅ 构建完成"

# 安装到 GOBIN
install:
	@echo "📦 安装 micro 命令到 GOBIN (包含短命令 'm')..."
	@./build.sh

# 运行测试
test:
	@echo "🧪 运行测试套件..."
	@./tests/run_tests.sh

# 运行基础测试
test-basic:
	@echo "🔧 运行基础测试..."
	@go test -v ./tests -run TestBasicFunctionality

# 运行工具函数测试
test-utils:
	@echo "🛠️ 运行工具函数测试..."
	@go test -v ./tests -run TestUtilsFunctionality

# 运行集成测试
test-integration:
	@echo "🔗 运行集成测试..."
	@go test -v ./tests -run TestCommandIntegration
	@go test -v ./tests -run TestInputProcessing
	@go test -v ./tests -run TestOutputProcessing

# 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
	@rm -f micro
	@rm -f coverage.out coverage.html
	@rm -rf build/
	@echo "✅ 清理完成"

# 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	@go fmt ./...
	@echo "✅ 格式化完成"

# 检查代码
lint:
	@echo "🔍 检查代码..."
	@go vet ./...
	@echo "✅ 检查完成"

# 更新依赖
deps:
	@echo "📦 更新依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖更新完成"

# 显示帮助信息
help:
	@echo "MicroDev 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build          构建 micro 命令"
	@echo "  install        安装 micro 命令到 GOBIN (包含短命令 'm')"
	@echo "  test           运行完整测试套件"
	@echo "  test-basic     运行基础测试"
	@echo "  test-utils     运行工具函数测试"
	@echo "  test-integration 运行集成测试"
	@echo "  clean          清理构建产物"
	@echo "  fmt            格式化代码"
	@echo "  lint           检查代码"
	@echo "  deps           更新依赖"
	@echo "  help           显示此帮助信息"
