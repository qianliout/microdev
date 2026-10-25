package youdao

import (
	"fmt"
	"strings"
)

type SyncParam struct {
	// 为空时同步 Dir。
	Path string `json:"path"`
	// 默认同步目录，单文件位于其中时共用该目录的映射文件。
	Dir string `json:"dir"`
	// 仅首次创建笔记时生效，为空表示放在云端根目录。
	Folder string `json:"folder"`
	// 仅单文件同步时允许指定，为空取文件名。
	Title  string `json:"title"`
	DryRun bool   `json:"dryRun"`
}

func (p *SyncParam) Serialize() *SyncParam {
	if p == nil {
		return &SyncParam{}
	}
	p.Path = strings.TrimSpace(p.Path)
	p.Dir = strings.TrimSpace(p.Dir)
	p.Folder = strings.TrimSpace(p.Folder)
	p.Title = strings.TrimSpace(p.Title)
	if p.Path == "" {
		p.Path = p.Dir
	}
	return p
}

func (p *SyncParam) Check() error {
	if p == nil {
		return fmt.Errorf("sync param is nil")
	}
	if p.Dir == "" {
		return fmt.Errorf("default dir is empty")
	}
	return nil
}

func (p *SyncParam) LogStr() string {
	return fmt.Sprintf("path=%s dir=%s folder=%s dryRun=%t", p.Path, p.Dir, p.Folder, p.DryRun)
}

type FileResult struct {
	// 相对同步目录的路径，无法计算时为绝对路径。
	Path   string
	Action string
	Note   string
	Err    error
}

type SyncSummary struct {
	Results []*FileResult
	Created int64
	Updated int64
	Skipped int64
	Failed  int64
}
