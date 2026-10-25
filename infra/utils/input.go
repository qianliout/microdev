package utils

import (
	"fmt"
	"io"
	"os"
	"strings"

	"microdev/infra/consts"
)

// fileExists 判断路径是否存在。
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadInput 按 "-" 读标准输入、已存在的普通文件读内容、其余按直接文本返回。
func ReadInput(arg string) (string, error) {
	if strings.TrimSpace(arg) == "" {
		return "", fmt.Errorf("empty input")
	}
	if arg == consts.StdinArg {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), nil
	}

	info, err := os.Stat(arg)
	isFile := err == nil && info.Mode().IsRegular()
	if !isFile {
		return arg, nil
	}
	data, err := os.ReadFile(arg)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}
