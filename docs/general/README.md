# 通用配置文档

本文档介绍 microdev 工具的通用配置、环境设置和最佳实践。

## 🔧 环境配置

### API 密钥配置

microdev 支持两种 API 服务：

#### 1. 阿里云 DashScope API
```bash
export DASHSCOPE_API_KEY="your-dashscope-api-key"
```

获取方式：
1. 访问 [阿里云 DashScope 控制台](https://dashscope.console.aliyun.com/)
2. 创建应用并获取 API 密钥
3. 设置环境变量

#### 2. 阿里百炼 API
```bash
export ALI_BAILIAN_API_KEY="your-bailian-api-key"
```

获取方式：
1. 访问 [阿里百炼控制台](https://bailian.console.aliyun.com/)
2. 创建应用并获取 API 密钥
3. 设置环境变量

### 环境变量持久化

#### Linux/macOS
```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
echo 'export DASHSCOPE_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc
```

#### Windows
```cmd
# 设置系统环境变量
setx DASHSCOPE_API_KEY "your-api-key"
```

## ⚙️ 默认配置

### LLM 配置
- **模型**: qwen-plus-latest
- **最大令牌数**: 2048
- **温度**: 
  - 提示词优化：0.7
  - 翻译任务：0.3
- **流式输出**: 启用

### 性能配置
- **并发数**: 3（翻译任务）
- **超时时间**: 30秒
- **重试次数**: 3次
- **重试间隔**: 1秒

### 日志配置
- **日志级别**: INFO
- **日志格式**: JSON
- **时间格式**: RFC3339

## 🏗️ 构建配置

### 基础构建
```bash
# 标准构建（仅内存存储）
go build -o micro ./cmd/micro
```

### 扩展构建
```bash
# 包含 MySQL 支持
go build -tags mysql -o micro ./cmd/micro

# 包含 Redis 支持
go build -tags redis -o micro ./cmd/micro

# 包含所有存储支持
go build -tags "mysql redis" -o micro ./cmd/micro
```

### 构建标签说明
- `mysql`: 启用 MySQL 存储支持
- `redis`: 启用 Redis 存储支持
- 无标签: 仅支持内存存储

## 📊 配置验证

### 检查 API 连接
```bash
# 测试 API 连接
micro p "test" --dry-run

# 查看配置信息
micro --version
```

### 验证构建标签
```bash
# 检查支持的存储类型
micro p --help | grep -i storage
```

## 🔍 调试配置

### 日志级别设置
```bash
# 设置调试级别
export LOG_LEVEL=DEBUG

# 设置错误级别
export LOG_LEVEL=ERROR
```

### 详细输出
```bash
# 启用详细输出
micro p "test" --verbose

# 查看调试信息
micro p "test" --debug
```

## 🚀 性能优化

### API 调用优化
- 合理设置并发数，避免API限流
- 使用流式输出减少等待时间
- 配置合适的超时时间

### 内存优化
- 定期清理会话数据
- 合理设置最大消息数
- 使用持久化存储减少内存占用

### 网络优化
- 配置代理服务器（如需要）
- 设置合适的重试策略
- 监控网络延迟

## 🔒 安全配置

### API 密钥安全
- 不要在代码中硬编码API密钥
- 使用环境变量存储敏感信息
- 定期轮换API密钥

### 文件权限
```bash
# 设置配置文件权限
chmod 600 ~/.microdev/config

# 设置日志文件权限
chmod 644 /var/log/microdev.log
```

### 网络安全
- 使用HTTPS连接
- 验证SSL证书
- 配置防火墙规则

## 📁 目录结构

### 配置目录
```
~/.microdev/
├── config.yaml          # 主配置文件
├── sessions/            # 会话数据目录
├── logs/               # 日志文件目录
└── cache/              # 缓存文件目录
```

### 项目目录
```
microdev/
├── cmd/                # 命令行入口
├── component/          # 核心组件
├── pkg/               # 公共包
├── tests/             # 测试文件
├── docs/              # 文档目录
└── build/             # 构建脚本
```

## 🧪 测试配置

### 运行测试
```bash
# 运行所有测试
go test ./tests/ -v

# 运行特定测试
go test ./tests/basic_test.go -v

# 运行测试并查看覆盖率
go test ./tests/ -cover
```

### 测试环境变量
```bash
# 设置测试API密钥
export TEST_API_KEY="test-key"

# 启用测试模式
export TEST_MODE=true
```

## 🔧 故障排除

### 常见问题

#### API 密钥错误
```
错误: API key not configured
解决: 设置 DASHSCOPE_API_KEY 或 ALI_BAILIAN_API_KEY 环境变量
```

#### 网络连接问题
```
错误: connection timeout
解决: 检查网络连接，配置代理（如需要）
```

#### 权限问题
```
错误: permission denied
解决: 检查文件权限，确保有读写权限
```

### 调试步骤
1. 检查环境变量配置
2. 验证API密钥有效性
3. 测试网络连接
4. 查看详细日志
5. 检查文件权限

## 📈 监控和维护

### 日志监控
```bash
# 查看实时日志
tail -f ~/.microdev/logs/app.log

# 搜索错误日志
grep ERROR ~/.microdev/logs/app.log
```

### 性能监控
- 监控API调用频率
- 跟踪响应时间
- 观察内存使用情况

### 定期维护
- 清理过期日志文件
- 更新API密钥
- 检查依赖更新

## 🔗 相关链接

- [返回主文档](../../README.md)
- [Prompt 子命令文档](../prompt/README.md)
- [Translate 子命令文档](../translate/README.md)

## 📞 支持和反馈

如果您遇到问题或有改进建议，请：
1. 查看相关文档
2. 检查常见问题解答
3. 提交 Issue 或 Pull Request
4. 联系维护团队
