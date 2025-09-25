# microdev 文档中心

欢迎来到 microdev 工具的文档中心！这里包含了所有子命令的详细使用指南和配置说明。

## 📚 文档结构

### 子命令文档
每个子命令都有独立的文档目录，包含完整的使用指南：

#### 🚀 [Prompt 子命令](prompt/README.md)
提示词优化工具，支持交互式会话和智能记忆管理。

**核心功能：**
- 智能提示词优化
- 交互式会话模式
- 对话历史记忆
- 上下文感知优化

**相关文档：**
- [交互式功能演示](prompt/interactive-demo.md)
- [详细日志输出演示](prompt/detailed-logging-demo.md)
- [会话存储架构设计](prompt/session-storage-architecture.md)
- [存储配置示例](prompt/storage-config-examples.md)

#### 🌐 [Translate 子命令](translate/README.md)
智能中英互译工具，支持批量翻译和多种文件格式。

**核心功能：**
- 自动语言检测
- 中英互译
- 批量文件翻译
- 并发处理

**特色功能：**
- 翻译幂等性
- 多格式支持
- 智能文件管理

#### ⚙️ [通用配置](general/README.md)
环境配置、构建选项和最佳实践指南。

**包含内容：**
- API 密钥配置
- 环境变量设置
- 构建配置选项
- 性能优化建议
- 故障排除指南

## 🎯 快速导航

### 新用户指南
1. **环境配置** → [通用配置文档](general/README.md#环境配置)
2. **基础使用** → [Prompt 使用指南](prompt/README.md#使用指南) | [Translate 使用指南](translate/README.md#使用指南)
3. **高级功能** → [交互式会话](prompt/interactive-demo.md) | [批量翻译](translate/README.md#目录批量翻译)

### 开发者指南
1. **架构设计** → [会话存储架构](prompt/session-storage-architecture.md)
2. **调试工具** → [详细日志输出](prompt/detailed-logging-demo.md)
3. **扩展配置** → [存储配置示例](prompt/storage-config-examples.md)

### 运维指南
1. **部署配置** → [构建配置](general/README.md#构建配置)
2. **性能优化** → [性能优化](general/README.md#性能优化)
3. **故障排除** → [故障排除](general/README.md#故障排除)

## 🔍 按功能查找

### 基础功能
- **提示词优化** → [Prompt 基本用法](prompt/README.md#基本用法)
- **文本翻译** → [Translate 基本用法](translate/README.md#基本用法)
- **环境配置** → [API 密钥配置](general/README.md#api-密钥配置)

### 高级功能
- **交互式会话** → [交互式功能演示](prompt/interactive-demo.md)
- **批量处理** → [目录批量翻译](translate/README.md#目录批量翻译)
- **存储后端** → [存储架构设计](prompt/session-storage-architecture.md)

### 调试和监控
- **日志系统** → [详细日志输出](prompt/detailed-logging-demo.md)
- **性能监控** → [性能优化](general/README.md#性能优化)
- **错误处理** → [故障排除](general/README.md#故障排除)

## 📖 使用场景

### 开发场景
- **AI 应用开发** → 使用 Prompt 优化提示词
- **国际化项目** → 使用 Translate 进行文档翻译
- **调试优化** → 使用日志系统分析问题

### 生产场景
- **内容创作** → 批量优化和翻译内容
- **文档管理** → 自动化文档国际化
- **系统集成** → 集成到 CI/CD 流程

### 学习场景
- **提示词工程** → 学习如何优化 AI 提示词
- **多语言处理** → 了解自动翻译技术
- **工具开发** → 学习 CLI 工具设计

## 🛠️ 工具和资源

### 配置工具
- [环境变量配置](general/README.md#环境变量持久化)
- [构建脚本配置](general/README.md#构建配置)
- [存储后端配置](prompt/storage-config-examples.md)

### 调试工具
- [日志分析工具](prompt/detailed-logging-demo.md#日志级别说明)
- [性能监控工具](general/README.md#性能监控)
- [测试工具](general/README.md#测试配置)

### 开发工具
- [API 测试工具](general/README.md#检查-api-连接)
- [构建验证工具](general/README.md#验证构建标签)
- [代码质量工具](../../tests/README.md)

## 🔄 更新日志

### 最新功能
- ✅ 交互式会话模式
- ✅ 智能记忆管理
- ✅ 详细日志输出
- ✅ 模块化存储架构
- ✅ 批量翻译功能

### 即将推出
- 🔄 更多存储后端支持
- 🔄 Web 界面
- 🔄 插件系统
- 🔄 更多语言支持

## 📞 获取帮助

### 文档问题
如果您在文档中发现问题或有改进建议：
1. 查看相关的子命令文档
2. 检查通用配置文档
3. 提交 Issue 描述问题
4. 提供改进建议

### 功能问题
如果您在使用过程中遇到问题：
1. 查看对应的故障排除指南
2. 检查日志输出信息
3. 验证环境配置
4. 联系技术支持

### 贡献文档
欢迎贡献文档改进：
1. Fork 项目仓库
2. 修改或添加文档
3. 提交 Pull Request
4. 参与文档审查

## 🔗 相关链接

- [项目主页](../../README.md)
- [源代码仓库](https://github.com/your-org/microdev)
- [问题反馈](https://github.com/your-org/microdev/issues)
- [贡献指南](../../CONTRIBUTING.md)

---

**提示：** 建议按照子命令分类阅读相关文档，每个子命令的文档都包含完整的使用指南和示例。
