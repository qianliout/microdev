# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Project summary
- MicroDev is a Go CLI with two primary features: prompt optimization and Chinese/English translation. It uses Cobra for CLI, zerolog for logging, and langchaingo (OpenAI-compatible) to call Qwen models via DashScope/Bailian. Docs and tests are included for both subcommands.

Prerequisites
- Go ≥ 1.24
- One of these environment variables must be set before running most commands:
  - export DASHSCOPE_API_KEY={{DASHSCOPE_API_KEY}}
  - or export ALI_BAILIAN_API_KEY={{ALI_BAILIAN_API_KEY}}
- Optional runtime tuning:
  - MICRO_MODEL_NAME (default qwen3-max), MICRO_MAX_TOKENS (default 2048), MICRO_TEMPERATURE (default 0.7)
  - MICRO_STREAM_OUTPUT=true|false (default true)
  - MICRO_CONCURRENCY=1..10 (default 3)

Common commands
- Build
  - make build
  - go build -o micro ./cmd/micro
  - With optional storage build tags (docs/general/README.md):
    - go build -tags mysql -o micro ./cmd/micro
    - go build -tags redis -o micro ./cmd/micro
    - go build -tags "mysql redis" -o micro ./cmd/micro
- Install binaries to GOBIN (also installs short alias m)
  - ./build.sh
  - make install
- Format and lint
  - make fmt      # go fmt ./...
  - make lint     # go vet ./...
- Tests
  - All tests: ./tests/run_tests.sh
  - All tests (Go): go test -v ./tests
  - Single test by name: go test -v ./tests -run TestName
  - Single file: go test -v ./tests/session_manager_test.go
  - Coverage: go test -coverprofile=coverage.out ./tests && go tool cover -html=coverage.out -o coverage.html

Minimal usage (to quickly validate binaries)
- Prompt: micro p "优化这个提示词"  (interactive: micro p -i)
- Translate: micro t "Hello World"  (interactive: micro t -i)

High-level architecture
- Entry point
  - cmd/micro/main.go defines the Cobra root command (micro) and adds subcommands from component/prompt and component/translate.
- Components (feature-oriented modules under component/)
  - prompt/
    - cmd/command.go wires CLI flags.
    - service/optimizer.go is the core “prompt optimization” logic. It builds system/user prompts, supports streaming and non-streaming output, and integrates a session Manager for interactive mode.
    - session/ implements conversation state with pluggable storage interfaces; memory storage is default. Files for MySQL/Redis storage exist behind build tags (see docs); actual implementations may be stubs pending future work.
    - model/types.go defines request/response types.
  - translate/
    - cmd/command.go wires CLI flags: file/dir modes, type filters, concurrency, force, interactive.
    - service/translator.go orchestrates translation: detects language, keeps Markdown structure, chunks long content via MarkdownService, supports idempotent outputs (e.g., _zh/_en suffix), and concurrency control via config.
    - service/processor.go handles directory/file traversal, filtering, and output path computation.
    - model/types.go defines request/response types.
- Shared packages (pkg/)
  - config/config.go loads env vars, sets defaults, validates availability of at least one API key, and exposes tunables (model, tokens, temperature, stream, concurrency with max 10).
  - llm/client.go wraps langchaingo’s OpenAI-compatible client to call Qwen models via DashScope/Bailian endpoints.
  - logger/logger.go is a thin zerolog wrapper with structured logs and user-facing helpers (UserInfo/UserError/...)
  - input/processor.go and output/processor.go centralize file/stdin/stdout handling for commands.
  - utils/language.go provides zh/en detection used by translate.
- Data flow
  - CLI (Cobra) → component service (prompt/translate) → LLM via langchaingo → output writer (streaming or buffered). Config is resolved first; logger provides structured telemetry and user messages.
- Tests and docs
  - tests/ contains GoConvey-based unit/integration tests. Connectivity tests are skipped if API keys are absent. Use run_tests.sh for a full local run with coverage.
  - docs/ provides detailed per-feature guides and general configuration (including build tags for storage backends).

Notes for future changes (observed from repo)
- LLM client creation logic appears in multiple places (pkg/llm and component services). Consider centralizing if you plan refactors.
- Storage backends for session (MySQL/Redis) are guarded by build tags; confirm implementations before enabling tags in CI.
