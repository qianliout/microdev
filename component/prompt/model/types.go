package model

// OptimizationRequest 提示词优化请求
type OptimizationRequest struct {
	OriginalPrompt string `json:"original_prompt"`
	Context        string `json:"context,omitempty"`
	Language       string `json:"language,omitempty"`
	SessionContext string `json:"session_context,omitempty"` // 会话上下文
}

// OptimizationResponse 提示词优化响应
type OptimizationResponse struct {
	OptimizedPrompt string `json:"optimized_prompt"`
	Improvements    string `json:"improvements,omitempty"`
	Suggestions     string `json:"suggestions,omitempty"`
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	MaxTokens     int     `json:"max_tokens"`
	Temperature   float64 `json:"temperature"`
	StreamOutput  bool    `json:"stream_output"`
	IncludeReason bool    `json:"include_reason"`
}

// InteractiveConfig 交互式配置
type InteractiveConfig struct {
	EnableSession  bool     `json:"enable_session"`  // 是否启用会话模式
	ExitCommands   []string `json:"exit_commands"`   // 退出命令列表
	WelcomeMessage string   `json:"welcome_message"` // 欢迎消息
	PromptPrefix   string   `json:"prompt_prefix"`   // 输入提示前缀
}
