# AGENTS.md — microdev

## 项目

个人长期运营的命令行微开发工具集，Go 编写，二进制名 `micro`（安装时同时链接短命令 `m`）。基于 cobra，大模型走通义千问（DashScope 的 OpenAI 兼容接口）。

| 子命令 | 别名 | 作用 |
| --- | --- | --- |
| `micro prompt` | `p` | 优化提示词；`-i` 进入带会话记忆的交互模式 |
| `micro translate` | `t` | 中英互译，输入可以是文本、文件路径或 `-`（标准输入），流式输出 |
| `micro file` | `f` | 检测二进制文件信息（格式、架构、是否 Go 构建等） |

- 入口：`cmd/micro/main.go`，只负责注册子命令与信号取消，不放业务。
- 配置：`component/config`，全部走环境变量，没有配置文件。至少需要 `DASHSCOPE_API_KEY` 或 `ALI_BAILIAN_API_KEY`；变量名与默认值见 `infra/consts/consts.go`。
- 构建与安装：`make build`、`make install`，其余见 `Makefile`。

## 目录结构

结构与 `rockcore/backend` 保持一致，新增代码必须落在对应位置。

```text
cmd/micro/main.go            程序入口
component/<name>/            按功能划分的组件，目前有 file、translate、prompt
  api/                       cobra 命令，只做参数与输出适配，相当于 HTTP 项目的 handler
  service/                   业务编排
  store/                     持久化（DAL），目前只有 prompt 用到
component/config/            环境变量配置加载
infra/                       不含业务的基础设施
  consts/                    项目级常量与提示词，所有常量必须放这里
  logger/                    结构化日志
  console/                   面向用户的终端输出，与日志分开
  rdb/                       数据库连接封装
  llm/                       大模型客户端
  utils/                     通用工具
  model/<领域>/              领域模型与 Param、Response，包名与目录名一致
  model/filter/              DAL 通用分页与排序条件
migrate/<driver>/            各数据库的 up.sql、drop.sql、truncate.sql
migrate/migrate.go           内嵌 up.sql，供默认内存库启动时自动建表
```

要点：

1. 模型不放在 `component/*/`，统一放 `infra/model/<领域>/`；引用时加别名，例如 `promptmodel "microdev/infra/model/prompt"`。
2. `api` 层只做适配：解析参数、读输入、组装依赖、调 service、输出结果。装配在各命令的 `RunE` 里完成，因为命令是惰性执行，`micro file` 不应因为缺 API Key 而失败。
3. 用户可见的输出走 `infra/console`，结构化日志走 `infra/logger`（默认输出到标准错误，避免污染译文）。
4. 不允许空目录或 stub 占位，建了目录就必须有真实内容。

## 开发环境

MacBook M1 (ARM64)，Go 1.25。本机 `CGO_ENABLED=0`，因此依赖必须是纯 Go 实现：SQLite 使用 `github.com/glebarez/sqlite`，不要换回依赖 cgo 的驱动。

## 开发规则

1. 以最优解为唯一目标：项目由个人长期运营，没有外部使用方，判断一次改动好坏的标准是「改完之后是不是当前能想到的最好实现」，而不是 diff 有多小。允许大范围重写与跨目录重构，不必事先请示；只有需要引入新依赖、或改动会破坏当前可运行状态时才先说明。
2. 可读性第一：命名自解释、职责单一、显式胜于隐式、线性平铺优先于过度抽象。宁可多写几行直白代码，不可牺牲可理解性。
3. 不背历史包袱：不保留兼容分支、适配层、废弃字段或过渡开关。发现更优的命名、分层或数据结构就直接改到位，调用方一并同步。重构必须一次做干净，禁止新旧两套实现并存。
4. 暂时未被引用的代码不要删：这是个人长期运营的工具集，`infra/utils` 里的路径与语言工具、`RWPipe`、`RedisSessionCache` 等备用代码可能以后还会用。它们不是免检区，仍要按 Go Skill 的规范维护；确实要删时先问。
5. 测试写 Go `_test.go`，用 testify，表驱动，放在被测代码旁边。测试自身不准备外部环境：需要真实大模型或外部数据库的用例不写进默认的 `go test ./...`；需要数据库的单测一律用 SQLite 内存库，DSN 带上 `t.Name()` 避免用例间串库。
6. 数据库 schema 以全量快照维护，不使用增量迁移文件：`migrate/<driver>/` 每种数据库只保留 `up.sql`（幂等全量建表）、`drop.sql`、`truncate.sql`。改 schema 直接改 `up.sql`，重建时先跑 `drop.sql`。`migrate.Statements` 按分号切分 `up.sql`，所以语句以分号结尾，字符串字面量里不要出现分号。
   - 与 rockcore 的唯一差异：CLI 默认使用 SQLite 内存库，无法手动建表，所以 `micro prompt` 启动时会自动执行内嵌的 `sqlite/up.sql`。外部数据库（MySQL）仍由开发者手动执行脚本，MySQL 需 8.0.13 及以上。
7. 新增配置项时，环境变量名与默认值进 `infra/consts`，解析与校验进 `component/config`；非法值回退到默认值，不要让程序因配置笔误而崩溃。
8. 底线：任何改动至少保证 `go build ./...` 与 `go vet ./...` 通过，并跑过相关包的 `go test`。改动涉及 `redis` 构建标签文件时，额外执行 `go build -tags redis ./...`。

## Go Skill

所有 Go 代码改动必须先走 `go` skill —— 唯一入口，按需路由到 `go-api-layer`、`go-code-style`、`go-model-hierarchy`、`go-query-dal`、`go-service-layer`、`go-test-writer`。注释与日志细则在 `go-code-style` 的 `references/` 下：代码注释必须中文，日志 Msg 必须英文。

说明：本项目没有 HTTP 层，`api` 目录里是 cobra 命令，对照 `go-api-layer` 时只取「只做适配、不写业务」这一条；`go-gin-openapi-json` 不适用。
