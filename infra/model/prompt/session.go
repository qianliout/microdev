package prompt

import (
	"fmt"
	"strings"

	"microdev/infra/consts"
	"microdev/infra/model/filter"
)

type Session struct {
	ID int64 `gorm:"primaryKey;column:id" json:"id"`

	// 压缩旧消息后得到的会话摘要。
	Summary string `gorm:"column:summary" json:"summary"`

	CreatedAt int64 `gorm:"autoCreateTime:milli;column:created_at" json:"createdAt"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli;column:updated_at" json:"updatedAt"`
	DeletedAt int64 `gorm:"column:deleted_at" json:"deletedAt"`
}

func (vi *Session) TableName() string { return "prompt_session" }

func (vi *Session) Serialize() *Session {
	if vi == nil {
		return &Session{}
	}
	vi.Summary = strings.TrimSpace(vi.Summary)
	return vi
}

func (vi *Session) Deserialize() *Session {
	if vi == nil {
		return &Session{}
	}
	return vi
}

func (vi *Session) Check() error {
	if vi == nil {
		return fmt.Errorf("session is nil")
	}
	return nil
}

type Message struct {
	ID int64 `gorm:"primaryKey;column:id" json:"id"`

	SessionID int64  `gorm:"column:session_id" json:"sessionId"`
	Role      string `gorm:"column:role" json:"role"`
	Content   string `gorm:"column:content" json:"content"`

	CreatedAt int64 `gorm:"autoCreateTime:milli;column:created_at" json:"createdAt"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli;column:updated_at" json:"updatedAt"`
	DeletedAt int64 `gorm:"column:deleted_at" json:"deletedAt"`
}

func (vi *Message) TableName() string { return "prompt_message" }

func (vi *Message) Serialize() *Message {
	if vi == nil {
		return &Message{}
	}
	vi.Role = strings.TrimSpace(vi.Role)
	return vi
}

func (vi *Message) Deserialize() *Message {
	if vi == nil {
		return &Message{}
	}
	return vi
}

func (vi *Message) Check() error {
	if vi == nil {
		return fmt.Errorf("message is nil")
	}
	if vi.SessionID <= 0 {
		return fmt.Errorf("sessionID must be positive")
	}
	isUser := vi.Role == consts.RoleUser
	isAssistant := vi.Role == consts.RoleAssistant
	if !isUser && !isAssistant {
		return fmt.Errorf("unsupported role %q", vi.Role)
	}
	return nil
}

type CreateSessionParam struct {
	Summary string `json:"summary"`
}

func (p *CreateSessionParam) Serialize() *CreateSessionParam {
	if p == nil {
		return &CreateSessionParam{}
	}
	p.Summary = strings.TrimSpace(p.Summary)
	return p
}

func (p *CreateSessionParam) Check() error {
	if p == nil {
		return fmt.Errorf("create session param is nil")
	}
	return nil
}

type SessionIDParam struct {
	SessionID int64 `json:"sessionId"`
}

func (p *SessionIDParam) Serialize() *SessionIDParam {
	if p == nil {
		return &SessionIDParam{}
	}
	return p
}

func (p *SessionIDParam) Check() error {
	if p == nil {
		return fmt.Errorf("session id param is nil")
	}
	if p.SessionID <= 0 {
		return fmt.Errorf("sessionID must be positive")
	}
	return nil
}

type AddMessageParam struct {
	SessionID int64  `json:"sessionId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

func (p *AddMessageParam) Serialize() *AddMessageParam {
	if p == nil {
		return &AddMessageParam{}
	}
	p.Role = strings.TrimSpace(p.Role)
	return p
}

func (p *AddMessageParam) Check() error {
	if p == nil {
		return fmt.Errorf("add message param is nil")
	}
	if p.SessionID <= 0 {
		return fmt.Errorf("sessionID must be positive")
	}
	isUser := p.Role == consts.RoleUser
	isAssistant := p.Role == consts.RoleAssistant
	if !isUser && !isAssistant {
		return fmt.Errorf("unsupported role %q", p.Role)
	}
	return nil
}

type SearchSessionParam struct {
	ID int64 `json:"id"`
	// 大于 0 时只查更新时间早于该毫秒时间戳的会话。
	UpdatedBefore int64          `json:"updatedBefore"`
	Filter        *filter.Filter `json:"filter"`
}

func (p *SearchSessionParam) Serialize() *SearchSessionParam {
	if p == nil {
		return &SearchSessionParam{}
	}
	return p
}

func (p *SearchSessionParam) Check() error {
	if p == nil {
		return fmt.Errorf("search session param is nil")
	}
	return nil
}

type SearchMessageParam struct {
	SessionID int64          `json:"sessionId"`
	Filter    *filter.Filter `json:"filter"`
}

func (p *SearchMessageParam) Serialize() *SearchMessageParam {
	if p == nil {
		return &SearchMessageParam{}
	}
	return p
}

func (p *SearchMessageParam) Check() error {
	if p == nil {
		return fmt.Errorf("search message param is nil")
	}
	if p.SessionID <= 0 {
		return fmt.Errorf("sessionID must be positive")
	}
	return nil
}

type CleanupSessionParam struct {
	ExpiredBefore int64 `json:"expiredBefore"`
}

func (p *CleanupSessionParam) Serialize() *CleanupSessionParam {
	if p == nil {
		return &CleanupSessionParam{}
	}
	return p
}

func (p *CleanupSessionParam) Check() error {
	if p == nil {
		return fmt.Errorf("cleanup session param is nil")
	}
	if p.ExpiredBefore <= 0 {
		return fmt.Errorf("expiredBefore must be positive")
	}
	return nil
}

type SessionInfo struct {
	ID           int64  `json:"id"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	MessageCount int64  `json:"messageCount"`
	HasSummary   string `json:"hasSummary"`
}

type SessionStats struct {
	SessionActive     string `json:"sessionActive"`
	SessionID         int64  `json:"sessionId"`
	TotalMessages     int64  `json:"totalMessages"`
	UserMessages      int64  `json:"userMessages"`
	AssistantMessages int64  `json:"assistantMessages"`
	HasSummary        string `json:"hasSummary"`
	SummaryLength     int64  `json:"summaryLength"`
}
