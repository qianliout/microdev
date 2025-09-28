# MicroDev - 微开发工具集

MicroDev 是一个基于 Go 语言开发的微开发工具集，提供提示词优化和中英互译功能，帮助开发者提高工作效率。

## 功能特性

### 🚀 提示词优化 (Prompt)
- 智能优化用户输入的提示词
- 支持文件输入和直接文本输入
- 基于大语言模型生成更好的提示词
- **🆕 交互式会话模式**：保持对话历史和上下文
- **🧠 智能记忆管理**：自动压缩长对话历史
- **🔄 上下文感知优化**：基于对话历史提供更精准的优化建议

### 🌐 中英互译 (Translate)
- 智能检测语言并进行中英互译
- 支持单个文件翻译
- 支持目录批量翻译
- 支持直接文本翻译
- 保持 Markdown 格式完整性
- 自动生成翻译后的文件
- **翻译幂等性**：避免重复翻译已存在的文件
- **智能输出路径**：翻译文件保存在源文件同目录下
- **并发翻译**：支持多文件并发处理，提高翻译效率
- **🆕 交互式翻译模式**：支持连续翻译会话

### 🚀 短命令支持 (New!)
- **短命令 `m`**：使用 `m` 代替 `micro`，节省输入时间
- **完全兼容**：所有 `micro` 命令都可以用 `m` 替代
- **提高效率**：日常使用更加便捷快速

```bash
# 传统命令
micro p "优化提示词"
micro t "翻译文本"
micro t -i

# 🆕 短命令（推荐）
m p "优化提示词"      # 节省 4 个字符
m t "翻译文本"        # 节省 4 个字符
m t -i               # 节省 4 个字符
```

## 安装

### 前置要求
- Go 1.24.1 或更高版本
- 阿里云 DashScope API 密钥或阿里百炼 API 密钥

### 编译安装

1. 克隆项目
```bash
git clone <repository-url>
cd microdev
```

2. 编译并安装

使用构建脚本：
```bash
./build.sh
```

或使用 Makefile：
```bash
# 构建项目
make build

# 安装到 GOBIN
make install

# 查看所有可用命令
make help
```

编译成功后，`micro` 命令和短命令 `m` 都将被安装到 `$GOBIN` 目录中。请确保 `$GOBIN` 在您的 `PATH` 环境变量中。

**🎉 安装完成后，您将获得两个命令：**
- `micro` - 完整命令名
- `m` - 短命令（推荐日常使用）

## 使用方法

### 基本语法
```bash
micro <子命令> [参数] [选项]
# 或使用短命令
m <子命令> [参数] [选项]
```

### 提示词优化 (prompt/p)

#### 基本用法
```bash
# 使用完整命令名
micro prompt "优化这个提示词"

# 使用简写
micro p "优化这个提示词"

# 🆕 使用短命令（推荐）
m p "优化这个提示词"

# 从文件读取
micro p /path/to/prompt.txt
m p /path/to/prompt.txt       # 🆕 使用短命令

# 🆕 交互式会话模式
micro p -i
m p -i                        # 🆕 使用短命令
```

#### 交互式会话模式特性
- **对话记忆**：自动记住当前会话中的完整对话历史
- **记忆压缩**：当对话轮次较多时，主动对之前的对话进行智能总结
- **持续对话**：保持对话连续性，直到用户明确表示退出
- **上下文感知**：基于对话历史提供更精准的优化建议
- **退出命令**：支持 `exit`、`quit`、`退出`、`q`、`bye`、`再见` 等命令

#### 示例
```bash
# 优化简单文本
micro p "帮我写一个Python函数"
m p "帮我写一个Python函数"    # 🆕 使用短命令

# 从文件优化
micro p ./prompts/my_prompt.txt
m p ./prompts/my_prompt.txt   # 🆕 使用短命令

# 🆕 进入交互式会话模式
micro p -i
m p -i                        # 🆕 使用短命令
# 然后可以连续输入多个提示词进行优化
# 系统会记住对话历史，提供更好的上下文感知优化
```

### 中英互译 (translate/t)

#### 基本用法
```bash
# 使用完整命令名
micro translate "Hello World"

# 使用简写
micro t "Hello World"

# 🆕 使用短命令（推荐）
m t "Hello World"

# 🆕 交互式翻译模式
m t -i
```

#### 翻译选项
- `-f, --file`: 明确指定为文件翻译模式
- `-d, --dir`: 明确指定为目录翻译模式
- `-t, --type`: 指定文件类型过滤（默认为 md）
- **🆕 `-i, --interactive`**: 启用交互式翻译模式

#### 示例

##### 直接文本翻译
```bash
# 中译英
micro t "你好世界"
m t "你好世界"          # 🆕 使用短命令

# 英译中
micro t "Hello World"
m t "Hello World"       # 🆕 使用短命令
```

##### 🆕 交互式翻译模式
```bash
# 进入交互式翻译模式
micro t -i
m t -i                  # 🆕 使用短命令

# 交互式会话示例：
# 🌍 请输入要翻译的文本: Hello World
# [原文]
# Hello World
#
# [译文]
# 你好世界
```

##### 单文件翻译
```bash
# 自动检测文件
micro t ./document.md
m t ./document.md       # 🆕 使用短命令

# 明确指定文件模式
micro t -f ./document.md
m t -f ./document.md    # 🆕 使用短命令
```

##### 目录批量翻译
```bash
# 翻译目录下所有 Markdown 文件
micro t -d ./docs
m t -d ./docs           # 🆕 使用短命令

# 翻译目录下所有 txt 文件
micro t -d ./docs -t txt
m t -d ./docs -t txt    # 🆕 使用短命令

# 翻译目录下所有 Go 文件
micro t -d ./src -t go
m t -d ./src -t go      # 🆕 使用短命令

# 使用并发翻译（5个并发）
micro t -d ./docs -c 5
m t -d ./docs -c 5      # 🆕 使用短命令

# 强制重新翻译已存在的文件
micro t -d ./docs --force
m t -d ./docs --force   # 🆕 使用短命令
```

#### 新功能特性

##### 🔄 翻译幂等性
- 自动检测翻译文件是否已存在（如 `document_zh.md`、`document_en.md`）
- 默认跳过已翻译的文件，避免重复翻译
- 使用 `--force` 参数强制重新翻译

##### ⚡ 并发翻译
- 支持并发调用大模型，提高翻译效率
- 通过 `-c` 参数或 `MICRO_CONCURRENCY` 环境变量控制并发数
- 默认并发数为 3，最大为 10

##### 📁 智能输出路径
- 翻译文件自动保存到源文件同目录下
- 根据翻译方向自动添加后缀：`_zh`（中文）或 `_en`（英文）
- 保持原文件扩展名不变

## 配置

### 环境变量配置

在使用前，需要设置以下环境变量：

#### 🔑 必需配置

```bash
# API 密钥（二选一）
export DASHSCOPE_API_KEY=your_dashscope_api_key
# 或
export ALI_BAILIAN_API_KEY=your_bailian_api_key
```

#### ⚙️ 可选配置

```bash
# 🆕 模型配置
export MICRO_MODEL_NAME=qwen-plus           # 指定使用的模型
export MICRO_MAX_TOKENS=1024                # 最大Token数 (1-8192)
export MICRO_TEMPERATURE=0.5                # 温度值 (0.0-2.0)

# 性能配置
export MICRO_CONCURRENCY=5                  # 并发数 (1-10，默认3)
export MICRO_STREAM_OUTPUT=true             # 流式输出 (true/false，默认true)
```

#### 🎯 支持的模型列表

```bash
# 通义千问系列
qwen3-max              # 🆕 默认模型，最新最强
qwen-plus              # 平衡性能和成本
qwen-plus-latest       # 最新版本
qwen-turbo             # 快速响应
qwen-turbo-latest      # 最新快速版本
qwen-max               # 最大模型
qwen-max-latest        # 最新最大模型

# 开源模型系列
qwen2.5-72b-instruct   # 72B 参数模型
qwen2.5-32b-instruct   # 32B 参数模型
qwen2.5-14b-instruct   # 14B 参数模型
qwen2.5-7b-instruct    # 7B 参数模型
```

#### 💡 使用示例

```bash
# 使用不同模型进行翻译
MICRO_MODEL_NAME=qwen-turbo m t "快速翻译"
MICRO_MODEL_NAME=qwen3-max m t "高质量翻译"

# 调整参数进行提示词优化
MICRO_MODEL_NAME=qwen-plus MICRO_TEMPERATURE=0.8 m p "创意提示词"

# 批量翻译时使用高并发
MICRO_CONCURRENCY=8 m t ./docs/ -d
```

### 获取 API 密钥

1. **DashScope API**: 访问 [阿里云 DashScope 控制台](https://dashscope.console.aliyun.com/) 获取 API 密钥
2. **阿里百炼 API**: 访问 [阿里百炼控制台](https://bailian.console.aliyun.com/) 获取 API 密钥

### 默认配置

- **模型**: qwen3-max（可通过 `MICRO_MODEL_NAME` 修改）
- **最大令牌数**: 2048（可通过 `MICRO_MAX_TOKENS` 修改）
- **温度**: 0.7（可通过 `MICRO_TEMPERATURE` 修改）
- **流式输出**: 启用（可通过 `MICRO_STREAM_OUTPUT` 修改）
- **并发数**: 3（可通过 `MICRO_CONCURRENCY` 修改）

## 项目结构

```text
microdev/
├── cmd/
│   └── micro/           # 主命令入口
├── component/           # 功能组件（按子命令组织）
│   ├── prompt/         # 提示词优化组件
│   └── translate/      # 翻译功能组件（包含翻译器和Markdown处理）
├── pkg/                 # 公共功能包（无相互依赖）
│   ├── config/         # 配置管理
│   ├── errors/         # 错误处理
│   ├── input/          # 输入处理
│   ├── llm/            # 大语言模型客户端
│   ├── logger/         # 日志管理
│   ├── output/         # 输出处理
│   └── utils/          # 工具函数
├── tests/              # 测试套件
├── build.sh            # 构建脚本
├── go.mod              # Go 模块文件
└── README.md           # 项目说明
```

## 开发

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试套件（推荐）
./tests/run_tests.sh
```

### 构建开发版本

```bash
go build -o micro ./cmd/micro
```

## 📚 文档

### 子命令文档
- 🚀 [Prompt 子命令](docs/prompt/README.md) - 提示词优化工具完整指南
- 🌐 [Translate 子命令](docs/translate/README.md) - 中英互译工具完整指南
- ⚙️ [通用配置](docs/general/README.md) - 环境配置和最佳实践

### 详细文档
- [文档中心](docs/README.md) - 完整的文档导航和索引
- [交互式功能演示](docs/prompt/interactive-demo.md) - 详细的交互模式使用指南
- [详细日志输出演示](docs/prompt/detailed-logging-demo.md) - 日志功能和调试指南
- [会话存储架构设计](docs/prompt/session-storage-architecture.md) - 模块化存储架构详解

## 🧪 测试

所有测试文件都位于 `tests/` 目录下，包含：

- **基础功能测试** (`basic_test.go`) - 核心组件和配置测试
- **集成测试** (`integration_test.go`) - 端到端工作流测试
- **LLM连接测试** (`llm_connectivity_test.go`) - 语言模型连接测试
- **会话管理测试** (`session_manager_test.go`) - 会话功能测试
- **工具函数测试** (`utils_test.go`) - 辅助函数测试

### 运行测试

```bash
# 运行所有测试
go test ./tests/ -v

# 运行特定测试
go test ./tests/session_manager_test.go -v

# 运行测试并查看覆盖率
go test ./tests/ -cover
```

## 🔍 日志和调试

系统提供详细的日志输出，帮助您了解：

- **📝 记忆内容**：用户输入和助手回复的详细记录
- **🧠 上下文构建**：如何构建和使用对话上下文
- **🗜️ 记忆压缩**：何时触发压缩以及压缩了什么内容
- **📤 LLM交互**：发送给语言模型的完整内容
- **⏱️ 性能监控**：响应时间和处理统计

查看 [详细日志输出演示](docs/detailed-logging-demo.md) 了解更多信息。

## 🆕 最新更新

### v2.0 新功能亮点

#### 🚀 短命令支持
- **新增短命令 `m`**：使用 `m` 代替 `micro`，提高输入效率
- **完全兼容**：所有功能保持不变，只是命令更短
- **推荐使用**：日常使用建议采用短命令

#### 🌐 交互式翻译模式
- **新增 `-i/--interactive` 参数**：支持连续翻译会话
- **智能退出**：支持多种退出命令（exit, quit, 退出等）
- **实时翻译**：输入即翻译，体验流畅

#### 📝 优化日志系统
- **结构化日志**：使用 zerolog 替代 fmt 打印
- **用户友好**：区分调试日志和用户消息
- **更好调试**：提供详细的系统运行信息

### 快速体验新功能

```bash
# 安装最新版本
./build.sh

# 体验短命令
m p "测试提示词优化"
m t "测试翻译功能"

# 体验交互式翻译
m t -i
```

## 🔧 项目改进点

### 🎯 当前已知改进点

#### 1. 代码重复和架构问题

**🔍 具体问题**：
- **LLM 客户端重复创建**：`pkg/llm/client.go` 和 `component/translate/service/translator.go` 中存在相同的 LLM 客户端创建逻辑（第 30-49 行）
- **配置硬编码**：`pkg/config/config.go` 中并发数限制硬编码为 10，缺乏灵活性
- **存储接口冗余**：`component/prompt/session/` 目录下有多个存储相关文件，但实际只使用内存存储

**🛠️ 改进方案**：
```go
// 1. 创建统一的 LLM 工厂
type LLMFactory struct {
    config *config.Config
    logger *logger.Logger
}

func (f *LLMFactory) CreateClient() (llms.Model, error) {
    // 统一的客户端创建逻辑
}

// 2. 配置文件支持
type Config struct {
    LLM struct {
        ModelName   string `yaml:"model_name"`
        MaxTokens   int    `yaml:"max_tokens"`
        Temperature float64 `yaml:"temperature"`
    } `yaml:"llm"`

    Performance struct {
        MaxConcurrency int `yaml:"max_concurrency"`
        Timeout        time.Duration `yaml:"timeout"`
    } `yaml:"performance"`
}
```

#### 2. 错误处理不一致

**🔍 具体问题**：
- **错误类型使用不统一**：某些地方直接返回 `fmt.Errorf`，某些地方使用 `errors.NewXXXError`
- **错误上下文缺失**：`pkg/errors/errors.go` 中的错误信息缺乏调用栈和上下文信息
- **用户友好性不足**：错误信息对普通用户不够友好，缺乏解决建议

**🛠️ 改进方案**：
```go
// 增强错误结构
type AppError struct {
    Type        ErrorType
    Message     string
    Cause       error
    Context     map[string]interface{} // 新增：错误上下文
    Suggestions []string               // 新增：解决建议
    Code        string                 // 新增：错误代码
}

// 使用示例
func (c *Config) Validate() error {
    if c.GetAPIKey() == "" {
        return errors.NewConfigError("API密钥未配置", nil).
            WithContext("env_vars", []string{"DASHSCOPE_API_KEY", "ALI_BAILIAN_API_KEY"}).
            WithSuggestion("请设置环境变量：export DASHSCOPE_API_KEY=your_key").
            WithCode("CONFIG_001")
    }
}
```

#### 3. 测试覆盖率和质量问题

**🔍 具体问题**：
- **测试覆盖率不足**：`tests/basic_test.go` 中的 `min` 函数（第 178-183 行）应该使用标准库
- **测试依赖外部服务**：真实 LLM 测试依赖网络，导致 CI/CD 不稳定
- **缺乏性能测试**：没有针对并发翻译、大文件处理的性能测试

**🛠️ 改进方案**：
```go
// 1. 使用 testify/mock 进行 LLM 模拟
type MockLLM struct {
    mock.Mock
}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
    args := m.Called(ctx, messages, options)
    return args.Get(0).(*llms.ContentResponse), args.Error(1)
}

// 2. 性能基准测试
func BenchmarkConcurrentTranslation(b *testing.B) {
    // 测试并发翻译性能
}

func BenchmarkLargeFileProcessing(b *testing.B) {
    // 测试大文件处理性能
}
```

#### 4. 内存和资源管理

**🔍 具体问题**：
- **会话内存泄漏风险**：`component/prompt/session/memory_storage.go` 中的内存存储没有 TTL 清理机制
- **HTTP 连接未复用**：每次 LLM 调用都创建新连接，效率低下
- **大文件处理内存占用**：翻译大文件时一次性加载到内存

**🛠️ 改进方案**：
```go
// 1. 会话 TTL 管理
type MemoryStorage struct {
    sessions map[string]*SessionWithTTL
    mutex    sync.RWMutex
    ticker   *time.Ticker // 定期清理过期会话
}

type SessionWithTTL struct {
    Session   *Session
    ExpiresAt time.Time
}

// 2. HTTP 连接池
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

// 3. 流式文件处理
func (t *TranslatorService) TranslateFileStream(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        // 逐行处理，避免大文件内存占用
    }
}
```

### 🚀 技术债务

#### 1. 构建和部署问题

**🔍 具体问题**：
- **构建标签未实现**：`component/prompt/session/storage.go` 中声明支持 MySQL/Redis 构建标签，但实际文件 `mysql_storage.go` 和 `redis_storage.go` 为空
- **二进制大小**：当前编译后约 15MB，包含未使用的依赖
- **交叉编译缺失**：`Makefile` 和 `build.sh` 不支持多平台编译

**🛠️ 改进方案**：
```bash
# 1. 实现真正的构建标签
# mysql_storage.go
//go:build mysql
// +build mysql

package session

import "database/sql"
// 实际的 MySQL 实现

# 2. 多平台构建支持
make build-all:
	GOOS=linux GOARCH=amd64 go build -o micro-linux-amd64 ./cmd/micro
	GOOS=darwin GOARCH=amd64 go build -o micro-darwin-amd64 ./cmd/micro
	GOOS=windows GOARCH=amd64 go build -o micro-windows-amd64.exe ./cmd/micro

# 3. 依赖精简
go mod tidy
go build -ldflags="-s -w" -o micro ./cmd/micro  # 减少二进制大小
```

#### 2. 配置和环境管理

**🔍 具体问题**：
- **环境变量硬编码**：`pkg/config/config.go` 中环境变量名硬编码，不支持前缀配置
- **配置验证缺失**：没有配置项的有效性验证（如温度范围、Token 数量限制）
- **默认值分散**：默认配置分散在多个文件中，难以维护

**🛠️ 改进方案**：
```go
// 配置验证
func (c *Config) Validate() error {
    if c.Temperature < 0 || c.Temperature > 2 {
        return errors.NewConfigError("温度值必须在 0-2 之间", nil)
    }

    if c.MaxTokens < 1 || c.MaxTokens > 8192 {
        return errors.NewConfigError("Token 数量必须在 1-8192 之间", nil)
    }

    if c.Concurrency < 1 || c.Concurrency > 50 {
        return errors.NewConfigError("并发数必须在 1-50 之间", nil)
    }

    return nil
}

// 环境变量前缀支持
type EnvConfig struct {
    Prefix string // 默认 "MICRO"
}

func (e *EnvConfig) GetAPIKey() string {
    return os.Getenv(e.Prefix + "_API_KEY")
}
```

#### 3. 日志和监控缺陷

**🔍 具体问题**：
- **日志级别混乱**：`pkg/logger/logger.go` 中同时使用 `fmt.Println` 和结构化日志
- **性能指标缺失**：没有 API 调用耗时、成功率等关键指标
- **错误追踪不足**：缺乏请求 ID 和调用链追踪

**🛠️ 改进方案**：
```go
// 统一日志接口
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, err error, fields ...Field)

    // 用户友好输出（不记录到日志文件）
    UserInfo(msg string)
    UserError(msg string)
}

// 性能监控
type Metrics struct {
    APICallDuration   prometheus.HistogramVec
    APICallTotal      prometheus.CounterVec
    TranslationErrors prometheus.CounterVec
}

// 请求追踪
type RequestContext struct {
    RequestID string
    UserID    string
    StartTime time.Time
}
```

#### 4. 安全和稳定性问题

**🔍 具体问题**：
- **API 密钥明文传输**：环境变量中的 API 密钥可能被进程列表泄露
- **文件路径注入**：`component/translate/cmd/command.go` 中文件路径未验证，存在路径遍历风险
- **并发安全问题**：`component/prompt/session/memory_storage.go` 中的 map 操作未加锁

**🛠️ 改进方案**：
```go
// 1. 安全的密钥管理
type SecureConfig struct {
    keyring keyring.Keyring
}

func (s *SecureConfig) GetAPIKey() (string, error) {
    return s.keyring.Get("microdev", "api_key")
}

// 2. 路径验证
func ValidateFilePath(path string) error {
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return errors.NewInputError("不允许路径遍历", nil)
    }

    absPath, err := filepath.Abs(cleanPath)
    if err != nil {
        return errors.NewInputError("无效的文件路径", err)
    }

    // 检查是否在允许的目录内
    if !strings.HasPrefix(absPath, allowedBasePath) {
        return errors.NewInputError("文件路径超出允许范围", nil)
    }

    return nil
}

// 3. 并发安全的内存存储
type SafeMemoryStorage struct {
    sessions sync.Map // 使用 sync.Map 替代 map + mutex
    stats    atomic.Value
}
```

## 🛠️ 开发指南

### 📋 新功能开发注意事项

#### 1. 架构原则和实际约束

**🏗️ 目录结构规范**：
```bash
# 添加新功能时必须遵循的结构
microdev/
├── component/
│   └── [new-feature]/          # 新功能目录
│       ├── cmd/
│       │   └── command.go      # 必须：实现 NewXXXCommand() *cobra.Command
│       ├── model/
│       │   └── types.go        # 可选：数据模型定义
│       ├── service/
│       │   └── service.go      # 必须：核心业务逻辑
│       └── README.md           # 必须：功能文档
├── pkg/                        # 禁止：不要在这里添加特定功能代码
└── tests/
    └── [new-feature]_test.go   # 必须：功能测试
```

**🔧 集成要求**：
```go
// 1. 在 cmd/micro/main.go 中注册新命令
import newFeatureCmd "microdev/component/new-feature/cmd"

func main() {
    rootCmd.AddCommand(newFeatureCmd.NewNewFeatureCommand())
}

// 2. 新命令必须实现标准接口
func NewNewFeatureCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "new-feature",
        Short: "简短描述",
        Long:  "详细描述",
        RunE:  runNewFeatureCommand,
    }

    // 添加通用标志
    cmd.Flags().BoolP("interactive", "i", false, "交互式模式")
    cmd.Flags().BoolP("verbose", "v", false, "详细输出")

    return cmd
}
```

#### 2. 代码质量强制要求

**🚨 必须遵循的规范**：
```go
// ❌ 错误示例：直接使用 fmt.Errorf
if err != nil {
    return fmt.Errorf("something went wrong: %v", err)
}

// ✅ 正确示例：使用项目错误类型
if err != nil {
    return errors.NewInputError("文件读取失败", err)
}

// ❌ 错误示例：直接使用 fmt.Println
fmt.Println("Processing file:", filename)

// ✅ 正确示例：使用统一日志
log.Info().
    Str("file", filename).
    Msg("开始处理文件")

// 用户输出使用专门方法
log.UserInfo(fmt.Sprintf("正在处理文件: %s", filename))
```

**📝 注释和文档要求**：
```go
// ✅ 公共函数必须有详细注释
// ProcessDocument 处理文档并返回结果
//
// 参数:
//   - filePath: 文档文件路径，必须是绝对路径
//   - options: 处理选项，可以为 nil 使用默认值
//
// 返回:
//   - *ProcessResult: 处理结果，包含状态和输出路径
//   - error: 处理过程中的错误，使用项目错误类型
//
// 示例:
//   result, err := ProcessDocument("/path/to/doc.md", nil)
//   if err != nil {
//       return err
//   }
func ProcessDocument(filePath string, options *ProcessOptions) (*ProcessResult, error) {
    // 实现
}
```

#### 3. 测试要求和覆盖率

**🧪 强制测试要求**：
```go
// 每个新功能必须包含以下测试类型

// 1. 单元测试 - 测试核心逻辑
func TestNewFeatureCore(t *testing.T) {
    Convey("新功能核心逻辑测试", t, func() {
        Convey("正常输入", func() {
            result, err := ProcessInput("valid input")
            So(err, ShouldBeNil)
            So(result, ShouldNotBeNil)
        })

        Convey("异常输入", func() {
            _, err := ProcessInput("")
            So(err, ShouldNotBeNil)
            So(err.(*errors.AppError).Type, ShouldEqual, errors.InputError)
        })
    })
}

// 2. 集成测试 - 测试命令行接口
func TestNewFeatureCommand(t *testing.T) {
    Convey("新功能命令测试", t, func() {
        cmd := NewNewFeatureCommand()
        So(cmd, ShouldNotBeNil)
        So(cmd.Use, ShouldEqual, "new-feature")

        // 测试标志
        So(cmd.Flags().Lookup("interactive"), ShouldNotBeNil)
        So(cmd.Flags().Lookup("verbose"), ShouldNotBeNil)
    })
}

// 3. 性能测试 - 对于 I/O 密集型功能
func BenchmarkNewFeaturePerformance(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ProcessInput("benchmark input")
    }
}
```

**📊 覆盖率要求**：
- 新功能代码覆盖率必须 ≥ 85%
- 核心业务逻辑覆盖率必须 ≥ 95%
- 错误处理路径覆盖率必须 ≥ 90%

```bash
# 检查覆盖率
go test -coverprofile=coverage.out ./component/new-feature/...
go tool cover -html=coverage.out -o coverage.html
# 覆盖率不达标的 PR 将被拒绝
```

#### 3. 测试要求
- **单元测试**：每个公共函数都需要测试
- **集成测试**：新命令需要端到端测试
- **性能测试**：涉及 I/O 操作的功能需要性能测试

#### 4. 文档要求
- **代码注释**：公共 API 必须有详细注释
- **README 更新**：新功能需要更新使用说明
- **变更日志**：记录重要变更和破坏性更改

### 🔍 开发最佳实践

#### 1. 强制开发流程

**📋 开发检查清单**：
```bash
# 🔴 必须：创建功能分支
git checkout -b feature/具体功能名称

# 🔴 必须：开发前先写测试（TDD）
touch tests/new_feature_test.go
# 先写失败的测试，再实现功能

# 🔴 必须：每次提交前运行完整检查
make fmt                    # 代码格式化
make lint                   # 代码检查
make test-basic            # 基础测试
make test-integration      # 集成测试

# 🔴 必须：检查测试覆盖率
go test -cover ./component/new-feature/...
# 覆盖率必须 ≥ 85%

# 🔴 必须：更新文档
# 1. 更新 component/new-feature/README.md
# 2. 更新主 README.md 的使用示例
# 3. 添加 docs/ 下的详细文档

# 🔴 必须：性能基准测试
go test -bench=. ./component/new-feature/...
```

#### 2. 性能和资源管理

**🚀 具体性能要求**：
```go
// ❌ 错误：每次调用都创建新的 HTTP 客户端
func CallAPI() error {
    client := &http.Client{Timeout: 30 * time.Second}
    // ...
}

// ✅ 正确：复用 HTTP 客户端
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}

func CallAPI() error {
    // 使用全局客户端
}

// ❌ 错误：没有超时控制的操作
func ProcessLargeFile(filename string) error {
    // 可能无限期阻塞
}

// ✅ 正确：所有 I/O 操作必须有超时
func ProcessLargeFile(filename string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    return processWithContext(ctx, filename)
}

// ❌ 错误：内存泄漏风险
func ProcessFiles(files []string) error {
    var results []Result
    for _, file := range files {
        result := processFile(file) // 可能积累大量内存
        results = append(results, result)
    }
    return nil
}

// ✅ 正确：流式处理大量数据
func ProcessFiles(files []string) error {
    for _, file := range files {
        if err := processFileStream(file); err != nil {
            return err
        }
        // 每个文件处理完立即释放内存
    }
    return nil
}
```

#### 3. 安全和输入验证

**🔒 强制安全检查**：
```go
// ✅ 所有用户输入必须验证
func ValidateInput(input string) error {
    // 1. 长度检查
    if len(input) == 0 {
        return errors.NewInputError("输入不能为空", nil)
    }
    if len(input) > 10000 {
        return errors.NewInputError("输入长度不能超过 10000 字符", nil)
    }

    // 2. 字符检查
    if !utf8.ValidString(input) {
        return errors.NewInputError("输入包含无效的 UTF-8 字符", nil)
    }

    // 3. 恶意内容检查
    dangerous := []string{"<script", "javascript:", "data:"}
    for _, pattern := range dangerous {
        if strings.Contains(strings.ToLower(input), pattern) {
            return errors.NewInputError("输入包含潜在危险内容", nil)
        }
    }

    return nil
}

// ✅ 文件路径必须验证
func ValidateFilePath(path string) error {
    // 1. 清理路径
    cleanPath := filepath.Clean(path)

    // 2. 检查路径遍历
    if strings.Contains(cleanPath, "..") {
        return errors.NewInputError("不允许路径遍历", nil)
    }

    // 3. 检查绝对路径
    if filepath.IsAbs(cleanPath) {
        return errors.NewInputError("不允许绝对路径", nil)
    }

    // 4. 检查文件扩展名
    allowedExts := []string{".md", ".txt", ".json"}
    ext := filepath.Ext(cleanPath)
    if !contains(allowedExts, ext) {
        return errors.NewInputError("不支持的文件类型", nil)
    }

    return nil
}

// ✅ API 密钥必须安全处理
func LogAPICall(endpoint string, apiKey string) {
    // ❌ 错误：记录完整 API 密钥
    // log.Info().Str("api_key", apiKey).Msg("API 调用")

    // ✅ 正确：只记录前几位用于调试
    maskedKey := apiKey[:min(len(apiKey), 8)] + "***"
    log.Info().
        Str("endpoint", endpoint).
        Str("api_key_prefix", maskedKey).
        Msg("API 调用")
}
```

#### 4. 错误处理和用户体验

**🎯 用户友好的错误处理**：
```go
// ✅ 提供具体的错误信息和解决建议
func TranslateFile(filePath string) error {
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return errors.NewFileError("文件不存在", err).
            WithSuggestion("请检查文件路径是否正确").
            WithSuggestion("确保文件存在且有读取权限").
            WithCode("FILE_001")
    }

    if !strings.HasSuffix(filePath, ".md") {
        return errors.NewInputError("不支持的文件格式", nil).
            WithSuggestion("当前只支持 .md 文件").
            WithSuggestion("请使用 Markdown 格式的文件").
            WithCode("FORMAT_001")
    }

    // API 调用失败时提供重试建议
    if err := callTranslationAPI(); err != nil {
        if isNetworkError(err) {
            return errors.NewNetworkError("网络连接失败", err).
                WithSuggestion("请检查网络连接").
                WithSuggestion("如果使用代理，请确保代理配置正确").
                WithSuggestion("稍后重试或联系管理员").
                WithCode("NETWORK_001")
        }
    }

    return nil
}
```

### 📊 监控和维护

#### 1. 性能监控
- 监控 API 调用延迟和成功率
- 跟踪内存使用和 CPU 占用
- 记录错误频率和类型

#### 2. 用户反馈
- 收集用户使用数据和反馈
- 分析常见错误和使用模式
- 持续优化用户体验

#### 3. 定期维护
- 每月更新依赖和安全补丁
- 季度性能评估和优化
- 年度架构回顾和重构

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

### 贡献指南
1. Fork 项目并创建功能分支
2. 遵循上述开发指南和代码规范
3. 确保所有测试通过
4. 提交 Pull Request 并详细描述变更

## 许可证

本项目采用 MIT 许可证。详情请参阅 LICENSE 文件。
