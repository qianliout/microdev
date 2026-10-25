package llm

import (
	"fmt"
	"io"
	"strings"

	"microdev/infra/consts"
)

type GenerateParam struct {
	System string    `json:"system"`
	User   string    `json:"user"`
	Output io.Writer `json:"-"`
}

func (p *GenerateParam) Serialize() *GenerateParam {
	if p == nil {
		return &GenerateParam{}
	}
	p.System = strings.TrimSpace(p.System)
	return p
}

func (p *GenerateParam) Check() error {
	if p == nil {
		return fmt.Errorf("generate param is nil")
	}
	if p.User == "" {
		return fmt.Errorf("user message is empty")
	}
	if p.Output == nil {
		return fmt.Errorf("output writer is nil")
	}
	return nil
}

type TranslateTextParam struct {
	Text      string    `json:"text"`
	Direction string    `json:"direction"`
	Output    io.Writer `json:"-"`
}

func (p *TranslateTextParam) Serialize() *TranslateTextParam {
	if p == nil {
		return &TranslateTextParam{}
	}
	p.Direction = strings.TrimSpace(p.Direction)
	return p
}

func (p *TranslateTextParam) Check() error {
	if p == nil {
		return fmt.Errorf("translate text param is nil")
	}
	if p.Text == "" {
		return fmt.Errorf("text is empty")
	}
	isZh2En := p.Direction == consts.DirectionZh2En
	isEn2Zh := p.Direction == consts.DirectionEn2Zh
	if !isZh2En && !isEn2Zh {
		return fmt.Errorf("unsupported direction %q", p.Direction)
	}
	if p.Output == nil {
		return fmt.Errorf("output writer is nil")
	}
	return nil
}

type OptimizePromptParam struct {
	Prompt string    `json:"prompt"`
	Output io.Writer `json:"-"`
}

func (p *OptimizePromptParam) Serialize() *OptimizePromptParam {
	if p == nil {
		return &OptimizePromptParam{}
	}
	return p
}

func (p *OptimizePromptParam) Check() error {
	if p == nil {
		return fmt.Errorf("optimize prompt param is nil")
	}
	if p.Prompt == "" {
		return fmt.Errorf("prompt is empty")
	}
	if p.Output == nil {
		return fmt.Errorf("output writer is nil")
	}
	return nil
}
