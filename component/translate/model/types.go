package model

// TranslationRequest 翻译请求
type TranslationRequest struct {
	Content    string `json:"content"`
	InputPath  string `json:"input_path,omitempty"`
	OutputPath string `json:"output_path,omitempty"`
	Language   string `json:"language,omitempty"` // zh2en 或 en2zh
}

// TranslationResponse 翻译响应
type TranslationResponse struct {
	TranslatedContent string `json:"translated_content"`
	OriginalContent   string `json:"original_content"`
	Language          string `json:"language"`
}

// TranslationTask 翻译任务
type TranslationTask struct {
	InputPath  string
	Content    string
	OutputPath string
}

// TranslationConfig 翻译配置
type TranslationConfig struct {
	MaxTokens     int     `json:"max_tokens"`
	Temperature   float64 `json:"temperature"`
	StreamOutput  bool    `json:"stream_output"`
	Concurrency   int     `json:"concurrency"`
	ForceRewrite  bool    `json:"force_rewrite"`
}

// ProcessingOptions 处理选项
type ProcessingOptions struct {
	Directory    bool   `json:"directory"`     // 是否处理目录
	File         bool   `json:"file"`          // 是否处理文件
	FileType     string `json:"file_type"`     // 文件类型过滤
	Force        bool   `json:"force"`         // 强制重新翻译
	Concurrency  int    `json:"concurrency"`   // 并发数
	OutputSuffix string `json:"output_suffix"` // 输出文件后缀
}

// Chunk Markdown内容块
type Chunk struct {
	Content string
	Level   int    // 标题层级，0表示非标题块
	Title   string // 标题文本
}

// FileInfo 文件信息
type FileInfo struct {
	Path         string
	Name         string
	Extension    string
	Size         int64
	IsDirectory  bool
	LastModified int64
}
