#!/bin/bash

# 测试运行脚本 - 使用GoConvey运行所有测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 开始运行MicroDev项目测试...${NC}"

# 进入项目根目录
cd "$(dirname "$0")/.."

echo -e "${YELLOW}📁 当前目录: $(pwd)${NC}"

# 检查Go环境
echo -e "${YELLOW}🔍 检查Go环境...${NC}"
go version

# 检查依赖
echo -e "${YELLOW}📦 检查依赖...${NC}"
go mod tidy

# 检查API密钥环境变量
echo -e "${YELLOW}🔑 检查API密钥环境变量...${NC}"
if [ -n "$DASHSCOPE_API_KEY" ]; then
    echo -e "${GREEN}✅ 检测到DASHSCOPE_API_KEY${NC}"
elif [ -n "$ALI_BAILIAN_API_KEY" ]; then
    echo -e "${GREEN}✅ 检测到ALI_BAILIAN_API_KEY${NC}"
else
    echo -e "${YELLOW}⚠️  未检测到API密钥环境变量${NC}"
    echo -e "${YELLOW}   真实LLM连接测试将被跳过${NC}"
    echo -e "${YELLOW}   要运行完整测试，请设置以下环境变量之一：${NC}"
    echo -e "${YELLOW}   - export DASHSCOPE_API_KEY=your_key${NC}"
    echo -e "${YELLOW}   - export ALI_BAILIAN_API_KEY=your_key${NC}"
fi

echo ""

# 运行基础测试
echo -e "${BLUE}🔧 运行基础功能测试...${NC}"
go test -v ./tests -run TestBasicFunctionality

echo ""

# 运行环境设置测试
echo -e "${BLUE}🌍 运行环境设置测试...${NC}"
go test -v ./tests -run TestEnvironmentSetup

echo ""

# 运行集成测试
echo -e "${BLUE}🔗 运行集成测试...${NC}"
go test -v ./tests -run TestCommandIntegration
go test -v ./tests -run TestInputProcessing
go test -v ./tests -run TestOutputProcessing
go test -v ./tests -run TestLanguageDetection
go test -v ./tests -run TestConfigValidation

echo ""

# 运行LLM连接性测试
echo -e "${BLUE}🤖 运行LLM连接性测试...${NC}"
go test -v ./tests -run TestLLMConnectivity

echo ""

# 运行端到端测试
echo -e "${BLUE}🎯 运行端到端测试...${NC}"
go test -v ./tests -run TestEndToEndWorkflow

echo ""

# 运行所有测试（包括真实连接测试）
echo -e "${BLUE}🚀 运行完整测试套件...${NC}"
go test -v ./tests

echo ""

# 检查测试覆盖率
echo -e "${BLUE}📊 生成测试覆盖率报告...${NC}"
go test -coverprofile=coverage.out ./tests
if [ -f coverage.out ]; then
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✅ 覆盖率报告已生成: coverage.html${NC}"
    
    # 显示覆盖率统计
    go tool cover -func=coverage.out | tail -1
else
    echo -e "${YELLOW}⚠️  未生成覆盖率文件${NC}"
fi

echo ""

# 运行GoConvey Web界面（可选）
echo -e "${YELLOW}💡 提示: 要启动GoConvey Web界面，请运行:${NC}"
echo -e "${YELLOW}   goconvey -port=8080${NC}"
echo -e "${YELLOW}   然后在浏览器中访问 http://localhost:8080${NC}"

echo ""
echo -e "${GREEN}🎉 测试运行完成！${NC}"

# 清理临时文件
if [ -f coverage.out ]; then
    echo -e "${YELLOW}🧹 清理临时文件...${NC}"
    # 保留coverage文件，用户可能需要查看
    echo -e "${YELLOW}   coverage.out 和 coverage.html 已保留${NC}"
fi
