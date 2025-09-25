# Python 重构 microdev 项目完整 Prompt

请基于以下详细规格，从零开始用 Python 重构 microdev 项目。这是一个 CLI 工具，包含提示词优化和翻译功能，具有交互式会话、智能记忆管理和模块化存储架构。

## 项目概述

### 核心功能
1. **Prompt 子命令** (`micro prompt` / `micro p`)：智能提示词优化工具
2. **Translate 子命令** (`micro translate` / `micro t`)：中英互译工具
3. **交互式会话模式**：支持对话历史和上下文记忆
4. **模块化存储架构**：支持内存、MySQL、Redis 存储后端

### 技术栈要求
- **语言**：Python 3.8+
- **CLI 框架**：Click 或 Typer
- **HTTP 客户端**：httpx 或 requests
- **配置管理**：pydantic + python-dotenv
- **日志系统**：loguru 或 structlog
- **数据库**：SQLAlchemy (MySQL)，redis-py (Redis)
- **测试框架**：pytest
- **代码质量**：black, isort, flake8, mypy

## 详细功能规格

### 1. Prompt 子命令功能

#### 基础优化功能
```python
# 使用示例
micro prompt "请帮我写一个函数"
micro p ./prompt.txt
echo "提示词" | micro p -
```

**核心特性：**
- 支持直接文本、文件输入、标准输入
- 基于大语言模型（阿里云 DashScope/百炼）优化提示词
- 流式输出显示优化结果
- 支持额外上下文信息

#### 交互式会话模式
```python
# 使用示例
micro p -i
micro p --interactive
```

**会话功能要求：**
1. **对话历史管理**
   - 自动记录用户输入和AI回复
   - 为每个会话生成唯一ID（格式：session_时间戳）
   - 支持会话统计（消息数、用户消息数、助手消息数）

2. **智能记忆压缩**
   - 默认保留最近8条消息（4轮对话）
   - 超过阈值时自动压缩早期对话为摘要
   - 摘要包含：讨论主题、关键活动、对话轮次
   - 压缩算法：保留最近消息，压缩较早消息

3. **上下文感知优化**
   - 构建优化上下文：会话摘要 + 最近对话历史
   - 发送给LLM的消息包含历史上下文
   - 基于对话历史提供更精准的优化建议

4. **会话控制**
   - 输入 'exit'、'quit'、'退出' 结束会话
   - 显示会话统计信息
   - 支持会话加载和恢复

#### 详细日志输出
**日志级别和内容：**
```python
# 用户消息记录
logger.info("📝 记录用户消息", 
    session_id=session_id,
    role="user", 
    content=content,
    content_length=len(content),
    total_messages=total_count
)

# 助手回复记录  
logger.info("🤖 记录助手回复",
    session_id=session_id,
    role="assistant",
    content=response,
    content_length=len(response),
    total_messages=total_count
)

# 上下文构建
logger.info("🧠 构建优化上下文",
    session_id=session_id,
    context_length=len(context),
    has_summary=bool(summary),
    recent_messages=recent_count,
    context_preview=context[:200]
)

# 记忆压缩
logger.info("🗜️ 触发记忆压缩",
    session_id=session_id,
    current_messages=current_count,
    max_messages=max_limit
)

logger.info("✅ 记忆压缩完成",
    compressed_messages=compressed_count,
    remaining_messages=remaining_count,
    new_summary_length=len(summary),
    summary_preview=summary[:100]
)

# LLM交互
logger.info("📤 发送给LLM的完整内容",
    system_prompt=system_prompt,
    user_message=user_message
)

logger.info("⏱️ LLM响应生成完成",
    duration=duration,
    response_length=len(response),
    response_preview=response[:200]
)
```

### 2. Translate 子命令功能

#### 基础翻译功能
```python
# 使用示例
micro translate "Hello, World!"
micro t "你好，世界！"
micro t ./document.md
micro t -d ./docs
```

**核心特性：**
- 自动语言检测（中文↔英文）
- 支持直接文本、文件、目录批量翻译
- 智能输出文件命名（document.md → document_en.md）

#### 高级功能
1. **翻译幂等性**
   - 检测目标文件是否已存在
   - 默认跳过已翻译文件
   - `--force` 参数强制重新翻译

2. **并发翻译**
   - 支持多文件并发处理
   - 默认并发数：3，可通过 `-c` 参数调整
   - 自动负载均衡，避免API限流

3. **多格式支持**
   - Markdown (.md)：保持格式
   - 文本文件 (.txt)：纯文本翻译
   - 代码文件 (.py, .js, .go)：翻译注释

#### 命令选项
```python
@click.command()
@click.argument('input_text', required=False)
@click.option('-d', '--directory', help='翻译目录')
@click.option('-t', '--type', help='文件类型过滤')
@click.option('-o', '--output', help='输出文件路径')
@click.option('-c', '--concurrency', default=3, help='并发数量')
@click.option('--force', is_flag=True, help='强制重新翻译')
def translate(input_text, directory, type, output, concurrency, force):
    pass
```

### 3. 模块化存储架构

#### 存储接口设计
```python
from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any
from datetime import datetime

class SessionStorage(ABC):
    @abstractmethod
    async def save_session(self, session: Session) -> None:
        pass
    
    @abstractmethod
    async def load_session(self, session_id: str) -> Optional[Session]:
        pass
    
    @abstractmethod
    async def save_message(self, session_id: str, message: Message) -> None:
        pass
    
    @abstractmethod
    async def load_messages(self, session_id: str, limit: int = 50) -> List[Message]:
        pass
    
    @abstractmethod
    async def update_session_summary(self, session_id: str, summary: str) -> None:
        pass
    
    @abstractmethod
    async def get_session_stats(self, session_id: str) -> SessionStats:
        pass
    
    @abstractmethod
    async def cleanup_expired_sessions(self, expired_before: datetime) -> int:
        pass
```

#### 存储实现

1. **内存存储** (默认)
```python
class MemoryStorage(SessionStorage):
    def __init__(self):
        self.sessions: Dict[str, Session] = {}
        self.messages: Dict[str, List[Message]] = {}
        self.lock = asyncio.Lock()
```

2. **MySQL 存储**
```python
class MySQLStorage(SessionStorage):
    def __init__(self, database_url: str):
        self.engine = create_async_engine(database_url)
        # 表结构：sessions, messages
```

3. **Redis 存储**
```python
class RedisStorage(SessionStorage):
    def __init__(self, redis_url: str):
        self.redis = redis.from_url(redis_url)
        # 键结构：session:{id}, messages:{id}
```

#### 存储工厂
```python
class StorageFactory:
    @staticmethod
    def create_storage(storage_type: str, **kwargs) -> SessionStorage:
        if storage_type == "memory":
            return MemoryStorage()
        elif storage_type == "mysql":
            return MySQLStorage(kwargs["database_url"])
        elif storage_type == "redis":
            return RedisStorage(kwargs["redis_url"])
        else:
            raise ValueError(f"Unsupported storage type: {storage_type}")
```

### 4. 数据模型

#### 核心数据结构
```python
from pydantic import BaseModel
from datetime import datetime
from typing import List, Optional

class Message(BaseModel):
    role: str  # "user" or "assistant"
    content: str
    timestamp: datetime

class Session(BaseModel):
    id: str
    messages: List[Message] = []
    summary: str = ""
    created_at: datetime
    updated_at: datetime

class SessionStats(BaseModel):
    session_active: bool
    session_id: Optional[str] = None
    total_messages: int
    user_messages: int
    assistant_messages: int
    has_summary: bool
    summary_length: int
```

### 5. 配置管理

#### 配置结构
```python
from pydantic import BaseSettings

class LLMConfig(BaseSettings):
    model: str = "qwen-plus-latest"
    max_tokens: int = 2048
    temperature: float = 0.7
    stream_output: bool = True

class StorageConfig(BaseSettings):
    type: str = "memory"  # memory, mysql, redis
    database_url: Optional[str] = None
    redis_url: Optional[str] = None
    ttl_hours: int = 24
    max_sessions: int = 100

class AppConfig(BaseSettings):
    # API 配置
    dashscope_api_key: Optional[str] = None
    bailian_api_key: Optional[str] = None
    
    # LLM 配置
    llm: LLMConfig = LLMConfig()
    
    # 存储配置
    storage: StorageConfig = StorageConfig()
    
    # 日志配置
    log_level: str = "INFO"
    
    class Config:
        env_file = ".env"
        env_nested_delimiter = "__"
```

### 6. 项目结构

```
microdev/
├── pyproject.toml              # 项目配置
├── README.md                   # 项目说明
├── .env.example               # 环境变量示例
├── microdev/                  # 主包
│   ├── __init__.py
│   ├── cli/                   # CLI 命令
│   │   ├── __init__.py
│   │   ├── main.py           # 主命令入口
│   │   ├── prompt.py         # prompt 子命令
│   │   └── translate.py      # translate 子命令
│   ├── core/                  # 核心业务逻辑
│   │   ├── __init__.py
│   │   ├── optimizer.py      # 提示词优化服务
│   │   ├── translator.py     # 翻译服务
│   │   └── llm_client.py     # LLM 客户端
│   ├── session/               # 会话管理
│   │   ├── __init__.py
│   │   ├── manager.py        # 会话管理器
│   │   ├── storage/          # 存储实现
│   │   │   ├── __init__.py
│   │   │   ├── base.py       # 存储接口
│   │   │   ├── memory.py     # 内存存储
│   │   │   ├── mysql.py      # MySQL 存储
│   │   │   └── redis.py      # Redis 存储
│   │   └── models.py         # 数据模型
│   ├── utils/                 # 工具函数
│   │   ├── __init__.py
│   │   ├── language.py       # 语言检测
│   │   ├── file_utils.py     # 文件处理
│   │   └── logging.py        # 日志配置
│   └── config.py             # 配置管理
├── tests/                     # 测试文件
│   ├── __init__.py
│   ├── test_prompt.py
│   ├── test_translate.py
│   ├── test_session.py
│   └── test_storage.py
└── docs/                      # 文档
    ├── README.md
    ├── prompt/
    ├── translate/
    └── general/
```

### 7. 关键实现要求

#### CLI 入口点
```python
# microdev/cli/main.py
import click
from microdev.cli.prompt import prompt_command
from microdev.cli.translate import translate_command

@click.group()
@click.version_option()
def cli():
    """microdev - AI-powered prompt optimization and translation tool"""
    pass

cli.add_command(prompt_command, name="prompt")
cli.add_command(translate_command, name="translate")
cli.add_command(prompt_command, name="p")  # 别名
cli.add_command(translate_command, name="t")  # 别名
```

#### 会话管理器
```python
# microdev/session/manager.py
class SessionManager:
    def __init__(self, storage: SessionStorage, max_messages: int = 8):
        self.storage = storage
        self.current_session: Optional[Session] = None
        self.max_messages = max_messages
    
    async def start_session(self) -> Session:
        session_id = f"session_{int(time.time())}"
        session = Session(
            id=session_id,
            created_at=datetime.now(),
            updated_at=datetime.now()
        )
        await self.storage.save_session(session)
        self.current_session = session
        return session
    
    async def add_user_message(self, content: str):
        # 实现用户消息添加逻辑
        pass
    
    async def add_assistant_message(self, content: str):
        # 实现助手消息添加逻辑
        pass
    
    async def get_context_for_optimization(self) -> str:
        # 实现上下文构建逻辑
        pass
    
    async def check_and_compress_memory(self):
        # 实现记忆压缩逻辑
        pass
```

### 8. 测试要求

#### 测试覆盖
- 单元测试：所有核心功能模块
- 集成测试：CLI 命令端到端测试
- 存储测试：各种存储后端测试
- 会话测试：会话管理和记忆压缩测试

#### 测试示例
```python
# tests/test_session.py
import pytest
from microdev.session.manager import SessionManager
from microdev.session.storage.memory import MemoryStorage

@pytest.mark.asyncio
async def test_session_creation():
    storage = MemoryStorage()
    manager = SessionManager(storage)
    
    session = await manager.start_session()
    assert session.id.startswith("session_")
    assert len(session.messages) == 0

@pytest.mark.asyncio
async def test_memory_compression():
    storage = MemoryStorage()
    manager = SessionManager(storage, max_messages=4)
    
    await manager.start_session()
    
    # 添加超过阈值的消息
    for i in range(6):
        await manager.add_user_message(f"用户消息 {i}")
        await manager.add_assistant_message(f"助手回复 {i}")
    
    # 验证记忆压缩
    assert len(manager.current_session.messages) <= 4
    assert manager.current_session.summary != ""
```

### 9. 部署和打包

#### pyproject.toml
```toml
[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"

[tool.poetry]
name = "microdev"
version = "0.1.0"
description = "AI-powered prompt optimization and translation tool"
authors = ["Your Name <your.email@example.com>"]

[tool.poetry.dependencies]
python = "^3.8"
click = "^8.0"
httpx = "^0.24"
pydantic = "^2.0"
python-dotenv = "^1.0"
loguru = "^0.7"
sqlalchemy = {extras = ["asyncio"], version = "^2.0", optional = true}
asyncpg = {version = "^0.28", optional = true}
redis = {version = "^4.5", optional = true}

[tool.poetry.extras]
mysql = ["sqlalchemy", "asyncpg"]
redis = ["redis"]
all = ["sqlalchemy", "asyncpg", "redis"]

[tool.poetry.scripts]
micro = "microdev.cli.main:cli"

[tool.poetry.group.dev.dependencies]
pytest = "^7.0"
pytest-asyncio = "^0.21"
black = "^23.0"
isort = "^5.12"
flake8 = "^6.0"
mypy = "^1.0"
```

### 10. 特殊要求

#### 错误处理
- 网络错误自动重试（最多3次）
- API 限流处理和退避策略
- 详细的错误日志记录
- 用户友好的错误消息

#### 性能优化
- 异步 I/O 处理
- 连接池管理
- 内存使用优化
- 缓存机制

#### 安全考虑
- API 密钥安全存储
- 输入验证和清理
- SQL 注入防护
- 敏感信息脱敏

### 11. LLM 集成实现

#### API 客户端
```python
# microdev/core/llm_client.py
import httpx
from typing import AsyncGenerator, Dict, Any
from microdev.config import AppConfig

class LLMClient:
    def __init__(self, config: AppConfig):
        self.config = config
        self.client = httpx.AsyncClient(timeout=30.0)

    async def generate_content(
        self,
        messages: List[Dict[str, str]],
        stream: bool = True
    ) -> AsyncGenerator[str, None]:
        """生成内容，支持流式输出"""

        # DashScope API 调用
        if self.config.dashscope_api_key:
            url = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
            headers = {
                "Authorization": f"Bearer {self.config.dashscope_api_key}",
                "Content-Type": "application/json"
            }
            payload = {
                "model": self.config.llm.model,
                "input": {"messages": messages},
                "parameters": {
                    "max_tokens": self.config.llm.max_tokens,
                    "temperature": self.config.llm.temperature,
                    "stream": stream
                }
            }

        # 百炼 API 调用
        elif self.config.bailian_api_key:
            # 实现百炼 API 调用逻辑
            pass

        async with self.client.stream("POST", url, headers=headers, json=payload) as response:
            if stream:
                async for line in response.aiter_lines():
                    if line.startswith("data: "):
                        data = json.loads(line[6:])
                        if "output" in data and "text" in data["output"]:
                            yield data["output"]["text"]
            else:
                result = await response.json()
                yield result["output"]["text"]
```

#### 系统提示词模板
```python
# microdev/core/prompts.py
PROMPT_OPTIMIZATION_SYSTEM = """你是一个专业的提示词优化专家，擅长将用户的原始提示词优化成更清晰、更有效的版本。

## 优化原则：
1. **清晰性**：使提示词更加明确和具体
2. **结构化**：合理组织信息层次
3. **完整性**：补充必要的上下文信息
4. **可操作性**：确保AI能够准确理解和执行

## 优化策略：
- 明确任务目标和期望输出
- 提供必要的背景信息和约束条件
- 使用结构化的格式（如分点、分段）
- 添加示例或模板（如果有助于理解）
- 优化语言表达，使其更加专业和准确

## 输出格式：
请按以下格式输出：

**优化后的提示词：**
[在这里提供优化后的完整提示词]

**主要改进点：**
1. [改进点1]
2. [改进点2]
3. [改进点3]

**使用建议：**
[提供使用这个优化后提示词的具体建议]"""

TRANSLATION_SYSTEM = """你是一个专业的中英互译专家，能够准确理解语言的细微差别并提供高质量的翻译。

## 翻译原则：
1. **准确性**：忠实原文意思，不添加或删减内容
2. **流畅性**：译文自然流畅，符合目标语言习惯
3. **一致性**：术语翻译保持一致
4. **格式保持**：保持原文的格式和结构

## 翻译要求：
- 中文内容翻译为英文
- 英文内容翻译为中文
- 保持Markdown格式不变
- 保持代码块不变
- 保持链接和图片不变

请直接输出翻译结果，不要添加任何解释或说明。"""
```

### 12. 具体实现示例

#### 交互式会话实现
```python
# microdev/cli/prompt.py
import asyncio
import click
from microdev.core.optimizer import PromptOptimizer
from microdev.session.manager import SessionManager
from microdev.utils.logging import get_logger

@click.command()
@click.argument('input_text', required=False)
@click.option('-i', '--interactive', is_flag=True, help='启用交互式会话模式')
@click.option('-c', '--context', help='额外上下文信息')
def prompt_command(input_text, interactive, context):
    """优化提示词工具"""
    asyncio.run(_prompt_main(input_text, interactive, context))

async def _prompt_main(input_text, interactive, context):
    logger = get_logger(__name__)
    optimizer = PromptOptimizer()

    if interactive:
        await _interactive_mode(optimizer, logger)
    else:
        await _single_optimization(optimizer, input_text, context, logger)

async def _interactive_mode(optimizer: PromptOptimizer, logger):
    """交互式模式实现"""
    session_manager = optimizer.session_manager
    session = await session_manager.start_session()

    logger.info(f"🚀 开始交互式会话", session_id=session.id)

    click.echo("🤖 欢迎使用提示词优化工具！")
    click.echo("💡 输入您的提示词，我将帮您优化")
    click.echo("🚪 输入 'exit'、'quit' 或 '退出' 结束会话")
    click.echo("-" * 50)

    while True:
        try:
            user_input = click.prompt("📝 请输入提示词", type=str)

            # 检查退出命令
            if user_input.lower() in ['exit', 'quit', '退出']:
                stats = await session_manager.get_session_stats()
                click.echo(f"\n📊 会话统计：")
                click.echo(f"   总消息数: {stats.total_messages}")
                click.echo(f"   用户消息: {stats.user_messages}")
                click.echo(f"   助手回复: {stats.assistant_messages}")
                click.echo("👋 感谢使用，再见！")
                break

            # 处理用户输入
            await session_manager.add_user_message(user_input)

            # 获取优化结果
            click.echo("\n🤖 正在优化中...")
            async for chunk in optimizer.optimize_with_session(user_input):
                click.echo(chunk, nl=False)

            click.echo("\n" + "-" * 50)

        except KeyboardInterrupt:
            click.echo("\n\n👋 会话已中断，再见！")
            break
        except Exception as e:
            logger.error(f"❌ 处理错误: {e}")
            click.echo(f"❌ 发生错误: {e}")
```

#### 记忆压缩算法
```python
# microdev/session/manager.py
async def _compress_memory(self):
    """智能记忆压缩算法"""
    if not self.current_session or len(self.current_session.messages) <= self.max_messages:
        return

    logger.info("🗜️ 触发记忆压缩",
        session_id=self.current_session.id,
        current_messages=len(self.current_session.messages),
        max_messages=self.max_messages
    )

    # 计算压缩点
    keep_count = self.max_messages // 2
    compress_count = len(self.current_session.messages) - keep_count

    # 分离消息
    messages_to_compress = self.current_session.messages[:compress_count]
    messages_to_keep = self.current_session.messages[compress_count:]

    # 生成摘要
    summary = await self._generate_summary(messages_to_compress)

    # 更新会话
    old_summary = self.current_session.summary
    if old_summary:
        # 合并旧摘要和新摘要
        combined_summary = f"{old_summary}\n\n{summary}"
        self.current_session.summary = combined_summary
    else:
        self.current_session.summary = summary

    self.current_session.messages = messages_to_keep

    # 保存到存储
    await self.storage.update_session_summary(
        self.current_session.id,
        self.current_session.summary
    )

    logger.info("✅ 记忆压缩完成",
        compressed_messages=compress_count,
        remaining_messages=len(messages_to_keep),
        new_summary_length=len(self.current_session.summary)
    )

async def _generate_summary(self, messages: List[Message]) -> str:
    """生成对话摘要"""
    if not messages:
        return ""

    # 提取关键信息
    topics = []
    key_points = []

    for msg in messages:
        if msg.role == "user":
            # 提取用户的主要需求
            topic = msg.content[:50] + "..." if len(msg.content) > 50 else msg.content
            topics.append(topic)
        elif msg.role == "assistant":
            # 提取助手回复的关键点
            if "优化后的提示词" in msg.content:
                key_points.append("提供了优化建议")
            if "主要改进点" in msg.content:
                key_points.append("说明了改进要点")

    # 构建摘要
    summary_parts = []

    if topics:
        main_topic = topics[-1]  # 使用最后一个主题作为主要讨论内容
        summary_parts.append(f"讨论主题：{main_topic}")

    if key_points:
        summary_parts.append(f"关键活动：{', '.join(set(key_points))}")

    summary_parts.append(f"对话轮次：{len(messages)}条消息")

    return "\n".join(summary_parts)
```

请基于以上完整详细的规格说明，从零开始创建一个功能完整、代码质量高、测试覆盖完整的 Python 版本 microdev 项目。确保所有功能都能正常工作，包括交互式会话、记忆管理、存储后端、详细日志等核心特性。
