// Package consts 定义项目级常量。
package consts

import "time"

// 二元状态在字段中的字符串取值。
const (
	TrueString  = "true"
	FalseString = "false"
)

// 翻译方向。
const (
	DirectionZh2En = "zh2en"
	DirectionEn2Zh = "en2zh"
)

// 会话消息角色。
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// 未识别信息的统一占位值。
const Unknown = "unknown"

// 文件格式。
const (
	FormatMachOFat = "mach-o.fat"
	FormatMachO    = "mach-o"
	FormatELF      = "elf"
	FormatPE       = "pe"
)

// 目标操作系统与架构。
const (
	OSDarwin  = "darwin"
	OSLinux   = "linux"
	OSWindows = "windows"
	OSFreeBSD = "freebsd"
	OSNetBSD  = "netbsd"
	OSSolaris = "solaris"

	ArchAMD64 = "amd64"
	ArchARM64 = "arm64"
	Arch386   = "386"
	ArchARM   = "arm"
)

// PE 文件头中的 Machine 取值。
const (
	PEMachineAMD64 uint16 = 0x8664
	PEMachine386   uint16 = 0x014c
	PEMachineARM   uint16 = 0x01c4
	PEMachineARM64 uint16 = 0xaa64
)

// Go 构建信息中的配置键。
const (
	BuildKeyGOOS       = "GOOS"
	BuildKeyGOARCH     = "GOARCH"
	BuildKeyCGOEnabled = "CGO_ENABLED"
)

// 环境变量名。
const (
	EnvDashScopeAPIKey  = "DASHSCOPE_API_KEY"
	EnvAliBailianAPIKey = "ALI_BAILIAN_API_KEY"
	EnvModelName        = "MICRO_MODEL_NAME"
	EnvMaxTokens        = "MICRO_MAX_TOKENS"
	EnvTemperature      = "MICRO_TEMPERATURE"
	EnvStreamOutput     = "MICRO_STREAM_OUTPUT"
	EnvConcurrency      = "MICRO_CONCURRENCY"
	EnvSessionDriver    = "MICRO_SESSION_DRIVER"
	EnvSessionDSN       = "MICRO_SESSION_DSN"
)

// 模型调用配置的默认值与上限。
const (
	DefaultModelName         = "qwen3-max"
	DefaultMaxTokens   int64 = 2048
	MaxMaxTokens       int64 = 8192
	DefaultTemperature       = 0.7
	MaxTemperature           = 2.0
	DefaultConcurrency int64 = 3
	MaxConcurrency     int64 = 10
	DashScopeBaseURL         = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

// 数据库驱动与会话存储默认值。
const (
	DBDriverSQLite = "sqlite"
	DBDriverMySQL  = "mysql"

	DefaultSessionDriver = DBDriverSQLite
	DefaultSessionDSN    = "file:micro_session?mode=memory&cache=shared"
)

// 翻译与会话的默认参数。
const (
	TranslateChunkSize int64 = 2000
	// 汉字占比超过该阈值时按中译英处理。
	ChineseRatioThreshold = 0.3
	// 会话消息超过该数量后触发记忆压缩。
	SessionMaxMessages int64 = 8
	// 构造优化上下文时带入的最近消息数。
	SessionRecentMessages int64 = 6
	// 会话摘要中单条用户需求的截取长度。
	SessionTopicLength int64 = 50
	// 会话列表默认的单次加载上限。
	SessionLoadLimit int64 = 1000
	// 会话默认保留时长。
	SessionTTL = 24 * time.Hour
)

// 数据库操作超时。
const (
	DBCreateTimeout = 10 * time.Second
	DBQueryTimeout  = 3 * time.Second
)

// 交互模式的退出命令。
var ExitCommands = []string{"exit", "quit", "退出", "q", "bye", "再见"}

// 翻译结果文件的后缀。
const (
	SuffixZh2En = "_en"
	SuffixEn2Zh = "_zh"
)

// 提示词优化请求的默认目标语言。
const DefaultPromptLanguage = "中文"

// 标准输入的占位参数。
const StdinArg = "-"

// 支持的大模型名称。
var SupportedModels = []string{
	"qwen3-max", "qwen-plus", "qwen-plus-latest",
	"qwen-turbo", "qwen-turbo-latest",
	"qwen-max", "qwen-max-latest",
	"qwen2.5-72b-instruct", "qwen2.5-32b-instruct",
	"qwen2.5-14b-instruct", "qwen2.5-7b-instruct",
	"qwen-flash",
}

// 文件权限中表示可执行的位。
const ExecPermMask uint32 = 0o111

// Redis 会话缓存的键格式，参数依次为会话 ID。
const (
	RedisSessionKeyFormat = "micro:session:%d"
	RedisMessageKeyFormat = "micro:session:messages:%d"
	RedisMessageIDKey     = "micro:session:message:id"
	RedisSessionIndexKey  = "micro:sessions"
	RedisSessionIDKey     = "micro:session:id"
)
