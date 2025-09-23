#!/bin/bash

# 编译脚本 - 编译 prompt 命令并安装到 GOBIN

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔨 开始编译 prompt 命令...${NC}"

# 检查 GOBIN 是否设置
if [ -z "$GOBIN" ]; then
    echo -e "${YELLOW}⚠️  GOBIN 未设置，使用默认值 \$GOPATH/bin${NC}"
    if [ -z "$GOPATH" ]; then
        GOBIN="$HOME/go/bin"
    else
        GOBIN="$GOPATH/bin"
    fi
fi

echo -e "${YELLOW}📁 GOBIN 目录: $GOBIN${NC}"

# 确保 GOBIN 目录存在
mkdir -p "$GOBIN"

# 进入项目根目录
cd "$(dirname "$0")"

# 编译 prompt 命令
echo -e "${YELLOW}🔨 编译 prompt 命令...${NC}"
go build -o "$GOBIN/prompt" ./cmd/prompt

# 检查编译是否成功
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 编译成功!${NC}"
    
    # 设置可执行权限
    chmod +x "$GOBIN/prompt"
    echo -e "${GREEN}✅ 已设置可执行权限${NC}"
    
    # 显示文件信息
    ls -la "$GOBIN/prompt"
    
    echo -e "${GREEN}🎉 prompt 命令已安装到: $GOBIN/prompt${NC}"
    echo -e "${YELLOW}💡 请确保 $GOBIN 在您的 PATH 中${NC}"
    echo -e "${YELLOW}💡 使用方法: prompt <input_content>${NC}"
    echo -e "${YELLOW}💡 示例: prompt 'hello world'${NC}"
    echo -e "${YELLOW}💡 示例: prompt /path/to/file.txt${NC}"
else
    echo -e "${RED}❌ 编译失败!${NC}"
    exit 1
fi
