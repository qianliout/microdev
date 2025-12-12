package input

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"microdev/pkg/logger"
)

// Processor 输入处理器
type Processor struct {
	logger *logger.Logger
}

// NewProcessor 创建新的输入处理器
func NewProcessor(log *logger.Logger) *Processor {
	return &Processor{
		logger: log,
	}
}

// Process 处理输入内容
// 如果输入是有效的文件路径，则读取文件内容
// 否则直接使用输入字符串作为内容
func (p *Processor) Process(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	p.logger.Debug().Str("input", input).Msg("处理输入")

	// 检查是否为文件路径
	if p.isValidFilePath(input) {
		p.logger.Info().Str("path", input).Msg("检测到文件路径，读取文件内容")
		return p.readFileContent(input)
	}

	// 直接使用输入字符串
	p.logger.Info().Msg("使用直接文本输入")
	return strings.TrimSpace(input), nil
}

// isValidFilePath 检查输入是否为有效的文件路径
func (p *Processor) isValidFilePath(input string) bool {
	// 清理路径
	cleanPath := filepath.Clean(input)

	// 检查文件是否存在
	info, err := os.Stat(cleanPath)
	if err != nil {
		p.logger.Debug().Str("path", cleanPath).Err(err).Msg("路径不存在或无法访问")
		return false
	}

	// 检查是否为常规文件（不是目录）
	if info.IsDir() {
		p.logger.Debug().Str("path", cleanPath).Msg("路径是目录，不是文件")
		return false
	}

	return true
}

// readFileContent 读取文件内容
func (p *Processor) readFileContent(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file %s failed: %v", filePath, err)
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read file %s failed: %v", filePath, err)
	}

	// 检查文件是否为空
	if len(content) == 0 {
		return "", fmt.Errorf("file %s is empty", filePath)
	}

	p.logger.Info().Int("len", len(content)).Msg("成功读取文件")
	return strings.TrimSpace(string(content)), nil
}
