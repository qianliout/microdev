package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"microdev/pkg/logger"
)

// Processor 输出处理器
type Processor struct {
	outputFile string
	file       *os.File
	logger     *logger.Logger
	isConsole  bool
}

// NewProcessor 创建新的输出处理器
func NewProcessor(outputFile string, log *logger.Logger) *Processor {
	return &Processor{
		outputFile: outputFile,
		logger:     log,
		isConsole:  outputFile == "",
	}
}

// Write 实现io.Writer接口，用于流式输出
func (p *Processor) Write(data []byte) (int, error) {
	// 如果是控制台输出，直接写入stdout
	if p.isConsole {
		return os.Stdout.Write(data)
	}

	// 如果文件还未打开，先打开文件
	if p.file == nil {
		if err := p.openFile(); err != nil {
			return 0, err
		}
	}

	// 同时写入文件和控制台（用于实时显示）
	n, err := p.file.Write(data)
	if err != nil {
		return n, err
	}

	// 也输出到控制台以便用户实时看到
	os.Stdout.Write(data)

	return n, nil
}

// WriteString 写入字符串
func (p *Processor) WriteString(s string) (int, error) {
	return p.Write([]byte(s))
}

// openFile 打开输出文件
func (p *Processor) openFile() error {
	if p.outputFile == "" {
		return fmt.Errorf("output file path is empty")
	}

	// 确保目录存在
	dir := filepath.Dir(p.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s failed: %v", dir, err)
	}

	// 以追加模式打开文件
	file, err := os.OpenFile(p.outputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open output file %s failed: %v", p.outputFile, err)
	}

	p.file = file
	p.logger.Info().Str("output", p.outputFile).Msg("输出文件已打开")

	return nil
}

// Close 关闭文件
func (p *Processor) Close() error {
	if p.file != nil {
		err := p.file.Close()
		p.file = nil
		if err != nil {
			return fmt.Errorf("close output file failed: %v", err)
		}
		p.logger.Info().Msg("输出文件已关闭")
	}
	return nil
}

// IsConsoleOutput 检查是否为控制台输出
func (p *Processor) IsConsoleOutput() bool {
	return p.isConsole
}

// GetOutputPath 获取输出文件路径
func (p *Processor) GetOutputPath() string {
	return p.outputFile
}

// Flush 刷新缓冲区
func (p *Processor) Flush() error {
	if p.file != nil {
		return p.file.Sync()
	}
	return nil
}

// StreamWriter 流式写入器接口
type StreamWriter interface {
	io.Writer
	WriteString(s string) (int, error)
	Flush() error
}
