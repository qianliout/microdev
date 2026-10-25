package consts

// 有道云笔记同步的环境变量名与默认值。
const (
	EnvYoudaoDir = "MICRO_YOUDAO_DIR"
	EnvYoudaoBin = "MICRO_YOUDAO_BIN"

	DefaultYoudaoDir = "~/youdao-notes"
	DefaultYoudaoBin = "youdaonote"
)

// 同步目录下记录本地文件与云端笔记对应关系的文件名。
const YoudaoMappingFile = ".micro-youdao.json"

// 仅同步该扩展名的文件，官方 update 只支持 Markdown 笔记。
const MarkdownExt = ".md"

// 以该前缀开头的文件与目录不参与同步。
const HiddenPrefix = "."

// 官方 youdaonote CLI 的子命令、参数与保存笔记时的类型。
const (
	YoudaoCmdSave   = "save"
	YoudaoCmdUpdate = "update"
	YoudaoFlagFile  = "--file"
	YoudaoFlagJSON  = "--json"
	YoudaoFlagData  = "-c"

	YoudaoNoteTypeMarkdown = "md"
	YoudaoTempFilePattern  = "micro-youdao-*.json"
)

// 同步结果中单个文件的动作。
const (
	SyncActionCreate = "create"
	SyncActionUpdate = "update"
	SyncActionSkip   = "skip"
	SyncActionFail   = "fail"
)

// 映射文件的权限。
const YoudaoMappingFileMode = 0o644
