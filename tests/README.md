# MicroDev 测试套件

本目录包含了 MicroDev 项目的完整测试套件，使用 [GoConvey](https://github.com/smartystreets/goconvey) 框架编写。

## 测试文件说明

### 1. `basic_test.go` - 基础功能测试
- **日志器测试**: 验证日志系统正常工作
- **错误处理测试**: 测试各种错误类型的创建和用户友好消息
- **配置管理测试**: 测试配置加载、默认值、环境变量读取等

### 2. `llm_connectivity_test.go` - LLM连接性测试
- **配置加载测试**: 测试不同环境变量配置下的行为
- **LLM客户端创建测试**: 测试客户端创建的成功和失败场景
- **翻译器创建测试**: 测试翻译器创建的各种情况
- **真实LLM连接测试**: 需要真实API密钥的集成测试

### 3. `integration_test.go` - 集成测试
- **命令集成测试**: 测试 Cobra 命令的创建和配置
- **输入处理测试**: 测试各种输入场景的处理
- **输出处理测试**: 测试控制台和文件输出
- **语言检测测试**: 测试中英文语言检测算法
- **配置验证测试**: 测试配置对象的验证逻辑
- **端到端工作流测试**: 完整流程的集成测试

## 运行测试

### 快速运行
使用提供的测试脚本：
```bash
./tests/run_tests.sh
```

### 手动运行特定测试
```bash
# 运行基础功能测试
go test -v ./tests -run TestBasicFunctionality

# 运行LLM连接性测试
go test -v ./tests -run TestLLMConnectivity

# 运行集成测试
go test -v ./tests -run TestCommandIntegration

# 运行所有测试
go test -v ./tests
```

### 生成覆盖率报告
```bash
go test -coverprofile=coverage.out ./tests
go tool cover -html=coverage.out -o coverage.html
```

## 环境变量配置

为了运行完整的测试套件（包括真实LLM连接测试），需要设置以下环境变量之一：

```bash
# 使用 DashScope API
export DASHSCOPE_API_KEY=your_dashscope_api_key

# 或使用阿里百炼API
export ALI_BAILIAN_API_KEY=your_bailian_api_key
```

如果未设置API密钥，真实连接测试将被跳过，但其他所有测试仍会正常运行。

## GoConvey Web界面

GoConvey 提供了一个优雅的Web界面来查看测试结果：

```bash
# 安装 GoConvey 命令行工具（如果尚未安装）
go install github.com/smartystreets/goconvey

# 启动Web界面
goconvey -port=8080
```

然后在浏览器中访问 `http://localhost:8080` 查看实时测试结果。

## 测试覆盖的功能

### ✅ 已测试的功能
- 配置管理和环境变量处理
- 错误处理和用户友好消息
- 日志系统
- 命令行接口（Cobra命令）
- 输入处理（文本、文件）
- 输出处理（控制台、文件）
- 语言检测算法
- LLM客户端和翻译器的创建
- 基本的集成流程

### 🔄 需要真实API密钥的测试
- 实际的LLM API调用
- 真实的提示词生成
- 真实的翻译功能
- 流式输出测试

### 📝 测试统计
- **总测试用例**: 8个主要测试函数
- **总断言数**: 115+ 个断言
- **测试类型**: 单元测试 + 集成测试
- **框架**: GoConvey (BDD风格)

## 故障排除

### 常见问题

1. **依赖问题**
   ```bash
   go mod tidy
   go get github.com/smartystreets/goconvey/convey
   ```

2. **API密钥未设置**
   - 真实连接测试会被跳过
   - 其他测试正常运行
   - 在日志中会显示相应提示

3. **网络问题**
   - 真实LLM测试设置了30秒超时
   - 超时不会导致测试失败
   - 会在日志中显示超时信息

## 贡献指南

添加新测试时请遵循以下规范：

1. 使用 GoConvey 的 BDD 风格
2. 测试函数以 `Test` 开头
3. 使用描述性的测试名称
4. 包含正面和负面测试用例
5. 适当使用 `t.Skip()` 跳过需要外部依赖的测试
6. 添加必要的清理代码（defer语句）

## 持续集成

这些测试设计为可以在CI/CD环境中运行：
- 不依赖外部服务的测试总是运行
- 需要API密钥的测试会自动跳过
- 所有测试都有合理的超时设置
- 生成标准的Go测试输出格式
