package translate

import (
	"fmt"
	"io"
	"strings"

	"microdev/infra/consts"
)

type TranslateParam struct {
	Text string `json:"text"`
	// 为空时由服务按文本内容自动判断。
	Direction string    `json:"direction"`
	Output    io.Writer `json:"-"`
}

func (p *TranslateParam) Serialize() *TranslateParam {
	if p == nil {
		return &TranslateParam{}
	}
	p.Direction = strings.TrimSpace(p.Direction)
	return p
}

func (p *TranslateParam) Check() error {
	if p == nil {
		return fmt.Errorf("translate param is nil")
	}
	if strings.TrimSpace(p.Text) == "" {
		return fmt.Errorf("text is empty")
	}
	if p.Output == nil {
		return fmt.Errorf("output writer is nil")
	}
	isZh2En := p.Direction == consts.DirectionZh2En
	isEn2Zh := p.Direction == consts.DirectionEn2Zh
	if p.Direction != "" && !isZh2En && !isEn2Zh {
		return fmt.Errorf("unsupported direction %q", p.Direction)
	}
	return nil
}
