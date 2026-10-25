@AGENTS.md

## Claude Code 补充

- 项目规则以 `AGENTS.md` 为唯一来源，本文件只放 Claude Code 专有的补充，不要在这里复述或另写一份规则。
- 改任何 Go 代码前先调用 `go` skill；改完必须跑 `goimports`、`go build ./...`、`go vet ./...` 和相关包的 `go test`。
