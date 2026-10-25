// Package service 编排本地 Markdown 到有道云笔记的同步。
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"microdev/component/youdao/store"
	"microdev/infra/consts"
	"microdev/infra/logger"
	youdaomodel "microdev/infra/model/youdao"
)

// noteClient 是 SyncerSrv 对云端笔记读写的最小依赖，由 *youdao.Client 满足。
type noteClient interface {
	Create(ctx context.Context, param *youdaomodel.CreateNoteParam) (string, error)
	Update(ctx context.Context, param *youdaomodel.UpdateNoteParam) error
}

type SyncerSrv struct {
	mappingStore *store.MappingStore
	noteCli      noteClient
	log          *logger.Logger
}

func NewSyncerSrv(mappingStore *store.MappingStore, noteCli noteClient) *SyncerSrv {
	return &SyncerSrv{
		mappingStore: mappingStore,
		noteCli:      noteCli,
		log:          logger.New(logger.WithModule("youdao"), logger.WithSubModule("service")),
	}
}

type syncTarget struct {
	// 映射文件所在目录，文件的映射键相对它计算。
	root  string
	files []string
}

type syncTask struct {
	root    string
	path    string
	param   *youdaomodel.SyncParam
	mapping *youdaomodel.Mapping
}

// Sync 把 param.Path 下有改动的 Markdown 同步到云端，单个文件失败不中断其余文件。
func (s *SyncerSrv) Sync(ctx context.Context, param *youdaomodel.SyncParam) (*youdaomodel.SyncSummary, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return nil, err
	}

	target, err := s.resolveTarget(param)
	if err != nil {
		return nil, err
	}
	mapping, err := s.mappingStore.Load(ctx, &youdaomodel.MappingParam{Root: target.root})
	if err != nil {
		s.log.Error(ctx).Err(err).Str("param", param.LogStr()).Msg("load mapping failed")
		return nil, err
	}

	started := time.Now()
	summary := &youdaomodel.SyncSummary{Results: make([]*youdaomodel.FileResult, 0, len(target.files))}
	for _, path := range target.files {
		res := s.syncFile(ctx, &syncTask{root: target.root, path: path, param: param, mapping: mapping})
		summary.Results = append(summary.Results, res)
		s.tally(summary, res)
	}

	s.log.Info(ctx).
		Str("param", param.LogStr()).
		Int("total", len(summary.Results)).
		Int64("created", summary.Created).
		Int64("updated", summary.Updated).
		Int64("skipped", summary.Skipped).
		Int64("failed", summary.Failed).
		Int64("costMs", time.Since(started).Milliseconds()).
		Msg("youdao sync finished")
	return summary, nil
}

func (s *SyncerSrv) tally(summary *youdaomodel.SyncSummary, res *youdaomodel.FileResult) {
	switch res.Action {
	case consts.SyncActionCreate:
		summary.Created++
	case consts.SyncActionUpdate:
		summary.Updated++
	case consts.SyncActionSkip:
		summary.Skipped++
	default:
		summary.Failed++
	}
}

func (s *SyncerSrv) resolveTarget(param *youdaomodel.SyncParam) (*syncTarget, error) {
	path, err := filepath.Abs(param.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve path %s: %w", param.Path, err)
	}
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("path %s not exist, create it or set %s", path, consts.EnvYoudaoDir)
	}
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if info.IsDir() {
		return s.dirTarget(param, path)
	}
	return s.fileTarget(param, path)
}

func (s *SyncerSrv) dirTarget(param *youdaomodel.SyncParam, dir string) (*syncTarget, error) {
	if param.Title != "" {
		return nil, fmt.Errorf("title can only be set when syncing a single file")
	}
	files, err := collectMarkdown(dir)
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", dir, err)
	}
	return &syncTarget{root: dir, files: files}, nil
}

// fileTarget 位于默认目录内的文件共用默认目录的映射，否则映射放在文件所在目录。
func (s *SyncerSrv) fileTarget(param *youdaomodel.SyncParam, file string) (*syncTarget, error) {
	if !isMarkdown(file) {
		return nil, fmt.Errorf("%s is not a %s file", file, consts.MarkdownExt)
	}
	dir, err := filepath.Abs(param.Dir)
	if err != nil {
		return nil, fmt.Errorf("resolve dir %s: %w", param.Dir, err)
	}
	rel, err := filepath.Rel(dir, file)
	escapes := rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
	if err != nil || escapes {
		return &syncTarget{root: filepath.Dir(file), files: []string{file}}, nil
	}
	return &syncTarget{root: dir, files: []string{file}}, nil
}

// syncFile 记录单个文件的同步结果，错误只在这里打一次日志。
func (s *SyncerSrv) syncFile(ctx context.Context, task *syncTask) *youdaomodel.FileResult {
	res, err := s.applyFile(ctx, task)
	if err == nil {
		return res
	}
	res.Action = consts.SyncActionFail
	res.Err = err
	s.log.Error(ctx).Err(err).Str("path", task.path).Str("param", task.param.LogStr()).Msg("sync note failed")
	return res
}

func (s *SyncerSrv) applyFile(ctx context.Context, task *syncTask) (*youdaomodel.FileResult, error) {
	res := &youdaomodel.FileResult{Path: task.path}
	rel, err := filepath.Rel(task.root, task.path)
	if err != nil {
		return res, fmt.Errorf("relative path of %s: %w", task.path, err)
	}
	key := filepath.ToSlash(rel)
	res.Path = key

	data, err := os.ReadFile(task.path)
	if err != nil {
		return res, fmt.Errorf("read note %s: %w", key, err)
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	entry := task.mapping.Notes[key]
	if entry != nil && entry.Hash == hash {
		res.Action = consts.SyncActionSkip
		return res, nil
	}
	if entry == nil {
		res.Action = consts.SyncActionCreate
		return res, s.createNote(ctx, task, key, data, hash)
	}
	res.Action = consts.SyncActionUpdate
	if task.param.Folder != "" {
		res.Note = "folder ignored, note already exists in cloud"
	}
	return res, s.updateNote(ctx, task, key, entry, data, hash)
}

func (s *SyncerSrv) createNote(ctx context.Context, task *syncTask, key string, data []byte, hash string) error {
	if task.param.DryRun {
		return nil
	}
	title := task.param.Title
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(task.path), filepath.Ext(task.path))
	}
	createParam := &youdaomodel.CreateNoteParam{Title: title, Content: string(data), FolderID: task.param.Folder}
	fileID, err := s.noteCli.Create(ctx, createParam)
	if err != nil {
		return fmt.Errorf("create note %s: %w", key, err)
	}

	task.mapping.Notes[key] = &youdaomodel.MappingEntry{FileID: fileID, Hash: hash}
	if err := s.saveMapping(ctx, task); err != nil {
		return fmt.Errorf("note %s created in cloud (fileId %s) but mapping not saved: %w", key, fileID, err)
	}
	return nil
}

func (s *SyncerSrv) updateNote(ctx context.Context, task *syncTask, key string,
	entry *youdaomodel.MappingEntry, data []byte, hash string) error {
	if task.param.DryRun {
		return nil
	}
	if err := s.noteCli.Update(ctx, &youdaomodel.UpdateNoteParam{FileID: entry.FileID, Content: string(data)}); err != nil {
		return fmt.Errorf("update note %s (fileId %s): %w", key, entry.FileID, err)
	}

	entry.Hash = hash
	if err := s.saveMapping(ctx, task); err != nil {
		return fmt.Errorf("note %s updated in cloud but mapping not saved: %w", key, err)
	}
	return nil
}

// saveMapping 每个文件同步成功后立即落盘，中途中断也不会丢失已创建笔记的映射。
func (s *SyncerSrv) saveMapping(ctx context.Context, task *syncTask) error {
	return s.mappingStore.Save(ctx, &youdaomodel.MappingParam{Root: task.root, Mapping: task.mapping})
}

func isMarkdown(path string) bool {
	return strings.EqualFold(filepath.Ext(path), consts.MarkdownExt)
}

// collectMarkdown 递归收集 root 下的 Markdown，跳过以点开头的文件与目录。
func collectMarkdown(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		hidden := path != root && strings.HasPrefix(d.Name(), consts.HiddenPrefix)
		if hidden && d.IsDir() {
			return filepath.SkipDir
		}
		if hidden || d.IsDir() || !isMarkdown(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}
