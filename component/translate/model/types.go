package model

import "io"

// TranslateRequest 翻译请求
type Translate struct {
	Input  io.ReadCloser      `json:"input"`
	Direct string             `json:"direction"` // zh2en 或 en2zh
	Output io.ReadWriteCloser `json:"output"`
}

const (
	DirectionZh2En = "zh2en"
	DirectionEn2Zh = "en2zh"
)

// RWPipe 读写管道，支持流式边写边读
// 使用 io.Pipe 实现，便于LLM流式输出与终端同步打印
type RWPipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

// NewRWPipe 创建新的读写管道
func NewRWPipe() *RWPipe {
	pr, pw := io.Pipe()
	return &RWPipe{r: pr, w: pw}
}

// Read 从管道读取数据
func (p *RWPipe) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

// Write 向管道写入数据
func (p *RWPipe) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

// Close 关闭读写端（用于资源回收）
func (p *RWPipe) Close() error {
	// 先关闭写端，通知读端EOF
	_ = p.w.Close()
	// 再关闭读端
	return p.r.Close()
}

// CloseWrite 仅关闭写端，触发读端EOF
func (p *RWPipe) CloseWrite() error {
	return p.w.Close()
}

// CloseRead 仅关闭读端
func (p *RWPipe) CloseRead() error {
	return p.r.Close()
}
