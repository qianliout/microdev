package prompt

import (
	"fmt"
	"io"
	"strings"
)

type OptimizeParam struct {
	OriginalPrompt string    `json:"originalPrompt"`
	Context        string    `json:"context"`
	Language       string    `json:"language"`
	SessionContext string    `json:"sessionContext"`
	Output         io.Writer `json:"-"`
}

func (p *OptimizeParam) Serialize() *OptimizeParam {
	if p == nil {
		return &OptimizeParam{}
	}
	p.Language = strings.TrimSpace(p.Language)
	return p
}

func (p *OptimizeParam) Check() error {
	if p == nil {
		return fmt.Errorf("optimize param is nil")
	}
	if strings.TrimSpace(p.OriginalPrompt) == "" {
		return fmt.Errorf("original prompt is empty")
	}
	if p.Output == nil {
		return fmt.Errorf("output writer is nil")
	}
	return nil
}

type OptimizeWithSessionParam struct {
	SessionID      int64     `json:"sessionId"`
	OriginalPrompt string    `json:"originalPrompt"`
	Context        string    `json:"context"`
	Language       string    `json:"language"`
	Output         io.Writer `json:"-"`
}

func (p *OptimizeWithSessionParam) Serialize() *OptimizeWithSessionParam {
	if p == nil {
		return &OptimizeWithSessionParam{}
	}
	p.Language = strings.TrimSpace(p.Language)
	return p
}

func (p *OptimizeWithSessionParam) Check() error {
	if p == nil {
		return fmt.Errorf("optimize with session param is nil")
	}
	if p.SessionID <= 0 {
		return fmt.Errorf("sessionID must be positive")
	}
	if strings.TrimSpace(p.OriginalPrompt) == "" {
		return fmt.Errorf("original prompt is empty")
	}
	if p.Output == nil {
		return fmt.Errorf("output writer is nil")
	}
	return nil
}

type OptimizeResponse struct {
	OptimizedPrompt string `json:"optimizedPrompt"`
	Improvements    string `json:"improvements"`
	Suggestions     string `json:"suggestions"`
}

type OptimizeConfig struct {
	MaxTokens     int64   `json:"maxTokens"`
	Temperature   float64 `json:"temperature"`
	StreamOutput  string  `json:"streamOutput"`
	IncludeReason string  `json:"includeReason"`
}

type InteractiveConfig struct {
	EnableSession  string   `json:"enableSession"`
	ExitCommands   []string `json:"exitCommands"`
	WelcomeMessage string   `json:"welcomeMessage"`
	PromptPrefix   string   `json:"promptPrefix"`
}
