# MicroDev - 微开发工具集

MicroDev 是一个基于 Go 语言开发的微开发工具集，提供提示词优化和中英互译功能，帮助开发者提高工作效率。

## 功能特性

### 🚀 提示词优化 (Prompt)
- 智能优化用户输入的提示词
- 支持文件输入和直接文本输入
- 基于大语言模型生成更好的提示词

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

编译成功后，`micro` 命令将被安装到 `$GOBIN` 目录中。请确保 `$GOBIN` 在您的 `PATH` 环境变量中。

## 使用方法

### 基本语法
```bash
micro <子命令> [参数] [选项]
```

### 提示词优化 (prompt/p)

#### 基本用法
```bash
# 使用完整命令名
micro prompt "优化这个提示词"

# 使用简写
micro p "优化这个提示词"

# 从文件读取
micro p /path/to/prompt.txt
```

#### 示例
```bash
# 优化简单文本
micro p "帮我写一个Python函数"

# 从文件优化
micro p ./prompts/my_prompt.txt
```

### 中英互译 (translate/t)

#### 基本用法
```bash
# 使用完整命令名
micro translate "Hello World"

# 使用简写
micro t "Hello World"
```

#### 翻译选项
- `-f, --file`: 明确指定为文件翻译模式
- `-d, --dir`: 明确指定为目录翻译模式
- `-t, --type`: 指定文件类型过滤（默认为 md）

#### 示例

##### 直接文本翻译
```bash
# 中译英
micro t "你好世界"

# 英译中
micro t "Hello World"
```

##### 单文件翻译
```bash
# 自动检测文件
micro t ./document.md

# 明确指定文件模式
micro t -f ./document.md
```

##### 目录批量翻译
```bash
# 翻译目录下所有 Markdown 文件
micro t -d ./docs

# 翻译目录下所有 txt 文件
micro t -d ./docs -t txt

# 翻译目录下所有 Go 文件
micro t -d ./src -t go

# 使用并发翻译（5个并发）
micro t -d ./docs -c 5

# 强制重新翻译已存在的文件
micro t -d ./docs --force
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

```bash
# 必需：API 密钥（二选一）
export DASHSCOPE_API_KEY=your_dashscope_api_key
# 或
export ALI_BAILIAN_API_KEY=your_bailian_api_key

# 可选：并发控制
export MICRO_CONCURRENCY=5  # 默认为3，最大为10
```

### 获取 API 密钥

1. **DashScope API**: 访问 [阿里云 DashScope 控制台](https://dashscope.console.aliyun.com/) 获取 API 密钥
2. **阿里百炼 API**: 访问 [阿里百炼控制台](https://bailian.console.aliyun.com/) 获取 API 密钥

### 默认配置

- **模型**: qwen-plus-latest
- **最大令牌数**: 2048
- **温度**: 0.7（提示词优化），0.3（翻译任务）
- **流式输出**: 启用

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

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

本项目采用 MIT 许可证。详情请参阅 LICENSE 文件。
