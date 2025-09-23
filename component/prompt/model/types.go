package model

// OptimizationRequest 提示词优化请求
type OptimizationRequest struct {
	OriginalPrompt string `json:"original_prompt"`
	Context        string `json:"context,omitempty"`
	Language       string `json:"language,omitempty"`
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
