package model

// Result 表示对目标二进制文件的分析结果。
//
// 字段来源说明：
//   - Format：通过解析文件格式得到（Mach-O/ELF/PE/unknown）。
//   - IsExecutable：是否属于可执行或可装载的格式（Mach-O: TypeExec/TypeDylib；ELF: ET_EXEC/ET_DYN）。
//   - HasExecPerms：文件权限是否包含可执行位（基于 os.Stat 的 Mode）。
//   - HasDynamicLibs：是否引用动态链接库（Mach-O 通过 ImportedLibraries；ELF 通过 DT_NEEDED）。
//   - GOOS/GOARCH：优先从 Go 构建信息（debug/buildinfo）读取；若非 Go 构建则根据二进制格式推断。
//     注意在 Mach-O fat（通用二进制）下，GOARCH 可能为多个架构，以逗号分隔（例如 "amd64,arm64"）。
//   - IsGoBinary/GoVersion/CGOEnabled：仅在检测到 Go 构建信息时填写；
//     其中 CGOEnabled 为字符串形式（"0"/"1"），保持与构建信息一致。
type Result struct {
	// Path 为输入文件的绝对路径（尽可能规范化为绝对路径，便于展示与追踪）。
	Path string

	// Format 为二进制文件格式：如 "mach-o"、"mach-o.fat"、"elf"、"pe"、"unknown"。
	Format string

	// IsExecutable 表示该文件是否为可执行或可装载格式（不等同于具有可执行权限）。
	IsExecutable bool

	// HasExecPerms 表示文件权限是否包含可执行位（针对当前文件系统权限）。
	HasExecPerms bool

	// HasDynamicLibs 表示该二进制是否在动态链接时依赖外部库（如 DT_NEEDED 或 LC_LOAD_DYLIB）。
	HasDynamicLibs bool

	// GOOS 为目标平台的操作系统，例如 "darwin"、"linux"、"windows"。
	GOOS string

	// GOARCH 为目标平台的架构，例如 "amd64"、"arm64"；
	// 在 Mach-O fat 二进制下可能包含多个架构，使用逗号分隔（如 "amd64,arm64"）。
	GOARCH string

	// IsGoBinary 指示该二进制是否包含 Go 构建信息（debug/buildinfo），便于进一步分析版本与设置。
	IsGoBinary bool

	// GoVersion 为 Go 构建版本号（例如 "go1.25.3"），仅在 IsGoBinary 为 true 时有效。
	GoVersion string

	// CGOEnabled 为 Go 构建时的 CGO 开关设置（"0" 或 "1"），仅在 IsGoBinary 为 true 时有效。
	CGOEnabled string
}
