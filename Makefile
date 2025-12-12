# MicroDev Makefile

.PHONY: build test clean install help

# 默认目标
all: build install clean

# 构建项目
build:
	@echo "🔨 构建 micro 命令..."
	@go build -o micro ./cmd/micro
	@echo "✅ 构建完成"

# 安装到 GOBIN
install: build
	@echo "📦 安装 micro 命令..."
	@INSTALL_PATH=$$(go env GOBIN); \
	if [ -z "$$INSTALL_PATH" ]; then \
		INSTALL_PATH=$$(go env GOPATH)/bin; \
	fi; \
	if [ -z "$$INSTALL_PATH" ]; then \
		echo "❌ 错误: GOBIN 和 GOPATH 均未设置。无法确定安装路径。"; \
		exit 1; \
	fi; \
	mkdir -p "$$INSTALL_PATH"; \
	cp ./micro "$$INSTALL_PATH/micro"; \
	ln -sf "$$INSTALL_PATH/micro" "$$INSTALL_PATH/m"; \
	echo "✅ micro 和 m 已成功安装到 $$INSTALL_PATH"

# 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
	@rm -f micro
	@rm -f coverage.out coverage.html
	@rm -rf build/
	@echo "✅ 清理完成"

# 显示帮助信息
help:
	@echo "MicroDev 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build          构建 micro 命令"
	@echo "  install        安装 micro 命令到 GOBIN (包含短命令 'm')"
