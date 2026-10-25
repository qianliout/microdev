package youdao

import (
	"fmt"
	"strings"
)

type CreateNoteParam struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	FolderID string `json:"folderId"`
}

func (p *CreateNoteParam) Serialize() *CreateNoteParam {
	if p == nil {
		return &CreateNoteParam{}
	}
	p.Title = strings.TrimSpace(p.Title)
	p.FolderID = strings.TrimSpace(p.FolderID)
	return p
}

func (p *CreateNoteParam) Check() error {
	if p == nil {
		return fmt.Errorf("create note param is nil")
	}
	if p.Title == "" {
		return fmt.Errorf("note title is empty")
	}
	return nil
}

type UpdateNoteParam struct {
	FileID  string `json:"fileId"`
	Content string `json:"content"`
}

func (p *UpdateNoteParam) Serialize() *UpdateNoteParam {
	if p == nil {
		return &UpdateNoteParam{}
	}
	p.FileID = strings.TrimSpace(p.FileID)
	return p
}

func (p *UpdateNoteParam) Check() error {
	if p == nil {
		return fmt.Errorf("update note param is nil")
	}
	if p.FileID == "" {
		return fmt.Errorf("note file id is empty")
	}
	return nil
}
