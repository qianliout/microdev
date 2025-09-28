#!/bin/bash

# 测试环境变量配置功能

set -e

echo "🧪 测试环境变量配置功能"
echo "================================"

# 确保命令可用
if ! command -v m &> /dev/null; then
    echo "❌ 命令 'm' 不可用，请先运行 make install"
    exit 1
fi

echo "✅ 命令 'm' 可用"

# 测试1: 默认配置
echo ""
echo "📋 测试1: 默认配置"
echo "预期: model=qwen3-max"
m t "默认配置测试" 2>&1 | grep "翻译服务初始化成功" || echo "❌ 未找到初始化日志"

# 测试2: 自定义模型名
echo ""
echo "📋 测试2: 自定义模型名"
echo "预期: model=qwen-plus"
if MICRO_MODEL_NAME=qwen-plus m t "自定义模型测试" 2>&1 | grep "qwen-plus" > /dev/null; then
    echo "✅ 模型名配置成功"
else
    echo "❌ 模型名配置失败"
fi

# 测试3: 无效模型名回退
echo ""
echo "📋 测试3: 无效模型名回退"
echo "预期: model=qwen3-max (回退到默认值)"
if MICRO_MODEL_NAME=invalid-model m t "无效模型测试" 2>&1 | grep "qwen3-max" > /dev/null; then
    echo "✅ 无效模型正确回退"
else
    echo "❌ 无效模型回退失败"
fi

# 测试4: 多个环境变量组合
echo ""
echo "📋 测试4: 多个环境变量组合"
echo "预期: model=qwen-turbo"
if MICRO_MODEL_NAME=qwen-turbo MICRO_MAX_TOKENS=1024 MICRO_TEMPERATURE=0.3 m t "组合配置测试" 2>&1 | grep "qwen-turbo" > /dev/null; then
    echo "✅ 组合配置成功"
else
    echo "❌ 组合配置失败"
fi

# 测试5: 支持的模型列表测试
echo ""
echo "📋 测试5: 支持的模型列表测试"
models=("qwen-plus" "qwen-turbo" "qwen-max" "qwen2.5-7b-instruct")

for model in "${models[@]}"; do
    echo "测试模型: $model"
    if MICRO_MODEL_NAME=$model m t "测试$model" 2>&1 | grep "$model" > /dev/null; then
        echo "✅ $model 支持"
    else
        echo "❌ $model 不支持"
    fi
done

echo ""
echo "🎉 环境变量配置测试完成！"
echo "================================"
echo ""
echo "💡 使用提示："
echo "export MICRO_MODEL_NAME=qwen-plus     # 设置模型"
echo "export MICRO_MAX_TOKENS=1024          # 设置最大Token数"
echo "export MICRO_TEMPERATURE=0.5          # 设置温度"
echo "export MICRO_CONCURRENCY=5            # 设置并发数"
echo "export MICRO_STREAM_OUTPUT=false      # 禁用流式输出"
