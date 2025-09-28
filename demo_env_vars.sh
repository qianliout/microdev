#!/bin/bash

# 环境变量配置功能演示脚本

echo "🎯 MicroDev 环境变量配置功能演示"
echo "=================================="
echo ""

# 确保命令可用
if ! command -v m &> /dev/null; then
    echo "❌ 命令 'm' 不可用，请先运行 make install"
    exit 1
fi

echo "✅ 命令 'm' 可用"
echo ""

echo "📋 1. 默认配置演示"
echo "-------------------"
echo "使用默认配置进行翻译："
echo "$ m t \"Hello World\""
m t "Hello World" 2>&1 | head -1
echo ""

echo "📋 2. 自定义模型演示"
echo "-------------------"
echo "使用 qwen-plus 模型："
echo "$ MICRO_MODEL_NAME=qwen-plus m t \"Hello World\""
MICRO_MODEL_NAME=qwen-plus m t "Hello World" 2>&1 | head -1
echo ""

echo "使用 qwen-turbo 模型（更快响应）："
echo "$ MICRO_MODEL_NAME=qwen-turbo m t \"Hello World\""
MICRO_MODEL_NAME=qwen-turbo m t "Hello World" 2>&1 | head -1
echo ""

echo "📋 3. 参数组合演示"
echo "-------------------"
echo "组合多个环境变量："
echo "$ MICRO_MODEL_NAME=qwen-plus MICRO_MAX_TOKENS=1024 MICRO_TEMPERATURE=0.3 m t \"Hello World\""
MICRO_MODEL_NAME=qwen-plus MICRO_MAX_TOKENS=1024 MICRO_TEMPERATURE=0.3 m t "Hello World" 2>&1 | head -1
echo ""

echo "📋 4. 无效配置处理演示"
echo "---------------------"
echo "使用无效模型名（会自动回退到默认值）："
echo "$ MICRO_MODEL_NAME=invalid-model m t \"Hello World\""
MICRO_MODEL_NAME=invalid-model m t "Hello World" 2>&1 | head -1
echo ""

echo "📋 5. 支持的模型列表"
echo "-------------------"
echo "当前支持的模型："
echo "• qwen3-max (默认) - 最新最强模型"
echo "• qwen-plus - 平衡性能和成本"
echo "• qwen-turbo - 快速响应"
echo "• qwen-max - 最大模型"
echo "• qwen2.5-72b-instruct - 72B 参数开源模型"
echo "• qwen2.5-32b-instruct - 32B 参数开源模型"
echo "• qwen2.5-14b-instruct - 14B 参数开源模型"
echo "• qwen2.5-7b-instruct - 7B 参数开源模型"
echo ""

echo "📋 6. 环境变量配置说明"
echo "---------------------"
echo "可配置的环境变量："
echo ""
echo "🔧 模型配置："
echo "  MICRO_MODEL_NAME     - 指定使用的模型（默认: qwen3-max）"
echo "  MICRO_MAX_TOKENS     - 最大Token数 (1-8192，默认: 2048)"
echo "  MICRO_TEMPERATURE    - 温度值 (0.0-2.0，默认: 0.7)"
echo ""
echo "⚡ 性能配置："
echo "  MICRO_CONCURRENCY    - 并发数 (1-10，默认: 3)"
echo "  MICRO_STREAM_OUTPUT  - 流式输出 (true/false，默认: true)"
echo ""
echo "🔑 API配置："
echo "  DASHSCOPE_API_KEY    - DashScope API 密钥"
echo "  ALI_BAILIAN_API_KEY  - 阿里百炼 API 密钥"
echo ""

echo "📋 7. 实用示例"
echo "---------------"
echo ""
echo "💡 快速翻译（使用 turbo 模型）："
echo "export MICRO_MODEL_NAME=qwen-turbo"
echo "m t \"需要快速翻译的文本\""
echo ""
echo "💡 高质量翻译（使用 max 模型）："
echo "export MICRO_MODEL_NAME=qwen3-max"
echo "m t \"需要高质量翻译的文本\""
echo ""
echo "💡 批量处理（高并发）："
echo "export MICRO_CONCURRENCY=8"
echo "m t ./docs/ -d"
echo ""
echo "💡 创意提示词优化（高温度）："
echo "export MICRO_TEMPERATURE=0.9"
echo "m p \"创意写作提示词\""
echo ""

echo "🎉 演示完成！"
echo "============="
echo ""
echo "💡 提示：您可以将常用的环境变量配置添加到 ~/.bashrc 或 ~/.zshrc 中："
echo "echo 'export MICRO_MODEL_NAME=qwen-plus' >> ~/.bashrc"
echo "echo 'export MICRO_CONCURRENCY=5' >> ~/.bashrc"
echo ""
echo "📚 更多信息请查看 README.md 中的环境变量配置部分"
