package file

import (
	"fmt"
	"strings"

	"microdev/infra/consts"
)

type AnalyzeParam struct {
	Path string `json:"path"`
}

func (p *AnalyzeParam) Serialize() *AnalyzeParam {
	if p == nil {
		return &AnalyzeParam{}
	}
	p.Path = strings.TrimSpace(p.Path)
	return p
}

func (p *AnalyzeParam) Check() error {
	if p == nil {
		return fmt.Errorf("analyze param is nil")
	}
	if p.Path == "" {
		return fmt.Errorf("path is empty")
	}
	return nil
}

type Result struct {
	// 输入文件的绝对路径。
	Path string `json:"path"`

	Format string `json:"format"`

	// 是否为可执行或可装载格式，不等同于文件权限含可执行位。
	IsExecutable string `json:"isExecutable"`

	HasExecPerms string `json:"hasExecPerms"`

	// 是否依赖外部动态库，Mach-O 看导入库，ELF 看 DT_NEEDED。
	HasDynamicLibs string `json:"hasDynamicLibs"`

	GOOS string `json:"goos"`

	// Mach-O fat 二进制包含多个架构时以逗号分隔，如 "amd64,arm64"。
	GOARCH string `json:"goarch"`

	IsGoBinary string `json:"isGoBinary"`

	// 仅 IsGoBinary 为 true 时有效。
	GoVersion string `json:"goVersion"`

	// 与构建信息一致，取值 "0" 或 "1"，仅 IsGoBinary 为 true 时有效。
	CGOEnabled string `json:"cgoEnabled"`
}

func (vi *Result) Serialize() *Result {
	if vi == nil {
		return &Result{}
	}
	vi.Path = defaultString(vi.Path, consts.Unknown)
	vi.Format = defaultString(vi.Format, consts.Unknown)
	vi.GOOS = defaultString(vi.GOOS, consts.Unknown)
	vi.GOARCH = defaultString(vi.GOARCH, consts.Unknown)
	vi.IsExecutable = defaultString(vi.IsExecutable, consts.FalseString)
	vi.HasExecPerms = defaultString(vi.HasExecPerms, consts.FalseString)
	vi.HasDynamicLibs = defaultString(vi.HasDynamicLibs, consts.FalseString)
	vi.IsGoBinary = defaultString(vi.IsGoBinary, consts.FalseString)
	if vi.IsGoBinary == consts.TrueString {
		vi.GoVersion = defaultString(vi.GoVersion, consts.Unknown)
		vi.CGOEnabled = defaultString(vi.CGOEnabled, consts.Unknown)
	}
	return vi
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
