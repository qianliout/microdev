# Makefile for microdev project

# 变量定义
PROJECT_NAME := microdev
PROMPT_CMD := prompt
TRANSLATE_CMD := translate

# Go 相关变量
GO := go
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
	GOBIN := $(shell go env GOPATH)/bin
endif
ifeq ($(GOBIN),)
	GOBIN := $(HOME)/go/bin
endif

# 构建相关变量
BUILD_DIR := build
DIST_DIR := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m

# 默认目标
.PHONY: all
all: clean build

# 帮助信息
.PHONY: help
help:
	@echo "$(BLUE)微服务开发工具包 - Makefile 使用说明$(NC)"
	@echo ""
	@echo "$(YELLOW)可用目标:$(NC)"
	@echo "  $(GREEN)build$(NC)           - 编译所有命令"
	@echo "  $(GREEN)prompt$(NC)          - 编译 prompt 命令"
	@echo "  $(GREEN)translate$(NC)       - 编译 translate 命令"
	@echo "  $(GREEN)install$(NC)         - 编译并安装到 GOBIN 目录"
	@echo "  $(GREEN)clean$(NC)           - 清理构建文件"
	@echo "  $(GREEN)test$(NC)            - 运行测试"
	@echo "  $(GREEN)lint$(NC)            - 运行代码检查"
	@echo "  $(GREEN)fmt$(NC)             - 格式化代码"
	@echo "  $(GREEN)deps$(NC)            - 安装/更新依赖"
	@echo "  $(GREEN)dist$(NC)            - 创建发布包"
	@echo "  $(GREEN)info$(NC)            - 显示项目信息"
	@echo "  $(GREEN)uninstall$(NC)       - 卸载命令"
	@echo "  $(GREEN)help$(NC)            - 显示此帮助信息"
	@echo ""
	@echo "$(YELLOW)环境变量:$(NC)"
	@echo "  $(GREEN)GOBIN$(NC)    - Go 二进制文件安装目录 (当前: $(GOBIN))"
	@echo "  $(GREEN)VERSION$(NC)  - 构建版本 (当前: $(VERSION))"
	@echo ""
	@echo "$(YELLOW)示例用法:$(NC)"
	@echo "  make install         # 编译并安装 prompt 命令"
	@echo "  make clean build     # 清理后重新构建"
	@echo "  make test            # 运行所有测试"

# 创建必要目录
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

# 编译 prompt 命令
.PHONY: prompt
prompt: $(BUILD_DIR)
	@echo "$(YELLOW)🔨 编译 $(PROMPT_CMD) 命令...$(NC)"
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(PROMPT_CMD) ./cmd/prompt
	@echo "$(GREEN)✅ $(PROMPT_CMD) 编译完成: $(BUILD_DIR)/$(PROMPT_CMD)$(NC)"

# 编译 translate 命令
.PHONY: translate
translate: $(BUILD_DIR)
	@echo "$(YELLOW)🔨 编译 $(TRANSLATE_CMD) 命令...$(NC)"
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(TRANSLATE_CMD) ./cmd/translate
	@echo "$(GREEN)✅ $(TRANSLATE_CMD) 编译完成: $(BUILD_DIR)/$(TRANSLATE_CMD)$(NC)"

# 编译所有命令
.PHONY: build
build: prompt translate
	@echo "$(GREEN)🎉 所有命令编译完成!$(NC)"

# 安装到 GOBIN 目录
.PHONY: install
install: build
	@echo "$(YELLOW)📦 安装命令到 $(GOBIN)...$(NC)"
	@mkdir -p $(GOBIN)
	@cp $(BUILD_DIR)/$(PROMPT_CMD) $(GOBIN)/
	@cp $(BUILD_DIR)/$(TRANSLATE_CMD) $(GOBIN)/
	@chmod +x $(GOBIN)/$(PROMPT_CMD)
	@chmod +x $(GOBIN)/$(TRANSLATE_CMD)
	@echo "$(GREEN)✅ 安装完成!$(NC)"
	@echo "$(BLUE)📍 已安装命令:$(NC)"
	@ls -la $(GOBIN)/$(PROMPT_CMD) $(GOBIN)/$(TRANSLATE_CMD)
	@echo ""
	@echo "$(YELLOW)💡 使用说明:$(NC)"
	@echo "  $(PROMPT_CMD) <input_content>    # 优化提示词"
	@echo "  $(TRANSLATE_CMD) <input_content> # 智能中英互译"

# 只安装 prompt 命令
.PHONY: install-prompt
install-prompt: prompt
	@echo "$(YELLOW)📦 安装 $(PROMPT_CMD) 到 $(GOBIN)...$(NC)"
	@mkdir -p $(GOBIN)
	@cp $(BUILD_DIR)/$(PROMPT_CMD) $(GOBIN)/
	@chmod +x $(GOBIN)/$(PROMPT_CMD)
	@echo "$(GREEN)✅ $(PROMPT_CMD) 安装完成: $(GOBIN)/$(PROMPT_CMD)$(NC)"

# 清理构建文件
.PHONY: clean
clean:
	@echo "$(YELLOW)🧹 清理构建文件...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@echo "$(GREEN)✅ 清理完成$(NC)"

# 运行测试
.PHONY: test
test:
	@echo "$(YELLOW)🧪 运行测试...$(NC)"
	@$(GO) test -v ./...
	@echo "$(GREEN)✅ 测试完成$(NC)"

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "$(YELLOW)🧪 运行测试并生成覆盖率报告...$(NC)"
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ 测试完成，覆盖率报告: coverage.html$(NC)"

# 代码检查
.PHONY: lint
lint:
	@echo "$(YELLOW)🔍 运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint 未安装，使用 go vet...$(NC)"; \
		$(GO) vet ./...; \
	fi
	@echo "$(GREEN)✅ 代码检查完成$(NC)"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "$(YELLOW)🎨 格式化代码...$(NC)"
	@$(GO) fmt ./...
	@echo "$(GREEN)✅ 代码格式化完成$(NC)"

# 安装/更新依赖
.PHONY: deps
deps:
	@echo "$(YELLOW)📦 更新依赖...$(NC)"
	@$(GO) mod tidy
	@$(GO) mod download
	@echo "$(GREEN)✅ 依赖更新完成$(NC)"

# 创建发布包
.PHONY: dist
dist: clean $(DIST_DIR)
	@echo "$(YELLOW)📦 创建发布包...$(NC)"
	# Linux amd64
	@echo "$(BLUE)构建 Linux amd64...$(NC)"
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(PROMPT_CMD)-linux-amd64 ./cmd/prompt
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(DEVKIT_CMD)-linux-amd64 ./cmd/devkit
	# Linux arm64
	@echo "$(BLUE)构建 Linux arm64...$(NC)"
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(PROMPT_CMD)-linux-arm64 ./cmd/prompt
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(DEVKIT_CMD)-linux-arm64 ./cmd/devkit
	# macOS amd64
	@echo "$(BLUE)构建 macOS amd64...$(NC)"
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(PROMPT_CMD)-darwin-amd64 ./cmd/prompt
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(DEVKIT_CMD)-darwin-amd64 ./cmd/devkit
	# macOS arm64
	@echo "$(BLUE)构建 macOS arm64...$(NC)"
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(PROMPT_CMD)-darwin-arm64 ./cmd/prompt
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(DEVKIT_CMD)-darwin-arm64 ./cmd/devkit
	# Windows amd64
	@echo "$(BLUE)构建 Windows amd64...$(NC)"
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(PROMPT_CMD)-windows-amd64.exe ./cmd/prompt
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(DEVKIT_CMD)-windows-amd64.exe ./cmd/devkit
	@echo "$(GREEN)✅ 发布包创建完成!$(NC)"
	@echo "$(BLUE)📦 发布文件:$(NC)"
	@ls -la $(DIST_DIR)/

# 显示项目信息
.PHONY: info
info:
	@echo "$(BLUE)项目信息:$(NC)"
	@echo "  项目名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  Go 版本: $(shell $(GO) version)"
	@echo "  GOBIN: $(GOBIN)"
	@echo "  构建目录: $(BUILD_DIR)"
	@echo "  发布目录: $(DIST_DIR)"

# 卸载命令
.PHONY: uninstall
uninstall:
	@echo "$(YELLOW)🗑️  卸载命令...$(NC)"
	@rm -f $(GOBIN)/$(PROMPT_CMD)
	@rm -f $(GOBIN)/$(TRANSLATE_CMD)
	@echo "$(GREEN)✅ 卸载完成$(NC)"

# 快速测试 prompt 命令 (需要设置 API key)
.PHONY: test-prompt
test-prompt: prompt
	@echo "$(YELLOW)🧪 测试 prompt 命令...$(NC)"
	@echo "$(BLUE)测试直接文本输入:$(NC)"
	@$(BUILD_DIR)/$(PROMPT_CMD) "hello world" || echo "$(RED)测试失败 - 请检查环境变量配置$(NC)"
	@echo "$(GREEN)✅ prompt 测试完成$(NC)"

# 开发模式 - 监听文件变化并自动重新编译 (需要安装 entr)
.PHONY: dev
dev:
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(YELLOW)👨‍💻 开发模式启动 (使用 Ctrl+C 退出)...$(NC)"; \
		find . -name "*.go" | entr -r make install-prompt; \
	else \
		echo "$(RED)❌ 需要安装 entr 工具用于文件监听$(NC)"; \
		echo "$(YELLOW)安装命令: brew install entr (macOS) 或 apt-get install entr (Ubuntu)$(NC)"; \
	fi