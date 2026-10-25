// Package store 持久化有道同步的本地映射文件。
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"microdev/infra/consts"
	youdaomodel "microdev/infra/model/youdao"
)

type MappingStore struct{}

func NewMappingStore() *MappingStore {
	return &MappingStore{}
}

// Load 读取 param.Root 下的映射文件，文件不存在时返回空映射。
func (dal *MappingStore) Load(_ context.Context, param *youdaomodel.MappingParam) (*youdaomodel.Mapping, error) {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return nil, err
	}

	path := mappingPath(param.Root)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return (*youdaomodel.Mapping)(nil).Deserialize(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read mapping file %s: %w", path, err)
	}

	mapping := &youdaomodel.Mapping{}
	if err := json.Unmarshal(data, mapping); err != nil {
		return nil, fmt.Errorf("mapping file %s is corrupted, fix or remove it: %w", path, err)
	}
	return mapping.Deserialize(), nil
}

// Save 把 param.Mapping 写入 param.Root 下的映射文件。
func (dal *MappingStore) Save(_ context.Context, param *youdaomodel.MappingParam) error {
	param = param.Serialize()
	if err := param.Check(); err != nil {
		return err
	}
	if param.Mapping == nil {
		return fmt.Errorf("mapping is nil")
	}

	path := mappingPath(param.Root)
	data, err := json.MarshalIndent(param.Mapping, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mapping file %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, consts.YoudaoMappingFileMode); err != nil {
		return fmt.Errorf("write mapping file %s: %w", path, err)
	}
	return nil
}

func mappingPath(root string) string {
	return filepath.Join(root, consts.YoudaoMappingFile)
}
