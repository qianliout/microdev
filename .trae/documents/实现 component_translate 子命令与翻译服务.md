## 目标
- 在 `component/translate` 下完成中英互译子命令与服务，实现一次只翻译一个文本，自动方向识别（英→中或中→英），并按照“原文/译文/音标/例句”结构化输出。
- 满足输入/输出均为 `io.ReadCloser` 的要求，考虑大文本分块与命令行流式输出。
- 代码分层：`cmd / model / service`，主要逻辑在 `service`；日志统一英文、注释中文；尽量复用 `pkg` 下现有能力。

## 现状与差距
- 已有文件：
  - `component/translate/cmd/command.go`：子命令骨架，`RunE` 未实现完整执行流程。
  - `component/translate/service/translator.go`：已有分块与流式写入逻辑，但输入处理与输出管道未完全打通（`Input` 忽略参数、`Output` 需配合管道）。
  - `component/translate/model/types.go`：`Translate{ Input io.ReadCloser, Output io.ReadWriteCloser, Direct string }` 已定义方向常量。
  - `pkg/llm/client.go`：封装 DashScope 兼容的 OpenAI 客户端，已支持流式/非流式，翻译 Prompt 已包含“原文/译文/音标/例句”。
  - `pkg/config/config.go`、`pkg/logger/logger.go`：环境与日志已就绪。
- 缺失：
  - 子命令 `RunE` 的完整参数处理（文本、文件、stdin）与输出管道搭建。
  - 一个 `ReadWriteCloser` 管道包装以满足“边写边读”的流式要求。
  - `service.Input` 的输入模式（文本/文件/stdin）与返回 `io.ReadCloser` 的实现。

## 设计方案
- 输入处理（单次翻译）
  - 支持三种输入：
    - 直接文本：`micro t "hello"`
    - 文件路径：`micro t ./a.txt`
    - 标准输入：`echo hello | micro t -`
  - 统一封装为 `io.ReadCloser`：
    - 文本：`io.NopCloser(strings.NewReader(text))`
    - 文件：`os.Open(path)`（`defer` 关闭）
    - stdin：`os.Stdin`（以只读形式使用）
- 输出流设计
  - 使用 `io.Pipe()` 构建读写两端，实现 `ReadWriteCloser` 包装（`pipeRW`），写端供 LLM 流式写入，读端由控制台打印进程持续 `io.Copy` 到 `stdout`。
  - 在 `RunE` 中并发启动：
    - goroutine A：调用 `service.Translate` 将结果写入 `pipeRW` 的写端；结束后关闭写端。
    - goroutine B：调用 `service.Output` 从读端连续输出到终端；读端遇到 EOF 结束。
- 语言方向识别
  - 沿用 `service.detectLang`（汉字→`zh2en`，否则 `en2zh`）；同时支持 `--direction` 覆盖（可选）。
- 分块策略
  - `[]rune` 维度分块，默认 `size=2000`；每块独立调用 `llm.TranslateText`，由流式回调将 token 写入输出。
  - 大文本在控制台平滑滚动输出；非流式模式则一次性写入。
- 结构化输出
  - 依赖 `pkg/llm/client.go:99-152` 的系统提示，确保输出包含：
    - 【原文】原始文本
    - 【译文】翻译结果
    - 【音标】关键词 IPA
    - 【例句】至少 2 组中英对照例句
- 日志与注释规范
  - 所有日志英文（`info/debug/error`），注释中文；服务层详细日志，非服务层少量日志；错误均处理或返回。

## 实施步骤（文件级别）
1. 完成 `component/translate/cmd/command.go` 的 `RunE` 执行流程：
  - 读取 `args[0]`，支持 `-`（stdin）、文件与直接文本。
  - 构建输入 `io.ReadCloser` 与管道输出 `ReadWriteCloser`（`pipeRW`）。
  - 创建 `model.Translate` 请求并调用 `service.Translate` 与 `service.Output`（并发、流式）。
  - 严格资源管理：`defer` 关闭输入与输出端；错误路径日志记录。
2. 增强 `component/translate/service/translator.go`：
  - 修正 `Input(in string)`：根据传入的 `in` 处理文本/文件/`-` 并返回 `io.ReadCloser`；不再无条件读取 `stdin`。
  - 保留现有 `Translate` 分块与流式写入逻辑；必要处增加边界与错误日志。
  - 保留 `Output` 的 `io.Copy` 到终端；确保在写端关闭后正确结束。
3. 在 `component/translate/model/types.go` 附加 `pipeRW` 实现（或在 `service` 内部声明）以满足 `io.ReadWriteCloser` 的需求。

## 使用示例
- 英译中：`micro t "Hello, world!"`
- 中译英：`micro t "你好，世界！"`
- 读取文件：`micro t ./text.txt`
- 从 stdin：`cat long.txt | micro t -`

## 环境变量
- `DASHSCOPE_API_KEY` 或 `ALI_BAILIAN_API_KEY`
- `MICRO_MODEL_NAME`（默认 `qwen3-max`）
- `MICRO_MAX_TOKENS`、`MICRO_TEMPERATURE`、`MICRO_STREAM_OUTPUT`（默认启用）、`MICRO_CONCURRENCY`

## 验证计划
- 运行示例命令验证英/中文方向、结构化输出与流式滚动。
- 使用超长文本（>10k rune）验证分块与连续输出。
- 在无流式模式（设置 `MICRO_STREAM_OUTPUT=false`）验证一次性输出。
- 错误场景：无 API Key、文件不存在、空输入；日志为英文、错误处理完整。

## 代码引用
- 子命令骨架：`component/translate/cmd/command.go:12-45`
- 翻译服务：`component/translate/service/translator.go:35-68`（分块与流式）、`:98-106`（方向识别）、`:108-123`（分块）
- LLM 翻译提示：`pkg/llm/client.go:99-152`（结构化输出）、`:154-175`（流式）、`:177-204`（非流式）
- 根命令集成：`cmd/micro/main.go:23-26`