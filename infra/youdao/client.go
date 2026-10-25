// Package youdao 封装官方 youdaonote CLI 的调用。
package youdao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"microdev/infra/consts"
	youdaomodel "microdev/infra/model/youdao"
)

type Client struct {
	bin string
}

func NewClient(bin string) *Client {
	return &Client{bin: bin}
}

// noteResponse 兼容 fileId 位于顶层或 data 下、字段名为 fileId 或 id 的输出。
type noteResponse struct {
	FileID string `json:"fileId"`
	ID     string `json:"id"`
	Data   struct {
		FileID string `json:"fileId"`
		ID     string `json:"id"`
	} `json:"data"`
}

// Create 通过 save --file 创建 Markdown 笔记，返回云端 fileId。
func (c *Client) Create(ctx context.Context, param *youdaomodel.CreateNoteParam) (string, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return "", err
	}

	note := map[string]string{
		"title":   param.Title,
		"content": param.Content,
		"type":    consts.YoudaoNoteTypeMarkdown,
	}
	if param.FolderID != "" {
		note["parentId"] = param.FolderID
	}
	body, err := json.Marshal(note)
	if err != nil {
		return "", fmt.Errorf("marshal note %q: %w", param.Title, err)
	}
	notePath, err := writeTempNote(body)
	if err != nil {
		return "", err
	}
	// 临时文件只含已落在本地的笔记内容，删除失败不影响结果。
	defer os.Remove(notePath)

	out, err := c.run(ctx, consts.YoudaoCmdSave, consts.YoudaoFlagFile, notePath, consts.YoudaoFlagJSON)
	if err != nil {
		return "", err
	}
	return parseFileID(out)
}

// Update 覆盖更新已有 Markdown 笔记的内容。
func (c *Client) Update(ctx context.Context, param *youdaomodel.UpdateNoteParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}
	_, err := c.run(ctx, consts.YoudaoCmdUpdate, param.FileID, consts.YoudaoFlagData, param.Content)
	return err
}

func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, c.bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if errors.Is(err, exec.ErrNotFound) {
		return nil, fmt.Errorf("%s not found, install the official youdaonote CLI and run "+
			"`youdaonote config set apiKey <API Key>`, or set %s: %w", c.bin, consts.EnvYoudaoBin, err)
	}
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return nil, fmt.Errorf("youdaonote %s failed: %w: %s", args[0], err, detail)
	}
	return stdout.Bytes(), nil
}

func writeTempNote(body []byte) (string, error) {
	file, err := os.CreateTemp("", consts.YoudaoTempFilePattern)
	if err != nil {
		return "", fmt.Errorf("create temp note file: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(body); err != nil {
		os.Remove(file.Name())
		return "", fmt.Errorf("write temp note file: %w", err)
	}
	return file.Name(), nil
}

func parseFileID(out []byte) (string, error) {
	var resp noteResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parse youdaonote output %q: %w", strings.TrimSpace(string(out)), err)
	}
	for _, id := range []string{resp.FileID, resp.ID, resp.Data.FileID, resp.Data.ID} {
		if id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("fileId not found in youdaonote output %q, the note may already exist in cloud",
		strings.TrimSpace(string(out)))
}
