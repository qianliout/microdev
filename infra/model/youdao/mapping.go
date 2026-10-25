package youdao

import "fmt"

// Mapping 以相对同步目录的斜杠路径为键，记录已同步的云端笔记。
type Mapping struct {
	Notes map[string]*MappingEntry `json:"notes"`
}

type MappingEntry struct {
	FileID string `json:"fileId"`
	// 最近一次同步成功的本地内容 sha256。
	Hash string `json:"hash"`
}

func (vi *Mapping) Deserialize() *Mapping {
	if vi == nil {
		return &Mapping{Notes: make(map[string]*MappingEntry)}
	}
	if vi.Notes == nil {
		vi.Notes = make(map[string]*MappingEntry)
	}
	return vi
}

// MappingParam 的 Mapping 仅在保存时必填。
type MappingParam struct {
	Root    string
	Mapping *Mapping
}

func (p *MappingParam) Serialize() *MappingParam {
	if p == nil {
		return &MappingParam{}
	}
	return p
}

func (p *MappingParam) Check() error {
	if p == nil {
		return fmt.Errorf("mapping param is nil")
	}
	if p.Root == "" {
		return fmt.Errorf("mapping root is empty")
	}
	return nil
}
