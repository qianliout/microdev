# 详细日志输出功能演示

根据您的反馈，我们已经为 prompt 命令的会话功能增加了详细且可读性高的日志输出。现在系统会清楚地告诉用户记忆了什么内容、上下文是什么，以及发送给LLM的具体内容。

## 日志功能特性

### 🎯 核心日志类型

1. **📝 用户消息记录**：记录用户输入的详细信息
2. **🤖 助手回复记录**：记录AI助手的回复内容
3. **🧠 上下文构建**：显示如何构建优化上下文
4. **🗜️ 记忆压缩**：展示记忆压缩的详细过程
5. **📤 LLM交互**：记录发送给LLM的完整内容
6. **⏱️ 性能监控**：显示响应时间和处理统计

### 📊 日志信息内容

#### 用户消息记录
```
📝 记录用户消息 
  session_id=session_1758783922 
  role=user 
  content="请帮我优化这个提示词" 
  content_length=24 
  total_messages=1
```

#### 助手回复记录
```
🤖 记录助手回复 
  session_id=session_1758783922 
  role=assistant 
  content="已提供优化建议" 
  content_length=18 
  total_messages=2
```

#### 上下文构建过程
```
🧠 构建优化上下文 
  session_id=session_1758783922 
  context_length=119 
  has_summary=false 
  recent_messages=4 
  context_preview="**最近对话：**\n\n用户: 第一个问题..."
```

#### 记忆压缩详情
```
🗜️ 触发记忆压缩 
  session_id=session_1758783922 
  current_messages=9 
  max_messages=8

📦 准备压缩消息 
  messages_to_compress=5 
  messages_to_keep=4

✅ 记忆压缩完成 
  compressed_messages=5 
  remaining_messages=4 
  new_summary_length=55 
  summary_preview="讨论主题：用户消息 3\n对话轮次：5条消息"
```

#### LLM交互记录
```
📤 发送给LLM的完整内容 
  system_prompt="你是一个专业的提示词优化专家..." 
  user_message="请优化以下提示词：..."

🤖 调用LLM生成响应

⏱️ LLM响应生成完成 
  duration=2.5s 
  choices_count=1

📥 收到LLM响应内容 
  choice_index=0 
  content_length=1024 
  content_preview="**优化后的提示词：**\n\n请为我设计..."
```

## 实际使用演示

### 场景1：首次对话
```bash
$ ./micro p -i
```

**日志输出：**
```
2025-09-25 15:05:22 INF 开始新的对话会话 session_id=session_1758783922
> 请帮我优化这个提示词：写一个函数

2025-09-25 15:05:22 INF 🚀 开始会话模式优化 
  original_prompt="请帮我优化这个提示词：写一个函数" 
  prompt_length=24

2025-09-25 15:05:22 INF 📝 记录用户消息 
  session_id=session_1758783922 
  role=user 
  content="请帮我优化这个提示词：写一个函数" 
  content_length=24 
  total_messages=1

2025-09-25 15:05:22 INF 🆕 首次对话，无历史上下文

2025-09-25 15:05:22 INF 📋 构建系统提示词 
  system_prompt_length=1024

2025-09-25 15:05:22 INF 📝 构建用户消息 
  user_message_length=156 
  includes_context=false 
  includes_session_context=false 
  user_message_preview="请优化以下提示词，使其更加清晰、具体和有效：\n\n原始提示词：\n请帮我优化这个提示词：写一个函数\n\n请按照以下格式输出：..."

2025-09-25 15:05:22 INF ⚙️ 配置LLM参数 
  temperature=0.7 
  max_tokens=2048 
  stream_output=true

2025-09-25 15:05:22 INF 📤 发送给LLM的完整内容 
  system_prompt="你是一个专业的提示词优化专家，擅长将用户的原始提示词优化成更清晰、更有效的版本..." 
  user_message="请优化以下提示词，使其更加清晰、具体和有效：\n\n原始提示词：\n请帮我优化这个提示词：写一个函数..."

2025-09-25 15:05:22 INF 🌊 使用流式输出模式
```

### 场景2：继续对话（有上下文）
```
> 需要支持排序功能

2025-09-25 15:05:25 INF 🚀 开始会话模式优化 
  original_prompt="需要支持排序功能" 
  prompt_length=21

2025-09-25 15:05:25 INF 📝 记录用户消息 
  session_id=session_1758783922 
  role=user 
  content="需要支持排序功能" 
  content_length=21 
  total_messages=3

2025-09-25 15:05:25 INF 💬 包含最近对话历史 
  session_id=session_1758783922 
  recent_messages_count=2 
  total_messages=3

2025-09-25 15:05:25 INF 🧠 构建优化上下文 
  session_id=session_1758783922 
  context_length=245 
  has_summary=false 
  recent_messages=2 
  context_preview="**最近对话：**\n\n用户: 请帮我优化这个提示词：写一个函数\n\n助手: **优化后的提示词：**\n\n请为我设计并实现一个具体的函数..."

2025-09-25 15:05:25 INF 🧠 使用会话上下文 
  context_length=245 
  context_preview="**最近对话：**\n\n用户: 请帮我优化这个提示词：写一个函数\n\n助手: **优化后的提示词：**\n\n请为我设计并实现一个具体的函数..."
```

### 场景3：记忆压缩触发
```
> 继续添加更多功能...

2025-09-25 15:05:30 INF 📝 记录用户消息 
  session_id=session_1758783922 
  role=user 
  content="继续添加更多功能..." 
  content_length=27 
  total_messages=9

2025-09-25 15:05:30 INF 🗜️ 触发记忆压缩 
  session_id=session_1758783922 
  current_messages=9 
  max_messages=8

2025-09-25 15:05:30 INF 📦 准备压缩消息 
  session_id=session_1758783922 
  messages_to_compress=5 
  messages_to_keep=4

2025-09-25 15:05:30 INF ✅ 记忆压缩完成 
  session_id=session_1758783922 
  compressed_messages=5 
  remaining_messages=4 
  had_previous_summary=false 
  new_summary_length=89 
  summary_preview="讨论主题：请帮我优化这个提示词：写一个函数\n关键活动：提供了优化建议、说明了改进要点\n对话轮次：5条消息"
```

## 日志级别说明

### INFO 级别日志
- 📝 用户消息记录
- 🤖 助手回复记录  
- 🧠 上下文构建
- 🗜️ 记忆压缩
- 📤 LLM交互
- ⏱️ 性能统计

### DEBUG 级别日志
- 📋 系统提示词构建
- 🔍 空上下文检查
- 详细的内部状态

### ERROR 级别日志
- ❌ LLM调用失败
- ❌ 存储操作失败
- ❌ 文件写入失败

## 日志字段说明

### 会话相关
- `session_id`: 会话唯一标识符
- `total_messages`: 当前会话总消息数
- `recent_messages`: 最近消息数量

### 内容相关
- `content`: 消息内容
- `content_length`: 内容长度
- `content_preview`: 内容预览（前200字符）
- `role`: 消息角色（user/assistant）

### 上下文相关
- `context_length`: 上下文总长度
- `has_summary`: 是否包含会话摘要
- `includes_context`: 是否包含额外上下文
- `includes_session_context`: 是否包含会话上下文

### 压缩相关
- `current_messages`: 当前消息数
- `max_messages`: 最大消息数阈值
- `messages_to_compress`: 待压缩消息数
- `messages_to_keep`: 保留消息数
- `compressed_messages`: 已压缩消息数
- `remaining_messages`: 剩余消息数
- `new_summary_length`: 新摘要长度
- `summary_preview`: 摘要预览

### 性能相关
- `duration`: 操作耗时
- `choices_count`: LLM返回选择数量
- `temperature`: LLM温度参数
- `max_tokens`: 最大令牌数
- `stream_output`: 是否流式输出

## 使用建议

### 开发调试
```bash
# 查看详细日志
./micro p -i

# 只查看错误日志
./micro p -i 2>/dev/null

# 保存日志到文件
./micro p -i 2>&1 | tee session.log
```

### 生产环境
- 建议设置日志级别为 INFO
- 定期清理日志文件
- 监控错误日志

### 性能分析
- 关注 `duration` 字段了解响应时间
- 监控 `content_length` 了解内容大小
- 观察 `compressed_messages` 了解压缩频率

通过这些详细的日志输出，您可以清楚地了解：
1. 系统记住了什么内容（用户输入和助手回复）
2. 如何构建和使用上下文信息
3. 何时触发记忆压缩以及压缩了什么
4. 发送给LLM的完整内容
5. 系统的性能表现

这些信息对于调试、优化和理解系统行为非常有价值！
